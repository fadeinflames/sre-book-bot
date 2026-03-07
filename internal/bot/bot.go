package bot

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"sync"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"sre-learning-bot/internal/app"
	"sre-learning-bot/internal/config"
)

type Service struct {
	logger     *slog.Logger
	cfg        config.Config
	store      *app.Store
	bot        *tgbotapi.BotAPI
	quizMu     sync.Mutex
	quizStates map[int64]*quizState
	reviewMu   sync.Mutex
	reviewFlow map[int64][]app.ReviewCard
}

// Тексты кнопок главного меню (ReplyKeyboard). При нажатии пользователь отправляет их как обычное сообщение.
const (
	btnRoadmap    = "📖 Roadmap"
	btnLesson     = "📚 Урок"
	btnLessonNext = "📄 Дальше"
	btnQuiz       = "❓ Квиз"
	btnReview     = "🔄 Повторение"
	btnProgress   = "📊 Прогресс"
	btnChecklists = "✅ Чеклисты"
	btnHelp       = "❓ Помощь"
)

type quizState struct {
	ModuleID   int64
	ModuleSlug string
	Questions  []app.QuizQuestion
	Index      int
	Answers    map[int64]string
}

func New(logger *slog.Logger, cfg config.Config, store *app.Store, botAPI *tgbotapi.BotAPI) *Service {
	return &Service{
		logger:     logger,
		cfg:        cfg,
		store:      store,
		bot:        botAPI,
		quizStates: map[int64]*quizState{},
		reviewFlow: map[int64][]app.ReviewCard{},
	}
}

func (s *Service) Run(ctx context.Context) error {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = int(s.cfg.PollTimeout.Seconds())
	updates := s.bot.GetUpdatesChan(u)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case update := <-updates:
			if update.Message != nil {
				s.handleMessage(ctx, update.Message)
			}
			if update.CallbackQuery != nil {
				s.handleCallback(ctx, update.CallbackQuery)
			}
		}
	}
}

func (s *Service) SendReminder(ctx context.Context, telegramID int64, dueCount int) error {
	msg := tgbotapi.NewMessage(telegramID, fmt.Sprintf("У тебя %d карточек на повторение. Запусти /review и закрой сессию за 5-7 минут.", dueCount))
	_, err := s.bot.Send(msg)
	return err
}

func (s *Service) handleMessage(ctx context.Context, msg *tgbotapi.Message) {
	if msg.From == nil {
		return
	}
	if len(s.cfg.AllowedTelegramIDs) > 0 {
		if _, ok := s.cfg.AllowedTelegramIDs[msg.From.ID]; !ok {
			s.sendText(msg.Chat.ID, "Доступ к боту ограничен.")
			return
		}
	}
	user, err := s.store.EnsureUser(
		ctx,
		msg.From.ID,
		msg.From.UserName,
		app.RoleJunior,
		s.cfg.AdminTelegramIDs,
		s.cfg.MentorTelegramIDs,
	)
	if err != nil {
		s.logger.Error("ensure user failed", "err", err)
		return
	}
	_ = s.store.LogEvent(ctx, user.ID, "telegram_message", map[string]any{"text": msg.Text})

	if !msg.IsCommand() {
		s.handleNonCommand(ctx, user, msg)
		return
	}

	cmd := msg.Command()
	args := strings.TrimSpace(msg.CommandArguments())

	switch cmd {
	case "start":
		s.sendTextWithMainMenu(msg.Chat.ID, s.quickStartText())
	case "help":
		s.sendTextWithMainMenu(msg.Chat.ID, s.quickStartText())
	case "roadmap":
		s.cmdRoadmap(ctx, user, msg.Chat.ID)
	case "lesson":
		s.cmdLesson(ctx, user, msg.Chat.ID)
	case "lesson_next":
		s.cmdLessonNext(ctx, user, msg.Chat.ID)
	case "done":
		s.cmdDone(ctx, user, msg.Chat.ID, args)
	case "quiz":
		s.cmdQuiz(ctx, user, msg.Chat.ID, args)
	case "review":
		s.cmdReview(ctx, user, msg.Chat.ID)
	case "progress":
		s.cmdProgress(ctx, user, msg.Chat.ID)
	case "checklists":
		s.cmdChecklists(ctx, msg.Chat.ID, args)
	case "sources":
		s.cmdSources(ctx, user, msg.Chat.ID, args)
	case "submit_checklist":
		s.cmdSubmitChecklist(ctx, user, msg.Chat.ID, args)
	case "mentor_report":
		s.cmdMentorReport(ctx, user, msg.Chat.ID)
	case "pending_reviews":
		s.cmdPendingChecklistReviews(ctx, user, msg.Chat.ID)
	case "review_submission":
		s.cmdReviewSubmission(ctx, user, msg.Chat.ID, args)
	default:
		s.sendText(msg.Chat.ID, "Неизвестная команда.")
	}
}

func (s *Service) handleNonCommand(ctx context.Context, user app.User, msg *tgbotapi.Message) {
	text := strings.TrimSpace(msg.Text)
	if text == "" {
		return
	}
	// Навигация по кнопкам главного меню
	switch text {
	case btnRoadmap:
		s.cmdRoadmap(ctx, user, msg.Chat.ID)
		return
	case btnLesson:
		s.cmdLesson(ctx, user, msg.Chat.ID)
		return
	case btnLessonNext:
		s.cmdLessonNext(ctx, user, msg.Chat.ID)
		return
	case btnQuiz:
		s.cmdQuiz(ctx, user, msg.Chat.ID, "")
		return
	case btnReview:
		s.cmdReview(ctx, user, msg.Chat.ID)
		return
	case btnProgress:
		s.cmdProgress(ctx, user, msg.Chat.ID)
		return
	case btnChecklists:
		s.cmdChecklists(ctx, msg.Chat.ID, "")
		return
	case btnHelp:
		s.sendTextWithMainMenu(msg.Chat.ID, s.quickStartText())
		return
	}
	if isOptionAnswer(text) {
		s.quizMu.Lock()
		state := s.quizStates[user.TelegramID]
		s.quizMu.Unlock()
		ans := strings.ToUpper(string(text[0]))
		if state != nil {
			s.acceptQuizAnswerAndAskNext(ctx, user, msg.Chat.ID, ans)
			return
		}
		s.acceptReviewAnswer(ctx, user, msg.Chat.ID, ans)
	}
}

func (s *Service) cmdRoadmap(ctx context.Context, user app.User, chatID int64) {
	roadmap, err := s.store.GetRoadmap(ctx, user.ID)
	if err != nil {
		s.sendText(chatID, "Не смог загрузить roadmap.")
		return
	}
	if len(roadmap) == 0 {
		s.sendText(chatID, "Контент пока не загружен.")
		return
	}
	var b strings.Builder
	for _, m := range roadmap {
		status := "in progress"
		if m.Completed {
			status = "done"
		}
		fmt.Fprintf(&b, "- %s (%s): lessons %d/%d, mastery %.0f%%\n", m.Slug, status, m.LessonCompleted, m.LessonTotal, m.MasteryScore*100)
	}
	b.WriteString("\nКак начать: /lesson\nПосле чтения урока: /done (lesson_id)\nДальше: /quiz (module_slug)")
	s.sendText(chatID, b.String())
}

func (s *Service) cmdLesson(ctx context.Context, user app.User, chatID int64) {
	s.sendNextLessonChunkOrCard(ctx, user, chatID, false)
}

func (s *Service) cmdLessonNext(ctx context.Context, user app.User, chatID int64) {
	s.sendNextLessonChunkOrCard(ctx, user, chatID, true)
}

const maxChunkLen = 3500

func (s *Service) sendNextLessonChunkOrCard(ctx context.Context, user app.User, chatID int64, onlyNextChunk bool) {
	roadmap, err := s.store.GetRoadmap(ctx, user.ID)
	if err != nil {
		s.sendText(chatID, "Ошибка загрузки урока.")
		return
	}
	var nextMod *app.ModuleProgress
	for i := range roadmap {
		if roadmap[i].NextLessonID > 0 {
			nextMod = &roadmap[i]
			break
		}
	}
	if nextMod == nil {
		s.sendText(chatID, "Все уроки по roadmap завершены.\nСледующий шаг: /review и /mentor_report (для ментора).")
		return
	}
	lessonID := nextMod.NextLessonID
	chunks, err := s.store.GetLessonChunks(ctx, lessonID)
	if err != nil {
		s.logger.Warn("get lesson chunks failed", "err", err, "lesson_id", lessonID)
	}
	if onlyNextChunk && len(chunks) == 0 {
		s.sendText(chatID, "У этого урока нет порций текста. Используй /lesson чтобы увидеть материалы.")
		return
	}
	progress, err := s.store.GetUserLessonReadingProgress(ctx, user.ID, lessonID)
	if err != nil {
		s.logger.Warn("get reading progress failed", "err", err)
		progress = 0
	}
	if len(chunks) > 0 && progress < len(chunks) {
		chunk := chunks[progress]
		if len(chunk) > maxChunkLen {
			chunk = chunk[:maxChunkLen] + "..."
		}
		header := fmt.Sprintf("Урок: %s\nmodule=%s, часть %d из %d\n\n", nextMod.NextLessonTitle, nextMod.Slug, progress+1, len(chunks))
		s.sendText(chatID, header+chunk)
		nextIdx := progress + 1
		_ = s.store.SetUserLessonReadingProgress(ctx, user.ID, lessonID, nextIdx)
		if nextIdx < len(chunks) {
			s.sendText(chatID, "Продолжить? /lesson_next")
			return
		}
		s.sendText(chatID, fmt.Sprintf("Конец урока.\nОтметь прохождение: /done %d\nПотом квиз: /quiz %s", lessonID, nextMod.Slug))
		return
	}
	// Нет фрагментов или всё уже прочитано — показываем карточку урока и ссылки на материалы
	resources, _ := s.store.GetLessonResources(ctx, lessonID)
	resourceText := "Материалы для этого урока пока не добавлены."
	if len(resources) > 0 {
		var rb strings.Builder
		rb.WriteString("Ссылки на источники:\n")
		for _, r := range resources {
			url := r.URL
			if strings.HasPrefix(url, "local://") {
				url = "(локальный файл: " + strings.TrimPrefix(url, "local://") + ")"
			}
			fmt.Fprintf(&rb, "- %s: %s\n", r.Title, url)
		}
		resourceText = strings.TrimSpace(rb.String())
	}
	text := fmt.Sprintf(
		"%s\nmodule=%s, lesson_id=%d\n\n%s\n\n%s\n\nЧто делать дальше:\n1) Отметь прохождение: /done %d\n2) Квиз модуля: /quiz %s\n3) Ещё ссылки: /sources %s",
		nextMod.NextLessonTitle,
		nextMod.Slug,
		lessonID,
		nextMod.NextLessonContent,
		resourceText,
		lessonID,
		nextMod.Slug,
		nextMod.Slug,
	)
	s.sendText(chatID, text)
}

func (s *Service) cmdDone(ctx context.Context, user app.User, chatID int64, args string) {
	lessonID, err := strconv.ParseInt(strings.TrimSpace(args), 10, 64)
	if err != nil || lessonID <= 0 {
		s.sendText(chatID, "Использование: /done (lesson_id)")
		return
	}
	if err := s.store.MarkLessonDone(ctx, user.ID, lessonID); err != nil {
		s.sendText(chatID, "Не удалось отметить урок.")
		return
	}
	_ = s.store.LogEvent(ctx, user.ID, "lesson_done", map[string]any{"lesson_id": lessonID})
	s.sendText(chatID, "Урок отмечен как пройденный.")
}

func (s *Service) cmdQuiz(ctx context.Context, user app.User, chatID int64, args string) {
	moduleSlug := strings.TrimSpace(args)
	if moduleSlug == "" {
		s.sendText(chatID, "Использование: /quiz (module_slug)")
		return
	}
	moduleID, qs, err := s.store.GetQuizQuestionsByModule(ctx, moduleSlug)
	if err != nil || len(qs) == 0 {
		s.sendText(chatID, "Не нашел вопросы для модуля.")
		return
	}
	state := &quizState{
		ModuleID:   moduleID,
		ModuleSlug: moduleSlug,
		Questions:  qs,
		Answers:    map[int64]string{},
	}
	s.quizMu.Lock()
	s.quizStates[user.TelegramID] = state
	s.quizMu.Unlock()

	s.sendText(chatID, fmt.Sprintf("Старт квиза по %s (%d вопросов). Выбери ответ кнопкой ниже.", moduleSlug, len(qs)))
	s.askQuizQuestion(chatID, state)
}

func (s *Service) askQuizQuestion(chatID int64, state *quizState) {
	if state.Index >= len(state.Questions) {
		return
	}
	q := state.Questions[state.Index]
	text := fmt.Sprintf("Q%d: %s\n\nA) %s\nB) %s\nC) %s\nD) %s", state.Index+1, q.Question, q.OptionA, q.OptionB, q.OptionC, q.OptionD)
	kb := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("A", "quiz:A"),
			tgbotapi.NewInlineKeyboardButtonData("B", "quiz:B"),
			tgbotapi.NewInlineKeyboardButtonData("C", "quiz:C"),
			tgbotapi.NewInlineKeyboardButtonData("D", "quiz:D"),
		),
	)
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ReplyMarkup = kb
	if _, err := s.bot.Send(msg); err != nil {
		s.logger.Error("send quiz question failed", "err", err, "chat_id", chatID)
	}
}

func (s *Service) acceptQuizAnswerAndAskNext(ctx context.Context, user app.User, chatID int64, ans string) {
	s.quizMu.Lock()
	state := s.quizStates[user.TelegramID]
	if state == nil || state.Index >= len(state.Questions) {
		s.quizMu.Unlock()
		return
	}
	q := state.Questions[state.Index]
	state.Answers[q.ID] = ans
	state.Index++
	done := state.Index >= len(state.Questions)
	s.quizMu.Unlock()

	if !done {
		s.askQuizQuestion(chatID, state)
		return
	}

	result, err := s.store.SubmitQuiz(ctx, user.ID, state.ModuleID, state.Answers)
	if err != nil {
		s.sendText(chatID, "Не удалось сохранить результаты квиза.")
		return
	}
	_ = s.store.LogEvent(ctx, user.ID, "quiz_submitted", map[string]any{"module_id": state.ModuleID, "score": result.Score, "max": result.MaxScore})

	var b strings.Builder
	fmt.Fprintf(&b, "Квиз завершен: %d/%d. Mastery %.0f%%\n", result.Score, result.MaxScore, result.Mastery*100)
	for _, d := range result.Details {
		if d.Correct {
			continue
		}
		fmt.Fprintf(&b, "- Q%d: неверно (%s), правильно %s\n  %s\n  %s\n", d.QuestionID, d.Selected, d.CorrectOpt, d.Explanation, d.SourceURL)
	}
	s.sendText(chatID, b.String())

	s.quizMu.Lock()
	delete(s.quizStates, user.TelegramID)
	s.quizMu.Unlock()
}

func (s *Service) cmdReview(ctx context.Context, user app.User, chatID int64) {
	cards, err := s.store.GetDueReviewCards(ctx, user.ID, 10)
	if err != nil {
		s.sendText(chatID, "Ошибка при загрузке карточек повторения.")
		return
	}
	if len(cards) == 0 {
		s.sendText(chatID, "Сейчас нет карточек на повторение.")
		return
	}
	s.reviewMu.Lock()
	s.reviewFlow[user.TelegramID] = cards
	s.reviewMu.Unlock()
	s.sendText(chatID, "Запускаю review-сессию. Ответь A/B/C/D.")
	s.askNextReviewCard(chatID, fmt.Sprintf("Review 1/%d", len(cards)), cards[0])
}

func (s *Service) acceptReviewAnswer(ctx context.Context, user app.User, chatID int64, ans string) {
	s.reviewMu.Lock()
	cards := s.reviewFlow[user.TelegramID]
	if len(cards) == 0 {
		s.reviewMu.Unlock()
		return
	}
	card := cards[0]
	remaining := cards[1:]
	s.reviewFlow[user.TelegramID] = remaining
	s.reviewMu.Unlock()

	correct := ans == card.Correct
	quality := 1
	if correct {
		quality = 5
	}
	if err := s.store.ApplyReviewScore(ctx, card.ID, quality); err != nil {
		s.sendText(chatID, "Не удалось сохранить результат review.")
		return
	}
	if correct {
		s.sendText(chatID, "Верно. Переходим дальше.")
	} else {
		s.sendText(chatID, fmt.Sprintf("Неверно. Правильный ответ: %s\n%s\nИсточник: %s", card.Correct, card.Explanation, card.SourceURL))
	}

	if len(remaining) == 0 {
		s.reviewMu.Lock()
		delete(s.reviewFlow, user.TelegramID)
		s.reviewMu.Unlock()
		s.sendText(chatID, "Review-сессия завершена.")
		return
	}
	title := fmt.Sprintf("Следующая карточка (%d осталось)", len(remaining))
	s.askNextReviewCard(chatID, title, remaining[0])
}

func (s *Service) askNextReviewCard(chatID int64, title string, card app.ReviewCard) {
	text := fmt.Sprintf("%s\n%s\nA) %s\nB) %s\nC) %s\nD) %s", title, card.Question, card.OptionA, card.OptionB, card.OptionC, card.OptionD)
	s.sendText(chatID, text)
}

func (s *Service) cmdProgress(ctx context.Context, user app.User, chatID int64) {
	roadmap, err := s.store.GetRoadmap(ctx, user.ID)
	if err != nil {
		s.sendText(chatID, "Не удалось получить прогресс.")
		return
	}
	var completed int
	var mastery float64
	for _, m := range roadmap {
		if m.Completed {
			completed++
		}
		mastery += m.MasteryScore
	}
	avg := 0.0
	if len(roadmap) > 0 {
		avg = mastery / float64(len(roadmap))
	}
	s.sendText(chatID, fmt.Sprintf("Модулей завершено: %d/%d\nСредний mastery: %.0f%%", completed, len(roadmap), avg*100))
}

func (s *Service) cmdChecklists(ctx context.Context, chatID int64, moduleSlug string) {
	moduleSlug = strings.TrimSpace(moduleSlug)
	if moduleSlug == "" {
		s.sendText(chatID, "Использование: /checklists (module_slug)")
		return
	}
	items, err := s.store.GetChecklistsByModule(ctx, moduleSlug)
	if err != nil || len(items) == 0 {
		s.sendText(chatID, "Чеклисты не найдены.")
		return
	}
	var b strings.Builder
	for _, c := range items {
		fmt.Fprintf(&b, "Checklist %d: %s\n", c.ID, c.Title)
		fmt.Fprintf(&b, "%s\n", c.Description)
		for i, it := range c.Items {
			fmt.Fprintf(&b, "%d) %s\n", i+1, it)
		}
		fmt.Fprintln(&b)
	}
	s.sendText(chatID, b.String())
}

func (s *Service) cmdSources(ctx context.Context, user app.User, chatID int64, moduleSlug string) {
	moduleSlug = strings.TrimSpace(moduleSlug)
	if moduleSlug == "" {
		roadmap, err := s.store.GetRoadmap(ctx, user.ID)
		if err != nil || len(roadmap) == 0 {
			s.sendText(chatID, "Использование: /sources (module_slug)\nСначала открой roadmap: /roadmap")
			return
		}
		var b strings.Builder
		b.WriteString("Укажи модуль: /sources (module_slug)\nДоступные модули:\n")
		for _, m := range roadmap {
			fmt.Fprintf(&b, "- %s\n", m.Slug)
		}
		s.sendText(chatID, b.String())
		return
	}
	resources, err := s.store.GetModuleResources(ctx, moduleSlug)
	if err != nil || len(resources) == 0 {
		s.sendText(chatID, "Источники не найдены.")
		return
	}
	var b strings.Builder
	fmt.Fprintf(&b, "Источники для %s:\n", moduleSlug)
	for _, r := range resources {
		fmt.Fprintf(&b, "- %s: %s\n", r.Title, r.URL)
	}
	s.sendText(chatID, b.String())
}

func (s *Service) quickStartText() string {
	return strings.Join([]string{
		"Привет! Я SRE Learning Bot.",
		"",
		"Используй кнопки ниже или команды:",
		"• Roadmap — учебный путь",
		"• Урок / Дальше — следующий урок по частям",
		"• Квиз — проверка знаний (нужен module_slug в чате)",
		"• Повторение — карточки",
		"• Прогресс — статистика",
		"• Чеклисты — по module_slug",
		"",
		"Команды: /done (lesson_id), /quiz (module_slug), /sources (module_slug), /submit_checklist (id) (notes)",
	}, "\n")
}

func (s *Service) mainMenuKeyboard() tgbotapi.ReplyKeyboardMarkup {
	kb := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(btnRoadmap),
			tgbotapi.NewKeyboardButton(btnLesson),
			tgbotapi.NewKeyboardButton(btnLessonNext),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(btnQuiz),
			tgbotapi.NewKeyboardButton(btnReview),
			tgbotapi.NewKeyboardButton(btnProgress),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(btnChecklists),
			tgbotapi.NewKeyboardButton(btnHelp),
		),
	)
	kb.ResizeKeyboard = true
	return kb
}

func (s *Service) sendTextWithMainMenu(chatID int64, text string) {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ReplyMarkup = s.mainMenuKeyboard()
	if _, err := s.bot.Send(msg); err != nil {
		s.logger.Error("send message with menu failed", "err", err, "chat_id", chatID)
	}
}

func (s *Service) cmdSubmitChecklist(ctx context.Context, user app.User, chatID int64, args string) {
	parts := strings.SplitN(strings.TrimSpace(args), " ", 2)
	if len(parts) < 2 {
		s.sendText(chatID, "Использование: /submit_checklist (checklist_id) (notes)")
		return
	}
	id, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		s.sendText(chatID, "Некорректный checklist_id.")
		return
	}
	sub, err := s.store.SubmitChecklist(ctx, user.ID, id, strings.TrimSpace(parts[1]))
	if err != nil {
		s.sendText(chatID, "Не удалось отправить чеклист.")
		return
	}
	_ = s.store.LogEvent(ctx, user.ID, "checklist_submitted", map[string]any{"submission_id": sub.ID})
	s.sendText(chatID, fmt.Sprintf("Чеклист отправлен. submission_id=%d", sub.ID))
}

func (s *Service) cmdMentorReport(ctx context.Context, user app.User, chatID int64) {
	if user.Role != app.RoleMentor && user.Role != app.RoleAdmin {
		s.sendText(chatID, "Команда доступна только ментору/admin.")
		return
	}
	reports, err := s.store.ListJuniorReports(ctx)
	if err != nil {
		s.sendText(chatID, "Ошибка генерации отчета.")
		return
	}
	if len(reports) == 0 {
		s.sendText(chatID, "Нет активных джунов.")
		return
	}
	var b strings.Builder
	for _, r := range reports {
		last := "-"
		if r.LastActivityUTC != nil {
			last = r.LastActivityUTC.UTC().Format(time.RFC3339)
		}
		fmt.Fprintf(&b, "junior %d (@%s)\n", r.TelegramID, r.Username)
		fmt.Fprintf(&b, "completion: %.0f%%, mastery: %.0f%%\n", r.CompletionPct, r.MasteryAvg)
		fmt.Fprintf(&b, "weak: %s\nactivity 7d/30d: %d/%d\nlast: %s\n\n", r.WeakModules, r.Activity7d, r.Activity30d, last)
	}
	s.sendText(chatID, b.String())
}

func (s *Service) cmdPendingChecklistReviews(ctx context.Context, user app.User, chatID int64) {
	if user.Role != app.RoleMentor && user.Role != app.RoleAdmin {
		s.sendText(chatID, "Команда доступна только ментору/admin.")
		return
	}
	list, err := s.store.ListPendingChecklistSubmissions(ctx)
	if err != nil {
		s.sendText(chatID, "Ошибка загрузки pending submissions.")
		return
	}
	if len(list) == 0 {
		s.sendText(chatID, "Нет pending submissions.")
		return
	}
	var b strings.Builder
	for _, it := range list {
		fmt.Fprintf(&b, "submission_id=%d user_id=%d checklist_id=%d notes=%q\n", it.ID, it.UserID, it.ChecklistID, it.Notes)
	}
	s.sendText(chatID, b.String())
}

func (s *Service) cmdReviewSubmission(ctx context.Context, user app.User, chatID int64, args string) {
	if user.Role != app.RoleMentor && user.Role != app.RoleAdmin {
		s.sendText(chatID, "Команда доступна только ментору/admin.")
		return
	}
	parts := strings.SplitN(strings.TrimSpace(args), " ", 3)
	if len(parts) < 2 {
		s.sendText(chatID, "Использование: /review_submission (id) approve|rework [comment]")
		return
	}
	submissionID, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		s.sendText(chatID, "Некорректный submission_id")
		return
	}
	action := strings.ToLower(parts[1])
	comment := ""
	if len(parts) == 3 {
		comment = parts[2]
	}
	approve := action == "approve"
	if action != "approve" && action != "rework" {
		s.sendText(chatID, "action должен быть approve или rework")
		return
	}
	if err := s.store.ReviewChecklist(ctx, user.ID, submissionID, approve, comment); err != nil {
		s.sendText(chatID, "Не удалось сохранить review.")
		return
	}
	s.sendText(chatID, "Review сохранен.")
}

func (s *Service) handleCallback(ctx context.Context, cb *tgbotapi.CallbackQuery) {
	if cb == nil || cb.Message == nil || cb.From == nil {
		return
	}
	if len(s.cfg.AllowedTelegramIDs) > 0 {
		if _, ok := s.cfg.AllowedTelegramIDs[cb.From.ID]; !ok {
			_, _ = s.bot.Request(tgbotapi.NewCallbackWithAlert(cb.ID, "Доступ ограничен."))
			return
		}
	}
	user, err := s.store.EnsureUser(
		ctx,
		cb.From.ID,
		cb.From.UserName,
		app.RoleJunior,
		s.cfg.AdminTelegramIDs,
		s.cfg.MentorTelegramIDs,
	)
	if err != nil {
		s.logger.Error("ensure user failed on callback", "err", err)
		_, _ = s.bot.Request(tgbotapi.NewCallback(cb.ID, ""))
		return
	}

	data := cb.Data
	if len(data) >= 6 && data[:5] == "quiz:" {
		opt := strings.ToUpper(string(data[5]))
		if opt != "A" && opt != "B" && opt != "C" && opt != "D" {
			_, _ = s.bot.Request(tgbotapi.NewCallback(cb.ID, ""))
			return
		}
		s.quizMu.Lock()
		state := s.quizStates[user.TelegramID]
		s.quizMu.Unlock()
		if state != nil {
			_, _ = s.bot.Request(tgbotapi.NewCallback(cb.ID, opt))
			edit := tgbotapi.NewEditMessageReplyMarkup(cb.Message.Chat.ID, cb.Message.MessageID, tgbotapi.InlineKeyboardMarkup{InlineKeyboard: [][]tgbotapi.InlineKeyboardButton{}})
			_, _ = s.bot.Request(edit)
			s.acceptQuizAnswerAndAskNext(ctx, user, cb.Message.Chat.ID, opt)
			return
		}
	}
	_, _ = s.bot.Request(tgbotapi.NewCallback(cb.ID, ""))
}

func (s *Service) sendText(chatID int64, text string) {
	msg := tgbotapi.NewMessage(chatID, text)
	if _, err := s.bot.Send(msg); err != nil {
		s.logger.Error("send message failed", "err", err, "chat_id", chatID)
	}
}

func isOptionAnswer(text string) bool {
	if text == "" {
		return false
	}
	switch strings.ToLower(string(text[0])) {
	case "a", "b", "c", "d":
		return true
	default:
		return false
	}
}
