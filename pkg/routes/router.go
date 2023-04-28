package routes

import (
	"github.com/golang-jwt/jwt/v4"
	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/sirupsen/logrus"
	"net/http"

	"github.com/vovanwin/api-my-site/config"
	"github.com/vovanwin/api-my-site/pkg/controller"
	"github.com/vovanwin/api-my-site/pkg/middleware"
	"github.com/vovanwin/api-my-site/pkg/services"

	"github.com/labstack/echo/v4"
	echomw "github.com/labstack/echo/v4/middleware"
)

// BuildRouter builds the router
func BuildRouter(c *services.Container) {
	log := logrus.New()

	// Статические файлы с надлежащим управлением кэшем
	// функциональная карта.File() следует использовать в шаблонах для добавления ключа кэша к URL-адресу, чтобы разбить кэш
	// после каждого перезапуска сервера
	c.Web.Group("", middleware.CacheControl(c.Config.Cache.Expiration.StaticFile)).
		Static(config.StaticPrefix, config.StaticDir)

	// Нестатическая группа маршрутов к файлам
	g := c.Web.Group("api")

	// Force HTTPS, if enabled
	if c.Config.HTTP.TLS.Enabled {
		g.Use(echomw.HTTPSRedirect())
	}

	g.Use(
		echomw.RemoveTrailingSlashWithConfig(echomw.TrailingSlashConfig{
			RedirectCode: http.StatusMovedPermanently,
		}),
		echomw.Recover(),
		echomw.Secure(),
		echomw.RequestID(),
		echomw.Gzip(),
		echomw.RequestLoggerWithConfig(echomw.RequestLoggerConfig{
			LogURI:    true,
			LogStatus: true,
			LogValuesFunc: func(c echo.Context, values echomw.RequestLoggerValues) error {
				log.WithFields(logrus.Fields{
					"URI":       values.URI,
					"status":    values.Status,
					"Method":    values.Method,
					"Error":     values.Error,
					"latency":   values.Latency,
					"RequestID": values.RequestID,
				}).Info("request")

				return nil
			},
		}),
		middleware.LogRequestID(),
		// Configure middleware with the custom claims type
		echojwt.WithConfig(echojwt.Config{
			NewClaimsFunc: func(ctx echo.Context) jwt.Claims {
				return new(services.JwtCustomClaims)
			},
			SigningKey: []byte(c.Config.App.EncryptionKey),
		}),
		echomw.TimeoutWithConfig(echomw.TimeoutConfig{
			Timeout: c.Config.App.Timeout,
		}),
		middleware.LoadAuthenticatedUser(c.Auth),
		middleware.ServeCachedPage(c.Cache),
	)

	// Base controller
	ctr := controller.NewController(c)

	// Example routes
	navRoutes(c, g, ctr)
	userRoutes(c, g, ctr)
}

func navRoutes(c *services.Container, g *echo.Group, ctr controller.Controller) {

	groupNav := g.Group("/user", ctr.AuthMiddleware())

	groupNav.GET("/test", func(c echo.Context) error {

		return c.JSON(http.StatusOK, "claims")
	}, middleware.RequireAuthentication())
}

func userRoutes(c *services.Container, g *echo.Group, ctr controller.Controller) {

	noAuth := g.Group("/auth", middleware.RequireNoAuthentication())
	login := login{Controller: ctr}
	noAuth.POST("/login", login.Post).Name = "login.post"

	register := register{Controller: ctr}
	noAuth.POST("/register", register.Post).Name = "register.post"
}
