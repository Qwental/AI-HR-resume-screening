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
	WeightSoft  int            `json:"weight_soft"` // Вес soft skills (0-100)
	WeightHard  int            `json:"weight_hard"` // Вес hard skills (0-100)
	WeightCase  int            `json:"weight_case"` // Вес кейсов/опыта (0-100)
}
