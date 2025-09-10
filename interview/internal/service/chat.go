package service

import (
	"context"
	"encoding/json"
	"fmt"
	"interview/internal/models"
	"sync"
	"time"
)

type ChatService interface {
	CreateSession(interviewID string, resumeID, vacancyID string) (*models.ChatSession, error)
	GetSession(interviewID string) (*models.ChatSession, error)
	AddCandidateMessage(interviewID string, msg *models.ChatMessage) (*models.ChatMessage, error) // Новый метод
	AddAIMessage(interviewID string, msg *models.ChatMessage) error                               // Новый метод
	GetMessages(interviewID string) ([]models.ChatMessage, error)
	GetStatus(interviewID string) (*models.InterviewStatus, error)
	CloseSession(interviewID string) error
}

type ChatServiceImpl struct {
	mu         sync.RWMutex
	sessions   map[string]*models.ChatSession
	aiSvc      AIService
	resumeSvc  ResumeService
	vacancySvc VacancyService
}

func NewChatService(aiSvc AIService, resumeSvc ResumeService, vacancySvc VacancyService) ChatService {
	return &ChatServiceImpl{
		sessions:   make(map[string]*models.ChatSession),
		aiSvc:      aiSvc,
		resumeSvc:  resumeSvc,
		vacancySvc: vacancySvc,
	}
}

func (s *ChatServiceImpl) AddCandidateMessage(interviewID string, msg *models.ChatMessage) (*models.ChatMessage, error) {
	err := s.addMessage(interviewID, msg)
	if err != nil {
		return nil, err
	}

	history, err := s.GetMessages(interviewID)
	if err != nil {
		return nil, fmt.Errorf("failed to get chat history: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()

	aiResponse, err := s.aiSvc.GenerateResponse(ctx, msg.Content, history, interviewID)
	if err != nil {
		errorMsg := &models.ChatMessage{
			Type:    models.MessageTypeError,
			Content: "Извините, произошла техническая ошибка. Пожалуйста, повторите ваш вопрос.",
			Sender:  "ai",
		}
		s.addMessage(interviewID, errorMsg)
		return errorMsg, fmt.Errorf("AI service error: %w", err)
	}

	if aiResponse.MessageType == "result" {
		aiMsg := &models.ChatMessage{
			Type:    models.MessageTypeResult,
			Content: aiResponse.Response,
			Sender:  "ai",
		}

		err = s.addMessage(interviewID, aiMsg)
		if err != nil {
			return nil, fmt.Errorf("failed to save result message: %w", err)
		}

		err = s.saveResultToResume(interviewID, aiResponse.Result)
		if err != nil {
			fmt.Printf("Failed to save result to resume: %v\n", err) // Логируем, но не останавливаем процесс
		}

		session, err := s.GetSession(interviewID)
		if err != nil {
			return nil, err
		}

		err = s.resumeSvc.UpdateStatus(ctx, session.ResumeID, "Прошел собеседование")
		if err != nil {
			return nil, err
		}

		// Закрываем сессию
		s.CloseSession(interviewID)

		return aiMsg, nil
	}

	aiMsg := &models.ChatMessage{
		Type:    models.MessageTypeQuestion,
		Content: aiResponse.Response,
		Sender:  "ai",
	}

	err = s.addMessage(interviewID, aiMsg)
	if err != nil {
		return nil, fmt.Errorf("failed to save AI message: %w", err)
	}

	return aiMsg, nil
}

func (s *ChatServiceImpl) saveResultToResume(interviewID, resultString string) error {
	session, err := s.GetSession(interviewID)
	if err != nil {
		return fmt.Errorf("failed to get session: %w", err)
	}

	if session.ResumeID == "" {
		return fmt.Errorf("no resume ID in session")
	}

	ctx := context.Background()

	// Конвертируем в map[string]interface{}
	var resultMap map[string]interface{}
	if resultString != "" {
		err = json.Unmarshal([]byte(resultString), &resultMap)
		if err != nil {
			return fmt.Errorf("invalid result JSON string: %w", err)
		}
	}

	// Передаем map
	err = s.resumeSvc.UpdateResult(ctx, session.ResumeID, resultMap)
	if err != nil {
		return fmt.Errorf("failed to update resume result: %w", err)
	}

	return nil
}

func (s *ChatServiceImpl) AddAIMessage(interviewID string, msg *models.ChatMessage) error {
	msg.Sender = "ai"
	return s.addMessage(interviewID, msg)
}

func (s *ChatServiceImpl) addMessage(interviewID string, msg *models.ChatMessage) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	session, exists := s.sessions[interviewID]
	if !exists {
		return fmt.Errorf("session not found")
	}
	if !session.IsActive {
		return fmt.Errorf("session is closed")
	}

	msg.ID = generateID()
	msg.InterviewID = interviewID

	session.Messages = append(session.Messages, *msg)
	session.MessageCount++

	return nil
}

func (s *ChatServiceImpl) CreateSession(interviewID string, resumeID, vacancyID string) (*models.ChatSession, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.sessions[interviewID]; exists {
		return nil, fmt.Errorf("session already exists")
	}

	ctx := context.Background()
	vacancy, err := s.vacancySvc.GetVacancy(ctx, vacancyID)
	if err != nil {
		return nil, fmt.Errorf("failed to get vacancy: %w", err)
	}

	var resumeText string
	if resumeID != "" {
		resume, err := s.resumeSvc.GetResume(ctx, resumeID)
		if err != nil {
			return nil, fmt.Errorf("failed to get resume: %w", err)
		}
		resumeText = resume.Text
	}

	session := &models.ChatSession{
		InterviewID:  interviewID,
		IsActive:     true,
		Messages:     []models.ChatMessage{},
		MessageCount: 0,
		VacancyID:    vacancyID,
		ResumeID:     resumeID,
	}

	vacancyJSON, err := json.Marshal(vacancy)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal vacancy: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	aiResponse, err := s.aiSvc.GenerateWelcomeResponse(ctx, string(vacancyJSON), resumeText, interviewID)
	if err != nil {
		welcomeMsg := models.ChatMessage{
			Type:    models.MessageTypeQuestion,
			Content: "Добро пожаловать на собеседование! Расскажите немного о себе.",
			Sender:  "ai",
		}
		welcomeMsg.ID = generateID()
		welcomeMsg.InterviewID = interviewID
		session.Messages = append(session.Messages, welcomeMsg)
		session.MessageCount = 1
	} else {
		welcomeMsg := models.ChatMessage{
			Type:    models.MessageTypeQuestion,
			Content: aiResponse.Response,
			Sender:  "ai",
		}
		welcomeMsg.ID = generateID()
		welcomeMsg.InterviewID = interviewID
		session.Messages = append(session.Messages, welcomeMsg)
		session.MessageCount = 1
	}

	s.sessions[interviewID] = session

	err = s.resumeSvc.UpdateStatus(ctx, resumeID, "Проходит собеседование")
	if err != nil {
		return nil, err
	}
	return session, nil
}

func (s *ChatServiceImpl) GetSession(interviewID string) (*models.ChatSession, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	session, exists := s.sessions[interviewID]
	if !exists {
		return nil, fmt.Errorf("session not found")
	}
	return session, nil
}

func (s *ChatServiceImpl) GetMessages(interviewID string) ([]models.ChatMessage, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	session, exists := s.sessions[interviewID]
	if !exists {
		return nil, fmt.Errorf("session not found")
	}
	return session.Messages, nil
}

func (s *ChatServiceImpl) GetStatus(interviewID string) (*models.InterviewStatus, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	session, exists := s.sessions[interviewID]
	if !exists {
		return nil, fmt.Errorf("session not found")
	}
	return &models.InterviewStatus{
		InterviewID:  session.InterviewID,
		Status:       ifStatus(session.IsActive),
		IsActive:     session.IsActive,
		MessageCount: session.MessageCount,
	}, nil
}

func (s *ChatServiceImpl) CloseSession(interviewID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	session, exists := s.sessions[interviewID]
	if !exists {
		return fmt.Errorf("session not found")
	}
	session.IsActive = false
	return nil
}

func generateID() string {
	return fmt.Sprintf("msg_%d", time.Now().UnixNano())
}

func ifStatus(active bool) string {
	if active {
		return "started"
	}
	return "completed"
}
