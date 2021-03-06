package backend

import (
	"html/template"
	"io"

	"github.com/gorilla/sessions"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
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
func (b *Backend) Start(p Params) {
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

	//Pass database & job queues to context
	e.Use(func(h echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			cc := Context{b, p, c}
			return h(cc)
		}
	})

	e.File("/", "web/frontend/index.html", loginRequired)
	e.Static("/static", "web/frontend/static/")

	e.GET("/leaks/api/report/events", getReportedEvents, basicAuthRequired)
	e.GET("/leaks/api/report/events/count", getNewReportCount, basicAuthRequired)
	e.GET("/leaks/api/report/frags/:datatype/:status", getFragments, loginRequired)
	e.GET("/leaks/api/report/count/:datatype/:status", getFragmentCount, loginRequired)
	e.GET("/leaks/api/report/info/:frag_id", getFragmentInfo, loginRequired)
	e.GET("/leaks/api/report/mark/:frag_id/:status", markFragment, loginRequired)

	e.GET("/leaks/api/settings", getSettings, loginRequired)
	e.POST("/leaks/api/settings", updateSettings, loginRequired)

	e.GET("/leaks/api/task/all/start", startAllTasks, basicAuthRequired)
	e.GET("/leaks/api/task/:task/:state", taskManager, loginRequired)
	e.GET("/leaks/api/task/available", tasksAvailable, loginRequired)

	e.GET("/login", loginPage)
	e.POST("/login", handleLogin)

	e.File("/*", "web/frontend/index.html", loginRequired)

	e.HideBanner = false
	e.Debug = false

	//e.Logger.Fatal(e.StartAutoTLS(":1234"))
	e.Logger.Fatal(e.Start(":1234"))
}
