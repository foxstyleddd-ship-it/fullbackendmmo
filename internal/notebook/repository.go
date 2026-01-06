package notebook

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type Repository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) *Repository {
	return &Repository{db: db}
}

// Notebook operations

func (r *Repository) CreateNotebook(ctx context.Context, notebook *SchoolNotebook) error {
	query := `
		INSERT INTO school_notebooks (character_id, title, content, subject)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at, updated_at
	`
	return r.db.QueryRowContext(ctx, query,
		notebook.CharacterID,
		notebook.Title,
		notebook.Content,
		notebook.Subject,
	).Scan(&notebook.ID, &notebook.CreatedAt, &notebook.UpdatedAt)
}

func (r *Repository) GetNotebookByID(ctx context.Context, id uuid.UUID) (*SchoolNotebook, error) {
	var notebook SchoolNotebook
	query := `
		SELECT id, character_id, title, content, subject, created_at, updated_at
		FROM school_notebooks
		WHERE id = $1 AND deleted_at IS NULL
	`
	err := r.db.GetContext(ctx, &notebook, query, id)
	return &notebook, err
}

func (r *Repository) GetNotebooksByCharacter(ctx context.Context, characterID uuid.UUID) ([]SchoolNotebook, error) {
	var notebooks []SchoolNotebook
	query := `
		SELECT id, character_id, title, content, subject, created_at, updated_at
		FROM school_notebooks
		WHERE character_id = $1 AND deleted_at IS NULL
		ORDER BY updated_at DESC
	`
	err := r.db.SelectContext(ctx, &notebooks, query, characterID)
	return notebooks, err
}

func (r *Repository) UpdateNotebook(ctx context.Context, id uuid.UUID, title, content string, subject sql.NullString) error {
	query := `
		UPDATE school_notebooks
		SET title = $1, content = $2, subject = $3, updated_at = NOW()
		WHERE id = $4 AND deleted_at IS NULL
	`
	_, err := r.db.ExecContext(ctx, query, title, content, subject, id)
	return err
}

func (r *Repository) DeleteNotebook(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE school_notebooks
		SET deleted_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL
	`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

// Grade operations

func (r *Repository) CreateGrade(ctx context.Context, grade *AcademicGrade) error {
	query := `
		INSERT INTO academic_grades (character_id, school_year, subject, grade_value, exam_type, score, teacher_comment, awarded_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, created_at, updated_at
	`
	return r.db.QueryRowContext(ctx, query,
		grade.CharacterID,
		grade.SchoolYear,
		grade.Subject,
		grade.GradeValue,
		grade.ExamType,
		grade.Score,
		grade.TeacherComment,
		grade.AwardedAt,
	).Scan(&grade.ID, &grade.CreatedAt, &grade.UpdatedAt)
}

func (r *Repository) GetGradesByCharacter(ctx context.Context, characterID uuid.UUID) ([]AcademicGrade, error) {
	var grades []AcademicGrade
	query := `
		SELECT id, character_id, school_year, subject, grade_value, exam_type, score, teacher_comment, awarded_at, created_at, updated_at
		FROM academic_grades
		WHERE character_id = $1
		ORDER BY school_year DESC, subject ASC
	`
	err := r.db.SelectContext(ctx, &grades, query, characterID)
	return grades, err
}

func (r *Repository) GetGradesByCharacterAndYear(ctx context.Context, characterID uuid.UUID, schoolYear int) ([]AcademicGrade, error) {
	var grades []AcademicGrade
	query := `
		SELECT id, character_id, school_year, subject, grade_value, exam_type, score, teacher_comment, awarded_at, created_at, updated_at
		FROM academic_grades
		WHERE character_id = $1 AND school_year = $2
		ORDER BY subject ASC
	`
	err := r.db.SelectContext(ctx, &grades, query, characterID, schoolYear)
	return grades, err
}

func (r *Repository) UpdateGrade(ctx context.Context, id uuid.UUID, gradeValue GradeValue, score sql.NullInt64, teacherComment sql.NullString) error {
	query := `
		UPDATE academic_grades
		SET grade_value = $1, score = $2, teacher_comment = $3, updated_at = NOW()
		WHERE id = $4
	`
	_, err := r.db.ExecContext(ctx, query, gradeValue, score, teacherComment, id)
	return err
}

func (r *Repository) DeleteGrade(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM academic_grades WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

// Report Card operations

func (r *Repository) CreateReportCard(ctx context.Context, reportCard *ReportCard) error {
	query := `
		INSERT INTO report_cards (character_id, school_year, overall_comment, headmaster_signature, issued_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at
	`
	return r.db.QueryRowContext(ctx, query,
		reportCard.CharacterID,
		reportCard.SchoolYear,
		reportCard.OverallComment,
		reportCard.HeadmasterSignature,
		reportCard.IssuedAt,
	).Scan(&reportCard.ID, &reportCard.CreatedAt)
}

func (r *Repository) GetReportCardByCharacterAndYear(ctx context.Context, characterID uuid.UUID, schoolYear int) (*ReportCard, error) {
	var reportCard ReportCard
	query := `
		SELECT id, character_id, school_year, overall_comment, headmaster_signature, issued_at, created_at
		FROM report_cards
		WHERE character_id = $1 AND school_year = $2
	`
	err := r.db.GetContext(ctx, &reportCard, query, characterID, schoolYear)
	return &reportCard, err
}

func (r *Repository) GetReportCardsByCharacter(ctx context.Context, characterID uuid.UUID) ([]ReportCard, error) {
	var reportCards []ReportCard
	query := `
		SELECT id, character_id, school_year, overall_comment, headmaster_signature, issued_at, created_at
		FROM report_cards
		WHERE character_id = $1
		ORDER BY school_year DESC
	`
	err := r.db.SelectContext(ctx, &reportCards, query, characterID)
	return reportCards, err
}

func (r *Repository) UpdateReportCard(ctx context.Context, id uuid.UUID, overallComment, headmasterSignature sql.NullString) error {
	query := `
		UPDATE report_cards
		SET overall_comment = $1, headmaster_signature = $2
		WHERE id = $3
	`
	_, err := r.db.ExecContext(ctx, query, overallComment, headmasterSignature, id)
	return err
}

func (r *Repository) DeleteReportCard(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM report_cards WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}
