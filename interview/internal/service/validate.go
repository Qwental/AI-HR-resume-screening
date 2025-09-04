package service

import (
	"fmt"
	"path/filepath"
	"strings"
)

func validateFileType(filename string) error {
	allowedExts := []string{".pdf", ".doc", ".docx", ".txt"}
	ext := strings.ToLower(filepath.Ext(filename))

	for _, allowed := range allowedExts {
		if ext == allowed {
			return nil
		}
	}
	return fmt.Errorf("unsupported file type: %s", ext)
}
