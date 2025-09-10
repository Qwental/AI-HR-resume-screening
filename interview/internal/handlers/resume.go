package handlers

import (
	"net/http"
	_ "strconv"

	"github.com/gin-gonic/gin"
	"interview/internal/models"
	"interview/internal/service"
)

type ResumeHandler struct {
	svc service.ResumeService
}

func NewResumeHandler(svc service.ResumeService) *ResumeHandler {
	return &ResumeHandler{svc: svc}
}

// POST /api/resumes
func (h *ResumeHandler) Create(c *gin.Context) {
	vacancyID := c.PostForm("vacancy_id")

	if vacancyID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "vacancy_id is required"})
		return
	}

	fileHeader, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file is required"})
		return
	}

	file, err := fileHeader.Open()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to open file"})
		return
	}
	defer file.Close()

	resume := &models.Resume{
		VacancyID: vacancyID,
	}

	err = h.svc.CreateResume(c.Request.Context(), resume, file, fileHeader.Filename)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, resume)
}

// GET /api/resumes/:id
func (h *ResumeHandler) GetByID(c *gin.Context) {
	id := c.Param("id")

	resume, err := h.svc.GetResume(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "resume not found"})
		return
	}

	c.JSON(http.StatusOK, resume)
}

// GET /api/resumes/:id/download
func (h *ResumeHandler) GetDownloadLink(c *gin.Context) {
	id := c.Param("id")

	resume, err := h.svc.GetResumeWithFileURL(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "resume not found"})
		return
	}

	if resume.FileURL == "" {
		c.JSON(http.StatusNotFound, gin.H{"error": "file not found"})
		return
	}

	c.JSON(http.StatusOK, resume)

}

// GET /api/vacancies/:vacancy_id/resumes
//func (h *ResumeHandler) GetByVacancy(c *gin.Context) {
//	vacancyID := c.Param("vacancy_id")
//
//	resumes, err := h.svc.GetResumesByVacancy(c.Request.Context(), vacancyID)
//	if err != nil {
//		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get resumes"})
//		return
//	}
//
//	c.JSON(http.StatusOK, gin.H{
//		"count":   len(resumes),
//		"resumes": resumes,
//	})
//}

// GET /api/vacancies/:id/resumes
func (h *ResumeHandler) GetByVacancy(c *gin.Context) {
	vacancyID := c.Param("id") // ← ИСПРАВЛЕНО: используй "id" вместо "vacancy_id"
	resumes, err := h.svc.GetResumesByVacancy(c.Request.Context(), vacancyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get resumes"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"count":   len(resumes),
		"resumes": resumes,
	})
}

// PUT /api/resumes/:id/status
func (h *ResumeHandler) UpdateStatus(c *gin.Context) {
	id := c.Param("id")

	var req struct {
		Status string                 `json:"status" binding:"required"`
		Result map[string]interface{} `json:"result"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request: " + err.Error()})
		return
	}

	err := h.svc.UpdateStatusAndResult(c.Request.Context(), id, req.Status, req.Result)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "status updated successfully"})
}

// DELETE /api/resumes/:id
func (h *ResumeHandler) Delete(c *gin.Context) {
	id := c.Param("id")

	err := h.svc.DeleteResume(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "resume deleted successfully"})
}

// DELETE /api/resumes/:id
