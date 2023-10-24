package ctrler

import (
	"github.com/zgwit/iot-master-next/src/model"
	"github.com/zgwit/iot-master-next/src/plugin"

	"github.com/gin-gonic/gin"
)

type MiddlewareCtrler struct {
	Middlewares *[]model.MiddlewareType
}

func (ctrler *MiddlewareCtrler) Init(middlewares *[]model.MiddlewareType) (err error) {

	ctrler.Middlewares = middlewares

	return
}

func (ctrler *MiddlewareCtrler) List(ctx *gin.Context) {

	plugin.HttpSuccess(ctx, "成功", ctrler.Middlewares)
}

func (ctrler *MiddlewareCtrler) Status(ctx *gin.Context) {

	status := map[string]model.KeyValueType{}

	for _, item := range *ctrler.Middlewares {
		status[item.Name] = item.Status()
	}

	plugin.HttpSuccess(ctx, "成功", status)
}

func (ctrler *MiddlewareCtrler) Update(ctx *gin.Context) {

	middleware_param := model.MiddlewareType{}

	if err := ctx.Bind(&middleware_param); err != nil {
		plugin.HttpFailure(ctx, "参数格式错误", plugin.REQUEST_QUERY_ERR, err)
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
		plugin.HttpFailure(ctx, "中间件不存在", plugin.REQUEST_FAIL, nil)
		return
	}

	if err_str := (*ctrler.Middlewares)[index].CheckFormat(); err_str != "" {
		plugin.HttpFailure(ctx, "参数格式错误", plugin.REQUEST_QUERY_ERR, err_str)
		return
	}

	middleware_param.Name = (*ctrler.Middlewares)[index].Name

	if err := (*ctrler.Middlewares)[index].Update(middleware_param); err != nil {
		plugin.HttpFailure(ctx, "请求失败，请稍后重试", plugin.REQUEST_SERVER_ERR, err)
		return
	}

	plugin.HttpSuccess(ctx, "成功", nil)
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
		plugin.HttpFailure(ctx, "中间件不存在", plugin.REQUEST_FAIL, nil)
		return
	}

	switch command {
	case "start":
		(*ctrler.Middlewares)[index].Start()
	case "stop":
		(*ctrler.Middlewares)[index].Stop()
	default:
		plugin.HttpFailure(ctx, "不支持的命令", plugin.REQUEST_FAIL, nil)
		return
	}

	plugin.HttpSuccess(ctx, "成功", nil)
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
		plugin.HttpFailure(ctx, "中间件不存在", plugin.REQUEST_FAIL, nil)
		return
	}

	contents, err := (*ctrler.Middlewares)[index].Log(100)

	if err != nil {
		plugin.HttpFailure(ctx, "请求失败，请稍后重试", plugin.REQUEST_SERVER_ERR, err)
		return
	}

	plugin.HttpSuccess(ctx, "成功", contents)
}
