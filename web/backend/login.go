package backend

import (
	"net/http"

	"github.com/gorilla/sessions"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/megamon/core/utils"
)

func loginPage(c echo.Context) error {
	if login := getLoginFromSession(c); login != "" {
		return c.Redirect(http.StatusFound, "/")
	}

	return c.Render(http.StatusOK, "login.html", nil)
}

func handleLogin(c echo.Context) error {
	login := c.FormValue("username")
	password := c.FormValue("password")

	if login == utils.Settings.AdminCredentials.Username && password == utils.Settings.AdminCredentials.Password {
		sess := loginSession(c, login)
		if err := sess.Save(c.Request(), c.Response()); err != nil {
			return c.Render(http.StatusUnprocessableEntity, "login.html", "error")
		}
		return c.Redirect(http.StatusFound, "/")
	}
	return c.Render(http.StatusOK, "login.html",
		struct {
			Error string
		}{"Incorrect username or password"},
	)
}

func loginSession(c echo.Context, login string) *sessions.Session {
	sess, _ := session.Get("session", c)
	sess.Values["username"] = login

	sess.Options = &sessions.Options{
		Path: "/",
	}

	sess.Values["username"] = login
	return sess
}

func loginRequired(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		if login := getLoginFromSession(c); login == "" {
			return c.Redirect(http.StatusFound, "/login")
		}
		return next(c)
	}
}

func getLoginFromSession(c echo.Context) string {
	sess, _ := session.Get("session", c)
	login, exists := sess.Values["username"]
	if !exists {
		return ""
	}
	return login.(string)
}
