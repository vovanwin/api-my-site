package context

import (
	"context"
	"errors"
)

const (
	// AuthenticatedUserKey является ли значение ключа используемым для хранения аутентифицированного пользователя в контексте
	AuthenticatedUserKey = "auth_user"

	// UserKey является ли значение ключа используемым для хранения пользователя в контексте
	UserKey = "user"

	// FormKey является ли ключевое значение, используемое для хранения формы в контексте
	FormKey = "form"

	// PasswordTokenKey является ли значение ключа используемым для хранения токена пароля в контексте
	PasswordTokenKey = "password_token"
)

// IsCanceledError определяет, вызвана ли ошибка отменой контекста
func IsCanceledError(err error) bool {
	return errors.Is(err, context.Canceled)
}
