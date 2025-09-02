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
	// Получаем данные из multipart form
	title := c.PostForm("title")
	description := c.PostForm("description")
	usersID := c.PostForm("users_id")

	// Веса с дефолтными значениями
	weightSoft, _ := strconv.Atoi(c.DefaultPostForm("weight_soft", "33"))
	weightHard, _ := strconv.Atoi(c.DefaultPostForm("weight_hard", "33"))
	weightCase, _ := strconv.Atoi(c.DefaultPostForm("weight_case", "34"))

	// Валидация обязательных полей
	if title == "" || usersID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "title и users_id обязательны"})
		return
	}

	// Получаем файл
	fileHeader, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "файл обязателен"})
		return
	}

	file, err := fileHeader.Open()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "не удалось открыть файл"})
		return
	}
	defer file.Close()

	// Создаем модель вакансии
	vacancy := &models.Vacancy{
		Title:       title,
		Description: &description,
		UsersID:     usersID,
		WeightSoft:  weightSoft,
		WeightHard:  weightHard,
		WeightCase:  weightCase,
	}

	// Создаем вакансию через cервис
	err = h.svc.CreateVacancy(c.Request.Context(), vacancy, file, fileHeader.Filename)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":          vacancy.ID,
		"title":       vacancy.Title,
		"description": vacancy.Description,
		"users_id":    vacancy.UsersID,
		"storage_key": vacancy.StorageKey,
		"weight_soft": vacancy.WeightSoft,
		"weight_hard": vacancy.WeightHard,
		"weight_case": vacancy.WeightCase,
		"created_at":  vacancy.CreatedAt,
	})
}

// GET /api/vacancies/:id
func (h *VacancyHandler) GetByID(c *gin.Context) {
	id := c.Param("id")

	vacancy, err := h.svc.GetVacancy(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "вакансия не найдена"})
		return
	}

	c.JSON(http.StatusOK, vacancy)
}

// GET /api/vacancies/:id/file - получение вакансии с presigned URL для скачивания файла
func (h *VacancyHandler) GetWithFileURL(c *gin.Context) {
	id := c.Param("id")

	vacancy, err := h.svc.GetVacancyWithFileURL(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "вакансия не найдена"})
		return
	}

	c.JSON(http.StatusOK, vacancy)
}

// PUT /api/vacancies/:id - обновление без файла
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "неверный запрос: " + err.Error()})
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

	c.JSON(http.StatusOK, gin.H{"message": "вакансия обновлена"})
}

// DELETE /api/vacancies/:id
func (h *VacancyHandler) Delete(c *gin.Context) {
	id := c.Param("id")

	err := h.svc.DeleteVacancy(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "вакансия удалена"})
}
