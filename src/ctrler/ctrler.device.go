package zgwit_ctrler

import (
	"encoding/json"
	zgwit_model "local/model"
	zgwit_plugin "local/plugin"

	"github.com/gin-gonic/gin"
)

type DeviceCtrler struct {
	Influx    *zgwit_plugin.Influx
	NsqClient *zgwit_plugin.NsqClient
}

func (ctrler *DeviceCtrler) Init(influx *zgwit_plugin.Influx, nsq_client *zgwit_plugin.NsqClient) (err error) {

	ctrler.Influx = influx
	ctrler.NsqClient = nsq_client

	return
}

func (ctrler *DeviceCtrler) List(ctx *gin.Context) {

	var (
		err error

		model_id *string
		devices  = []zgwit_model.Device{}
	)

	if id := ctx.Query("id"); id != "" {
		model_id = &id
	}

	if devices, err = zgwit_model.DeviceFetchList(model_id); err != nil {
		zgwit_plugin.HttpFailure(ctx, "请求失败，请稍后重试", zgwit_plugin.REQUEST_SERVER_ERR, err)
		return
	}

	zgwit_plugin.HttpSuccess(ctx, "成功", devices)
}

func (ctrler *DeviceCtrler) AttributeRealtime(ctx *gin.Context) {

	var (
		model_id   = ctx.Query("model_id")
		device_ids = []string{}

		datas = zgwit_model.DeviceAttributeRealtime{}
	)

	if err := json.Unmarshal([]byte(ctx.Query("device_ids")), &device_ids); model_id == "" || err != nil {
		zgwit_plugin.HttpFailure(ctx, "参数格式错误", zgwit_plugin.REQUEST_QUERY_ERR, err)
		return
	}

	if err := datas.Read(ctrler.Influx, model_id, device_ids); err != nil {
		zgwit_plugin.HttpFailure(ctx, "请求失败，请稍后重试", zgwit_plugin.REQUEST_SERVER_ERR, err)
		return
	}

	zgwit_plugin.HttpSuccess(ctx, "成功", datas)
}

func (ctrler *DeviceCtrler) AttributeHistory(ctx *gin.Context) {

	var (
		model_id      = ctx.Query("model_id")
		device_ids    = []string{}
		attribute_ids = []string{}
		page          = zgwit_model.Page{}

		datas = zgwit_model.DeviceAttributeHistory{}
	)

	if err := json.Unmarshal([]byte(ctx.Query("device_ids")), &device_ids); model_id == "" || err != nil {
		zgwit_plugin.HttpFailure(ctx, "参数格式错误", zgwit_plugin.REQUEST_QUERY_ERR, err)
		return
	}

	if err := json.Unmarshal([]byte(ctx.Query("attribute_ids")), &attribute_ids); model_id == "" || err != nil {
		zgwit_plugin.HttpFailure(ctx, "参数格式错误", zgwit_plugin.REQUEST_QUERY_ERR, err)
		return
	}

	if err := ctx.BindQuery(&page); err != nil || page.Size <= 0 || page.Offset <= 0 {
		zgwit_plugin.HttpFailure(ctx, "参数格式错误", zgwit_plugin.REQUEST_QUERY_ERR, err)
		return
	}

	if err := datas.Read(ctrler.Influx, model_id, device_ids, attribute_ids, page); err != nil {
		zgwit_plugin.HttpFailure(ctx, "请求失败，请稍后重试", zgwit_plugin.REQUEST_SERVER_ERR, err)
		return
	}

	zgwit_plugin.HttpSuccess(ctx, "成功", datas)
}

func (ctrler *DeviceCtrler) EventRealtime(ctx *gin.Context) {

	var (
		model_id   = ctx.Query("model_id")
		device_ids = []string{}

		datas = zgwit_model.DeviceEventRealtime{}
	)

	if err := json.Unmarshal([]byte(ctx.Query("device_ids")), &device_ids); model_id == "" || err != nil {
		zgwit_plugin.HttpFailure(ctx, "参数格式错误", zgwit_plugin.REQUEST_QUERY_ERR, err)
		return
	}

	if err := datas.Read(ctrler.Influx, model_id, device_ids); err != nil {
		zgwit_plugin.HttpFailure(ctx, "请求失败，请稍后重试", zgwit_plugin.REQUEST_SERVER_ERR, err)
		return
	}

	zgwit_plugin.HttpSuccess(ctx, "成功", datas)
}

func (ctrler *DeviceCtrler) EventHistory(ctx *gin.Context) {

	var (
		model_id   = ctx.Query("model_id")
		device_ids = []string{}
		event_ids  = []string{}
		page       = zgwit_model.Page{}

		datas = zgwit_model.DeviceEventHistory{}
	)

	if err := json.Unmarshal([]byte(ctx.Query("event_ids")), &event_ids); model_id == "" || err != nil {
		zgwit_plugin.HttpFailure(ctx, "参数格式错误", zgwit_plugin.REQUEST_QUERY_ERR, err)
		return
	}

	if err := json.Unmarshal([]byte(ctx.Query("event_ids")), &event_ids); model_id == "" || err != nil {
		zgwit_plugin.HttpFailure(ctx, "参数格式错误", zgwit_plugin.REQUEST_QUERY_ERR, err)
		return
	}

	if err := ctx.BindQuery(&page); err != nil || page.Size <= 0 || page.Offset <= 0 {
		zgwit_plugin.HttpFailure(ctx, "参数格式错误", zgwit_plugin.REQUEST_QUERY_ERR, err)
		return
	}

	if err := datas.Read(ctrler.Influx, model_id, device_ids, event_ids, page); err != nil {
		zgwit_plugin.HttpFailure(ctx, "请求失败，请稍后重试", zgwit_plugin.REQUEST_SERVER_ERR, err)
		return
	}

	zgwit_plugin.HttpSuccess(ctx, "成功", datas)
}
