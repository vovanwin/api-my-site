package middleware

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/vovanwin/api-my-site/ent"
	"net/http"

	"github.com/vovanwin/api-my-site/pkg/context"
	"github.com/vovanwin/api-my-site/pkg/services"

	"github.com/labstack/echo/v4"
)

// LoadAuthenticatedUser загружает аутентифицированного пользователя, если таковой имеется, и сохраняет в контексте
func LoadAuthenticatedUser(authClient *services.AuthClient) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			u, err := authClient.GetAuthenticatedUser(c)

			switch err.(type) {
			case *ent.NotFoundError:
				logrus.Warn("авторизованный пользователь не найден")
			case services.NotAuthenticatedError:
			case nil:
				c.Set(context.AuthenticatedUserKey, u)
				logrus.Infof("авторизованный пользователь, загруженный в контекст: %d", u.ID)
			default:
				return echo.NewHTTPError(
					http.StatusInternalServerError,
					fmt.Sprintf("ошибка при запросе аутентифицированного пользователя: %v", err),
				)
			}

			return next(c)
		}
	}
}

// RequireAuthentication требуется, чтобы пользователь прошел аутентификацию для продолжения
func RequireAuthentication() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if u := c.Get(context.AuthenticatedUserKey); u == nil {
				return echo.NewHTTPError(http.StatusUnauthorized)
			}

			return next(c)
		}
	}
}

// RequireNoAuthentication требуется, чтобы пользователь не проходил аутентификацию для продолжения
func RequireNoAuthentication() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if u := c.Get(context.AuthenticatedUserKey); u != nil {
				return echo.NewHTTPError(http.StatusForbidden)
			}

			return next(c)
		}
	}
}
