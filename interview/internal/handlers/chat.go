package handlers

import (
	"github.com/gin-gonic/gin"
	"interview/internal/models"
	"interview/internal/service"
	"net/http"
)

type ChatHandler struct {
	chatSvc      service.ChatService
	interviewSvc service.InterviewService
}

func NewChatHandler(chatSvc service.ChatService, interviewSvc service.InterviewService) *ChatHandler {
	return &ChatHandler{
		chatSvc:      chatSvc,
		interviewSvc: interviewSvc,
	}
}

// POST /interview/:token/message
func (h *ChatHandler) SendMessage(c *gin.Context) {
	token := c.Param("token")

	// Получаем интервью по токену
	interview, err := h.interviewSvc.GetInterviewByToken(c.Request.Context(), token)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Interview not found"})
		return
	}

	// Проверяем статус интервью
	if interview.Status != "started" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Interview not active"})
		return
	}

	var req struct {
		Content string `json:"content" binding:"required,min=1,max=2000"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request", "details": err.Error()})
		return
	}

	// Создаем или получаем сессию чата
	session, err := h.getOrCreateSession(interview.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to initialize chat session"})
		return
	}

	if !session.IsActive {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Interview completed"})
		return
	}

	candidateMsg := &models.ChatMessage{
		Type:    models.MessageTypeText,
		Content: req.Content,
		Sender:  "candidate",
	}

	// Добавляем сообщение кандидата и получаем ответ AI
	aiMsg, err := h.chatSvc.AddCandidateMessage(interview.ID, candidateMsg)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"candidate_message": candidateMsg,
		"ai_response":       aiMsg,
	})
}

// GET /interview/:token/messages
func (h *ChatHandler) GetMessages(c *gin.Context) {
	token := c.Param("token")

	interview, err := h.interviewSvc.GetInterviewByToken(c.Request.Context(), token)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Interview not found"})
		return
	}

	msgs, err := h.chatSvc.GetMessages(interview.ID)
	if err != nil {
		// Если сессии нет, возвращаем пустой массив
		c.JSON(http.StatusOK, gin.H{"messages": []models.ChatMessage{}})
		return
	}

	c.JSON(http.StatusOK, gin.H{"messages": msgs})
}

// GET /interview/:token/status
func (h *ChatHandler) GetStatus(c *gin.Context) {
	token := c.Param("token")

	interview, err := h.interviewSvc.GetInterviewByToken(c.Request.Context(), token)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Interview not found"})
		return
	}

	status, err := h.chatSvc.GetStatus(interview.ID)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"interview_id":  interview.ID,
			"status":        interview.Status,
			"is_active":     interview.Status == "started",
			"message_count": 0,
		})
		return
	}

	c.JSON(http.StatusOK, status)
}

func (h *ChatHandler) getOrCreateSession(interviewID string) (*models.ChatSession, error) {
	session, err := h.chatSvc.GetSession(interviewID)
	if err != nil {
		return h.chatSvc.CreateSession(interviewID)
	}
	return session, nil
}
