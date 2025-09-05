package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"interview/internal/models"
	"interview/internal/service"
)

type VacancyHandler struct {
	svc service.VacancyService
}

func NewVacancyHandler(svc service.VacancyService) *VacancyHandler {
	return &VacancyHandler{svc: svc}
}

// POST /api/vacancies
func (h *VacancyHandler) Create(c *gin.Context) {
	title := c.PostForm("title")
	description := c.PostForm("description")
	usersID := c.PostForm("users_id")

	weightSoft, _ := strconv.Atoi(c.DefaultPostForm("weight_soft", "33"))
	weightHard, _ := strconv.Atoi(c.DefaultPostForm("weight_hard", "33"))
	weightCase, _ := strconv.Atoi(c.DefaultPostForm("weight_case", "34"))

	if title == "" || usersID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "title and users_id are required"})
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

	vacancy := &models.Vacancy{
		Title:       title,
		Description: &description,
		UsersID:     usersID,
		WeightSoft:  weightSoft,
		WeightHard:  weightHard,
		WeightCase:  weightCase,
	}

	err = h.svc.CreateVacancy(c.Request.Context(), vacancy, file, fileHeader.Filename)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, vacancy)
}

// GET /api/vacancies - get all vacancies
func (h *VacancyHandler) GetAll(c *gin.Context) {
	vacancies, err := h.svc.GetAllVacancies(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get vacancies"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"count":     len(vacancies),
		"vacancies": vacancies,
	})
}

// GET /api/vacancies/:id
func (h *VacancyHandler) GetByID(c *gin.Context) {
	id := c.Param("id")

	vacancy, err := h.svc.GetVacancy(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "vacancy not found"})
		return
	}

	c.JSON(http.StatusOK, vacancy)
}

// GET /api/vacancies/:id/download - get file download link
func (h *VacancyHandler) GetDownloadLink(c *gin.Context) {
	id := c.Param("id")

	vacancy, err := h.svc.GetVacancyWithFileURL(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "vacancy not found"})
		return
	}

	if vacancy.FileURL == "" {
		c.JSON(http.StatusNotFound, gin.H{"error": "file not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"download_url": vacancy.FileURL,
		"expires_in":   3600, // 1 hour in seconds
	})
}

// PUT /api/vacancies/:id - update without file
func (h *VacancyHandler) Update(c *gin.Context) {
	id := c.Param("id")

	var req struct {
		Title       string  `json:"title" binding:"required"`
		Description *string `json:"description"`
		WeightSoft  int     `json:"weight_soft" binding:"min=0,max=100"`
		WeightHard  int     `json:"weight_hard" binding:"min=0,max=100"`
		WeightCase  int     `json:"weight_case" binding:"min=0,max=100"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request: " + err.Error()})
		return
	}

	vacancy := &models.Vacancy{
		ID:          id,
		Title:       req.Title,
		Description: req.Description,
		WeightSoft:  req.WeightSoft,
		WeightHard:  req.WeightHard,
		WeightCase:  req.WeightCase,
	}

	err := h.svc.UpdateVacancy(c.Request.Context(), vacancy)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "vacancy updated successfully"})
}

// PUT /api/vacancies/:id/file - update with new file
func (h *VacancyHandler) UpdateWithFile(c *gin.Context) {
	id := c.Param("id")

	title := c.PostForm("title")
	description := c.PostForm("description")

	weightSoft, _ := strconv.Atoi(c.DefaultPostForm("weight_soft", "33"))
	weightHard, _ := strconv.Atoi(c.DefaultPostForm("weight_hard", "33"))
	weightCase, _ := strconv.Atoi(c.DefaultPostForm("weight_case", "34"))

	if title == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "title is required"})
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

	vacancy := &models.Vacancy{
		ID:          id,
		Title:       title,
		Description: &description,
		WeightSoft:  weightSoft,
		WeightHard:  weightHard,
		WeightCase:  weightCase,
	}

	err = h.svc.UpdateVacancyWithFile(c.Request.Context(), vacancy, file, fileHeader.Filename)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "vacancy and file updated successfully"})
}

// DELETE /api/vacancies/:id
func (h *VacancyHandler) Delete(c *gin.Context) {
	id := c.Param("id")

	err := h.svc.DeleteVacancy(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "vacancy deleted successfully"})
}
