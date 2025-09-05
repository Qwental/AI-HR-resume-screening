package storage

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
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

// ✅ НОВАЯ ВЕРСИЯ - Исправленная инициализация S3 клиента
func NewS3Client(endpoint, region, accessKey, secretKey string) *s3.Client {
	// Кастомный resolver для MinIO endpoint
	customResolver := aws.EndpointResolverWithOptionsFunc(
		func(service, region string, options ...interface{}) (aws.Endpoint, error) {
			if service == s3.ServiceID {
				return aws.Endpoint{
					URL:           endpoint,
					SigningRegion: region,
				}, nil
			}
			return aws.Endpoint{}, fmt.Errorf("unknown endpoint requested")
		})

	// Загружаем конфигурацию AWS
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithEndpointResolverWithOptions(customResolver),
		config.WithRegion(region),
		config.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(accessKey, secretKey, ""),
		),
	)
	if err != nil {
		panic(fmt.Sprintf("failed to load S3 config: %v", err))
	}

	// Создаем S3 клиент с PathStyle для MinIO
	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.UsePathStyle = true
	})

	return client
}

// ✅ НОВАЯ ВЕРСИЯ - с автоматическим созданием bucket
func NewS3Storage(client *s3.Client, bucketName string, region string) *S3Storage {
	storage := &S3Storage{
		client:     client,
		bucketName: bucketName,
		region:     region,
	}

	// 🚀 Автоматически создаем bucket при инициализации
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := storage.CreateBucketIfNotExists(ctx); err != nil {
		log.Printf("⚠️ Warning: Could not create bucket %s: %v", bucketName, err)
		log.Printf("📁 Bucket will be created on first upload")
	}

	return storage
}

// ✅ НОВЫЙ МЕТОД - Автоматическое создание bucket
func (s *S3Storage) CreateBucketIfNotExists(ctx context.Context) error {
	// Проверяем, существует ли bucket
	_, err := s.client.HeadBucket(ctx, &s3.HeadBucketInput{
		Bucket: aws.String(s.bucketName),
	})
	if err == nil {
		log.Printf("✅ Bucket '%s' already exists", s.bucketName)
		return nil // bucket уже существует
	}

	// Создаем bucket
	log.Printf("🛠️ Creating bucket '%s'...", s.bucketName)
	_, err = s.client.CreateBucket(ctx, &s3.CreateBucketInput{
		Bucket: aws.String(s.bucketName),
	})

	if err != nil {
		return fmt.Errorf("failed to create bucket: %w", err)
	}

	log.Printf("✅ Bucket '%s' created successfully", s.bucketName)
	return nil
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

// Определение Content-Type по содержимому файла
func (s *S3Storage) detectContentType(filename string, file io.Reader) (string, io.Reader, error) {
	// Сначала пробуем определить по расширению
	contentType := s.getContentTypeByExtension(filename)

	// Если получили универсальный тип, пробуем определить по содержимому
	if contentType == "application/octet-stream" {
		buffer := make([]byte, 512)
		n, err := file.Read(buffer)
		if err != nil && err != io.EOF {
			return contentType, file, fmt.Errorf("failed to read file for content type detection: %w", err)
		}

		// Определяем тип по содержимому
		detectedType := http.DetectContentType(buffer[:n])
		if detectedType != "application/octet-stream" {
			contentType = detectedType
		}

		// Создаем новый reader, объединяя прочитанные байты с остальным файлом
		newReader := io.MultiReader(bytes.NewReader(buffer[:n]), file)
		return contentType, newReader, nil
	}

	return contentType, file, nil
}

// Универсальный метод загрузки файлов
func (s *S3Storage) UploadFile(ctx context.Context, file io.Reader, filename, folder string) (string, error) {
	if filename == "" {
		return "", fmt.Errorf("filename cannot be empty")
	}

	fileID := uuid.New().String()
	key := fmt.Sprintf("%s/%s", folder, fileID)

	// Определяем Content-Type
	contentType, processedFile, err := s.detectContentType(filename, file)
	if err != nil {
		return "", fmt.Errorf("failed to detect content type: %w", err)
	}

	_, err = s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.bucketName),
		Key:         aws.String(key),
		Body:        processedFile,
		ContentType: aws.String(contentType),
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload file to S3: %w", err)
	}

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

// ✅ ИСПРАВЛЕННЫЙ МЕТОД - для совместимости
func (s *S3Storage) UploadVacancy(ctx context.Context, file io.Reader, filename string) (string, error) {
	return s.UploadVacancyFile(ctx, file, filename)
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

// ✅ ДОПОЛНИТЕЛЬНЫЙ МЕТОД - для совместимости с существующим кодом
func (s *S3Storage) GetFile(ctx context.Context, key string) (io.ReadCloser, error) {
	return s.DownloadFile(ctx, key)
}

/*
// ❌ СТАРАЯ ВЕРСИЯ - ЗАКОММЕНТИРОВАННАЯ
func NewS3Client(endpoint, region, accessKey, secretKey string) *s3.Client {
	customResolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
		return aws.Endpoint{
			URL:           endpoint,
			SigningRegion: region,
			Source:        aws.EndpointSourceCustom,
		}, nil
	})

	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(region),
		config.WithEndpointResolverWithOptions(customResolver),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKey, secretKey, "")),
	)

	if err != nil {
		panic(fmt.Errorf("failed to load AWS config: %w", err))
	}

	return s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.UsePathStyle = true
	})
}

func NewS3Storage(client *s3.Client, bucketName string, region string) *S3Storage {
	return &S3Storage{client: client, bucketName: bucketName, region: region}
}
*/
