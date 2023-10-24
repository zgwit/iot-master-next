package ctrler

import (
	"encoding/json"

	"github.com/zgwit/iot-master-next/src/model"
	"github.com/zgwit/iot-master-next/src/plugin"

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

func (ctrler *DeviceCtrler) List(ctx *gin.Context) {

	var (
		err error

		model_id *string
		devices  = []model.DeviceType{}
	)

	if id := ctx.Query("id"); id != "" {
		model_id = &id
	}

	if devices, err = model.DeviceFetchList(model_id); err != nil {
		plugin.HttpFailure(ctx, "请求失败，请稍后重试", plugin.REQUEST_SERVER_ERR, err)
		return
	}

	plugin.HttpSuccess(ctx, "成功", devices)
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
