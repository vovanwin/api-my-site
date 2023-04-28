package controller

import (
	"fmt"
	"github.com/golang-jwt/jwt/v4"
	echojwt "github.com/labstack/echo-jwt/v4"
	"net/http"

	"github.com/vovanwin/api-my-site/pkg/services"

	"github.com/labstack/echo/v4"
)

// Controller предоставляет базовую функциональность и зависимости для маршрутов.
// Предлагаемый шаблон заключается в том, чтобы встроить контроллер в каждую отдельную структуру маршрута и использовать
// маршрутизатор для внедрения контейнера, чтобы ваши маршруты имели доступ к службам внутри контейнера
type Controller struct {
	// Container хранит контейнер служб, содержащий зависимости
	Container *services.Container
}

// NewController creates a new Controller
func NewController(c *services.Container) Controller {
	return Controller{
		Container: c,
	}
}

// Fail является помощником для сбоя запроса, возвращая ошибку 500 и регистрируя ошибку
func (c *Controller) Fail(err error, log string) error {
	return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("%s: %v", log, err))
}

func (c *Controller) AuthMiddleware() echo.MiddlewareFunc {
	return echojwt.WithConfig(echojwt.Config{
		NewClaimsFunc: func(ctx echo.Context) jwt.Claims {
			return new(services.JwtCustomClaims)
		},
		SigningKey: []byte(c.Container.Config.App.EncryptionKey),
	})
}
