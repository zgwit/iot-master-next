package ctrler

import (
	"github.com/zgwit/iot-master-next/src/model"

	"github.com/zgwit/iot-master-next/src/plugin"

	"github.com/gin-gonic/gin"
)

type ApplicationCtrler struct {
	Applications *[]model.ApplicationType
}

func (ctrler *ApplicationCtrler) Init(applications *[]model.ApplicationType) (err error) {

	ctrler.Applications = applications

	return
}

func (ctrler *ApplicationCtrler) List(ctx *gin.Context) {

	plugin.HttpSuccess(ctx, "成功", ctrler.Applications)
}

func (ctrler *ApplicationCtrler) Status(ctx *gin.Context) {

	status := map[string]model.KeyValueType{}

	for _, item := range *ctrler.Applications {
		status[item.Name] = item.Status()
	}

	plugin.HttpSuccess(ctx, "成功", status)
}

func (ctrler *ApplicationCtrler) Update(ctx *gin.Context) {

	application_param := model.ApplicationType{}

	if err := ctx.Bind(&application_param); err != nil {
		plugin.HttpFailure(ctx, "参数格式错误", plugin.REQUEST_QUERY_ERR, err)
		return
	}

	exist, index := false, 0

	for index = range *ctrler.Applications {
		if (*ctrler.Applications)[index].Name == application_param.Name {
			exist = true
			break
		}
	}

	if !exist {
		plugin.HttpFailure(ctx, "中间件不存在", plugin.REQUEST_FAIL, nil)
		return
	}

	if err_str := (*ctrler.Applications)[index].CheckFormat(); err_str != "" {
		plugin.HttpFailure(ctx, "参数格式错误", plugin.REQUEST_QUERY_ERR, err_str)
		return
	}

	application_param.Name = (*ctrler.Applications)[index].Name

	if err := (*ctrler.Applications)[index].Update(application_param); err != nil {
		plugin.HttpFailure(ctx, "请求失败，请稍后重试", plugin.REQUEST_SERVER_ERR, err)
		return
	}

	plugin.HttpSuccess(ctx, "成功", nil)
}

func (ctrler *ApplicationCtrler) Control(ctx *gin.Context) {

	command := ctx.Query("command")
	name := ctx.Query("name")

	exist, index := false, 0

	for index = range *ctrler.Applications {
		if (*ctrler.Applications)[index].Name == name {
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
		(*ctrler.Applications)[index].Start()
	case "stop":
		(*ctrler.Applications)[index].Stop()
	default:
		plugin.HttpFailure(ctx, "不支持的命令", plugin.REQUEST_FAIL, nil)
		return
	}

	plugin.HttpSuccess(ctx, "成功", nil)
}

func (ctrler *ApplicationCtrler) Log(ctx *gin.Context) {

	name := ctx.Query("name")

	exist, index := false, 0

	for index = range *ctrler.Applications {
		if (*ctrler.Applications)[index].Name == name {
			exist = true
			break
		}
	}

	if !exist {
		plugin.HttpFailure(ctx, "中间件不存在", plugin.REQUEST_FAIL, nil)
		return
	}

	contents, err := (*ctrler.Applications)[index].Log(100)

	if err != nil {
		plugin.HttpFailure(ctx, "请求失败，请稍后重试", plugin.REQUEST_SERVER_ERR, err)
		return
	}

	plugin.HttpSuccess(ctx, "成功", contents)
}
