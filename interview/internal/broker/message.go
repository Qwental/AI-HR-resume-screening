package broker

import (
	"gorm.io/datatypes"
)

// ResumeMessage описывает сообщение для обработки резюме
type ResumeMessage struct {
	ID          string         `json:"id"`
	VacancyID   string         `json:"vacancy_id"`
	TextResume  datatypes.JSON `json:"text_resume_jsonb"`
	TextVacancy datatypes.JSON `json:"text_vacancy_jsonb"`
}
