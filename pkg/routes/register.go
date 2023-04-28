package routes

import (
	"github.com/sirupsen/logrus"
	"net/http"

	"github.com/vovanwin/api-my-site/pkg/controller"

	"github.com/labstack/echo/v4"
)

type (
	register struct {
		controller.Controller
	}

	registerForm struct {
		Name            string `form:"name" validate:"required"`
		Email           string `form:"email" validate:"required,email"`
		Password        string `form:"password" validate:"required"`
		ConfirmPassword string `form:"password-confirm" validate:"required,eqfield=Password"`
		Submission      controller.FormSubmission
	}
)

func (c *register) Post(ctx echo.Context) error {
	var form registerForm

	// Parse the form values
	if err := ctx.Bind(&form); err != nil {
		return ctx.JSON(http.StatusBadRequest, map[string]interface{}{
			"message": "не удается разобрать регистрационную форму",
		})
	}

	if err := form.Submission.Process(ctx, form); err != nil {
		return ctx.JSON(http.StatusBadRequest, map[string]interface{}{
			"message": "не удается обработать отправку формы",
		})
	}
	logrus.Infof("ДО: %s", "Ошибки")

	if form.Submission.HasErrors() {
		return ctx.JSON(http.StatusUnprocessableEntity, form.Submission.GetAllFieldErrors())
	}
	logrus.Infof("После: %s", "Ошибки")

	// Hash the password
	pwHash, err := c.Container.Auth.HashPassword(form.Password)
	if err != nil {
		return ctx.JSON(http.StatusUnprocessableEntity, map[string]interface{}{
			"message": "не удается хешировать пароль",
		})
	}

	// Attempt creating the user
	u, err := c.Container.ORM.User.
		Create().
		SetName(form.Name).
		SetEmail(form.Email).
		SetPassword(pwHash).
		Save(ctx.Request().Context())

	switch err.(type) {
	case nil:
		logrus.Infof("создан пользователь: %s", u.Name)
	case *ent.ConstraintError:
		return ctx.JSON(http.StatusBadRequest, map[string]interface{}{
			"message": "Пользователь с этим адресом электронной почты уже существует. Пожалуйста, войдите в систему.",
		})
	default:
		logrus.Errorf("Ошибка создания пользователя: %s", err)
		return ctx.JSON(http.StatusBadRequest, map[string]interface{}{
			"message": "не удалось создать пользователя.",
		})
	}
	// Log the user in
	token, err := c.Container.Auth.Login(ctx, u.ID)
	if err != nil {
		logrus.Errorf("не удается войти в систему: %v", err)
		return ctx.JSON(http.StatusOK, map[string]interface{}{
			"message": "Ваша учетная запись была создана.",
		})
	}

	// Send the verification email
	//c.sendVerificationEmail(ctx, u)

	return ctx.JSON(http.StatusOK, map[string]interface{}{
		"token": token,
	})
}

func (c *register) sendVerificationEmail(ctx echo.Context, usr *ent.User) {
	// Generate a token
	//token, err := c.Container.Auth.GenerateEmailVerificationToken(usr.Email)
	//if err != nil {
	//	logrus.Errorf("не удается сгенерировать токен подтверждения электронной почты: %v", err)
	//	return
	//}
	//
	//// Send the email
	////url := ctx.Echo().Reverse("verify_email", token)
	////err = c.Container.Mail.
	////	Compose().
	////	To(usr.Email).
	////	Subject("Подтвердите свой адрес электронной почты").
	////	Body(fmt.Sprintf("Нажмите здесь, чтобы подтвердить свой адрес электронной почты: %s", url)).
	////	Send(ctx)
	//
	//if err != nil {
	//	logrus.Errorf("не удается отправить ссылку для подтверждения по электронной почте: %v", err)
	//	return
	//}
}
