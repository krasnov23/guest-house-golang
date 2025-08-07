package models

import "github.com/krasnov23/guest-house-golang/internal/forms"

// Данные которые мы будет отправлять из handlers в наши шаблоны
type TemplateData struct {
	StringMap       map[string]string
	IntMap          map[string]int
	FloatMap        map[string]float32
	Data            map[string]interface{}
	CSRFToken       string
	Flash           string
	Warning         string
	Error           string
	Form            *forms.Form
	IsAuthenticated int
}
