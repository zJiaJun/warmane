package route

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/zJiajun/warmane/engine"
	"html/template"
	"net/http"
	"time"
)

type MenuItem struct {
	Text       string
	Link       string
	Active     bool
	ActiveFunc ActiveFunc
}

var menuItems = []*MenuItem{
	{Text: "Home", Link: "/", ActiveFunc: func(name string) bool {
		return name == "index"
	}},
	{Text: "Points", Link: "/points", ActiveFunc: func(name string) bool {
		return name == "points"
	}},
	{Text: "Trade", Link: "/trade", ActiveFunc: func(name string) bool {
		return name == "trade"
	}},
	{Text: "Account", Link: "/account", ActiveFunc: func(name string) bool {
		return name == "account"
	}},
}

type ActiveFunc func(string) bool

func New(engine *engine.Engine) *gin.Engine {
	r := gin.New()
	{
		r.Use(logger(), recovery())
		r.NoRoute(notFound())
		r.NoMethod(notAllowed())
		r.SetFuncMap(template.FuncMap{
			"formatAsDate": formatAsDate,
		})
		r.LoadHTMLGlob("templates/*")
		r.Static("/assets", "./assets")
		r.StaticFile("/favicon.ico", "./assets/favicon.ico")
	}
	r.GET("/", index())

	points := r.Group("/points")
	{
		points.GET("", listAccountDetailsHandler(engine))
		points.GET("/:id/refresh", refreshAccountDetailsHandler(engine))
		points.GET("/:id/collect", collectAccountPointsHandler(engine))
	}

	account := r.Group("/account")
	{
		account.GET("", listAccountHandler(engine))
		account.POST("", createAccountHandler(engine))
		account.PUT("/:id", updateAccountHandler(engine))
		account.DELETE("/:id", deleteAccountHandler(engine))
		account.GET("/:id/check", checkAccountHandler(engine))
	}
	return r
}

func activeMenusItems(name string) []*MenuItem {
	for _, item := range menuItems {
		item.Active = item.ActiveFunc(name)
	}
	return menuItems
}

func logger() gin.HandlerFunc {
	return gin.LoggerWithConfig(gin.LoggerConfig{})
}

func recovery() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, err any) {
		withHTMLData(c, "error", gin.H{
			"Code":  http.StatusInternalServerError,
			"Title": "Something went wrong",
			"Msg":   fmt.Sprintf("%v", err),
		})
	})
}

func notFound() gin.HandlerFunc {
	return func(c *gin.Context) {
		withHTMLData(c, "error", gin.H{
			"Code":  http.StatusNotFound,
			"Title": "You're lost",
			"Msg":   "The page you are looking for was not found.",
		})
	}
}

func notAllowed() gin.HandlerFunc {
	return func(c *gin.Context) {
		withHTMLData(c, "error", gin.H{
			"Code":  http.StatusMethodNotAllowed,
			"Title": "You're lost",
			"Msg":   http.StatusText(http.StatusMethodNotAllowed),
		})
	}
}

func index() gin.HandlerFunc {
	return func(c *gin.Context) {
		withHTMLData(c, "index", gin.H{})
	}
}

func withHTMLData(c *gin.Context, name string, data any) {
	c.HTML(http.StatusOK, name+".html", gin.H{
		"Menu": activeMenusItems(name),
		"Data": data,
	})
}

func withJSONData(c *gin.Context, data any, msg string) {
	c.JSON(http.StatusOK, &responseMessage{
		Code: http.StatusOK,
		Data: data,
		Msg:  msg,
	})
}

func withJsonError(c *gin.Context, code int, message any) {
	c.JSON(http.StatusOK, &responseMessage{
		Code: code,
		Data: "Something went wrong",
		Msg:  "Please try again later or contact the administrator. Reason: " + fmt.Sprintf("%v", message),
	})
}

func formatAsDate(t time.Time) string {
	year, month, day := t.Date()
	return fmt.Sprintf("%d-%02d-%02d %02d:%02d:%02d", year, month, day, t.Hour(), t.Minute(), t.Second())
}

type responseMessage struct {
	Code int    `json:"code,omitempty"`
	Data any    `json:"data,omitempty"`
	Msg  string `json:"msg,omitempty"`
}
