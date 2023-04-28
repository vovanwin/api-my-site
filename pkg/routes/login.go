package routes

import (
	"github.com/sirupsen/logrus"
	"net/http"
	"strings"

	"github.com/vovanwin/api-my-site/ent/user"
	"github.com/vovanwin/api-my-site/pkg/controller"

	"github.com/labstack/echo/v4"
)

type (
	login struct {
		controller.Controller
	}

	loginForm struct {
		Email      string `form:"email" validate:"required,email"`
		Password   string `form:"password" validate:"required"`
		Submission controller.FormSubmission
	}
)

func (c *login) Post(ctx echo.Context) error {
	var form loginForm

	authFailed := func() error {
		form.Submission.SetFieldError("Email", "")
		form.Submission.SetFieldError("Password", "")
		return ctx.JSON(http.StatusBadRequest, map[string]interface{}{
			"message": "Неверные учетные данные. Пожалуйста, попробуйте снова",
		})
	}

	// Parse the form values
	if err := ctx.Bind(&form); err != nil {
		logrus.Infof("не удается разобрать форму входа в систему: %s", err)
		return ctx.JSON(http.StatusBadRequest, map[string]interface{}{
			"message": "Неверные учетные данные. Пожалуйста, попробуйте снова",
		})
	}

	if err := form.Submission.Process(ctx, form); err != nil {
		logrus.Infof("\"не удается обработать отправку формы\": %s", err)
		return ctx.JSON(http.StatusBadRequest, map[string]interface{}{
			"message": "Неверные учетные данные. Пожалуйста, попробуйте снова",
		})
	}

	if form.Submission.HasErrors() {
		return ctx.JSON(http.StatusBadRequest, map[string]interface{}{
			"message": "Неверные учетные данные. Пожалуйста, попробуйте снова",
		})
	}

	// Попытка загрузить пользователя
	u, err := c.Container.ORM.User.
		Query().
		Where(user.Email(strings.ToLower(form.Email))).
		Only(ctx.Request().Context())

	switch err.(type) {
	case *ent.NotFoundError:
		return authFailed()
	case nil:
	default:
		return ctx.JSON(http.StatusBadRequest, map[string]interface{}{
			"message": "ошибка при запросе пользователя во время входа в систему",
		})
	}

	// Проверьте правильность пароля
	err = c.Container.Auth.CheckPassword(form.Password, u.Password)
	if err != nil {
		return authFailed()
	}

	// Войдите в систему пользователя
	token, err := c.Container.Auth.Login(ctx, u.ID)
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, map[string]interface{}{
			"message": "Неверные учетные данные. Пожалуйста, попробуйте снова",
		})
	}

	return ctx.JSON(http.StatusOK, map[string]interface{}{
		"token": token,
	})
}
