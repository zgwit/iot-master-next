package ctrler

import (
	"encoding/json"
	"sort"
	"strings"
	"time"

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

func device_list(model_id *string, labels []model.NameValueType) (devices []model.DeviceType, err error) {

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

		if model_id != nil && device.ModelId != *model_id {
			continue
		}

		for _, label_param := range labels {
			for _, label_device := range device.Labels {

				if label_param.Name == label_device.Name && label_param.Value != label_device.Value {
					goto NEXT
				}
			}
		}

		device.Id = device_id
		devices = append(devices, device)
	NEXT:
	}

	return
}

func (ctrler *DeviceCtrler) List(ctx *gin.Context) {

	var (
		err error

		model_id *string
		labels   []model.NameValueType

		devices = []model.DeviceType{}
	)

	if id := ctx.Query("model_id"); id != "" {
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

	sort.Sort(model.DevciceByCreateTime(devices))

	plugin.HttpSuccess(ctx, "成功", devices)
}

func device_table(ctx *gin.Context, table *model.DeviceType) (result bool) {

	if err := ctx.Bind(table); err != nil {
		return
	}

	if len(table.Labels) == 0 {
		table.Labels = []model.NameValueType{}
	}

	if table.Name == "" || table.Id == "" || table.ModelId == "" {
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

	table.CreateTime = time.Now().Unix()
	table.Config = nil

	if err := device_write(&table); err != nil {
		plugin.HttpFailure(ctx, "请求失败，请稍后重试", plugin.REQUEST_SERVER_ERR, err)
		return
	}

	plugin.HttpSuccess(ctx, "成功", nil)
}

func device_find(id string, table *model.DeviceType) (err error) {

	return utils.ReadFileToObject("./config/device/"+id+".txt", table)
}

func (ctrler *DeviceCtrler) Find(ctx *gin.Context) {

	device_id := ctx.Query("id")
	table := model.DeviceType{}

	if device_id == "" {
		plugin.HttpFailure(ctx, "参数格式错误", plugin.REQUEST_QUERY_ERR, "编号为空")
		return
	}

	if err := device_find(device_id, &table); err != nil {
		plugin.HttpFailure(ctx, "设备不存在", plugin.REQUEST_SERVER_ERR, err)
		return
	}

	plugin.HttpSuccess(ctx, "成功", table)
}

func device_delete(device_id string) (err error) {

	return utils.RemoveFile("./config/device/" + device_id + ".txt")
}

func (ctrler *DeviceCtrler) Delete(ctx *gin.Context) {

	device_id := ctx.Query("id")

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

	table.CreateTime = table_db.CreateTime
	table_db.Name = table.Name
	table_db.Labels = table.Labels

	if err := device_write(&table_db); err != nil {
		plugin.HttpFailure(ctx, "请求失败，请稍后重试", plugin.REQUEST_SERVER_ERR, err)
		return
	}

	plugin.HttpSuccess(ctx, "成功", nil)
}

func (ctrler *DeviceCtrler) Config(ctx *gin.Context) {

	var (
		id       = ctx.Query("id")
		table    any
		table_db = model.DeviceType{}
	)

	if err := ctx.Bind(&table); err != nil || id == "" {
		plugin.HttpFailure(ctx, "参数格式错误", plugin.REQUEST_QUERY_ERR, nil)
		return
	}

	if err := device_find(id, &table_db); err != nil {
		plugin.HttpFailure(ctx, "机型不存在", plugin.REQUEST_FAIL, err)
		return
	}

	table_db.Config = table

	if err := device_write(&table_db); err != nil {
		plugin.HttpFailure(ctx, "请求失败，请稍后重试", plugin.REQUEST_SERVER_ERR, err)
		return
	}

	plugin.HttpSuccess(ctx, "成功", nil)
}

func (ctrler *DeviceCtrler) AttributeRealtime(ctx *gin.Context) {

	var (
		filter = model.DataFilterType{}
		datas  = model.DeviceAttributeRealtimeType{}
	)

	if err := json.Unmarshal([]byte(ctx.Query("filter")), &filter); err != nil {
		plugin.HttpFailure(ctx, "参数格式错误", plugin.REQUEST_QUERY_ERR, err)
		return
	}

	if err := datas.Read(ctrler.Influx, filter); err != nil {
		plugin.HttpFailure(ctx, "请求失败，请稍后重试", plugin.REQUEST_SERVER_ERR, err)
		return
	}

	plugin.HttpSuccess(ctx, "成功", datas)
}

func (ctrler *DeviceCtrler) AttributeHistory(ctx *gin.Context) {

	var (
		filter = model.DataFilterType{}
		page   = model.PageType{}

		datas = model.DeviceAttributeHistoryType{}
	)

	if err := json.Unmarshal([]byte(ctx.Query("filter")), &filter); err != nil {
		plugin.HttpFailure(ctx, "参数格式错误", plugin.REQUEST_QUERY_ERR, err)
		return
	}

	if err := ctx.BindQuery(&page); err != nil || page.Size <= 0 || page.Offset <= 0 {
		plugin.HttpFailure(ctx, "参数格式错误", plugin.REQUEST_QUERY_ERR, err)
		return
	}

	if err := datas.Read(ctrler.Influx, filter, page); err != nil {
		plugin.HttpFailure(ctx, "请求失败，请稍后重试", plugin.REQUEST_SERVER_ERR, err)
		return
	}

	plugin.HttpSuccess(ctx, "成功", datas)
}

func (ctrler *DeviceCtrler) AttributeRealtimeDelete(ctx *gin.Context) {

	model_id, device_id := ctx.Query("model_id"), ctx.Query("device_id")

	if model_id == "" || device_id == "" {
		plugin.HttpFailure(ctx, "参数格式错误", plugin.REQUEST_QUERY_ERR, nil)
		return
	}

	if err := model.DeleteDeviceAttributeRealtime(ctrler.Influx, model_id, device_id); err != nil {
		plugin.HttpFailure(ctx, "请求失败，请稍后重试", plugin.REQUEST_SERVER_ERR, err)
		return
	}

	plugin.HttpSuccess(ctx, "成功", nil)
}

func (ctrler *DeviceCtrler) AttributeHistoryDelete(ctx *gin.Context) {

	model_id, device_id := ctx.Query("model_id"), ctx.Query("device_id")

	if model_id == "" || device_id == "" {
		plugin.HttpFailure(ctx, "参数格式错误", plugin.REQUEST_QUERY_ERR, nil)
		return
	}

	if err := model.DeleteDeviceAttributeHistory(ctrler.Influx, model_id, device_id); err != nil {
		plugin.HttpFailure(ctx, "请求失败，请稍后重试", plugin.REQUEST_SERVER_ERR, err)
		return
	}

	plugin.HttpSuccess(ctx, "成功", nil)
}

func (ctrler *DeviceCtrler) EventRealtime(ctx *gin.Context) {

	var (
		filter = model.DataFilterType{}
		datas  = model.DeviceEventRealtimeType{}
	)

	if err := json.Unmarshal([]byte(ctx.Query("filter")), &filter); err != nil {
		plugin.HttpFailure(ctx, "参数格式错误", plugin.REQUEST_QUERY_ERR, err)
		return
	}

	if err := datas.Read(ctrler.Influx, filter); err != nil {
		plugin.HttpFailure(ctx, "请求失败，请稍后重试", plugin.REQUEST_SERVER_ERR, err)
		return
	}

	plugin.HttpSuccess(ctx, "成功", datas)
}

func (ctrler *DeviceCtrler) EventHistory(ctx *gin.Context) {

	var (
		filter = model.DataFilterType{}
		page   = model.PageType{}

		datas = model.DeviceEventHistoryType{}
	)

	if err := json.Unmarshal([]byte(ctx.Query("filter")), &filter); err != nil {
		plugin.HttpFailure(ctx, "参数格式错误", plugin.REQUEST_QUERY_ERR, err)
		return
	}

	if err := ctx.BindQuery(&page); err != nil {
		plugin.HttpFailure(ctx, "参数格式错误", plugin.REQUEST_QUERY_ERR, err)
		return
	}

	if err := datas.Read(ctrler.Influx, filter, page); err != nil {
		plugin.HttpFailure(ctx, "请求失败，请稍后重试", plugin.REQUEST_SERVER_ERR, err)
		return
	}

	plugin.HttpSuccess(ctx, "成功", datas)
}

func (ctrler *DeviceCtrler) EventRealtimeDelete(ctx *gin.Context) {

	model_id, device_id := ctx.Query("model_id"), ctx.Query("device_id")

	if model_id == "" || device_id == "" {
		plugin.HttpFailure(ctx, "参数格式错误", plugin.REQUEST_QUERY_ERR, nil)
		return
	}

	if err := model.DeleteDeviceEventRealtime(ctrler.Influx, model_id, device_id); err != nil {
		plugin.HttpFailure(ctx, "请求失败，请稍后重试", plugin.REQUEST_SERVER_ERR, err)
		return
	}

	plugin.HttpSuccess(ctx, "成功", nil)
}

func (ctrler *DeviceCtrler) EventHistoryDelete(ctx *gin.Context) {

	model_id, device_id := ctx.Query("model_id"), ctx.Query("device_id")

	if model_id == "" || device_id == "" {
		plugin.HttpFailure(ctx, "参数格式错误", plugin.REQUEST_QUERY_ERR, nil)
		return
	}

	if err := model.DeleteDeviceEventHistory(ctrler.Influx, model_id, device_id); err != nil {
		plugin.HttpFailure(ctx, "请求失败，请稍后重试", plugin.REQUEST_SERVER_ERR, err)
		return
	}

	plugin.HttpSuccess(ctx, "成功", nil)
}

func (ctrler *DeviceCtrler) Activetime(ctx *gin.Context) {

	device_activetime := model.DeviceActivetimeType{}

	if err := device_activetime.Read(ctrler.Influx, []string{}); err != nil {
		plugin.HttpFailure(ctx, "请求失败，请稍后重试", plugin.REQUEST_SERVER_ERR, err)
		return
	}

	plugin.HttpSuccess(ctx, "成功", device_activetime)
}
