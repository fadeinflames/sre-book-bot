package app

import "time"

type UserRole string

const (
	RoleJunior UserRole = "junior"
	RoleMentor UserRole = "mentor"
	RoleAdmin  UserRole = "admin"
)

type User struct {
	ID         int64
	TelegramID int64
	Username   string
	Role       UserRole
	Status     string
}

type ModuleProgress struct {
	ModuleID          int64
	Slug              string
	Title             string
	Description       string
	OrderIndex        int
	LessonTotal       int
	LessonCompleted   int
	QuizBestScore     int
	QuizMaxScore      int
	MasteryScore      float64
	Completed         bool
	NextLessonID      int64
	NextLessonTitle   string
	NextLessonContent string
}

type QuizQuestion struct {
	ID          int64
	ModuleID    int64
	Question    string
	OptionA     string
	OptionB     string
	OptionC     string
	OptionD     string
	Correct     string
	Explanation string
	SourceURL   string
}

type QuizResult struct {
	AttemptID int64
	Score     int
	MaxScore  int
	Mastery   float64
	Details   []QuizAnswerDetail
}

type QuizAnswerDetail struct {
	QuestionID  int64
	Selected    string
	Correct     bool
	CorrectOpt  string
	Explanation string
	SourceURL   string
}

type ChecklistInfo struct {
	ID          int64
	ModuleID    int64
	ModuleSlug  string
	Title       string
	Description string
	Items       []string
}

type LearningResource struct {
	Title string
	URL   string
}

type ChecklistSubmission struct {
	ID            int64
	UserID        int64
	ChecklistID   int64
	Status        string
	Notes         string
	ReviewComment string
	SubmittedAt   time.Time
	ReviewedAt    *time.Time
}

type ReviewCard struct {
	ID            int64
	TopicKey      string
	QuestionID    int64
	Question      string
	OptionA       string
	OptionB       string
	OptionC       string
	OptionD       string
	Correct       string
	Explanation   string
	SourceURL     string
	IntervalStage int
}

type JuniorReport struct {
	TelegramID      int64
	Username        string
	CompletionPct   float64
	MasteryAvg      float64
	WeakModules     string
	Activity7d      int
	Activity30d     int
	LastActivityUTC *time.Time
}
