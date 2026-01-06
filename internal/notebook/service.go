package notebook

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

// Notebook operations

func (s *Service) CreateNotebook(ctx context.Context, req CreateNotebookRequest) (*SchoolNotebook, error) {
	notebook := &SchoolNotebook{
		CharacterID: req.CharacterID,
		Title:       req.Title,
		Content:     req.Content,
	}

	if req.Subject != nil && *req.Subject != "" {
		notebook.Subject = sql.NullString{String: *req.Subject, Valid: true}
	}

	if err := s.repo.CreateNotebook(ctx, notebook); err != nil {
		return nil, fmt.Errorf("failed to create notebook: %w", err)
	}

	return notebook, nil
}

func (s *Service) GetNotebook(ctx context.Context, id uuid.UUID) (*SchoolNotebook, error) {
	notebook, err := s.repo.GetNotebookByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get notebook: %w", err)
	}
	return notebook, nil
}

func (s *Service) GetCharacterNotebooks(ctx context.Context, characterID uuid.UUID) ([]SchoolNotebook, error) {
	notebooks, err := s.repo.GetNotebooksByCharacter(ctx, characterID)
	if err != nil {
		return nil, fmt.Errorf("failed to get character notebooks: %w", err)
	}
	return notebooks, nil
}

func (s *Service) UpdateNotebook(ctx context.Context, id uuid.UUID, req UpdateNotebookRequest) error {
	// Get existing notebook
	notebook, err := s.repo.GetNotebookByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get notebook: %w", err)
	}

	// Update fields
	title := notebook.Title
	content := notebook.Content
	subject := notebook.Subject

	if req.Title != nil {
		title = *req.Title
	}
	if req.Content != nil {
		content = *req.Content
	}
	if req.Subject != nil {
		if *req.Subject == "" {
			subject = sql.NullString{Valid: false}
		} else {
			subject = sql.NullString{String: *req.Subject, Valid: true}
		}
	}

	if err := s.repo.UpdateNotebook(ctx, id, title, content, subject); err != nil {
		return fmt.Errorf("failed to update notebook: %w", err)
	}

	return nil
}

func (s *Service) DeleteNotebook(ctx context.Context, id uuid.UUID) error {
	if err := s.repo.DeleteNotebook(ctx, id); err != nil {
		return fmt.Errorf("failed to delete notebook: %w", err)
	}
	return nil
}

// Grade operations

func (s *Service) CreateGrade(ctx context.Context, req CreateGradeRequest) (*AcademicGrade, error) {
	grade := &AcademicGrade{
		CharacterID: req.CharacterID,
		SchoolYear:  req.SchoolYear,
		Subject:     SchoolSubject(req.Subject),
		GradeValue:  GradeValue(req.GradeValue),
		ExamType:    req.ExamType,
		AwardedAt:   time.Now(),
	}

	if grade.ExamType == "" {
		grade.ExamType = string(ExamContinuous)
	}

	if req.Score != nil {
		grade.Score = sql.NullInt64{Int64: int64(*req.Score), Valid: true}
	}

	if req.TeacherComment != nil && *req.TeacherComment != "" {
		grade.TeacherComment = sql.NullString{String: *req.TeacherComment, Valid: true}
	}

	if err := s.repo.CreateGrade(ctx, grade); err != nil {
		return nil, fmt.Errorf("failed to create grade: %w", err)
	}

	return grade, nil
}

func (s *Service) GetCharacterGrades(ctx context.Context, characterID uuid.UUID) ([]AcademicGrade, error) {
	grades, err := s.repo.GetGradesByCharacter(ctx, characterID)
	if err != nil {
		return nil, fmt.Errorf("failed to get character grades: %w", err)
	}
	return grades, nil
}

func (s *Service) GetCharacterGradesByYear(ctx context.Context, characterID uuid.UUID, schoolYear int) ([]AcademicGrade, error) {
	grades, err := s.repo.GetGradesByCharacterAndYear(ctx, characterID, schoolYear)
	if err != nil {
		return nil, fmt.Errorf("failed to get character grades by year: %w", err)
	}
	return grades, nil
}

func (s *Service) DeleteGrade(ctx context.Context, id uuid.UUID) error {
	if err := s.repo.DeleteGrade(ctx, id); err != nil {
		return fmt.Errorf("failed to delete grade: %w", err)
	}
	return nil
}

// Report Card operations

func (s *Service) CreateReportCard(ctx context.Context, req CreateReportCardRequest) (*ReportCardWithGrades, error) {
	reportCard := &ReportCard{
		CharacterID: req.CharacterID,
		SchoolYear:  req.SchoolYear,
		IssuedAt:    time.Now(),
	}

	if req.OverallComment != nil && *req.OverallComment != "" {
		reportCard.OverallComment = sql.NullString{String: *req.OverallComment, Valid: true}
	}

	if req.HeadmasterSignature != nil && *req.HeadmasterSignature != "" {
		reportCard.HeadmasterSignature = sql.NullString{String: *req.HeadmasterSignature, Valid: true}
	}

	if err := s.repo.CreateReportCard(ctx, reportCard); err != nil {
		return nil, fmt.Errorf("failed to create report card: %w", err)
	}

	// Get grades for this year
	grades, err := s.repo.GetGradesByCharacterAndYear(ctx, req.CharacterID, req.SchoolYear)
	if err != nil {
		return nil, fmt.Errorf("failed to get grades: %w", err)
	}

	return &ReportCardWithGrades{
		ReportCard: *reportCard,
		Grades:     grades,
	}, nil
}

func (s *Service) GetReportCard(ctx context.Context, characterID uuid.UUID, schoolYear int) (*ReportCardWithGrades, error) {
	reportCard, err := s.repo.GetReportCardByCharacterAndYear(ctx, characterID, schoolYear)
	if err != nil {
		return nil, fmt.Errorf("failed to get report card: %w", err)
	}

	grades, err := s.repo.GetGradesByCharacterAndYear(ctx, characterID, schoolYear)
	if err != nil {
		return nil, fmt.Errorf("failed to get grades: %w", err)
	}

	return &ReportCardWithGrades{
		ReportCard: *reportCard,
		Grades:     grades,
	}, nil
}

func (s *Service) GetAllReportCards(ctx context.Context, characterID uuid.UUID) ([]ReportCardWithGrades, error) {
	reportCards, err := s.repo.GetReportCardsByCharacter(ctx, characterID)
	if err != nil {
		return nil, fmt.Errorf("failed to get report cards: %w", err)
	}

	result := make([]ReportCardWithGrades, len(reportCards))
	for i, rc := range reportCards {
		grades, err := s.repo.GetGradesByCharacterAndYear(ctx, characterID, rc.SchoolYear)
		if err != nil {
			return nil, fmt.Errorf("failed to get grades for year %d: %w", rc.SchoolYear, err)
		}
		result[i] = ReportCardWithGrades{
			ReportCard: rc,
			Grades:     grades,
		}
	}

	return result, nil
}

func (s *Service) DeleteReportCard(ctx context.Context, id uuid.UUID) error {
	if err := s.repo.DeleteReportCard(ctx, id); err != nil {
		return fmt.Errorf("failed to delete report card: %w", err)
	}
	return nil
}
