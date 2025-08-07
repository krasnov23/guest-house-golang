package forms

type errors map[string][]string

// Добавление ошибки для определенного поля
func (e errors) Add(field, message string) {
	e[field] = append(e[field], message)
}

// Получение сообщения об ошибке
func (e errors) Get(field string) string {
	es := e[field]
	if len(es) == 0 {
		return ""
	}

	return es[0]
}
