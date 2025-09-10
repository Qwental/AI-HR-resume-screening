package service

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/streadway/amqp"
	"interview/internal/models"
)

type AIService interface {
	GenerateResponse(ctx context.Context, userInput string, conversation []models.ChatMessage, interviewID string) (*AIResponse, error)
	GenerateWelcomeResponse(ctx context.Context, vacancyJSON, resumeText, interviewID string) (*AIResponse, error)
}

type AIResponse struct {
	RequestID   string `json:"request_id,omitempty"`
	Response    string `json:"response"`
	MessageType string `json:"message_type"`
	Error       string `json:"error,omitempty"`
	Result      string `json:"result,omitempty"`
}

type AIRequest struct {
	RequestID    string               `json:"request_id"`
	UserMessage  string               `json:"user_message"`
	Conversation []models.ChatMessage `json:"conversation,omitempty"`
	InterviewID  string               `json:"interview_id"`
	VacancyJSON  string               `json:"vacancy_json,omitempty"`
	ResumeText   string               `json:"resume_text,omitempty"`
	Action       string               `json:"action,omitempty"`
}

type AIServiceImpl struct {
	rabbitmq *RabbitMQService
}

func NewAIService(rabbitmq *RabbitMQService) AIService {
	return &AIServiceImpl{rabbitmq: rabbitmq}
}

func (a *AIServiceImpl) GenerateResponse(
	ctx context.Context,
	userInput string,
	conversation []models.ChatMessage,
	interviewID string,
) (*AIResponse, error) {
	req := &AIRequest{
		UserMessage:  userInput,
		Conversation: conversation,
		InterviewID:  interviewID,
	}

	brokerResp, err := a.rabbitmq.SendAIRequest(ctx, req)
	if err != nil {
		return nil, err
	}

	brokerResp.RequestID = ""
	return brokerResp, nil
}

func (a *AIServiceImpl) GenerateWelcomeResponse(
	ctx context.Context,
	vacancyJSON, resumeText, interviewID string,
) (*AIResponse, error) {
	req := &AIRequest{
		UserMessage:  "Привет",
		Conversation: []models.ChatMessage{},
		InterviewID:  interviewID,
		VacancyJSON:  vacancyJSON,
		ResumeText:   resumeText,
	}

	brokerResp, err := a.rabbitmq.SendAIRequest(ctx, req)
	if err != nil {
		return nil, err
	}

	brokerResp.RequestID = ""
	return brokerResp, nil
}

type RabbitMQService struct {
	conn            *amqp.Connection
	channel         *amqp.Channel
	requestQueue    string
	responseQueue   string
	pendingRequests map[string]chan *AIResponse
	mu              sync.RWMutex
}

func NewRabbitMQService(amqpURL string) (*RabbitMQService, error) {
	conn, err := amqp.Dial(amqpURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}
	ch, err := conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("failed to open channel: %w", err)
	}

	requestQueue := "ai_requests"
	responseQueue := "ai_responses"

	_, err = ch.QueueDeclare(requestQueue, true, false, false, false, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to declare request queue: %w", err)
	}
	_, err = ch.QueueDeclare(responseQueue, true, false, false, false, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to declare response queue: %w", err)
	}

	service := &RabbitMQService{
		conn:            conn,
		channel:         ch,
		requestQueue:    requestQueue,
		responseQueue:   responseQueue,
		pendingRequests: make(map[string]chan *AIResponse),
	}

	go service.startResponseListener()
	return service, nil
}

func (r *RabbitMQService) SendAIRequest(ctx context.Context, request *AIRequest) (*AIResponse, error) {
	request.RequestID = generateRequestID()

	r.mu.Lock()
	respChan := make(chan *AIResponse, 1)
	r.pendingRequests[request.RequestID] = respChan
	r.mu.Unlock()

	body, err := json.Marshal(request)
	if err != nil {
		r.mu.Lock()
		delete(r.pendingRequests, request.RequestID)
		r.mu.Unlock()
		return nil, err
	}

	err = r.channel.Publish(
		"", r.requestQueue, false, false,
		amqp.Publishing{
			ContentType:   "application/json",
			Body:          body,
			ReplyTo:       r.responseQueue,
			CorrelationId: request.RequestID,
		},
	)
	if err != nil {
		r.mu.Lock()
		delete(r.pendingRequests, request.RequestID)
		r.mu.Unlock()
		return nil, err
	}

	select {
	case resp := <-respChan:
		r.mu.Lock()
		delete(r.pendingRequests, request.RequestID)
		r.mu.Unlock()
		return resp, nil
	case <-ctx.Done():
		r.mu.Lock()
		delete(r.pendingRequests, request.RequestID)
		r.mu.Unlock()
		return nil, fmt.Errorf("timeout or canceled")
	}
}

func (r *RabbitMQService) startResponseListener() {
	msgs, err := r.channel.Consume(
		r.responseQueue,
		"",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		fmt.Printf("Failed to register response consumer: %v\n", err)
		return
	}

	for msg := range msgs {
		var resp AIResponse
		err := json.Unmarshal(msg.Body, &resp)
		if err != nil {
			fmt.Printf("Failed to unmarshal AI response: %v\n", err)
			continue
		}

		r.mu.RLock()
		ch, exists := r.pendingRequests[resp.RequestID]
		r.mu.RUnlock()

		if exists {
			ch <- &resp
		}
	}
}

func (r *RabbitMQService) Close() error {
	if r.channel != nil {
		r.channel.Close()
	}
	if r.conn != nil {
		return r.conn.Close()
	}
	return nil
}

func generateRequestID() string {
	return fmt.Sprintf("req_%d", time.Now().UnixNano())
}
