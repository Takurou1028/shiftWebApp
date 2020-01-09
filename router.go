package main

import(	
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/labstack/echo-contrib/session"
	"github.com/gorilla/sessions"
	"html/template"
	"io"
	"shift-webapp/controllers"
)

type TemplateRenderer struct{
	templates *template.Template
}

func (t *TemplateRenderer) Render(w io.Writer, name string, data interface{}, c echo.Context) error{
	return t.templates.ExecuteTemplate(w, name, data)
}

func newRouter() *echo.Echo{
	e := echo.New()
	renderer := &TemplateRenderer{
		templates: template.Must(template.ParseFiles("views/finishing_signup.html", "views/owner.html",
		 "views/registration.html", "views/user.html", "views/logout.html", "views/shiftlist.html")),
	}
	e.Renderer = renderer

	e.Pre(controllers.MethodOverride)	//FormをDELETE,PUTに対応させる（_method要素から読み取り）

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(session.Middleware(sessions.NewCookieStore([]byte("secret"))))	//セッション用


	e.File("/", "views/index.html")
	e.File("/signup", "views/signup.html")
	e.POST("/signup", controllers.Signup)
	e.File("/login", "views/login.html")
	e.POST("/login", controllers.Login)
	e.GET("/logout_", controllers.Logout)
	
	e.GET("/owner", controllers.ShowOwnerPage)
//	e.POST("/owner", controllers.ShowOwnerPage)
	e.GET("/owner/registration", controllers.CreateUserList)
	e.POST("/owner/registration", controllers.RegisterUser)
	e.DELETE("/owner/registration", controllers.DeleteUser)
	e.GET("/owner/shift", controllers.CreateShiftList)

	e.GET("/user", controllers.ShowUserPage)
	e.POST("/user", controllers.SubmitShift)
	return e
}