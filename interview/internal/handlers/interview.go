package handlers

import (
	"net/http"
	_ "time"

	"github.com/gin-gonic/gin"
	"interview/internal/models"
	"interview/internal/service"
)

type Handler struct {
	svc service.InterviewService
}

func NewHandler(svc service.InterviewService) *Handler { return &Handler{svc: svc} }

/* -------- POST /api/admin/interviews -------- */

func (h *Handler) Create(c *gin.Context) {
	var req struct {
		ResumeID  *string `json:"resume_id,omitempty"`
		VacancyID string  `json:"vacancy_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad request"})
		return
	}
	intv := &models.Interview{ResumeID: req.ResumeID, VacancyID: req.VacancyID}

	scheme := "http"
	if c.Request.TLS != nil {
		scheme = "https"
	}
	base := scheme + "://" + c.Request.Host

	if err := h.svc.CreateInterview(c.Request.Context(), intv, base); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "server error"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{
		"id":            intv.ID,
		"vacancy_id":    intv.VacancyID,
		"resume_id":     intv.ResumeID,
		"status":        intv.Status,
		"interview_url": *intv.URLToken,
		"scheduled_at":  intv.ScheduledAt,
		"expires_at":    intv.ExpiresAt,
		"created_at":    intv.CreatedAt,
	})
}

/* -------- GET /api/interview/:token -------- */

func (h *Handler) GetByToken(c *gin.Context) {
	tok := c.Param("token")
	intv, err := h.svc.GetInterviewByToken(c.Request.Context(), tok)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"status":       intv.Status,
		"scheduled_at": intv.ScheduledAt,
		"expires_at":   intv.ExpiresAt,
		"started_at":   intv.StartedAt,
	})
}

/* -------- POST /api/interview/:token/start -------- */

func (h *Handler) StartByToken(c *gin.Context) {
	if h.svc.StartInterview(c.Request.Context(), c.Param("token")) != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot start"})
		return
	}
	c.Status(http.StatusOK)
}

/* -------- POST /api/interview/:token/finish -------- */

func (h *Handler) FinishByToken(c *gin.Context) {
	if h.svc.FinishInterview(c.Request.Context(), c.Param("token")) != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot finish"})
		return
	}
	c.Status(http.StatusOK)
}

/* -------- Роутер -------- */

func SetupRouter(s service.InterviewService) *gin.Engine {
	r := gin.Default()
	h := NewHandler(s)

	api := r.Group("/api")
	{
		api.POST("/admin/interviews", h.Create)
		api.GET("/interview/:token", h.GetByToken)
		api.POST("/interview/:token/start", h.StartByToken)
		api.POST("/interview/:token/finish", h.FinishByToken)
	}
	return r
}
