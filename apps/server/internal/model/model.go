package model

import (
	"database/sql/driver"
	"fmt"
	"time"
)

// ─── Submission status constants ────────────────────────────────────────────

const (
	SubmissionStatusPending      = "PENDING"
	SubmissionStatusInQueue      = "IN_QUEUE"
	SubmissionStatusProcessing   = "PROCESSING"
	SubmissionStatusAccepted     = "ACCEPTED"
	SubmissionStatusWrongAnswer  = "WRONG_ANSWER"
	SubmissionStatusTimeLimit    = "TIME_LIMIT_EXCEEDED"
	SubmissionStatusMemoryLimit  = "MEMORY_LIMIT_EXCEEDED"
	SubmissionStatusRuntimeError = "RUNTIME_ERROR"
	SubmissionStatusCompileError = "COMPILE_ERROR"
	SubmissionStatusSystemError  = "SYSTEM_ERROR"
)

// ─── Difficulty constants ──────────────────────────────────────────────────

const (
	DifficultyEasy   = "EASY"
	DifficultyMedium = "MEDIUM"
	DifficultyHard   = "HARD"
)

// ─── Stage constants ───────────────────────────────────────────────────────

const (
	StageHot100 = "hot100"
)

// ─── User status constants ──────────────────────────────────────────────────

const (
	UserStatusPending  = "PENDING"
	UserStatusActive   = "ACTIVE"
	UserStatusDisabled = "DISABLED"
)

// ─── Progress state constants ──────────────────────────────────────────────

const (
	ProgressNotStarted = "NOT_STARTED"
	ProgressAttempted  = "ATTEMPTED"
	ProgressSolved     = "SOLVED"
)

// ─── User ──────────────────────────────────────────────────────────────────

// User represents a registered user.
type User struct {
	ID              string     `gorm:"type:char(36);primaryKey" json:"id"`
	Email           string     `gorm:"type:varchar(320);uniqueIndex;not null" json:"email"`
	Username        string     `gorm:"type:varchar(32);uniqueIndex;not null" json:"username"`
	PasswordHash    string     `gorm:"type:text;not null" json:"-"`
	EmailVerifiedAt *time.Time `gorm:"type:datetime(6)" json:"email_verified_at"`
	Status          string     `gorm:"type:varchar(20);not null;default:'PENDING'" json:"status"`
	CreatedAt       time.Time  `gorm:"type:datetime(6);autoCreateTime" json:"created_at"`
	UpdatedAt       time.Time  `gorm:"type:datetime(6);autoUpdateTime" json:"updated_at"`
}

func (User) TableName() string { return "users" }

// ─── Problem ───────────────────────────────────────────────────────────────

// Problem represents a coding problem in the Hot 100 list.
type Problem struct {
	ID             string    `gorm:"type:char(36);primaryKey" json:"id"`
	Slug           string    `gorm:"type:varchar(80);uniqueIndex;not null" json:"slug"`
	OrderIndex     int       `gorm:"uniqueIndex;not null" json:"order_index"`
	Title          string    `gorm:"type:varchar(120);not null" json:"title"`
	Difficulty     string    `gorm:"type:varchar(20);not null" json:"difficulty"`
	Stage          string    `gorm:"type:varchar(40);not null" json:"stage"`
	StatementMD    string    `gorm:"type:text;not null;column:statement_md" json:"statement_md"`
	InputFormatMD  string    `gorm:"type:text;not null;column:input_format_md" json:"input_format_md"`
	OutputFormatMD string    `gorm:"type:text;not null;column:output_format_md" json:"output_format_md"`
	ConstraintsMD  string    `gorm:"type:text;not null;column:constraints_md" json:"constraints_md"`
	HintsMD        string    `gorm:"type:text;column:hints_md" json:"hints_md"`
	TimeLimitMs    int       `gorm:"not null;default:1000" json:"time_limit_ms"`
	MemoryLimitKb  int       `gorm:"not null;default:262144" json:"memory_limit_kb"`
	IsPublished    bool      `gorm:"not null;default:false" json:"is_published"`
	CreatedAt      time.Time `gorm:"type:datetime(6);autoCreateTime" json:"created_at"`
	UpdatedAt      time.Time `gorm:"type:datetime(6);autoUpdateTime" json:"updated_at"`

	Tags      []Tag      `gorm:"many2many:problem_tags" json:"tags,omitempty"`
	TestCases []TestCase `gorm:"foreignKey:ProblemID" json:"test_cases,omitempty"`
}

func (Problem) TableName() string { return "problems" }

// ─── Tag ───────────────────────────────────────────────────────────────────

// Tag represents a problem classification tag.
type Tag struct {
	ID   string `gorm:"type:char(36);primaryKey" json:"id"`
	Slug string `gorm:"type:varchar(40);uniqueIndex;not null" json:"slug"`
	Name string `gorm:"type:varchar(60);not null" json:"name"`
}

func (Tag) TableName() string { return "tags" }

// ─── ProblemTag ────────────────────────────────────────────────────────────

// ProblemTag is the join table for the many-to-many relationship between problems and tags.
type ProblemTag struct {
	ProblemID string `gorm:"type:char(36);primaryKey;not null" json:"problem_id"`
	TagID     string `gorm:"type:char(36);primaryKey;not null" json:"tag_id"`
}

func (ProblemTag) TableName() string { return "problem_tags" }

// ─── TestCase ──────────────────────────────────────────────────────────────

// TestCase represents a single test case for a problem.
type TestCase struct {
	ID             string    `gorm:"type:char(36);primaryKey" json:"id"`
	ProblemID      string    `gorm:"type:char(36);not null;uniqueIndex:idx_problem_order" json:"problem_id"`
	OrderIndex     int       `gorm:"not null;uniqueIndex:idx_problem_order" json:"order_index"`
	InputData      string    `gorm:"type:mediumtext;not null;column:input_data" json:"input_data"`
	ExpectedOutput string    `gorm:"type:mediumtext;not null;column:expected_output" json:"expected_output"`
	IsSample       bool      `gorm:"not null;default:false" json:"is_sample"`
	ExplanationMD  string    `gorm:"type:text;column:explanation_md" json:"explanation_md"`
	CreatedAt      time.Time `gorm:"type:datetime(6);autoCreateTime" json:"created_at"`
}

func (TestCase) TableName() string { return "test_cases" }

// ─── LanguageConfig ────────────────────────────────────────────────────────

// LanguageConfig stores supported programming language configuration.
type LanguageConfig struct {
	Key                string `gorm:"type:varchar(20);primaryKey" json:"key"`
	DisplayName        string `gorm:"type:varchar(40);not null;column:display_name" json:"display_name"`
	Judge0LanguageName string `gorm:"type:varchar(120);not null;column:judge0_language_name" json:"judge0_language_name"`
	Judge0LanguageID   *int   `gorm:"column:judge0_language_id" json:"judge0_language_id"`
	EditorLanguage     string `gorm:"type:varchar(30);not null;column:editor_language" json:"editor_language"`
	SourceTemplate     string `gorm:"type:text;not null;column:source_template" json:"source_template"`
	Enabled            bool   `gorm:"not null;default:true" json:"enabled"`
}

func (LanguageConfig) TableName() string { return "language_configs" }

// ─── Draft ─────────────────────────────────────────────────────────────────

// Draft stores a user's unsaved source code for a problem+language combination.
type Draft struct {
	UserID      string    `gorm:"type:char(36);primaryKey;not null" json:"user_id"`
	ProblemID   string    `gorm:"type:char(36);primaryKey;not null" json:"problem_id"`
	LanguageKey string    `gorm:"type:varchar(20);primaryKey;not null;column:language_key" json:"language_key"`
	SourceCode  string    `gorm:"type:text;not null;column:source_code" json:"source_code"`
	UpdatedAt   time.Time `gorm:"type:datetime(6);autoUpdateTime" json:"updated_at"`
}

func (Draft) TableName() string { return "drafts" }

// ─── Submission ────────────────────────────────────────────────────────────

// Submission represents a user's code submission for judging.
type Submission struct {
	ID              string     `gorm:"type:char(36);primaryKey" json:"id"`
	UserID          string     `gorm:"type:char(36);not null;index" json:"user_id"`
	ProblemID       string     `gorm:"type:char(36);not null;index" json:"problem_id"`
	LanguageKey     string     `gorm:"type:varchar(20);not null;column:language_key" json:"language_key"`
	SourceCode      string     `gorm:"type:text;not null;column:source_code" json:"source_code"`
	Status          string     `gorm:"type:varchar(20);not null;default:'PENDING';index" json:"status"`
	PassedCases     int        `gorm:"not null;default:0;column:passed_cases" json:"passed_cases"`
	TotalCases      int        `gorm:"not null;default:0;column:total_cases" json:"total_cases"`
	TimeMs          *int       `gorm:"column:time_ms" json:"time_ms"`
	MemoryKb        *int       `gorm:"column:memory_kb" json:"memory_kb"`
	CompilerOutput  string     `gorm:"type:text;column:compiler_output" json:"compiler_output"`
	ErrorMessage    string     `gorm:"type:text;column:error_message" json:"error_message"`
	StreamMessageID string     `gorm:"type:varchar(64);column:stream_message_id" json:"stream_message_id"`
	EnqueuedAt      *time.Time `gorm:"type:datetime(6);column:enqueued_at" json:"enqueued_at"`
	ClaimedAt       *time.Time `gorm:"type:datetime(6);column:claimed_at" json:"claimed_at"`
	RetryCount      int        `gorm:"not null;default:0;column:retry_count" json:"retry_count"`
	CreatedAt       time.Time  `gorm:"type:datetime(6);autoCreateTime" json:"created_at"`
	JudgedAt        *time.Time `gorm:"type:datetime(6);column:judged_at" json:"judged_at"`

	CaseResults []SubmissionCaseResult `gorm:"foreignKey:SubmissionID" json:"case_results,omitempty"`
}

func (Submission) TableName() string { return "submissions" }

// ─── SubmissionCaseResult ──────────────────────────────────────────────────

// SubmissionCaseResult stores the result of a single test case within a submission.
type SubmissionCaseResult struct {
	ID             string `gorm:"type:char(36);primaryKey" json:"id"`
	SubmissionID   string `gorm:"type:char(36);not null;index;column:submission_id" json:"submission_id"`
	CaseIndex      int    `gorm:"not null;column:case_index" json:"case_index"`
	Status         string `gorm:"type:varchar(20);not null" json:"status"`
	TimeMs         *int   `gorm:"column:time_ms" json:"time_ms"`
	MemoryKb       *int   `gorm:"column:memory_kb" json:"memory_kb"`
	ActualOutput   string `gorm:"type:text;column:actual_output" json:"actual_output"`
	ExpectedOutput string `gorm:"type:text;column:expected_output" json:"expected_output"`
	IsSample       bool   `gorm:"not null;default:false" json:"is_sample"`
}

func (SubmissionCaseResult) TableName() string { return "submission_case_results" }

// ─── UserProblemProgress ──────────────────────────────────────────────────

// UserProblemProgress tracks a user's progress on a specific problem.
type UserProblemProgress struct {
	UserID          string     `gorm:"type:char(36);primaryKey;not null" json:"user_id"`
	ProblemID       string     `gorm:"type:char(36);primaryKey;not null" json:"problem_id"`
	State           string     `gorm:"type:varchar(20);not null;default:'NOT_STARTED'" json:"state"`
	AttemptCount    int        `gorm:"not null;default:0" json:"attempt_count"`
	FirstACAt       *time.Time `gorm:"type:datetime(6);column:first_ac_at" json:"first_ac_at"`
	LastSubmittedAt *time.Time `gorm:"type:datetime(6);column:last_submitted_at" json:"last_submitted_at"`
}

func (UserProblemProgress) TableName() string { return "user_problem_progress" }

// ─── NullTime ──────────────────────────────────────────────────────────────

// NullTime wraps time.Time for nullable datetime fields.
type NullTime struct {
	Time  time.Time
	Valid bool
}

// Scan implements the sql.Scanner interface.
func (nt *NullTime) Scan(value interface{}) error {
	if value == nil {
		nt.Valid = false
		return nil
	}
	switch v := value.(type) {
	case time.Time:
		nt.Time = v
		nt.Valid = true
	default:
		return fmt.Errorf("cannot scan %T into NullTime", value)
	}
	return nil
}

// Value implements the driver.Valuer interface.
func (nt NullTime) Value() (driver.Value, error) {
	if !nt.Valid {
		return nil, nil
	}
	return nt.Time, nil
}
