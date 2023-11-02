package plugin

import (
	"net/http"

	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
)

const (
	REQUEST_SUCCESS       = 200 // 请求成功
	REQUEST_FAIL          = 411 // 请求失败
	REQUEST_QUERY_ERR     = 412 // 参数格式错误
	REQUEST_SERVER_ERR    = 413 // 服务执行异常
	REQUEST_RESPONSE_ERR  = 414 // 返回数据格式错误
	REQUEST_TOKEN_OVERDUE = 886 // 令牌失效
)

type HttpServerConfig struct {
	Url      string `form:"url" bson:"url" json:"url"`
	Mode     string `form:"mode" bson:"mode" json:"mode"`
	HtmlPath string `form:"html_path" bson:"html_path" json:"html_path"`
}

type HttpServer struct {
	Router *gin.Engine

	Config HttpServerConfig
}

func NewHttpServer(config HttpServerConfig) (http_server HttpServer) {

	http_server.Config = config

	if http_server.Config.Mode == "" {
		http_server.Config.Mode = "release"
	}

	gin.SetMode(http_server.Config.Mode)

	http_server.Router = gin.New()

	http_server.Router.Use(func(ctx *gin.Context) {

		method := ctx.Request.Method

		ctx.Header("Access-Control-Allow-Origin", "*")
		ctx.Header("Access-Control-Allow-Headers", "Content-Type,AccessToken,X-CSRF-Token, Authorization, Token, x-token")
		ctx.Header("Access-Control-Allow-Methods", "POST, GET, OPTIONS, DELETE, PATCH, PUT")
		ctx.Header("Access-Control-Expose-Headers", "Content-Length, Access-Control-Allow-Origin, Access-Control-Allow-Headers, Content-Type")
		ctx.Header("Access-Control-Allow-Credentials", "true")

		if method == "OPTIONS" {
			ctx.AbortWithStatus(http.StatusNoContent)
		}
	})

	if http_server.Config.HtmlPath != "" {

		http_server.Router.Static("/static", http_server.Config.HtmlPath+"./static")
		http_server.Router.StaticFile("/", http_server.Config.HtmlPath+"./index.html")

		http_server.Router.NoRoute(func(c *gin.Context) {
			c.File(http_server.Config.HtmlPath + "/index.html")
		})

		http_server.Router.Use(static.Serve("/", static.LocalFile(http_server.Config.HtmlPath+"/index.html", false)))
	}

	return
}

func (http_server *HttpServer) Running() (err error) {

	return http_server.Router.Run(http_server.Config.Url)
}

func HttpDefault(ctx *gin.Context, msg string, code int, data interface{}) {

	ctx.JSON(http.StatusOK, map[string]interface{}{
		"msg":  msg,
		"code": code,
		"data": data,
	})
}

func HttpFailure(ctx *gin.Context, msg string, code int, data interface{}) {

	message := map[string]interface{}{
		"msg":  msg,
		"code": code,
	}

	if data != nil {
		message["data"] = data
	}

	ctx.JSON(http.StatusOK, message)
}

func HttpSuccess(ctx *gin.Context, msg string, data interface{}) {

	message := map[string]interface{}{
		"msg":  msg,
		"code": REQUEST_SUCCESS,
	}

	if data != nil {
		message["data"] = data
	}

	ctx.JSON(http.StatusOK, message)
}
