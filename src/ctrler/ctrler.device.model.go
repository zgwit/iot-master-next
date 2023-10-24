package ctrler

import (
	"github.com/zgwit/iot-master-next/src/model"
	"github.com/zgwit/iot-master-next/src/plugin"

	"github.com/zgwit/iot-master-next/src/utils"

	"github.com/gin-gonic/gin"
)

type DeviceModelCtrler struct {
	Influx    *plugin.Influx
	NsqClient *plugin.NsqClient
}

func (ctrler *DeviceModelCtrler) Init(influx *plugin.Influx, nsq_client *plugin.NsqClient) (err error) {

	ctrler.Influx = influx
	ctrler.NsqClient = nsq_client

	return
}

func dto_info(ctx *gin.Context, device_model *model.DeviceModelType) (result bool) {

	if err := ctx.Bind(device_model); err != nil {
		return
	}

	if len(device_model.Attributes) == 0 {
		device_model.Attributes = map[string]model.DeviceModelAttributeType{}
	}

	if len(device_model.Events) == 0 {
		device_model.Events = map[string]model.DeviceModelEventType{}
	}

	if len(device_model.Actions) == 0 {
		device_model.Actions = map[string]model.DeviceModelActionType{}
	}

	if device_model.Name == "" || device_model.Id == "" || device_model.Drive == "" {
		return
	}

	for _, item := range device_model.Attributes {
		if item.Name == "" {
			return
		}
		if item.DataType != "number" && item.DataType != "string" {
			return
		}
	}

	for _, item := range device_model.Events {
		if item.Name == "" {
			return
		}
	}

	for _, item := range device_model.Actions {
		if item.Name == "" {
			return
		}
	}

	return true
}

func (ctrler *DeviceModelCtrler) List(ctx *gin.Context) {

	var (
		err error

		device_models = []model.DeviceModelType{}
	)

	if device_models, err = model.ModelFetchList(); err != nil {
		plugin.HttpFailure(ctx, "请求失败，请稍后重试", plugin.REQUEST_SERVER_ERR, err)
		return
	}

	plugin.HttpSuccess(ctx, "成功", device_models)
}

func (ctrler *DeviceModelCtrler) Create(ctx *gin.Context) {

	var (
		err          error
		device_model = model.DeviceModelType{}
	)

	if !dto_info(ctx, &device_model) {
		plugin.HttpFailure(ctx, "参数格式错误", plugin.REQUEST_QUERY_ERR, err)
		return
	}

	if model.ModelExist(device_model.Id) {
		plugin.HttpFailure(ctx, "机型编号已存在", plugin.REQUEST_FAIL, err)
		return
	}

	device_model.DriveConfig = nil

	if err := device_model.Create(); err != nil {
		plugin.HttpFailure(ctx, "请求失败，请稍后重试", plugin.REQUEST_SERVER_ERR, err)
		return
	}

	plugin.HttpSuccess(ctx, "成功", nil)
}

func (ctrler *DeviceModelCtrler) Update(ctx *gin.Context) {

	var (
		device_model    = model.DeviceModelType{}
		device_model_db = model.DeviceModelType{}
	)

	if !dto_info(ctx, &device_model) {
		plugin.HttpFailure(ctx, "参数格式错误", plugin.REQUEST_QUERY_ERR, nil)
		return
	}

	if err := model.ModelFind(device_model.Id, &device_model_db); err != nil {
		plugin.HttpFailure(ctx, "机型不存在", plugin.REQUEST_FAIL, err)
		return
	}

	device_model.DriveConfig = device_model_db.DriveConfig

	if err := utils.WriteObjectToFile(device_model.Id, &device_model); err != nil {
		plugin.HttpFailure(ctx, "请求失败，请稍后重试", plugin.REQUEST_SERVER_ERR, err)
		return
	}

	plugin.HttpSuccess(ctx, "成功", nil)
}

func (ctrler *DeviceModelCtrler) Delete(ctx *gin.Context) {

	var (
		model_id = ctx.Query("id")
	)

	if model_id == "" {
		plugin.HttpFailure(ctx, "参数格式错误", plugin.REQUEST_QUERY_ERR, nil)
		return
	}

	if err := model.DeleteModelAttributeRealtime(ctrler.Influx, model_id); err != nil {
		plugin.HttpFailure(ctx, "请求失败，请稍后重试", plugin.REQUEST_SERVER_ERR, err)
		return
	}

	if err := model.DeleteModelAttributeHistory(ctrler.Influx, model_id); err != nil {
		plugin.HttpFailure(ctx, "请求失败，请稍后重试", plugin.REQUEST_SERVER_ERR, err)
		return
	}

	if err := model.DeleteModelEventRealtime(ctrler.Influx, model_id); err != nil {
		plugin.HttpFailure(ctx, "请求失败，请稍后重试", plugin.REQUEST_SERVER_ERR, err)
		return
	}

	if err := model.DeleteModelEventHistory(ctrler.Influx, model_id); err != nil {
		plugin.HttpFailure(ctx, "请求失败，请稍后重试", plugin.REQUEST_SERVER_ERR, err)
		return
	}

	if devices, err := model.DeviceFetchList(&model_id); err == nil {
		if len(devices) > 0 {
			plugin.HttpFailure(ctx, "存在使用此机型的设备", plugin.REQUEST_FAIL, len(devices))
			return
		}

	} else {
		plugin.HttpFailure(ctx, "请求失败，请稍后重试", plugin.REQUEST_SERVER_ERR, err)
		return
	}

	plugin.HttpSuccess(ctx, "成功", nil)
}

func (ctrler *DeviceModelCtrler) Config(ctx *gin.Context) {

	var (
		model_id        = ctx.Query("id")
		drive_config    any
		device_model_db = model.DeviceModelType{}
	)

	if err := ctx.Bind(&drive_config); err != nil || model_id == "" {
		plugin.HttpFailure(ctx, "参数格式错误", plugin.REQUEST_QUERY_ERR, nil)
		return
	}

	if err := model.ModelFind(model_id, &device_model_db); err != nil {
		plugin.HttpFailure(ctx, "机型不存在", plugin.REQUEST_FAIL, err)
		return
	}

	device_model_db.DriveConfig = drive_config

	if err := utils.WriteObjectToFile(model_id, &device_model_db); err != nil {
		plugin.HttpFailure(ctx, "请求失败，请稍后重试", plugin.REQUEST_SERVER_ERR, err)
		return
	}

	plugin.HttpSuccess(ctx, "成功", nil)
}
