package ctrler

import (
	"strings"

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

func device_model_list() (models []model.DeviceModelType, err error) {

	var (
		device_model_ids []string
	)

	models = []model.DeviceModelType{}

	if device_model_ids, err = utils.GetDirFileNames2("./config/device.model"); err != nil {
		return
	}

	for _, device_model_id := range device_model_ids {

		model := model.DeviceModelType{Id: strings.Replace(device_model_id, ".txt", "", -1)}

		if err = utils.ReadFileToObject("./config/device.model/"+device_model_id+".txt", &model); err != nil {
			continue
		}

		model.Id = device_model_id
		models = append(models, model)
	}

	return
}

func (ctrler *DeviceModelCtrler) List(ctx *gin.Context) {

	var (
		err error

		device_models = []model.DeviceModelType{}
	)

	if device_models, err = device_model_list(); err != nil {
		plugin.HttpFailure(ctx, "请求失败，请稍后重试", plugin.REQUEST_SERVER_ERR, err)
		return
	}

	plugin.HttpSuccess(ctx, "成功", device_models)
}

func device_model_table(ctx *gin.Context, table *model.DeviceModelType) (result bool) {

	if err := ctx.Bind(table); err != nil {
		return
	}

	if len(table.Attributes) == 0 {
		table.Attributes = map[string]model.DeviceModelAttributeType{}
	}

	if len(table.Events) == 0 {
		table.Events = map[string]model.DeviceModelEventType{}
	}

	if len(table.Actions) == 0 {
		table.Actions = map[string]model.DeviceModelActionType{}
	}

	if table.Name == "" || table.Id == "" || table.Drive == "" {
		return
	}

	for _, item := range table.Attributes {
		if item.Name == "" {
			return
		}
		if item.DataType != "number" && item.DataType != "string" {
			return
		}
	}

	for _, item := range table.Events {
		if item.Name == "" {
			return
		}
	}

	for _, item := range table.Actions {
		if item.Name == "" {
			return
		}
	}

	return true
}

func device_model_exist(model_id string) (result bool) {

	return utils.FileExist("./config/device.model/" + model_id + ".txt")
}

func device_model_write(device_model *model.DeviceModelType) (err error) {

	return utils.WriteObjectToFile2("./config/device.model/"+device_model.Id+".txt", device_model)
}

func (ctrler *DeviceModelCtrler) Create(ctx *gin.Context) {

	var (
		err   error
		table = model.DeviceModelType{}
	)

	if !device_model_table(ctx, &table) {
		plugin.HttpFailure(ctx, "参数格式错误", plugin.REQUEST_QUERY_ERR, err)
		return
	}

	if device_model_exist(table.Id) {
		plugin.HttpFailure(ctx, "机型编号已存在", plugin.REQUEST_FAIL, err)
		return
	}

	table.DriveConfig = nil

	if err := device_model_write(&table); err != nil {
		plugin.HttpFailure(ctx, "请求失败，请稍后重试", plugin.REQUEST_SERVER_ERR, err)
		return
	}

	plugin.HttpSuccess(ctx, "成功", nil)
}

func device_model_find(model_id string, table *model.DeviceModelType) (err error) {

	return utils.ReadFileToObject("./config/device.model/"+model_id+".txt", table)
}

func (ctrler *DeviceModelCtrler) Find(ctx *gin.Context) {

	device_id := ctx.Query("id")
	table := model.DeviceModelType{}

	if device_id == "" {
		plugin.HttpFailure(ctx, "参数格式错误", plugin.REQUEST_QUERY_ERR, "编号为空")
		return
	}

	if err := device_model_find(device_id, &table); err != nil {
		plugin.HttpFailure(ctx, "请求失败，请稍后重试", plugin.REQUEST_SERVER_ERR, err)
		return
	}

	plugin.HttpSuccess(ctx, "成功", table)
}

func (ctrler *DeviceModelCtrler) Update(ctx *gin.Context) {

	var (
		table    = model.DeviceModelType{}
		table_db = model.DeviceModelType{}
	)

	if !device_model_table(ctx, &table) {
		plugin.HttpFailure(ctx, "参数格式错误", plugin.REQUEST_QUERY_ERR, nil)
		return
	}

	if err := device_model_find(table.Id, &table_db); err != nil {
		plugin.HttpFailure(ctx, "机型不存在", plugin.REQUEST_FAIL, err)
		return
	}

	table.Drive = table_db.Drive
	table.DriveConfig = table_db.DriveConfig

	if err := device_model_write(&table); err != nil {
		plugin.HttpFailure(ctx, "请求失败，请稍后重试", plugin.REQUEST_SERVER_ERR, err)
		return
	}

	plugin.HttpSuccess(ctx, "成功", nil)
}

func device_model_delete(model_id string) (err error) {

	return utils.RemoveFile("./config/device.model/" + model_id + ".txt")
}

func (ctrler *DeviceModelCtrler) Delete(ctx *gin.Context) {

	model_id := ctx.Query("id")

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

	if devices, err := device_list(&model_id, map[string]string{}); err != nil {
		plugin.HttpFailure(ctx, "请求失败，请稍后重试", plugin.REQUEST_SERVER_ERR, err)
		return

	} else if len(devices) > 0 {
		plugin.HttpFailure(ctx, "存在使用此机型的设备", plugin.REQUEST_FAIL, len(devices))
		return
	}

	if err := device_model_delete(model_id); err != nil {
		plugin.HttpFailure(ctx, "请求失败，请稍后重试", plugin.REQUEST_SERVER_ERR, err)
		return
	}

	plugin.HttpSuccess(ctx, "成功", nil)
}

func (ctrler *DeviceModelCtrler) Config(ctx *gin.Context) {

	var (
		model_id = ctx.Query("id")
		table    any
		table_db = model.DeviceModelType{}
	)

	if err := ctx.Bind(&table); err != nil || model_id == "" {
		plugin.HttpFailure(ctx, "参数格式错误", plugin.REQUEST_QUERY_ERR, nil)
		return
	}

	if err := device_model_find(model_id, &table_db); err != nil {
		plugin.HttpFailure(ctx, "机型不存在", plugin.REQUEST_FAIL, err)
		return
	}

	table_db.DriveConfig = table

	if err := device_model_write(&table_db); err != nil {
		plugin.HttpFailure(ctx, "请求失败，请稍后重试", plugin.REQUEST_SERVER_ERR, err)
		return
	}

	plugin.HttpSuccess(ctx, "成功", nil)
}
