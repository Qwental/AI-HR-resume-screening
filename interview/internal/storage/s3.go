package storage

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/google/uuid"
)

type S3Storage struct {
	client     *s3.Client
	bucketName string
	region     string
}

func NewS3Client(endpoint, region, accessKey, secretKey string) *s3.Client {
	// Проверяем, что endpoint не пустой
	if endpoint == "" {
		panic("endpoint cannot be empty")
	}

	// Проверяем формат URL
	if !strings.HasPrefix(endpoint, "http://") && !strings.HasPrefix(endpoint, "https://") {
		panic("endpoint must start with http:// or https://")
	}

	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(region),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKey, secretKey, "")),
	)
	if err != nil {
		panic(fmt.Errorf("failed to load AWS config: %w", err))
	}

	return s3.NewFromConfig(cfg, func(o *s3.Options) {
		// Используем BaseEndpoint вместо устаревшего EndpointResolver
		o.BaseEndpoint = aws.String(endpoint)
		o.UsePathStyle = true
	})
}

func NewS3Storage(client *s3.Client, bucketName string, region string) *S3Storage {
	return &S3Storage{client: client, bucketName: bucketName, region: region}
}

// Определение Content-Type по расширению файла с fallback значениями
func (s *S3Storage) getContentTypeByExtension(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))

	contentTypeMap := map[string]string{
		".pdf":  "application/pdf",
		".doc":  "application/msword",
		".docx": "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
		".txt":  "text/plain",
		".rtf":  "application/rtf",
		".jpg":  "image/jpeg",
		".jpeg": "image/jpeg",
		".png":  "image/png",
		".gif":  "image/gif",
		".zip":  "application/zip",
		".rar":  "application/x-rar-compressed",
	}

	if contentType, exists := contentTypeMap[ext]; exists {
		return contentType
	}

	// Пробуем стандартную библиотеку mime
	if contentType := mime.TypeByExtension(ext); contentType != "" {
		return contentType
	}

	return "application/octet-stream"
}

func (s *S3Storage) DownloadFile(ctx context.Context, key string) (io.ReadCloser, error) {
	if key == "" {
		return nil, fmt.Errorf("storage key cannot be empty")
	}

	result, err := s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to download file from S3: %w", err)
	}

	return result.Body, nil
}

// Определение Content-Type по содержимому файла - ИСПРАВЛЕННАЯ ВЕРСИЯ
func (s *S3Storage) detectContentType(filename string, file io.Reader) (string, io.Reader, error) {
	// Сначала пробуем определить по расширению
	contentType := s.getContentTypeByExtension(filename)

	// Читаем весь файл в память для создания seekable reader
	data, err := io.ReadAll(file)
	if err != nil {
		return contentType, file, fmt.Errorf("failed to read file: %w", err)
	}

	// Если получили универсальный тип, пробуем определить по содержимому
	if contentType == "application/octet-stream" {
		// Определяем тип по содержимому (используем первые 512 байт)
		detectedType := http.DetectContentType(data)
		if detectedType != "application/octet-stream" {
			contentType = detectedType
		}
	}

	// Возвращаем seekable reader
	return contentType, bytes.NewReader(data), nil
}

// Универсальный метод загрузки файлов
func (s *S3Storage) UploadFile(ctx context.Context, file io.Reader, filename, folder string) (string, error) {
	if filename == "" {
		return "", fmt.Errorf("filename cannot be empty")
	}

	fileID := uuid.New().String()
	key := fmt.Sprintf("%s/%s", folder, fileID)

	// Определяем Content-Type и получаем seekable reader
	contentType, processedFile, err := s.detectContentType(filename, file)
	if err != nil {
		return "", fmt.Errorf("failed to detect content type: %w", err)
	}

	// Логируем информацию о запросе
	fmt.Printf("=== S3 UPLOAD DEBUG ===\n")
	fmt.Printf("Bucket: %s\n", s.bucketName)
	fmt.Printf("Key: %s\n", key)
	fmt.Printf("Region: %s\n", s.region)
	fmt.Printf("Expected URL: %s/%s/%s\n", "https://s3.cloud.ru", s.bucketName, key)
	fmt.Printf("=======================\n")

	_, err = s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.bucketName),
		Key:         aws.String(key),
		Body:        processedFile,
		ContentType: aws.String(contentType),
	})

	if err != nil {
		fmt.Printf("S3 Upload Error: %v\n", err)
		return "", fmt.Errorf("failed to upload file to S3: %w", err)
	}

	fmt.Printf("✅ Upload successful!\n")
	return key, nil
}

// Загрузка резюме
func (s *S3Storage) UploadResume(ctx context.Context, file io.Reader, filename string) (string, error) {
	return s.UploadFile(ctx, file, filename, "resumes")
}

// Загрузка файла вакансии
func (s *S3Storage) UploadVacancyFile(ctx context.Context, file io.Reader, filename string) (string, error) {
	return s.UploadFile(ctx, file, filename, "vacancies")
}

// Генерация presigned URL
func (s *S3Storage) GeneratePresignedURL(ctx context.Context, key string, expiration time.Duration) (string, error) {
	if key == "" {
		return "", fmt.Errorf("key cannot be empty")
	}

	presignClient := s3.NewPresignClient(s.client)

	request, err := presignClient.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(key),
	}, func(opts *s3.PresignOptions) {
		opts.Expires = expiration
	})

	if err != nil {
		return "", fmt.Errorf("failed to generate presigned URL: %w", err)
	}

	return request.URL, nil
}

// Удаление файла
func (s *S3Storage) DeleteFile(ctx context.Context, key string) error {
	if key == "" {
		return fmt.Errorf("key cannot be empty")
	}

	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(key),
	})

	if err != nil {
		return fmt.Errorf("failed to delete file from S3: %w", err)
	}

	return nil
}

// Проверка существования файла
func (s *S3Storage) FileExists(ctx context.Context, key string) (bool, error) {
	if key == "" {
		return false, fmt.Errorf("key cannot be empty")
	}

	_, err := s.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(key),
	})

	if err != nil {
		var nsk *types.NoSuchKey
		if errors.As(err, &nsk) {
			return false, nil
		}
		return false, fmt.Errorf("failed to check file existence: %w", err)
	}

	return true, nil
}
