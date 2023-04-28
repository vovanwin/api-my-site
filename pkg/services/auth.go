package services

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/vovanwin/api-my-site/ent"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/vovanwin/api-my-site/config"
	"github.com/vovanwin/api-my-site/ent/passwordtoken"
	"github.com/vovanwin/api-my-site/ent/user"
	"github.com/vovanwin/api-my-site/pkg/context"

	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
)

const (
	// authSessionName stores the name of the session which contains authentication data
	authSessionName = "ua"

	// authSessionKeyUserID stores the key used to store the user ID in the session
	authSessionKeyUserID = "user_id"

	// authSessionKeyAuthenticated stores the key used to store the authentication status in the session
	authSessionKeyAuthenticated = "authenticated"
)

// NotAuthenticatedError is an error returned when a user is not authenticated
type NotAuthenticatedError struct{}

// Error implements the error interface.
func (e NotAuthenticatedError) Error() string {
	return "user not authenticated"
}

// InvalidPasswordTokenError is an error returned when an invalid token is provided
type InvalidPasswordTokenError struct{}

// Error implements the error interface.
func (e InvalidPasswordTokenError) Error() string {
	return "invalid password token"
}

// AuthClient is the client that handles authentication requests
type AuthClient struct {
	config *config.Config
	orm    *ent.Client
}

// NewAuthClient creates a new authentication client
func NewAuthClient(cfg *config.Config, orm *ent.Client) *AuthClient {
	return &AuthClient{
		config: cfg,
		orm:    orm,
	}
}

// Login logs in a user of a given ID
func (c *AuthClient) Login(ctx echo.Context, userID int) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(c.config.App.EmailVerificationTokenExpiration).Unix(),
	})

	return token.SignedString([]byte(c.config.App.EncryptionKey))
}

// Logout выводит запрашивающего пользователя из системы
func (c *AuthClient) Logout(ctx echo.Context) error {

	return nil
}

type JwtCustomClaims struct {
	UserId int `json:"user_id"`
	jwt.RegisteredClaims
}

// GetAuthenticatedUserID возвращает идентификатор аутентифицированного пользователя, если пользователь вошел в систему
func (c *AuthClient) GetAuthenticatedUserID(ctx echo.Context) (int, error) {
	if ctx.Get("user") == nil {
		return 0, NotAuthenticatedError{}
	}
	userToken := ctx.Get("user").(*jwt.Token)
	claims := userToken.Claims.(*JwtCustomClaims)
	userId := claims.UserId
	return userId, nil
}

// GetAuthenticatedUser возвращает аутентифицированного пользователя, если пользователь вошел в систему
func (c *AuthClient) GetAuthenticatedUser(ctx echo.Context) (*ent.User, error) {
	if userID, err := c.GetAuthenticatedUserID(ctx); err == nil {
		return c.orm.User.Query().
			Where(user.ID(userID)).
			Only(ctx.Request().Context())
	}

	return nil, NotAuthenticatedError{}
}

// HashPassword возвращает хэш заданного пароля
func (c *AuthClient) HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// CheckPassword проверьте, соответствует ли данный пароль заданному хэшу
func (c *AuthClient) CheckPassword(password, hash string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}

// GeneratePasswordResetToken генерирует токен сброса пароля для данного пользователя.
// В целях безопасности сам токен не хранится в базе данных, а скорее
// хэш токена, точно указывающий, как обрабатываются пароли. Этот метод возвращает оба
// сгенерированный токен, а также объект token, который содержит только хэш.
func (c *AuthClient) GeneratePasswordResetToken(ctx echo.Context, userID int) (string, *ent.PasswordToken, error) {
	// Сгенерируйте токен, который будет указан в URL, но не в базе данных
	token, err := c.RandomToken(c.config.App.PasswordToken.Length)
	if err != nil {
		return "", nil, err
	}

	// Хэшируйте токен, который будет храниться в базе данных
	hash, err := c.HashPassword(token)
	if err != nil {
		return "", nil, err
	}

	// Create and save the password reset token
	pt, err := c.orm.PasswordToken.
		Create().
		SetHash(hash).
		SetUserID(userID).
		Save(ctx.Request().Context())

	return token, pt, err
}

// GetValidPasswordToken возвращает действительный объект токена пароля с не истекшим сроком действия для данного
// пользователя, идентификатора токена и токена-метки.
// Поскольку фактический токен не хранится в базе данных в целях безопасности, если найден соответствующий объект токена пароля,
// хэш предоставленного токена сравнивается с хэшем, хранящимся в базе данных, для проверки.
func (c *AuthClient) GetValidPasswordToken(ctx echo.Context, userID, tokenID int, token string) (*ent.PasswordToken, error) {
	// Ensure expired tokens are never returned
	expiration := time.Now().Add(-c.config.App.PasswordToken.Expiration)

	// Запрос для поиска объекта password token, который соответствует заданному пользователю и идентификатору токена
	pt, err := c.orm.PasswordToken.
		Query().
		Where(passwordtoken.ID(tokenID)).
		Where(passwordtoken.HasUserWith(user.ID(userID))).
		Where(passwordtoken.CreatedAtGTE(expiration)).
		Only(ctx.Request().Context())

	switch err.(type) {
	case *ent.NotFoundError:
	case nil:
		// Проверьте токен на соответствие хэшу
		if err := c.CheckPassword(token, pt.Hash); err == nil {
			return pt, nil
		}
	default:
		if !context.IsCanceledError(err) {
			return nil, err
		}
	}

	return nil, InvalidPasswordTokenError{}
}

// DeletePasswordTokens удаляет все токены пароля в базе данных, относящиеся к данному пользователю.
// Это должно быть вызвано после успешного сброса пароля.
func (c *AuthClient) DeletePasswordTokens(ctx echo.Context, userID int) error {
	_, err := c.orm.PasswordToken.
		Delete().
		Where(passwordtoken.HasUserWith(user.ID(userID))).
		Exec(ctx.Request().Context())

	return err
}

// RandomToken генерирует случайную строку токена заданной длины
func (c *AuthClient) RandomToken(length int) (string, error) {
	b := make([]byte, (length/2)+1)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	token := hex.EncodeToString(b)
	return token[:length], nil
}

// GenerateEmailVerificationToken генерирует токен подтверждения электронной почты для данного адреса электронной
// почты с помощью JWT, который устанавливается на срок действия в зависимости от продолжительности, сохраненной в конфигурации
func (c *AuthClient) GenerateEmailVerificationToken(email string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"email": email,
		"exp":   time.Now().Add(c.config.App.EmailVerificationTokenExpiration).Unix(),
	})

	return token.SignedString([]byte(c.config.App.EncryptionKey))
}

// ValidateEmailVerificationToken проверяет токен подтверждения электронной почты и возвращает связанный адрес
// электронной почты, если токен действителен и срок его действия не истек
func (c *AuthClient) ValidateEmailVerificationToken(token string) (string, error) {
	t, err := jwt.Parse(token, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("неожиданный способ подписи: %v", t.Header["alg"])
		}

		return []byte(c.config.App.EncryptionKey), nil
	})

	if err != nil {
		return "", err
	}

	if claims, ok := t.Claims.(jwt.MapClaims); ok && t.Valid {
		return claims["email"].(string), nil
	}

	return "", errors.New("недействительный или просроченный токен")
}
