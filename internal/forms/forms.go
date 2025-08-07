package forms

import (
	"fmt"
	"github.com/asaskevich/govalidator"
	"net/url"
	"strings"
)

// Пользовательская структура формы, встраивает объектurl.Values
type Form struct {
	url.Values // Метод GET тянется из данной структуры по причине встраивания,если бы было errors а не Error errors то errors был бы тоже встроен в Form
	Errors     errors
}

// Инициаилизация новый структуры Form
func New(data url.Values) *Form {
	return &Form{
		data,
		errors(map[string][]string{}),
	}
}

func (f *Form) Required(fields ...string) {
	for _, field := range fields {
		value := f.Get(field) // В данном случае метод Get тянется из url.Values
		if strings.TrimSpace(value) == "" {
			f.Errors.Add(field, "This field cannot be blank.")
		}
	}
}

// Прооверка действительно ли есть пришедшее с фронта поле и оно не пустое
func (f *Form) Has(field string) bool {

	// Получение поля из пришедшего запроса
	x := f.Get(field)

	if x == "" {
		f.Errors.Add(field, "This field cannot be blank")
		return false
	}

	return true
}

// MinLength checks for string minimum length
func (f *Form) MinLength(field string, length int) {
	x := f.Get(field)
	if len(x) < length {
		f.Errors.Add(field, fmt.Sprintf("This field must be at least %d characters.", length))
	}
}

func (f *Form) IsEmail(field string) {
	if !govalidator.IsEmail(f.Get(field)) {
		f.Errors.Add(field, "This field must be a valid email address.")
	}
}

func (f *Form) Valid() bool {
	return len(f.Errors) == 0
}
