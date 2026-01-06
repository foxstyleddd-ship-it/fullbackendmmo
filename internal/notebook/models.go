package notebook

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

// SchoolSubject represents a school subject
type SchoolSubject string

const (
	SubjectCharms                 SchoolSubject = "charms"
	SubjectTransfiguration        SchoolSubject = "transfiguration"
	SubjectPotions                SchoolSubject = "potions"
	SubjectDefenseAgainstDarkArts SchoolSubject = "defense_against_dark_arts"
	SubjectHerbology              SchoolSubject = "herbology"
	SubjectAstronomy              SchoolSubject = "astronomy"
	SubjectHistoryOfMagic         SchoolSubject = "history_of_magic"
	SubjectCareOfMagicalCreatures SchoolSubject = "care_of_magical_creatures"
	SubjectDivination             SchoolSubject = "divination"
	SubjectAncientRunes           SchoolSubject = "ancient_runes"
	SubjectArithmancy             SchoolSubject = "arithmancy"
	SubjectMuggleStudies          SchoolSubject = "muggle_studies"
	SubjectFlying                 SchoolSubject = "flying"
)

// GradeValue represents academic grade values
type GradeValue string

const (
	GradeOutstanding        GradeValue = "outstanding"          // O - Optimal
	GradeExceedsExpectations GradeValue = "exceeds_expectations" // E - Effort Exceptionnel
	GradeAcceptable         GradeValue = "acceptable"           // A - Acceptable
	GradePoor               GradeValue = "poor"                 // P - Piètre
	GradeDreadful           GradeValue = "dreadful"             // D - Désolant
	GradeTroll              GradeValue = "troll"                // T - Troll
)

// ExamType represents the type of examination
type ExamType string

const (
	ExamContinuous ExamType = "continuous" // Contrôle continu
	ExamMidterm    ExamType = "midterm"    // Examen de mi-année
	ExamFinal      ExamType = "final"      // Examen final
	ExamOWLs       ExamType = "owls"       // Ordinary Wizarding Levels (BUSEs)
	ExamNEWTs      ExamType = "newts"      // Nastily Exhausting Wizarding Tests (ASPICs)
)

// SchoolNotebook represents a student's notebook for taking notes
type SchoolNotebook struct {
	ID          uuid.UUID      `db:"id" json:"id"`
	CharacterID uuid.UUID      `db:"character_id" json:"character_id"`
	Title       string         `db:"title" json:"title"`
	Content     string         `db:"content" json:"content"`
	Subject     sql.NullString `db:"subject" json:"subject,omitempty"`
	CreatedAt   time.Time      `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time      `db:"updated_at" json:"updated_at"`
	DeletedAt   sql.NullTime   `db:"deleted_at" json:"-"`
}

// AcademicGrade represents a grade for a subject in a specific year
type AcademicGrade struct {
	ID             uuid.UUID     `db:"id" json:"id"`
	CharacterID    uuid.UUID     `db:"character_id" json:"character_id"`
	SchoolYear     int           `db:"school_year" json:"school_year"`
	Subject        SchoolSubject `db:"subject" json:"subject"`
	GradeValue     GradeValue    `db:"grade_value" json:"grade_value"`
	ExamType       string        `db:"exam_type" json:"exam_type"`
	Score          sql.NullInt64 `db:"score" json:"score,omitempty"`
	TeacherComment sql.NullString `db:"teacher_comment" json:"teacher_comment,omitempty"`
	AwardedAt      time.Time     `db:"awarded_at" json:"awarded_at"`
	CreatedAt      time.Time     `db:"created_at" json:"created_at"`
	UpdatedAt      time.Time     `db:"updated_at" json:"updated_at"`
}

// ReportCard represents an end-of-year report card
type ReportCard struct {
	ID                  uuid.UUID      `db:"id" json:"id"`
	CharacterID         uuid.UUID      `db:"character_id" json:"character_id"`
	SchoolYear          int            `db:"school_year" json:"school_year"`
	OverallComment      sql.NullString `db:"overall_comment" json:"overall_comment,omitempty"`
	HeadmasterSignature sql.NullString `db:"headmaster_signature" json:"headmaster_signature,omitempty"`
	IssuedAt            time.Time      `db:"issued_at" json:"issued_at"`
	CreatedAt           time.Time      `db:"created_at" json:"created_at"`
}

// ReportCardWithGrades combines a report card with its grades
type ReportCardWithGrades struct {
	ReportCard
	Grades []AcademicGrade `json:"grades"`
}

// CreateNotebookRequest represents a request to create a notebook
type CreateNotebookRequest struct {
	CharacterID uuid.UUID `json:"character_id" binding:"required"`
	Title       string    `json:"title" binding:"required,min=1,max=128"`
	Content     string    `json:"content"`
	Subject     *string   `json:"subject,omitempty"`
}

// UpdateNotebookRequest represents a request to update a notebook
type UpdateNotebookRequest struct {
	Title   *string `json:"title,omitempty"`
	Content *string `json:"content,omitempty"`
	Subject *string `json:"subject,omitempty"`
}

// CreateGradeRequest represents a request to create a grade
type CreateGradeRequest struct {
	CharacterID    uuid.UUID `json:"character_id" binding:"required"`
	SchoolYear     int       `json:"school_year" binding:"required,min=1,max=7"`
	Subject        string    `json:"subject" binding:"required"`
	GradeValue     string    `json:"grade_value" binding:"required"`
	ExamType       string    `json:"exam_type"`
	Score          *int      `json:"score,omitempty"`
	TeacherComment *string   `json:"teacher_comment,omitempty"`
}

// CreateReportCardRequest represents a request to create a report card
type CreateReportCardRequest struct {
	CharacterID         uuid.UUID `json:"character_id" binding:"required"`
	SchoolYear          int       `json:"school_year" binding:"required,min=1,max=7"`
	OverallComment      *string   `json:"overall_comment,omitempty"`
	HeadmasterSignature *string   `json:"headmaster_signature,omitempty"`
}
