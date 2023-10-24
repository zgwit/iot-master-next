package zgwit_ctrler

import (
	zgwit_model "local/model"
	zgwit_plugin "local/plugin"

	"github.com/gin-gonic/gin"
)

type MiddlewareCtrler struct {
	Middlewares *[]zgwit_model.Middleware
}

func (ctrler *MiddlewareCtrler) Init(middlewares *[]zgwit_model.Middleware) (err error) {

	ctrler.Middlewares = middlewares

	return
}

func (ctrler *MiddlewareCtrler) List(ctx *gin.Context) {

	zgwit_plugin.HttpSuccess(ctx, "成功", ctrler.Middlewares)
}

func (ctrler *MiddlewareCtrler) Status(ctx *gin.Context) {

	status := map[string]zgwit_model.KeyValue{}

	for _, item := range *ctrler.Middlewares {
		status[item.Name] = item.Status()
	}

	zgwit_plugin.HttpSuccess(ctx, "成功", status)
}

func (ctrler *MiddlewareCtrler) Update(ctx *gin.Context) {

	middleware_param := zgwit_model.Middleware{}

	if err := ctx.Bind(&middleware_param); err != nil {
		zgwit_plugin.HttpFailure(ctx, "参数格式错误", zgwit_plugin.REQUEST_QUERY_ERR, err)
		return
	}

	exist, index := false, 0

	for index = range *ctrler.Middlewares {
		if (*ctrler.Middlewares)[index].Name == middleware_param.Name {
			exist = true
			break
		}
	}

	if !exist {
		zgwit_plugin.HttpFailure(ctx, "中间件不存在", zgwit_plugin.REQUEST_FAIL, nil)
		return
	}

	if err_str := (*ctrler.Middlewares)[index].CheckFormat(); err_str != "" {
		zgwit_plugin.HttpFailure(ctx, "参数格式错误", zgwit_plugin.REQUEST_QUERY_ERR, err_str)
		return
	}

	middleware_param.Name = (*ctrler.Middlewares)[index].Name

	if err := (*ctrler.Middlewares)[index].Update(middleware_param); err != nil {
		zgwit_plugin.HttpFailure(ctx, "请求失败，请稍后重试", zgwit_plugin.REQUEST_SERVER_ERR, err)
		return
	}

	zgwit_plugin.HttpSuccess(ctx, "成功", nil)
}

func (ctrler *MiddlewareCtrler) Control(ctx *gin.Context) {

	command := ctx.Query("command")
	name := ctx.Query("name")

	exist, index := false, 0

	for index = range *ctrler.Middlewares {
		if (*ctrler.Middlewares)[index].Name == name {
			exist = true
			break
		}
	}

	if !exist {
		zgwit_plugin.HttpFailure(ctx, "中间件不存在", zgwit_plugin.REQUEST_FAIL, nil)
		return
	}

	switch command {
	case "start":
		(*ctrler.Middlewares)[index].Start()
	case "stop":
		(*ctrler.Middlewares)[index].Stop()
	default:
		zgwit_plugin.HttpFailure(ctx, "不支持的命令", zgwit_plugin.REQUEST_FAIL, nil)
		return
	}

	zgwit_plugin.HttpSuccess(ctx, "成功", nil)
}

func (ctrler *MiddlewareCtrler) Log(ctx *gin.Context) {

	name := ctx.Query("name")

	exist, index := false, 0

	for index = range *ctrler.Middlewares {
		if (*ctrler.Middlewares)[index].Name == name {
			exist = true
			break
		}
	}

	if !exist {
		zgwit_plugin.HttpFailure(ctx, "中间件不存在", zgwit_plugin.REQUEST_FAIL, nil)
		return
	}

	contents, err := (*ctrler.Middlewares)[index].Log(100)

	if err != nil {
		zgwit_plugin.HttpFailure(ctx, "请求失败，请稍后重试", zgwit_plugin.REQUEST_SERVER_ERR, err)
		return
	}

	zgwit_plugin.HttpSuccess(ctx, "成功", contents)
}
