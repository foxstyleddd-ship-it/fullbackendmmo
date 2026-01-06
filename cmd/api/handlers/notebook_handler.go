package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/hp-mmo/backend/internal/notebook"
	"github.com/hp-mmo/backend/internal/pkg/middleware"
)

type NotebookHandler struct {
	notebookService *notebook.Service
}

func NewNotebookHandler(notebookService *notebook.Service) *NotebookHandler {
	return &NotebookHandler{
		notebookService: notebookService,
	}
}

// CreateNotebook creates a new notebook
func (h *NotebookHandler) CreateNotebook(c *gin.Context) {
	var req notebook.CreateNotebookRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.ErrorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body", err.Error())
		return
	}

	createdNotebook, err := h.notebookService.CreateNotebook(c.Request.Context(), req)
	if err != nil {
		middleware.ErrorResponse(c, http.StatusInternalServerError, "CREATE_FAILED", "Failed to create notebook", err.Error())
		return
	}

	middleware.SuccessResponse(c, http.StatusCreated, gin.H{
		"notebook": createdNotebook,
		"message":  "Notebook created successfully",
	})
}

// GetNotebook gets a notebook by ID
func (h *NotebookHandler) GetNotebook(c *gin.Context) {
	notebookIDStr := c.Param("id")
	notebookID, err := uuid.Parse(notebookIDStr)
	if err != nil {
		middleware.ErrorResponse(c, http.StatusBadRequest, "INVALID_UUID", "Invalid notebook ID", err.Error())
		return
	}

	notebookData, err := h.notebookService.GetNotebook(c.Request.Context(), notebookID)
	if err != nil {
		middleware.ErrorResponse(c, http.StatusNotFound, "NOTEBOOK_NOT_FOUND", "Notebook not found", err.Error())
		return
	}

	middleware.SuccessResponse(c, http.StatusOK, gin.H{
		"notebook": notebookData,
	})
}

// GetCharacterNotebooks gets all notebooks for a character
func (h *NotebookHandler) GetCharacterNotebooks(c *gin.Context) {
	characterIDStr := c.Param("id")
	characterID, err := uuid.Parse(characterIDStr)
	if err != nil {
		middleware.ErrorResponse(c, http.StatusBadRequest, "INVALID_UUID", "Invalid character ID", err.Error())
		return
	}

	notebooks, err := h.notebookService.GetCharacterNotebooks(c.Request.Context(), characterID)
	if err != nil {
		middleware.ErrorResponse(c, http.StatusInternalServerError, "QUERY_FAILED", "Failed to get notebooks", err.Error())
		return
	}

	middleware.SuccessResponse(c, http.StatusOK, gin.H{
		"notebooks": notebooks,
		"count":     len(notebooks),
	})
}

// UpdateNotebook updates a notebook
func (h *NotebookHandler) UpdateNotebook(c *gin.Context) {
	notebookIDStr := c.Param("id")
	notebookID, err := uuid.Parse(notebookIDStr)
	if err != nil {
		middleware.ErrorResponse(c, http.StatusBadRequest, "INVALID_UUID", "Invalid notebook ID", err.Error())
		return
	}

	var req notebook.UpdateNotebookRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.ErrorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body", err.Error())
		return
	}

	if err := h.notebookService.UpdateNotebook(c.Request.Context(), notebookID, req); err != nil {
		middleware.ErrorResponse(c, http.StatusInternalServerError, "UPDATE_FAILED", "Failed to update notebook", err.Error())
		return
	}

	middleware.SuccessResponse(c, http.StatusOK, gin.H{
		"message": "Notebook updated successfully",
	})
}

// DeleteNotebook deletes a notebook
func (h *NotebookHandler) DeleteNotebook(c *gin.Context) {
	notebookIDStr := c.Param("id")
	notebookID, err := uuid.Parse(notebookIDStr)
	if err != nil {
		middleware.ErrorResponse(c, http.StatusBadRequest, "INVALID_UUID", "Invalid notebook ID", err.Error())
		return
	}

	if err := h.notebookService.DeleteNotebook(c.Request.Context(), notebookID); err != nil {
		middleware.ErrorResponse(c, http.StatusInternalServerError, "DELETE_FAILED", "Failed to delete notebook", err.Error())
		return
	}

	middleware.SuccessResponse(c, http.StatusOK, gin.H{
		"message": "Notebook deleted successfully",
	})
}

// CreateGrade creates a new academic grade
func (h *NotebookHandler) CreateGrade(c *gin.Context) {
	var req notebook.CreateGradeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.ErrorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body", err.Error())
		return
	}

	createdGrade, err := h.notebookService.CreateGrade(c.Request.Context(), req)
	if err != nil {
		middleware.ErrorResponse(c, http.StatusInternalServerError, "CREATE_FAILED", "Failed to create grade", err.Error())
		return
	}

	middleware.SuccessResponse(c, http.StatusCreated, gin.H{
		"grade":   createdGrade,
		"message": "Grade created successfully",
	})
}

// GetCharacterGrades gets all grades for a character
func (h *NotebookHandler) GetCharacterGrades(c *gin.Context) {
	characterIDStr := c.Param("id")
	characterID, err := uuid.Parse(characterIDStr)
	if err != nil {
		middleware.ErrorResponse(c, http.StatusBadRequest, "INVALID_UUID", "Invalid character ID", err.Error())
		return
	}

	// Check if year filter is provided
	yearStr := c.Query("year")
	if yearStr != "" {
		year, err := strconv.Atoi(yearStr)
		if err != nil || year < 1 || year > 7 {
			middleware.ErrorResponse(c, http.StatusBadRequest, "INVALID_YEAR", "Invalid school year (must be 1-7)", "")
			return
		}

		grades, err := h.notebookService.GetCharacterGradesByYear(c.Request.Context(), characterID, year)
		if err != nil {
			middleware.ErrorResponse(c, http.StatusInternalServerError, "QUERY_FAILED", "Failed to get grades", err.Error())
			return
		}

		middleware.SuccessResponse(c, http.StatusOK, gin.H{
			"grades":      grades,
			"count":       len(grades),
			"school_year": year,
		})
		return
	}

	// Get all grades
	grades, err := h.notebookService.GetCharacterGrades(c.Request.Context(), characterID)
	if err != nil {
		middleware.ErrorResponse(c, http.StatusInternalServerError, "QUERY_FAILED", "Failed to get grades", err.Error())
		return
	}

	middleware.SuccessResponse(c, http.StatusOK, gin.H{
		"grades": grades,
		"count":  len(grades),
	})
}

// DeleteGrade deletes a grade
func (h *NotebookHandler) DeleteGrade(c *gin.Context) {
	gradeIDStr := c.Param("id")
	gradeID, err := uuid.Parse(gradeIDStr)
	if err != nil {
		middleware.ErrorResponse(c, http.StatusBadRequest, "INVALID_UUID", "Invalid grade ID", err.Error())
		return
	}

	if err := h.notebookService.DeleteGrade(c.Request.Context(), gradeID); err != nil {
		middleware.ErrorResponse(c, http.StatusInternalServerError, "DELETE_FAILED", "Failed to delete grade", err.Error())
		return
	}

	middleware.SuccessResponse(c, http.StatusOK, gin.H{
		"message": "Grade deleted successfully",
	})
}

// CreateReportCard creates a new report card
func (h *NotebookHandler) CreateReportCard(c *gin.Context) {
	var req notebook.CreateReportCardRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.ErrorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body", err.Error())
		return
	}

	reportCard, err := h.notebookService.CreateReportCard(c.Request.Context(), req)
	if err != nil {
		middleware.ErrorResponse(c, http.StatusInternalServerError, "CREATE_FAILED", "Failed to create report card", err.Error())
		return
	}

	middleware.SuccessResponse(c, http.StatusCreated, gin.H{
		"report_card": reportCard,
		"message":     "Report card created successfully",
	})
}

// GetReportCard gets a report card for a specific year
func (h *NotebookHandler) GetReportCard(c *gin.Context) {
	characterIDStr := c.Param("id")
	characterID, err := uuid.Parse(characterIDStr)
	if err != nil {
		middleware.ErrorResponse(c, http.StatusBadRequest, "INVALID_UUID", "Invalid character ID", err.Error())
		return
	}

	yearStr := c.Param("year")
	year, err := strconv.Atoi(yearStr)
	if err != nil || year < 1 || year > 7 {
		middleware.ErrorResponse(c, http.StatusBadRequest, "INVALID_YEAR", "Invalid school year (must be 1-7)", "")
		return
	}

	reportCard, err := h.notebookService.GetReportCard(c.Request.Context(), characterID, year)
	if err != nil {
		middleware.ErrorResponse(c, http.StatusNotFound, "NOT_FOUND", "Report card not found", err.Error())
		return
	}

	middleware.SuccessResponse(c, http.StatusOK, gin.H{
		"report_card": reportCard,
	})
}

// GetAllReportCards gets all report cards for a character
func (h *NotebookHandler) GetAllReportCards(c *gin.Context) {
	characterIDStr := c.Param("id")
	characterID, err := uuid.Parse(characterIDStr)
	if err != nil {
		middleware.ErrorResponse(c, http.StatusBadRequest, "INVALID_UUID", "Invalid character ID", err.Error())
		return
	}

	reportCards, err := h.notebookService.GetAllReportCards(c.Request.Context(), characterID)
	if err != nil {
		middleware.ErrorResponse(c, http.StatusInternalServerError, "QUERY_FAILED", "Failed to get report cards", err.Error())
		return
	}

	middleware.SuccessResponse(c, http.StatusOK, gin.H{
		"report_cards": reportCards,
		"count":        len(reportCards),
	})
}

// DeleteReportCard deletes a report card
func (h *NotebookHandler) DeleteReportCard(c *gin.Context) {
	reportCardIDStr := c.Param("id")
	reportCardID, err := uuid.Parse(reportCardIDStr)
	if err != nil {
		middleware.ErrorResponse(c, http.StatusBadRequest, "INVALID_UUID", "Invalid report card ID", err.Error())
		return
	}

	if err := h.notebookService.DeleteReportCard(c.Request.Context(), reportCardID); err != nil {
		middleware.ErrorResponse(c, http.StatusInternalServerError, "DELETE_FAILED", "Failed to delete report card", err.Error())
		return
	}

	middleware.SuccessResponse(c, http.StatusOK, gin.H{
		"message": "Report card deleted successfully",
	})
}
