package services

import (
	"bytes"
	"fmt"
	"html/template"
	"path/filepath"
	"time"
)

type TemplateService interface {
	RenderWelcomeEmail(name string) (string, error)
	RenderOTPEmail(code string, expiresAt time.Time) (string, error)
}

type templateService struct {
	templateDir string
}

type WelcomeEmailData struct {
	Name string
}

type OTPEmailData struct {
	Code      string
	ExpiresAt string
}

func NewTemplateService(templateDir string) TemplateService {
	return &templateService{
		templateDir: templateDir,
	}
}

func (s *templateService) RenderWelcomeEmail(name string) (string, error) {
	data := WelcomeEmailData{
		Name: name,
	}
	return s.loadTemplate("welcome.html", data)
}

func (s *templateService) RenderOTPEmail(code string, expiresAt time.Time) (string, error) {
	data := OTPEmailData{
		Code:      code,
		ExpiresAt: expiresAt.Format("02/01/2006 15:04:05"),
	}
	return s.loadTemplate("otp.html", data)
}

func (s *templateService) loadTemplate(filename string, data any) (string, error) {
	templatePath := filepath.Join(s.templateDir, filename)
	tmpl, err := template.ParseFiles(templatePath)
	if err != nil {
		return "", fmt.Errorf("failed to parse %s template: %w", filename, err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute %s template: %w", filename, err)
	}
	return buf.String(), nil
}
