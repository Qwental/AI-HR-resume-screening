package service

import (
	"encoding/json"
	"fmt"
	"github.com/guylaor/goword"
	"log"
	"strings"
)

type Job struct {
	Статус           string `json:"status"`
	Название         string `json:"title"`
	Регион           string `json:"region"`
	Город            string `json:"city"`
	Адрес            string `json:"address"`
	ТипТрудового     string `json:"employment_type"`
	ТипЗанятости     string `json:"work_type"`
	График           string `json:"schedule"`
	Доход            string `json:"income"`
	ОкладМакс        string `json:"salary_max"`
	ОкладМин         string `json:"salary_min"`
	ГодоваяПремия    string `json:"annual_bonus"`
	ТипПремирования  string `json:"bonus_type"`
	Обязанности      string `json:"responsibilities"`
	Требования       string `json:"requirements"`
	Образование      string `json:"education"`
	Опыт             string `json:"experience"`
	ЗнаниеПрограмм   string `json:"software_skills"`
	НавыкиКомпьютера string `json:"computer_skills"`
	ИностранныеЯзыки string `json:"languages"`
	УровеньЯзыка     string `json:"language_level"`
	Командировки     string `json:"business_trips"`
	ДопИнформация    string `json:"additional_info"`
}

func extraction_vacancy() {
	text, err := goword.ParseText("1.docx")
	if err != nil {
		log.Panic(err)
	}

	lines := strings.Split(text, "\n")
	for i := range lines {
		lines[i] = strings.TrimSpace(lines[i])
	}

	job := Job{}
	keyMap := map[string]*string{
		"Статус":                       &job.Статус,
		"Название":                     &job.Название,
		"Регион":                       &job.Регион,
		"Город":                        &job.Город,
		"Адрес":                        &job.Адрес,
		"Тип трудового":                &job.ТипТрудового,
		"Тип занятости":                &job.ТипЗанятости,
		"Текст график работы":          &job.График,
		"Доход (руб/мес)":              &job.Доход,
		"Оклад макс. (руб/мес)":        &job.ОкладМакс,
		"Оклад мин. (руб/мес)":         &job.ОкладМин,
		"Годовая премия (%)":           &job.ГодоваяПремия,
		"Тип премирования. Описание":   &job.ТипПремирования,
		"Обязанности (для публикации)": &job.Обязанности,
		"Требования (для публикации)":  &job.Требования,
		"Уровень образования":          &job.Образование,
		"Требуемый опыт работы":        &job.Опыт,
		"Знание специальных программ":  &job.ЗнаниеПрограмм,
		"Навыки работы на компьютере":  &job.НавыкиКомпьютера,
		"Знание иностранных языков":    &job.ИностранныеЯзыки,
		"Уровень владения языка":       &job.УровеньЯзыка,
		"Наличие командировок":         &job.Командировки,
		"Дополнительная информация":    &job.ДопИнформация,
	}

	var currentKey *string
	var buffer []string

	for _, line := range lines {
		if ptr, ok := keyMap[line]; ok {
			// сохранили предыдущий ключ
			if currentKey != nil {
				*currentKey = strings.Join(buffer, " ")
			}
			currentKey = ptr
			buffer = nil
		} else if currentKey != nil {
			buffer = append(buffer, line)
		}
	}

	// сохранить последнее поле
	if currentKey != nil {
		*currentKey = strings.Join(buffer, " ")
	}

	jsonData, err := json.MarshalIndent(job, "", "  ")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(string(jsonData))
}
