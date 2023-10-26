package ctrler

import (
	"encoding/json"
	"strings"

	"github.com/zgwit/iot-master-next/src/model"
	"github.com/zgwit/iot-master-next/src/plugin"
	"github.com/zgwit/iot-master-next/src/utils"

	"github.com/gin-gonic/gin"
)

type DeviceCtrler struct {
	Influx    *plugin.Influx
	NsqClient *plugin.NsqClient
}

func (ctrler *DeviceCtrler) Init(influx *plugin.Influx, nsq_client *plugin.NsqClient) (err error) {

	ctrler.Influx = influx
	ctrler.NsqClient = nsq_client

	return
}

func device_list(model_id *string, labels map[string]string) (devices []model.DeviceType, err error) {

	var (
		device_ids []string
	)

	devices = []model.DeviceType{}

	if device_ids, err = utils.GetDirFileNames2("./config/device"); err != nil {
		return
	}

	for _, device_id := range device_ids {

		device_id := strings.Replace(device_id, ".txt", "", -1)

		device := model.DeviceType{}

		if err = utils.ReadFileToObject("./config/device/"+device_id+".txt", &device); err != nil {
			continue
		}

		if device.ModelId != *model_id {
			continue
		}

		for key := range labels {
			if labels[key] != device.Labels[key] {
				goto NEXT
			}
		}

		devices = append(devices, device)
	NEXT:
	}

	return
}

func (ctrler *DeviceCtrler) List(ctx *gin.Context) {

	var (
		err error

		model_id *string
		labels   map[string]string

		devices = []model.DeviceType{}
	)

	if id := ctx.Query("id"); id != "" {
		model_id = &id
	}

	if err := json.Unmarshal([]byte(ctx.Query("labels")), &labels); err != nil {
		plugin.HttpFailure(ctx, "参数格式错误", plugin.REQUEST_QUERY_ERR, err)
		return
	}

	if devices, err = device_list(model_id, labels); err != nil {
		plugin.HttpFailure(ctx, "请求失败，请稍后重试", plugin.REQUEST_SERVER_ERR, err)
		return
	}

	plugin.HttpSuccess(ctx, "成功", devices)
}

func device_table(ctx *gin.Context, table *model.DeviceType) (result bool) {

	if err := ctx.Bind(table); err != nil {
		return
	}

	if len(table.Labels) == 0 {
		table.Labels = map[string]string{}
	}

	if table.Name == "" || table.Id == "" || table.ModelId == "" || table.Drive == "" {
		return
	}

	return true
}

func device_exist(device_id string) (result bool) {

	return utils.FileExist("./config/device/" + device_id + ".txt")
}

func device_write(table *model.DeviceType) (err error) {

	return utils.WriteObjectToFile2("./config/device/"+table.Id+".txt", table)
}

func (ctrler *DeviceCtrler) Create(ctx *gin.Context) {

	var (
		err   error
		table = model.DeviceType{}
	)

	if !device_table(ctx, &table) {
		plugin.HttpFailure(ctx, "参数格式错误", plugin.REQUEST_QUERY_ERR, err)
		return
	}

	if device_exist(table.Id) {
		plugin.HttpFailure(ctx, "设备编号已存在", plugin.REQUEST_FAIL, err)
		return
	}

	table.DriveConfig = nil

	if err := device_write(&table); err != nil {
		plugin.HttpFailure(ctx, "请求失败，请稍后重试", plugin.REQUEST_SERVER_ERR, err)
		return
	}

	plugin.HttpSuccess(ctx, "成功", nil)
}

func device_delete(device_id string) (err error) {

	return utils.RemoveFile("./config/device/" + device_id + ".txt")
}

func (ctrler *DeviceCtrler) Delete(ctx *gin.Context) {

	device_id := ctx.Query("device_ids")

	if device_id == "" {
		plugin.HttpFailure(ctx, "参数格式错误", plugin.REQUEST_QUERY_ERR, "编号为空")
		return
	}

	if err := device_delete(device_id); err != nil {
		plugin.HttpFailure(ctx, "请求失败，请稍后重试", plugin.REQUEST_SERVER_ERR, err)
		return
	}

	plugin.HttpSuccess(ctx, "成功", nil)
}

func device_find(model_id string, table *model.DeviceType) (err error) {

	return utils.ReadFileToObject("./config/device/"+model_id+".txt", table)
}

func (ctrler *DeviceCtrler) Update(ctx *gin.Context) {

	var (
		table    = model.DeviceType{}
		table_db = model.DeviceType{}
	)

	if !device_table(ctx, &table) {
		plugin.HttpFailure(ctx, "参数格式错误", plugin.REQUEST_QUERY_ERR, nil)
		return
	}

	if err := device_find(table.Id, &table_db); err != nil {
		plugin.HttpFailure(ctx, "设备不存在", plugin.REQUEST_FAIL, err)
		return
	}

	table.Drive = table_db.Drive
	table.DriveConfig = table_db.DriveConfig

	if err := device_write(&table); err != nil {
		plugin.HttpFailure(ctx, "请求失败，请稍后重试", plugin.REQUEST_SERVER_ERR, err)
		return
	}

	plugin.HttpSuccess(ctx, "成功", nil)
}

func (ctrler *DeviceCtrler) Config(ctx *gin.Context) {

	var (
		model_id = ctx.Query("id")
		table    any
		table_db = model.DeviceType{}
	)

	if err := ctx.Bind(&table); err != nil || model_id == "" {
		plugin.HttpFailure(ctx, "参数格式错误", plugin.REQUEST_QUERY_ERR, nil)
		return
	}

	if err := device_find(model_id, &table_db); err != nil {
		plugin.HttpFailure(ctx, "机型不存在", plugin.REQUEST_FAIL, err)
		return
	}

	table_db.DriveConfig = table

	if err := device_write(&table_db); err != nil {
		plugin.HttpFailure(ctx, "请求失败，请稍后重试", plugin.REQUEST_SERVER_ERR, err)
		return
	}

	plugin.HttpSuccess(ctx, "成功", nil)
}

func (ctrler *DeviceCtrler) AttributeRealtime(ctx *gin.Context) {

	var (
		model_id   = ctx.Query("model_id")
		device_ids = []string{}

		datas = model.DeviceAttributeRealtimeType{}
	)

	if err := json.Unmarshal([]byte(ctx.Query("device_ids")), &device_ids); model_id == "" || err != nil {
		plugin.HttpFailure(ctx, "参数格式错误", plugin.REQUEST_QUERY_ERR, err)
		return
	}

	if err := datas.Read(ctrler.Influx, model_id, device_ids); err != nil {
		plugin.HttpFailure(ctx, "请求失败，请稍后重试", plugin.REQUEST_SERVER_ERR, err)
		return
	}

	plugin.HttpSuccess(ctx, "成功", datas)
}

func (ctrler *DeviceCtrler) AttributeHistory(ctx *gin.Context) {

	var (
		model_id      = ctx.Query("model_id")
		device_ids    = []string{}
		attribute_ids = []string{}
		page          = model.PageType{}

		datas = model.DeviceAttributeHistoryType{}
	)

	if err := json.Unmarshal([]byte(ctx.Query("device_ids")), &device_ids); model_id == "" || err != nil {
		plugin.HttpFailure(ctx, "参数格式错误", plugin.REQUEST_QUERY_ERR, err)
		return
	}

	if err := json.Unmarshal([]byte(ctx.Query("attribute_ids")), &attribute_ids); model_id == "" || err != nil {
		plugin.HttpFailure(ctx, "参数格式错误", plugin.REQUEST_QUERY_ERR, err)
		return
	}

	if err := ctx.BindQuery(&page); err != nil || page.Size <= 0 || page.Offset <= 0 {
		plugin.HttpFailure(ctx, "参数格式错误", plugin.REQUEST_QUERY_ERR, err)
		return
	}

	if err := datas.Read(ctrler.Influx, model_id, device_ids, attribute_ids, page); err != nil {
		plugin.HttpFailure(ctx, "请求失败，请稍后重试", plugin.REQUEST_SERVER_ERR, err)
		return
	}

	plugin.HttpSuccess(ctx, "成功", datas)
}

func (ctrler *DeviceCtrler) EventRealtime(ctx *gin.Context) {

	var (
		model_id   = ctx.Query("model_id")
		device_ids = []string{}

		datas = model.DeviceEventRealtimeType{}
	)

	if err := json.Unmarshal([]byte(ctx.Query("device_ids")), &device_ids); model_id == "" || err != nil {
		plugin.HttpFailure(ctx, "参数格式错误", plugin.REQUEST_QUERY_ERR, err)
		return
	}

	if err := datas.Read(ctrler.Influx, model_id, device_ids); err != nil {
		plugin.HttpFailure(ctx, "请求失败，请稍后重试", plugin.REQUEST_SERVER_ERR, err)
		return
	}

	plugin.HttpSuccess(ctx, "成功", datas)
}

func (ctrler *DeviceCtrler) EventHistory(ctx *gin.Context) {

	var (
		model_id   = ctx.Query("model_id")
		device_ids = []string{}
		event_ids  = []string{}
		page       = model.PageType{}

		datas = model.DeviceEventHistoryType{}
	)

	if err := json.Unmarshal([]byte(ctx.Query("event_ids")), &event_ids); model_id == "" || err != nil {
		plugin.HttpFailure(ctx, "参数格式错误", plugin.REQUEST_QUERY_ERR, err)
		return
	}

	if err := json.Unmarshal([]byte(ctx.Query("event_ids")), &event_ids); model_id == "" || err != nil {
		plugin.HttpFailure(ctx, "参数格式错误", plugin.REQUEST_QUERY_ERR, err)
		return
	}

	if err := ctx.BindQuery(&page); err != nil || page.Size <= 0 || page.Offset <= 0 {
		plugin.HttpFailure(ctx, "参数格式错误", plugin.REQUEST_QUERY_ERR, err)
		return
	}

	if err := datas.Read(ctrler.Influx, model_id, device_ids, event_ids, page); err != nil {
		plugin.HttpFailure(ctx, "请求失败，请稍后重试", plugin.REQUEST_SERVER_ERR, err)
		return
	}

	plugin.HttpSuccess(ctx, "成功", datas)
}
