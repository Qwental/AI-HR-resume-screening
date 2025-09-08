package models

type MessageType string

const (
	MessageTypeStart    MessageType = "start"
	MessageTypeAnswer   MessageType = "answer"
	MessageTypeQuestion MessageType = "question"
	MessageTypeResult   MessageType = "result"
	MessageTypeError    MessageType = "error"
)

type ChatMessage struct {
	ID          string      `json:"id"`
	InterviewID string      `json:"interview_id"`
	Type        MessageType `json:"type"`
	Content     string      `json:"content"`
	Sender      string      `json:"sender"` // "candidate" или "ai"
}

type ChatSession struct {
	InterviewID  string        `json:"interview_id"`
	Messages     []ChatMessage `json:"messages"`
	IsActive     bool          `json:"is_active"`
	MessageCount int           `json:"message_count"`
	VacancyID    string        `json:"vacancy_id"`
	ResumeID     string        `json:"resume_id"`
}

type InterviewStatus struct {
	InterviewID  string `json:"interview_id"`
	Status       string `json:"status"`
	IsActive     bool   `json:"is_active"`
	MessageCount int    `json:"message_count"`
}
