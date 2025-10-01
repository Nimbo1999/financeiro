package services

import "time"

type TemplateService interface {
	RenderWelcomeEmail(name string) (string, error)
	RenderOTPEmail(code string, expiresAt time.Time) (string, error)
}

type templateService struct {
	templateDir string
}

func NewTemplateService(templateDir string) TemplateService {
	return &templateService{
		templateDir: templateDir,
	}
}

func (s *templateService) RenderWelcomeEmail(name string) (string, error) {
	// Will be implemented in later steps
	return "", nil
}

func (s *templateService) RenderOTPEmail(code string, expiresAt time.Time) (string, error) {
	// Will be implemented in later steps
	return "", nil
}
