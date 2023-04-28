package controller

import (
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
)

// FormSubmission представляет состояние отправки формы, не включая саму форму
type FormSubmission struct {
	// IsSubmitted указывает, была ли отправлена форма
	IsSubmitted bool

	// Errors сохраняет фрагмент строк сообщения об ошибке с ключом, заданным именем поля структуры формы
	Errors map[string][]string
}

// Process обрабатывает отправку формы
func (f *FormSubmission) Process(ctx echo.Context, form interface{}) error {
	f.Errors = make(map[string][]string)
	f.IsSubmitted = true

	// Валидация формы
	if err := ctx.Validate(form); err != nil {
		f.setErrorMessages(err)
	}

	return nil
}

// HasErrors указывает, есть ли в отправке какие-либо ошибки при проверке
func (f FormSubmission) HasErrors() bool {
	if f.Errors == nil {
		return false
	}
	return len(f.Errors) > 0
}

// FieldHasErrors указывает, есть ли в данном поле формы какие-либо ошибки проверки
func (f FormSubmission) FieldHasErrors(fieldName string) bool {
	return len(f.GetFieldErrors(fieldName)) > 0
}

// SetFieldError задает сообщение об ошибке для заданного имени поля
func (f *FormSubmission) SetFieldError(fieldName string, message string) {
	if f.Errors == nil {
		f.Errors = make(map[string][]string)
	}
	f.Errors[fieldName] = append(f.Errors[fieldName], message)
}

// GetFieldErrors возвращает ошибки для заданного имени поля
func (f FormSubmission) GetFieldErrors(fieldName string) []string {
	if f.Errors == nil {
		return []string{}
	}
	return f.Errors[fieldName]
}

// GetFieldErrors возвращает ошибки для заданного имени поля
func (f FormSubmission) GetAllFieldErrors() map[string][]string {
	return f.Errors
}

// IsDone указывает, считается ли отправка выполненной, то есть когда она была отправлена
// and there are no errors.
func (f FormSubmission) IsDone() bool {
	return f.IsSubmitted && !f.HasErrors()
}

// setErrorMessages устанавливает сообщения об ошибках при отправке для всех полей, проверка которых не удалась
func (f *FormSubmission) setErrorMessages(err error) {
	// Прямо сейчас поддерживается только это
	ves, ok := err.(validator.ValidationErrors)
	if !ok {
		return
	}

	for _, ve := range ves {
		var message string

		// Предоставлять лучшие сообщения об ошибках в зависимости от тега неудачной проверки
		// Это должно быть расширено по мере использования дополнительных тегов в вашей проверке
		switch ve.Tag() {
		case "required":
			message = "Это поле является обязательным."
		case "email":
			message = "Введите действительный адрес электронной почты."
		case "eqfield":
			message = "Не соответствует."
		default:
			message = "Недопустимое значение."
		}

		// Add the error
		f.SetFieldError(ve.Field(), message)
	}
}
