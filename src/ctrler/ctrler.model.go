package zgwit_ctrler

import (
	zgwit_model "local/model"
	zgwit_plugin "local/plugin"
	zgwit_utils "local/utils"

	"github.com/gin-gonic/gin"
)

type ModelCtrler struct {
	Influx    *zgwit_plugin.Influx
	NsqClient *zgwit_plugin.NsqClient
}

func (ctrler *ModelCtrler) Init(influx *zgwit_plugin.Influx, nsq_client *zgwit_plugin.NsqClient) (err error) {

	ctrler.Influx = influx
	ctrler.NsqClient = nsq_client

	return
}

func dto_info(ctx *gin.Context, model *zgwit_model.Model) (result bool) {

	if err := ctx.Bind(model); err != nil {
		return
	}

	if len(model.Attributes) == 0 {
		model.Attributes = map[string]zgwit_model.ModelAttribute{}
	}

	if len(model.Events) == 0 {
		model.Events = map[string]zgwit_model.ModelEvent{}
	}

	if len(model.Actions) == 0 {
		model.Actions = map[string]zgwit_model.ModelAction{}
	}

	if model.Name == "" || model.Id == "" || model.Drive == "" {
		return
	}

	for _, item := range model.Attributes {
		if item.Name == "" {
			return
		}
		if item.DataType != "number" && item.DataType != "string" {
			return
		}
	}

	for _, item := range model.Events {
		if item.Name == "" {
			return
		}
	}

	for _, item := range model.Actions {
		if item.Name == "" {
			return
		}
	}

	return true
}

func (ctrler *ModelCtrler) List(ctx *gin.Context) {

	var (
		err error

		models = []zgwit_model.Model{}
	)

	if models, err = zgwit_model.ModelFetchList(); err != nil {
		zgwit_plugin.HttpFailure(ctx, "请求失败，请稍后重试", zgwit_plugin.REQUEST_SERVER_ERR, err)
		return
	}

	zgwit_plugin.HttpSuccess(ctx, "成功", models)
}

func (ctrler *ModelCtrler) Create(ctx *gin.Context) {

	var (
		err   error
		model = zgwit_model.Model{}
	)

	if !dto_info(ctx, &model) {
		zgwit_plugin.HttpFailure(ctx, "参数格式错误", zgwit_plugin.REQUEST_QUERY_ERR, err)
		return
	}

	if zgwit_model.ModelExist(model.Id) {
		zgwit_plugin.HttpFailure(ctx, "机型编号已存在", zgwit_plugin.REQUEST_FAIL, err)
		return
	}

	model.DriveConfig = nil

	if err := model.Create(); err != nil {
		zgwit_plugin.HttpFailure(ctx, "请求失败，请稍后重试", zgwit_plugin.REQUEST_SERVER_ERR, err)
		return
	}

	zgwit_plugin.HttpSuccess(ctx, "成功", nil)
}

func (ctrler *ModelCtrler) Update(ctx *gin.Context) {

	var (
		model    = zgwit_model.Model{}
		model_db = zgwit_model.Model{}
	)

	if !dto_info(ctx, &model) {
		zgwit_plugin.HttpFailure(ctx, "参数格式错误", zgwit_plugin.REQUEST_QUERY_ERR, nil)
		return
	}

	if err := zgwit_model.ModelFind(model.Id, &model_db); err != nil {
		zgwit_plugin.HttpFailure(ctx, "机型不存在", zgwit_plugin.REQUEST_FAIL, err)
		return
	}

	model.DriveConfig = model_db.DriveConfig

	if err := zgwit_utils.WriteObjectToFile(model.Id, &model); err != nil {
		zgwit_plugin.HttpFailure(ctx, "请求失败，请稍后重试", zgwit_plugin.REQUEST_SERVER_ERR, err)
		return
	}

	zgwit_plugin.HttpSuccess(ctx, "成功", nil)
}

func (ctrler *ModelCtrler) Delete(ctx *gin.Context) {

	var (
		model_id = ctx.Query("id")
	)

	if model_id == "" {
		zgwit_plugin.HttpFailure(ctx, "参数格式错误", zgwit_plugin.REQUEST_QUERY_ERR, nil)
		return
	}

	if err := zgwit_model.DeleteModelAttributeRealtime(ctrler.Influx, model_id); err != nil {
		zgwit_plugin.HttpFailure(ctx, "请求失败，请稍后重试", zgwit_plugin.REQUEST_SERVER_ERR, err)
		return
	}

	if err := zgwit_model.DeleteModelAttributeHistory(ctrler.Influx, model_id); err != nil {
		zgwit_plugin.HttpFailure(ctx, "请求失败，请稍后重试", zgwit_plugin.REQUEST_SERVER_ERR, err)
		return
	}

	if err := zgwit_model.DeleteModelEventRealtime(ctrler.Influx, model_id); err != nil {
		zgwit_plugin.HttpFailure(ctx, "请求失败，请稍后重试", zgwit_plugin.REQUEST_SERVER_ERR, err)
		return
	}

	if err := zgwit_model.DeleteModelEventHistory(ctrler.Influx, model_id); err != nil {
		zgwit_plugin.HttpFailure(ctx, "请求失败，请稍后重试", zgwit_plugin.REQUEST_SERVER_ERR, err)
		return
	}

	if devices, err := zgwit_model.DeviceFetchList(&model_id); err == nil {
		if len(devices) > 0 {
			zgwit_plugin.HttpFailure(ctx, "存在使用此机型的设备", zgwit_plugin.REQUEST_FAIL, len(devices))
			return
		}

	} else {
		zgwit_plugin.HttpFailure(ctx, "请求失败，请稍后重试", zgwit_plugin.REQUEST_SERVER_ERR, err)
		return
	}

	zgwit_plugin.HttpSuccess(ctx, "成功", nil)
}

func (ctrler *ModelCtrler) Config(ctx *gin.Context) {

	var (
		model_id     = ctx.Query("id")
		drive_config any
		model_db     = zgwit_model.Model{}
	)

	if err := ctx.Bind(&drive_config); err != nil || model_id == "" {
		zgwit_plugin.HttpFailure(ctx, "参数格式错误", zgwit_plugin.REQUEST_QUERY_ERR, nil)
		return
	}

	if err := zgwit_model.ModelFind(model_id, &model_db); err != nil {
		zgwit_plugin.HttpFailure(ctx, "机型不存在", zgwit_plugin.REQUEST_FAIL, err)
		return
	}

	model_db.DriveConfig = drive_config

	if err := zgwit_utils.WriteObjectToFile(model_id, &model_db); err != nil {
		zgwit_plugin.HttpFailure(ctx, "请求失败，请稍后重试", zgwit_plugin.REQUEST_SERVER_ERR, err)
		return
	}

	zgwit_plugin.HttpSuccess(ctx, "成功", nil)
}
