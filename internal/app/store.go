package app

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type Store struct {
	db    *pgxpool.Pool
	redis *redis.Client
}

func NewStore(db *pgxpool.Pool, redis *redis.Client) *Store {
	return &Store{db: db, redis: redis}
}

func (s *Store) EnsureUser(
	ctx context.Context,
	telegramID int64,
	username string,
	defaultRole UserRole,
	adminIDs map[int64]struct{},
	mentorIDs map[int64]struct{},
) (User, error) {
	if _, ok := adminIDs[telegramID]; ok {
		defaultRole = RoleAdmin
	} else if _, ok := mentorIDs[telegramID]; ok {
		defaultRole = RoleMentor
	}

	var user User
	row := s.db.QueryRow(ctx, `
		INSERT INTO users(telegram_id, username, role, status)
		VALUES ($1, $2, $3, 'active')
		ON CONFLICT (telegram_id) DO UPDATE SET username = EXCLUDED.username
		RETURNING id, telegram_id, username, role, status
	`, telegramID, username, defaultRole)
	if err := row.Scan(&user.ID, &user.TelegramID, &user.Username, &user.Role, &user.Status); err != nil {
		return User{}, fmt.Errorf("ensure user: %w", err)
	}
	return user, nil
}

func (s *Store) LogEvent(ctx context.Context, userID int64, eventType string, payload map[string]any) error {
	raw, _ := json.Marshal(payload)
	_, err := s.db.Exec(ctx, `INSERT INTO events(user_id, event_type, payload) VALUES ($1, $2, $3)`, userID, eventType, raw)
	return err
}

func (s *Store) GetRoadmap(ctx context.Context, userID int64) ([]ModuleProgress, error) {
	rows, err := s.db.Query(ctx, `
		SELECT m.id, m.slug, m.title, m.description, m.order_index,
		       COALESCE(lp.lesson_completed_count, 0),
		       COALESCE(lp.total_lessons, lesson_count.total),
		       COALESCE(lp.quiz_best_score, 0),
		       COALESCE(lp.quiz_max_score, quiz_count.total),
		       COALESCE(lp.mastery_score, 0),
		       COALESCE(lp.completed, false)
		FROM modules m
		LEFT JOIN learning_progress lp ON lp.user_id = $1 AND lp.module_id = m.id
		LEFT JOIN (SELECT module_id, COUNT(*)::int AS total FROM lessons GROUP BY module_id) lesson_count ON lesson_count.module_id = m.id
		LEFT JOIN (SELECT module_id, COUNT(*)::int AS total FROM quiz_questions GROUP BY module_id) quiz_count ON quiz_count.module_id = m.id
		ORDER BY m.order_index
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("select modules: %w", err)
	}
	defer rows.Close()

	var out []ModuleProgress
	for rows.Next() {
		var item ModuleProgress
		if err := rows.Scan(
			&item.ModuleID, &item.Slug, &item.Title, &item.Description, &item.OrderIndex,
			&item.LessonCompleted, &item.LessonTotal, &item.QuizBestScore, &item.QuizMaxScore, &item.MasteryScore, &item.Completed,
		); err != nil {
			return nil, fmt.Errorf("scan module progress: %w", err)
		}
		if err := s.fillNextLesson(ctx, userID, &item); err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func (s *Store) fillNextLesson(ctx context.Context, userID int64, item *ModuleProgress) error {
	row := s.db.QueryRow(ctx, `
		SELECT l.id, l.title, l.content
		FROM lessons l
		WHERE l.module_id = $1
		  AND NOT EXISTS (
		      SELECT 1 FROM lesson_completions lc
		      WHERE lc.user_id = $2 AND lc.lesson_id = l.id
		  )
		ORDER BY l.order_index
		LIMIT 1
	`, item.ModuleID, userID)
	if err := row.Scan(&item.NextLessonID, &item.NextLessonTitle, &item.NextLessonContent); err != nil {
		if err == pgx.ErrNoRows {
			return nil
		}
		return fmt.Errorf("next lesson: %w", err)
	}
	return nil
}

func (s *Store) MarkLessonDone(ctx context.Context, userID, lessonID int64) error {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var moduleID int64
	if err := tx.QueryRow(ctx, `SELECT module_id FROM lessons WHERE id = $1`, lessonID).Scan(&moduleID); err != nil {
		return fmt.Errorf("find lesson module: %w", err)
	}

	if _, err := tx.Exec(ctx, `
		INSERT INTO lesson_completions(user_id, lesson_id)
		VALUES ($1, $2) ON CONFLICT DO NOTHING
	`, userID, lessonID); err != nil {
		return fmt.Errorf("insert lesson completion: %w", err)
	}

	if err := s.recalcProgressTx(ctx, tx, userID, moduleID); err != nil {
		return err
	}
	if err := tx.Commit(ctx); err != nil {
		return err
	}
	return nil
}

func (s *Store) GetQuizQuestionsByModule(ctx context.Context, moduleSlug string) (int64, []QuizQuestion, error) {
	var moduleID int64
	if err := s.db.QueryRow(ctx, `SELECT id FROM modules WHERE slug = $1`, moduleSlug).Scan(&moduleID); err != nil {
		return 0, nil, err
	}
	rows, err := s.db.Query(ctx, `
		SELECT id, module_id, question, option_a, option_b, option_c, option_d, correct_option, explanation, source_url
		FROM quiz_questions
		WHERE module_id = $1
		ORDER BY id
	`, moduleID)
	if err != nil {
		return 0, nil, err
	}
	defer rows.Close()

	var list []QuizQuestion
	for rows.Next() {
		var q QuizQuestion
		if err := rows.Scan(&q.ID, &q.ModuleID, &q.Question, &q.OptionA, &q.OptionB, &q.OptionC, &q.OptionD, &q.Correct, &q.Explanation, &q.SourceURL); err != nil {
			return 0, nil, err
		}
		list = append(list, q)
	}
	return moduleID, list, rows.Err()
}

func (s *Store) SubmitQuiz(ctx context.Context, userID, moduleID int64, answers map[int64]string) (QuizResult, error) {
	questions, err := s.quizQuestionsByModuleID(ctx, moduleID)
	if err != nil {
		return QuizResult{}, err
	}
	if len(questions) == 0 {
		return QuizResult{}, fmt.Errorf("no questions found")
	}

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return QuizResult{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var attemptID int64
	if err := tx.QueryRow(ctx, `INSERT INTO quiz_attempts(user_id, module_id, score, max_score) VALUES ($1,$2,0,$3) RETURNING id`, userID, moduleID, len(questions)).Scan(&attemptID); err != nil {
		return QuizResult{}, err
	}

	score := 0
	details := make([]QuizAnswerDetail, 0, len(questions))
	for _, q := range questions {
		selected := strings.ToUpper(strings.TrimSpace(answers[q.ID]))
		if selected == "" {
			selected = "-"
		}
		correct := selected == q.Correct
		if correct {
			score++
		}
		if _, err := tx.Exec(ctx, `
			INSERT INTO quiz_answers(attempt_id, question_id, selected_option, is_correct)
			VALUES ($1,$2,$3,$4)
		`, attemptID, q.ID, selected, correct); err != nil {
			return QuizResult{}, err
		}
		if err := s.upsertReviewItemTx(ctx, tx, userID, moduleID, q, correct); err != nil {
			return QuizResult{}, err
		}
		details = append(details, QuizAnswerDetail{
			QuestionID: q.ID, Selected: selected, Correct: correct, CorrectOpt: q.Correct,
			Explanation: q.Explanation, SourceURL: q.SourceURL,
		})
	}

	if _, err := tx.Exec(ctx, `UPDATE quiz_attempts SET score = $2 WHERE id = $1`, attemptID, score); err != nil {
		return QuizResult{}, err
	}

	if err := s.recalcProgressTx(ctx, tx, userID, moduleID); err != nil {
		return QuizResult{}, err
	}

	var mastery float64
	if err := tx.QueryRow(ctx, `SELECT mastery_score FROM learning_progress WHERE user_id = $1 AND module_id = $2`, userID, moduleID).Scan(&mastery); err != nil {
		return QuizResult{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return QuizResult{}, err
	}
	return QuizResult{
		AttemptID: attemptID,
		Score:     score,
		MaxScore:  len(questions),
		Mastery:   mastery,
		Details:   details,
	}, nil
}

func (s *Store) quizQuestionsByModuleID(ctx context.Context, moduleID int64) ([]QuizQuestion, error) {
	rows, err := s.db.Query(ctx, `
		SELECT id, module_id, question, option_a, option_b, option_c, option_d, correct_option, explanation, source_url
		FROM quiz_questions
		WHERE module_id = $1 ORDER BY id
	`, moduleID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []QuizQuestion
	for rows.Next() {
		var q QuizQuestion
		if err := rows.Scan(&q.ID, &q.ModuleID, &q.Question, &q.OptionA, &q.OptionB, &q.OptionC, &q.OptionD, &q.Correct, &q.Explanation, &q.SourceURL); err != nil {
			return nil, err
		}
		list = append(list, q)
	}
	return list, rows.Err()
}

func (s *Store) upsertReviewItemTx(ctx context.Context, tx pgx.Tx, userID, moduleID int64, question QuizQuestion, correct bool) error {
	stage := 0
	nextAt := time.Now().UTC()
	if correct {
		stage = 1
		nextAt = time.Now().UTC().Add(24 * time.Hour)
	}
	_, err := tx.Exec(ctx, `
		INSERT INTO review_items(user_id, module_id, topic_key, question_id, next_review_at, interval_stage, last_score)
		VALUES ($1,$2,$3,$4,$5,$6,$7)
		ON CONFLICT (user_id, topic_key) DO UPDATE
		SET module_id = EXCLUDED.module_id,
		    question_id = EXCLUDED.question_id,
		    next_review_at = LEAST(review_items.next_review_at, EXCLUDED.next_review_at),
		    interval_stage = CASE WHEN EXCLUDED.last_score = 1 THEN GREATEST(review_items.interval_stage, 1) ELSE 0 END,
		    last_score = EXCLUDED.last_score
	`, userID, moduleID, fmt.Sprintf("quiz:%d", question.ID), question.ID, nextAt, stage, boolToInt(correct))
	return err
}

func (s *Store) recalcProgressTx(ctx context.Context, tx pgx.Tx, userID, moduleID int64) error {
	var lessonDone, lessonTotal int
	if err := tx.QueryRow(ctx, `
		SELECT
			(SELECT COUNT(*) FROM lesson_completions lc JOIN lessons l ON l.id = lc.lesson_id WHERE lc.user_id = $1 AND l.module_id = $2) AS done,
			(SELECT COUNT(*) FROM lessons WHERE module_id = $2) AS total
	`, userID, moduleID).Scan(&lessonDone, &lessonTotal); err != nil {
		return err
	}

	var quizBest, quizMax int
	if err := tx.QueryRow(ctx, `
		SELECT COALESCE(MAX(score), 0), COALESCE(MAX(max_score), 0)
		FROM quiz_attempts
		WHERE user_id = $1 AND module_id = $2
	`, userID, moduleID).Scan(&quizBest, &quizMax); err != nil {
		return err
	}

	lessonRatio := 0.0
	if lessonTotal > 0 {
		lessonRatio = float64(lessonDone) / float64(lessonTotal)
	}
	quizRatio := 0.0
	if quizMax > 0 {
		quizRatio = float64(quizBest) / float64(quizMax)
	}
	mastery := round2(0.4*lessonRatio + 0.6*quizRatio)
	completed := mastery >= 0.7 && lessonDone == lessonTotal && lessonTotal > 0

	_, err := tx.Exec(ctx, `
		INSERT INTO learning_progress(user_id, module_id, lesson_completed_count, total_lessons, quiz_best_score, quiz_max_score, mastery_score, completed)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
		ON CONFLICT (user_id, module_id) DO UPDATE
		SET lesson_completed_count = EXCLUDED.lesson_completed_count,
		    total_lessons = EXCLUDED.total_lessons,
		    quiz_best_score = EXCLUDED.quiz_best_score,
		    quiz_max_score = EXCLUDED.quiz_max_score,
		    mastery_score = EXCLUDED.mastery_score,
		    completed = EXCLUDED.completed,
		    updated_at = now()
	`, userID, moduleID, lessonDone, lessonTotal, quizBest, quizMax, mastery, completed)
	return err
}

func (s *Store) GetChecklistsByModule(ctx context.Context, moduleSlug string) ([]ChecklistInfo, error) {
	rows, err := s.db.Query(ctx, `
		SELECT c.id, c.module_id, m.slug, c.title, c.description
		FROM checklists c
		JOIN modules m ON m.id = c.module_id
		WHERE m.slug = $1
		ORDER BY c.id
	`, moduleSlug)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []ChecklistInfo
	for rows.Next() {
		var item ChecklistInfo
		if err := rows.Scan(&item.ID, &item.ModuleID, &item.ModuleSlug, &item.Title, &item.Description); err != nil {
			return nil, err
		}
		itemRows, err := s.db.Query(ctx, `SELECT item_text FROM checklist_items WHERE checklist_id = $1 ORDER BY order_index`, item.ID)
		if err != nil {
			return nil, err
		}
		for itemRows.Next() {
			var text string
			if err := itemRows.Scan(&text); err != nil {
				itemRows.Close()
				return nil, err
			}
			item.Items = append(item.Items, text)
		}
		itemRows.Close()
		out = append(out, item)
	}
	return out, rows.Err()
}

func (s *Store) GetModuleResources(ctx context.Context, moduleSlug string) ([]LearningResource, error) {
	rows, err := s.db.Query(ctx, `
		SELECT DISTINCT r.title, r.url
		FROM resources r
		JOIN lessons l ON l.id = r.lesson_id
		JOIN modules m ON m.id = l.module_id
		WHERE m.slug = $1
		ORDER BY r.title
	`, moduleSlug)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []LearningResource
	for rows.Next() {
		var item LearningResource
		if err := rows.Scan(&item.Title, &item.URL); err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func (s *Store) GetLessonResources(ctx context.Context, lessonID int64) ([]LearningResource, error) {
	rows, err := s.db.Query(ctx, `
		SELECT title, url
		FROM resources
		WHERE lesson_id = $1
		ORDER BY title
	`, lessonID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []LearningResource
	for rows.Next() {
		var item LearningResource
		if err := rows.Scan(&item.Title, &item.URL); err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func (s *Store) SubmitChecklist(ctx context.Context, userID, checklistID int64, notes string) (ChecklistSubmission, error) {
	var submission ChecklistSubmission
	err := s.db.QueryRow(ctx, `
		INSERT INTO checklist_submissions(user_id, checklist_id, status, notes)
		VALUES ($1,$2,'submitted',$3)
		RETURNING id, user_id, checklist_id, status, notes, COALESCE(review_comment, ''), submitted_at, reviewed_at
	`, userID, checklistID, notes).Scan(
		&submission.ID, &submission.UserID, &submission.ChecklistID, &submission.Status, &submission.Notes,
		&submission.ReviewComment, &submission.SubmittedAt, &submission.ReviewedAt,
	)
	if err != nil {
		return ChecklistSubmission{}, err
	}
	return submission, nil
}

func (s *Store) ReviewChecklist(ctx context.Context, reviewerUserID int64, submissionID int64, approve bool, comment string) error {
	status := "rework"
	if approve {
		status = "approved"
	}
	_, err := s.db.Exec(ctx, `
		UPDATE checklist_submissions
		SET status = $2, reviewed_at = now(), reviewed_by = $3, review_comment = $4
		WHERE id = $1
	`, submissionID, status, reviewerUserID, comment)
	return err
}

func (s *Store) ListPendingChecklistSubmissions(ctx context.Context) ([]ChecklistSubmission, error) {
	rows, err := s.db.Query(ctx, `
		SELECT id, user_id, checklist_id, status, notes, COALESCE(review_comment,''), submitted_at, reviewed_at
		FROM checklist_submissions
		WHERE status = 'submitted'
		ORDER BY submitted_at
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []ChecklistSubmission
	for rows.Next() {
		var item ChecklistSubmission
		if err := rows.Scan(&item.ID, &item.UserID, &item.ChecklistID, &item.Status, &item.Notes, &item.ReviewComment, &item.SubmittedAt, &item.ReviewedAt); err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func (s *Store) GetDueReviewCards(ctx context.Context, userID int64, limit int) ([]ReviewCard, error) {
	rows, err := s.db.Query(ctx, `
		SELECT r.id, r.topic_key, r.question_id, q.question, q.option_a, q.option_b, q.option_c, q.option_d, q.correct_option, q.explanation, q.source_url, r.interval_stage
		FROM review_items r
		JOIN quiz_questions q ON q.id = r.question_id
		WHERE r.user_id = $1
		  AND r.next_review_at <= now()
		ORDER BY r.next_review_at
		LIMIT $2
	`, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []ReviewCard
	for rows.Next() {
		var c ReviewCard
		if err := rows.Scan(&c.ID, &c.TopicKey, &c.QuestionID, &c.Question, &c.OptionA, &c.OptionB, &c.OptionC, &c.OptionD, &c.Correct, &c.Explanation, &c.SourceURL, &c.IntervalStage); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

func (s *Store) ApplyReviewScore(ctx context.Context, reviewID int64, quality int) error {
	if quality < 0 {
		quality = 0
	}
	if quality > 5 {
		quality = 5
	}
	var stage int
	if err := s.db.QueryRow(ctx, `SELECT interval_stage FROM review_items WHERE id = $1`, reviewID).Scan(&stage); err != nil {
		return err
	}

	intervals := []time.Duration{
		24 * time.Hour,
		3 * 24 * time.Hour,
		7 * 24 * time.Hour,
		14 * 24 * time.Hour,
		30 * 24 * time.Hour,
	}

	switch {
	case quality <= 2:
		stage = 0
	case quality == 3:
		if stage > 0 {
			stage--
		}
	default:
		if stage < len(intervals)-1 {
			stage++
		}
	}

	nextAt := time.Now().UTC().Add(intervals[stage])
	_, err := s.db.Exec(ctx, `
		UPDATE review_items
		SET next_review_at = $2, interval_stage = $3, last_score = $4, updated_at = now()
		WHERE id = $1
	`, reviewID, nextAt, stage, quality)
	return err
}

func (s *Store) DueReviewCounts(ctx context.Context) (map[int64]int, error) {
	rows, err := s.db.Query(ctx, `
		SELECT u.telegram_id, COUNT(*)::int
		FROM review_items r
		JOIN users u ON u.id = r.user_id
		WHERE r.next_review_at <= now()
		GROUP BY u.telegram_id
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := map[int64]int{}
	for rows.Next() {
		var tid int64
		var count int
		if err := rows.Scan(&tid, &count); err != nil {
			return nil, err
		}
		out[tid] = count
	}
	return out, rows.Err()
}

func (s *Store) ShouldSendReminder(ctx context.Context, telegramID int64, dayKey string) (bool, error) {
	key := fmt.Sprintf("srs:remind:%d:%s", telegramID, dayKey)
	ok, err := s.redis.SetNX(ctx, key, "1", 24*time.Hour).Result()
	if err != nil {
		return false, err
	}
	return ok, nil
}

func (s *Store) ListJuniorReports(ctx context.Context) ([]JuniorReport, error) {
	rows, err := s.db.Query(ctx, `
		WITH ev AS (
			SELECT user_id,
			       COUNT(*) FILTER (WHERE created_at >= now() - interval '7 days') AS c7,
			       COUNT(*) FILTER (WHERE created_at >= now() - interval '30 days') AS c30,
			       MAX(created_at) AS last_at
			FROM events
			GROUP BY user_id
		)
		SELECT u.telegram_id, COALESCE(u.username, ''), 
		       COALESCE(AVG(CASE WHEN lp.total_lessons > 0 THEN lp.lesson_completed_count::float / lp.total_lessons ELSE 0 END), 0) AS completion,
		       COALESCE(AVG(lp.mastery_score), 0) AS mastery,
		       COALESCE(ev.c7, 0), COALESCE(ev.c30, 0), ev.last_at
		FROM users u
		LEFT JOIN learning_progress lp ON lp.user_id = u.id
		LEFT JOIN ev ON ev.user_id = u.id
		WHERE u.role = 'junior' AND u.status = 'active'
		GROUP BY u.id, u.telegram_id, u.username, ev.c7, ev.c30, ev.last_at
		ORDER BY mastery ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []JuniorReport
	for rows.Next() {
		var r JuniorReport
		if err := rows.Scan(&r.TelegramID, &r.Username, &r.CompletionPct, &r.MasteryAvg, &r.Activity7d, &r.Activity30d, &r.LastActivityUTC); err != nil {
			return nil, err
		}
		r.CompletionPct = round2(r.CompletionPct * 100)
		r.MasteryAvg = round2(r.MasteryAvg * 100)
		r.WeakModules, _ = s.weakModules(ctx, r.TelegramID, 3)
		out = append(out, r)
	}
	return out, rows.Err()
}

func (s *Store) weakModules(ctx context.Context, telegramID int64, n int) (string, error) {
	rows, err := s.db.Query(ctx, `
		SELECT m.slug
		FROM learning_progress lp
		JOIN users u ON u.id = lp.user_id
		JOIN modules m ON m.id = lp.module_id
		WHERE u.telegram_id = $1
		ORDER BY lp.mastery_score ASC
		LIMIT $2
	`, telegramID, n)
	if err != nil {
		return "", err
	}
	defer rows.Close()
	var modules []string
	for rows.Next() {
		var slug string
		if err := rows.Scan(&slug); err != nil {
			return "", err
		}
		modules = append(modules, slug)
	}
	sort.Strings(modules)
	return strings.Join(modules, ", "), rows.Err()
}

func (s *Store) GetUserByTelegramID(ctx context.Context, telegramID int64) (User, error) {
	var u User
	err := s.db.QueryRow(ctx, `SELECT id, telegram_id, COALESCE(username,''), role, status FROM users WHERE telegram_id = $1`, telegramID).
		Scan(&u.ID, &u.TelegramID, &u.Username, &u.Role, &u.Status)
	return u, err
}

func boolToInt(v bool) int {
	if v {
		return 1
	}
	return 0
}

func round2(v float64) float64 {
	return math.Round(v*100) / 100
}
