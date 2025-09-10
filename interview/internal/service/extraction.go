package service

import (
	"archive/zip"
	"bytes"
	"context"
	"fmt"
	"github.com/guylaor/goword"
	"interview/internal/storage"
	"io"
	"log"
	"os"
	"regexp"
	"strings"

	"github.com/nguyenthenguyen/docx" // MIT License
)

type FileType int

const (
	FileTypeDOCX FileType = iota
	FileTypeDOC
	FileTypePDF
	FileTypeTXT
	FileTypeUnknown
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

// DocxDocument представляет структуру DOCX документа
type DocxDocument struct {
	Body DocxBody `xml:"body"`
}

type DocxBody struct {
	Paragraphs []DocxParagraph `xml:"p"`
}

type DocxParagraph struct {
	Runs []DocxRun `xml:"r"`
}

type DocxRun struct {
	Text []DocxText `xml:"t"`
}

type DocxText struct {
	Value string `xml:",chardata"`
}

func ExtractVacancyFromS3Key(ctx context.Context, storage *storage.S3Storage, storageKey string) (*Job, error) {
	reader, err := storage.DownloadFile(ctx, storageKey)
	if err != nil {
		return nil, fmt.Errorf("failed to download file from S3: %w", err)
	}
	defer reader.Close()

	tempFile, err := os.CreateTemp("", "extraction_*.docx")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tempFile.Name())
	defer tempFile.Close()

	_, err = io.Copy(tempFile, reader)
	if err != nil {
		return nil, fmt.Errorf("failed to copy file: %w", err)
	}

	text, err := goword.ParseText(tempFile.Name())
	if err != nil {
		return nil, fmt.Errorf("failed to parse docx file: %w", err)
	}

	lines := strings.Split(text, "\n")
	for i := range lines {
		lines[i] = strings.TrimSpace(lines[i])
	}

	job := &Job{}

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
		if line == "" {
			continue
		}

		if ptr, ok := keyMap[line]; ok {
			if currentKey != nil {
				*currentKey = strings.Join(buffer, " ")
			}
			currentKey = ptr
			buffer = nil
		} else if currentKey != nil {
			buffer = append(buffer, line)
		}
	}

	if currentKey != nil {
		*currentKey = strings.Join(buffer, " ")
	}

	return job, nil
}

//	func extractResumeFromDocx(file io.Reader) (string, error) {
//		tempFile, err := os.CreateTemp("", "extract_*.docx")
//		if err != nil {
//			return "", fmt.Errorf("failed to create temp file: %w", err)
//		}
//		defer os.Remove(tempFile.Name())
//		defer tempFile.Close()
//
//		_, err = io.Copy(tempFile, file)
//		if err != nil {
//			return "", fmt.Errorf("failed to copy file: %w", err)
//		}
//
//		text, err := goword.ParseText(tempFile.Name())
//		if err != nil {
//			return "", fmt.Errorf("failed to parse docx: %w", err)
//		}
//
//		return strings.TrimSpace(text), nil
//	}
//
// detectFileType определяет тип файла по содержимому
func detectFileType(data []byte) FileType {
	if len(data) < 4 {
		return FileTypeUnknown
	}

	// DOCX файлы начинаются с ZIP signature
	if bytes.Equal(data[:4], []byte{0x50, 0x4B, 0x03, 0x04}) {
		return FileTypeDOCX
	}

	// PDF файлы начинаются с "%PDF"
	if bytes.Equal(data[:4], []byte{0x25, 0x50, 0x44, 0x46}) {
		return FileTypePDF
	}

	// Проверяем на текстовый файл
	if isTextFile(data) {
		return FileTypeTXT
	}

	return FileTypeUnknown
}

// extractFromTXT извлекает текст из обычного текстового файла
func extractFromTXT(fileData []byte) (string, error) {
	return strings.TrimSpace(string(fileData)), nil
}

// isTextFile проверяет, является ли файл текстовым
func isTextFile(data []byte) bool {
	if len(data) == 0 {
		return false
	}

	sample := data
	if len(sample) > 512 {
		sample = sample[:512]
	}

	for _, b := range sample {
		if b == 0 || (b < 32 && b != 9 && b != 10 && b != 13) {
			return false
		}
	}
	return true
}

// getFileTypeName возвращает название типа файла
func getFileTypeName(fileType FileType) string {
	switch fileType {
	case FileTypeDOCX:
		return "DOCX"
	case FileTypePDF:
		return "PDF"
	case FileTypeTXT:
		return "TXT"
	default:
		return "Unknown"
	}
}

// ExtractTextFromFile извлекает текст из файла с использованием только бесплатных библиотек
func ExtractTextFromFile(fileData []byte, filename string) (string, error) {
	if len(fileData) == 0 {
		return "", fmt.Errorf("пустой файл")
	}

	fileType := detectFileType(fileData)

	log.Printf("🔍 Обнаружен тип файла: %s для %s", getFileTypeName(fileType), filename)

	switch fileType {
	case FileTypeDOCX:
		return extractFromDOCXFree(fileData)
	case FileTypePDF:
		//return extractFromPDFFree(fileData)
		return "", fmt.Errorf("неподдерживаемый формат файла: %s", getFileTypeName(fileType))

	case FileTypeTXT:
		return extractFromTXT(fileData)
	default:
		return "", fmt.Errorf("неподдерживаемый формат файла: %s", getFileTypeName(fileType))
	}
}

// extractFromDOCXFree извлекает текст из DOCX используя бесплатные библиотеки
func extractFromDOCXFree(fileData []byte) (string, error) {
	// Способ 1: Используем github.com/nguyenthenguyen/docx
	tempFile, err := os.CreateTemp("", "extract_*.docx")
	if err != nil {
		return "", fmt.Errorf("не удалось создать временный файл: %w", err)
	}
	defer os.Remove(tempFile.Name())
	defer tempFile.Close()

	if _, err := tempFile.Write(fileData); err != nil {
		return "", fmt.Errorf("не удалось записать временный файл: %w", err)
	}
	tempFile.Close()

	// Пытаемся использовать docx библиотеку
	doc, err := docx.ReadDocxFile(tempFile.Name())
	if err == nil {
		defer doc.Close()

		docText := doc.Editable()
		if docText != nil {
			text := docText.GetContent()
			if text != "" {
				return strings.TrimSpace(text), nil
			}
		}
	}

	log.Printf("⚠️ docx библиотека не смогла обработать файл: %v", err)

	// Способ 2: Прямое извлечение из ZIP архива (самописный парсер)
	if text, err := extractDOCXFromZip(fileData); err == nil {
		return text, nil
	} else {
		log.Printf("⚠️ ZIP extraction не удалась: %v", err)
	}

	return "", fmt.Errorf("не удалось извлечь текст из DOCX файла")
}

// extractDOCXFromZip извлекает текст из DOCX как из ZIP архива
func extractDOCXFromZip(fileData []byte) (string, error) {
	reader := bytes.NewReader(fileData)

	zipReader, err := zip.NewReader(reader, int64(len(fileData)))
	if err != nil {
		return "", fmt.Errorf("не является валидным ZIP архивом: %w", err)
	}

	// Ищем document.xml файл
	for _, file := range zipReader.File {
		if file.Name == "word/document.xml" {
			rc, err := file.Open()
			if err != nil {
				continue
			}
			defer rc.Close()

			content, err := io.ReadAll(rc)
			if err != nil {
				continue
			}

			// Парсим XML и извлекаем текст
			return parseDocumentXML(content), nil
		}
	}

	return "", fmt.Errorf("не найден document.xml в архиве")
}

// parseDocumentXML парсит document.xml и извлекает текст
func parseDocumentXML(xmlContent []byte) string {
	// Простое регулярное выражение для извлечения текста между <w:t> тегами
	re := regexp.MustCompile(`<w:t[^>]*>(.*?)</w:t>`)
	matches := re.FindAllStringSubmatch(string(xmlContent), -1)

	var result strings.Builder
	for _, match := range matches {
		if len(match) > 1 {
			// Декодируем XML entities
			text := strings.ReplaceAll(match[1], "&lt;", "<")
			text = strings.ReplaceAll(text, "&gt;", ">")
			text = strings.ReplaceAll(text, "&amp;", "&")
			text = strings.ReplaceAll(text, "&quot;", "\"")
			text = strings.ReplaceAll(text, "&apos;", "'")

			result.WriteString(text)
			result.WriteString(" ")
		}
	}

	return strings.TrimSpace(result.String())
}

// extractFromPDFFree извлекает текст из PDF используя бесплатные библиотеки
//func extractFromPDFFree(fileData []byte) (string, error) {
//	tempFile, err := os.CreateTemp("", "extract_*.pdf")
//	if err != nil {
//		return "", err
//	}
//	defer os.Remove(tempFile.Name())
//	defer tempFile.Close()
//
//	if _, err := tempFile.Write(fileData); err != nil {
//		return "", err
//	}
//	tempFile.Close()
//
//	// Извлекаем текст
//	text, err := api.ExtractTextFile(tempFile.Name(), nil, nil)
//	if err != nil {
//		return "", fmt.Errorf("pdfcpu extraction failed: %w", err)
//	}
//
//	return strings.TrimSpace(string(text)), nil
//}
