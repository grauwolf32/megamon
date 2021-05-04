package backend

import (
	"html/template"
	"io"

	"github.com/gorilla/sessions"
	"github.com/labstack/echo"
	"github.com/labstack/echo-contrib/session"
	"github.com/megamon/core/leaks/models"
	"github.com/megamon/core/utils"
)

//Template : custom template type
type Template struct {
	templates *template.Template
}

//Backend : backend instance
type Backend struct {
	DBManager models.Manager
}

//Render : render template function
func (t *Template) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

//Start : initialize backend
func (b *Backend) Start() {
	e := echo.New()
	t := &Template{
		templates: template.Must(template.ParseGlob("web/frontend/templates/*")),
	}
	secret := utils.RandBytes(20)
	cookieStore := sessions.NewCookieStore(secret)
	csMiddleware := session.Middleware(cookieStore)
	e.Use(csMiddleware)
	e.Renderer = t
	b.DBManager.Init()

	//Pass database to context
	e.Use(func(h echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			cc := Context{b, c}
			return h(cc)
		}
	})

	e.File("/", "web/frontend/index.html", loginRequired)
	e.Static("/static", "web/frontend/static/")

	e.File("/leaks/controls", "web/frontend/index.html", loginRequired)
	e.File("/leaks/settings", "web/frontend/index.html", loginRequired)
	e.File("/leaks/github", "web/frontend/index.html", loginRequired)
	e.File("/leaks/gist", "web/frontend/index.html", loginRequired)

	e.GET("/leaks/api/report/frags/:datatype/:status", getReports, loginRequired)
	e.GET("/leaks/api/report/info/:frag_id", getFragmentInfo, loginRequired)
	e.GET("/leaks/api/report/mark/:frag_id/:status", markFragment, loginRequired)

	e.GET("/leaks/api/settings", getSettings, loginRequired)
	e.POST("/leaks/api/settings", updateSettings, loginRequired)

	e.GET("/leaks/api/regexp", getRegexps, loginRequired)
	e.GET("/leaks/api/regexp/remove/:id", delRegexp, loginRequired)
	e.POST("/leaks/api/regexp", addRegexp, loginRequired)

	e.GET("/leaks/api/keywords", getKeywords, loginRequired)
	e.GET("/leaks/api/keywords/remove/:id", delKeyword, loginRequired)
	e.POST("/leaks/api/keywords", addKeyword, loginRequired)

	e.GET("/login", loginPage)
	e.POST("/login", handleLogin)

	e.HideBanner = true
	e.Debug = false

	//e.Logger.Fatal(e.StartAutoTLS(":1234"))
	e.Logger.Fatal(e.Start(":1234"))
}
