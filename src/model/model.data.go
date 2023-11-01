package model

import (
	"encoding/json"
	"fmt"

	"github.com/zgwit/iot-master-next/src/plugin"
	"github.com/zgwit/iot-master-next/src/utils"
)

type AttributeWriteType struct {
	Time int64 `form:"time" bson:"time" json:"time"`

	ModelId  string `form:"model_id" bson:"model_id" json:"model_id"`
	DeviceId string `form:"device_id" bson:"device_id" json:"device_id"`

	Points KeyValueType `form:"points" bson:"points" json:"points"`
}

func (attribute_write *AttributeWriteType) Scripting(javascript *plugin.Javascript, script string) (attributes KeyValueType, events KeyValueBoolType, err error) {

	var (
		ok bool

		script_result plugin.JavaScriptResult
		result        KeyValueType
	)

	if script_result, err = javascript.Function("handle",
		`function handle(points) { let attributes = {}, events = {};`+script+`;return { attributes, events } }`,
		attribute_write.Points); err != nil {
		return
	}

	if result, ok = script_result.Export().(map[string]interface{}); !ok {
		err = fmt.Errorf("return format wrong")
		return
	}

	if attributes, ok = result["attributes"].(map[string]interface{}); !ok {
		err = fmt.Errorf("attributes format wrong")
		return
	}

	events = KeyValueBoolType{}

	if _events, ok := result["events"].(map[string]interface{}); ok {

		for event_id := range _events {

			if value, ok := _events[event_id].(bool); ok {
				events[event_id] = value
			}
		}

	} else {
		err = fmt.Errorf("attributes format wrong")
		return
	}

	return
}

type EventWriteType struct {
	Type string `form:"type" bson:"type" json:"type"`

	ModelId  string `form:"model_id" bson:"model_id" json:"model_id"`
	DeviceId string `form:"device_id" bson:"device_id" json:"device_id"`
	EventId  string `form:"event_id" bson:"event_id" json:"event_id"`

	BeginTime int64 `form:"begin_time" bson:"begin_time" json:"begin_time"`
	EndTime   int64 `form:"end_time" bson:"end_time" json:"end_time"`
}

// device_id.time.data_id.value
type DeviceAttributeHistoryType map[string]map[int64]KeyValueType

func DeleteDeviceAttributeHistory(influx *plugin.Influx, model_id string, device_id string) (err error) {

	return influx.DeleteByTag("attribute_history", model_id, ":device_id", device_id)
}

func DeleteModelAttributeHistory(influx *plugin.Influx, model_id string) (err error) {

	return influx.DeleteByMeasurement("attribute_history", model_id)
}

func (datas DeviceAttributeHistoryType) Write(influx *plugin.Influx, model *DeviceModelType) (err error) {

	defer utils.ErrorRecover("DeviceAttributeHistoryType")

	batch := plugin.NewInfluxBatch(influx, "attribute_history")

	for device_id, datas := range datas {

		for time, data := range datas {

			tags := map[string]string{":device_id": device_id}

			fields := KeyValueType{}

			for attribute_id, value_interface := range data {

				index, attribute := -1, DeviceModelAttributeType{}

				for idx := range model.Attributes {
					if model.Attributes[idx].Id == attribute_id {
						index = idx
						attribute = model.Attributes[idx]
						break
					}
				}

				if index == -1 {
					continue
				}

				if attribute.LastestOnly {
					continue
				}

				switch value := value_interface.(type) {
				case float64:
					if attribute.DataType == "number" {
						fields[attribute_id] = value
					}
				case int64:
					if attribute.DataType == "number" {
						fields[attribute_id] = float64(value)
					}
				case string:
					if attribute.DataType == "string" {
						fields[attribute_id] = value
					}
				case bool:
					if attribute.DataType == "number" {
						fields[attribute_id] = utils.GetNumberBool(value)
					}
				}
			}

			batch.AddPoint(model.Id, tags, fields, time)
		}
	}

	return batch.Write()
}

func (datas *DeviceAttributeHistoryType) Read(influx *plugin.Influx, model_id string, device_ids, attribute_ids []string, page PageType) (err error) {

	var (
		query_data     = []map[string]interface{}{}
		flux_device    = ""
		flux_attribute = ""
	)

	for idx, device_id := range device_ids {

		if idx > 0 {
			flux_device += "or "
		}

		flux_device += fmt.Sprintf(`r[":device_id"] == "%s" `, device_id)
	}

	if len(device_ids) > 0 {
		flux_device = `and (` + flux_device + `)`
	}

	for idx, attribute_id := range attribute_ids {

		if idx > 0 {
			flux_attribute += "or "
		}

		flux_attribute += fmt.Sprintf(`r["_field"] == "%s" `, attribute_id)
	}

	if len(device_ids) > 0 {
		flux_attribute = `and (` + flux_attribute + `)`
	}

	if page.Stop == 0 && page.Start == 0 {
		page.Stop = 66666666666
	}

	cmd := fmt.Sprintf(`
		from(bucket: "attribute_history")
		|> range(start: %d, stop: %d)
		|> filter(fn: (r) => { return r["_measurement"] == "%s" %s %s })
		|> sort(columns: ["_time"], desc: %t)
		|> limit(n: %d, offset: %d)
		|> drop(columns: ["_start", "_stop", "_measurement"])
		|> yield()
		`,
		page.Start, page.Stop,
		model_id, flux_device, flux_attribute,
		page.Desc,
		page.Size, page.Offset-1,
	)

	if query_data, err = influx.Query(cmd); err != nil {
		return
	}

	for _, item := range query_data {

		device_id, data_id, value, time, ok := "", "", interface{}(nil), int64(0), false

		if device_id, ok = item[":device_id"].(string); !ok {
			continue
		}

		if data_id, ok = item["_field"].(string); !ok {
			continue
		}

		if time, ok = item["time"].(int64); !ok {
			continue
		}

		if value, ok = item["_value"]; !ok {
			continue
		}

		// map[string]map[int64]KeyValueType

		if _, ok = (*datas)[device_id]; !ok {
			(*datas)[device_id] = map[int64]KeyValueType{}
		}

		if _, ok = (*datas)[device_id][time]; !ok {
			(*datas)[device_id][time] = KeyValueType{}
		}

		(*datas)[device_id][time][data_id] = value
	}

	return
}

// device_id.begin_time.event_id.end_time
type DeviceEventHistoryType map[string]map[int64]KeyValueNumberType

func DeleteDeviceEventHistory(influx *plugin.Influx, model_id string, device_id string) (err error) {

	return influx.DeleteByTag("event_history", model_id, ":device_id", device_id)
}

func DeleteModelEventHistory(influx *plugin.Influx, model_id string) (err error) {

	return influx.DeleteByMeasurement("event_history", model_id)
}

func (datas DeviceEventHistoryType) Write(influx *plugin.Influx, model *DeviceModelType) (err error) {

	defer utils.ErrorRecover("DeviceEventHistoryType")

	batch := plugin.NewInfluxBatch(influx, "event_history")

	for device_id, datas := range datas {

		for begin_time, data := range datas {

			tags := map[string]string{":device_id": device_id}

			fields := KeyValueType{}

			for event_id, end_time := range data {

				exist := false

				for index := range model.Events {
					if model.Events[index].Id == event_id {
						exist = true
						break
					}
				}

				if !exist {
					continue
				}

				fields[event_id] = end_time
			}

			batch.AddPoint(model.Id, tags, fields, begin_time)
		}
	}

	return batch.Write()
}

func (datas *DeviceEventHistoryType) Read(influx *plugin.Influx, model_id string, device_ids, event_ids []string, page PageType) (err error) {

	var (
		query_data  = []map[string]interface{}{}
		flux_device = ""
		flex_event  = ""
	)

	for idx, device_id := range device_ids {

		if idx > 0 {
			flux_device += "or "
		}

		flux_device += fmt.Sprintf(`r[":device_id"] == "%s" `, device_id)
	}

	if len(device_ids) > 0 {
		flux_device = `and (` + flux_device + `)`
	}

	for idx, event_id := range event_ids {

		if idx > 0 {
			flex_event += "or "
		}

		flex_event += fmt.Sprintf(`r["_field"] == "%s" `, event_id)
	}

	if len(device_ids) > 0 {
		flex_event = `and (` + flex_event + `)`
	}

	if page.Stop == 0 && page.Start == 0 {
		page.Stop = 66666666666
	}

	cmd := fmt.Sprintf(`
		from(bucket: "event_history")
		|> range(start: %d, stop: %d)
		|> filter(fn: (r) => { return r["_measurement"] == "%s" %s %s })
		|> sort(columns: ["_time"], desc: %t)
		|> limit(n: %d, offset: %d)
		|> drop(columns: ["_start", "_stop", "_measurement"])
		|> yield()
		`,
		page.Start, page.Stop,
		model_id, flux_device, flex_event,
		page.Desc,
		page.Size, page.Offset-1,
	)

	if query_data, err = influx.Query(cmd); err != nil {
		return
	}

	for _, item := range query_data {

		device_id, event_id, end_time, begin_time, ok := "", "", int64(0), int64(0), false

		if device_id, ok = item[":device_id"].(string); !ok {
			continue
		}

		if event_id, ok = item["_field"].(string); !ok {
			continue
		}

		if begin_time, ok = item["time"].(int64); !ok {
			continue
		}

		if end_time, ok = item["_value"].(int64); !ok {
			continue
		}

		if _, ok = (*datas)[device_id]; !ok {
			(*datas)[device_id] = map[int64]KeyValueNumberType{}
		}

		if _, ok = (*datas)[device_id][begin_time]; !ok {
			(*datas)[device_id][begin_time] = KeyValueNumberType{}
		}

		(*datas)[device_id][begin_time][event_id] = end_time
	}

	return
}

// device_id.data_id.{time, value}
type DeviceAttributeRealtimeType map[string]map[string]TimeValueType

func DeleteDeviceAttributeRealtime(influx *plugin.Influx, model_id string, device_id string) (err error) {

	return influx.DeleteByTag("attribute_realtime", model_id, ":device_id", device_id)
}

func DeleteModelAttributeRealtime(influx *plugin.Influx, model_id string) (err error) {

	return influx.DeleteByMeasurement("attribute_realtime", model_id)
}

func (datas DeviceAttributeRealtimeType) Write(influx *plugin.Influx, model *DeviceModelType) (err error) {

	defer utils.ErrorRecover("DeviceAttributeRealtimeType")

	batch := plugin.NewInfluxBatch(influx, "attribute_realtime")

	for device_id, attriable_data := range datas {

		tags := map[string]string{":device_id": device_id}

		fields := KeyValueType{}

		for attribute_id, data := range attriable_data {

			exist := false

			for index := range model.Attributes {
				if model.Attributes[index].Id == attribute_id {
					exist = true
					break
				}
			}

			if !exist {
				continue
			}

			if data_byte, err := json.Marshal(&data); err == nil {
				fields[attribute_id] = string(data_byte)
			}
		}

		batch.AddPoint(model.Id, tags, fields, 0)
	}

	err = batch.Write()

	if err != nil {
		fmt.Println("write err:", model.Id, err)
	}

	return
}

func (datas *DeviceAttributeRealtimeType) Read(influx *plugin.Influx, model_id string, device_ids []string) (err error) {

	var (
		query_data    = []map[string]interface{}{}
		flux_deviceid = ""
	)

	for idx, device_id := range device_ids {

		if idx > 0 {
			flux_deviceid += "or "
		}

		flux_deviceid += fmt.Sprintf(`r[":device_id"] == "%s" `, device_id)
	}

	if len(device_ids) > 0 {
		flux_deviceid = `and (` + flux_deviceid + `)`
	}

	cmd := fmt.Sprintf(`
		from(bucket: "attribute_realtime")
		|> range(start: 0)
		|> filter(fn: (r) => { return r["_measurement"] == "%s" %s })
		|> drop(columns: ["_start", "_stop", "_measurement"])
		|> yield()
		`,
		model_id, flux_deviceid)

	if query_data, err = influx.Query(cmd); err != nil {
		return
	}

	for _, item := range query_data {

		device_id, dataid, value, ok := "", "", TimeValueType{}, false

		if device_id, ok = item[":device_id"].(string); !ok {
			continue
		}

		if dataid, ok = item["_field"].(string); !ok {
			continue
		}

		if _value, ok := item["_value"].(string); !ok {
			continue
		} else {
			if err := json.Unmarshal([]byte(_value), &value); err != nil {
				continue
			}
		}

		if _, ok = (*datas)[device_id]; !ok {
			(*datas)[device_id] = map[string]TimeValueType{}
		}

		(*datas)[device_id][dataid] = value
	}

	return
}

// device_id.event_id.{time, value}
type DeviceEventRealtimeType map[string]map[string]int64

func DeleteDeviceEventRealtime(influx *plugin.Influx, model_id string, device_id string) (err error) {

	return influx.DeleteByTag("event_realtime", model_id, ":device_id", device_id)
}

func DeleteModelEventRealtime(influx *plugin.Influx, model_id string) (err error) {

	return influx.DeleteByMeasurement("event_realtime", model_id)
}

func (datas DeviceEventRealtimeType) Write(influx *plugin.Influx, model *DeviceModelType) (err error) {

	defer utils.ErrorRecover("DeviceEventRealtimeType")

	batch := plugin.NewInfluxBatch(influx, "event_realtime")

	for device_id, attriable_data := range datas {

		tags := map[string]string{":device_id": device_id}
		fields := KeyValueType{}

		for event_id, data := range attriable_data {

			exist := false

			for index := range model.Events {
				if model.Events[index].Id == event_id {
					exist = true
					break
				}
			}

			if !exist {
				continue
			}

			fields[event_id] = data
		}

		batch.AddPoint(model.Id, tags, fields, 0)
	}

	return batch.Write()
}

func (datas *DeviceEventRealtimeType) Read(influx *plugin.Influx, model_id string, device_ids []string) (err error) {

	var (
		query_data  = []map[string]interface{}{}
		flux_string = ""
	)

	for idx, device_id := range device_ids {

		if idx > 0 {
			flux_string += "or "
		}

		flux_string += fmt.Sprintf(`r[":device_id"] == "%s" `, device_id)
	}

	if len(device_ids) > 0 {
		flux_string = `and (` + flux_string + `)`
	}

	cmd := fmt.Sprintf(`
		from(bucket: "event_realtime")
		|> range(start: 0)
		|> filter(fn: (r) => { return r["_measurement"] == "%s" %s })
		|> drop(columns: ["_start", "_stop", "_measurement"])
		|> yield()
		`,
		model_id, flux_string)

	if query_data, err = influx.Query(cmd); err != nil {
		return
	}

	for _, item := range query_data {

		device_id, event_id, value, ok := "", "", int64(0), false

		if device_id, ok = item[":device_id"].(string); !ok {
			continue
		}

		if event_id, ok = item["_field"].(string); !ok {
			continue
		}

		if value, ok = item["_value"].(int64); !ok {
			continue
		}

		if _, ok = (*datas)[device_id]; !ok {
			(*datas)[device_id] = map[string]int64{}
		}

		(*datas)[device_id][event_id] = value
	}

	return
}

// device_id.time
type DeviceActivetimeType map[string]int64

func (datas DeviceActivetimeType) Write(influx *plugin.Influx) (err error) {

	defer utils.ErrorRecover("DeviceActivetimeType")

	batch := plugin.NewInfluxBatch(influx, "activetime")

	for device_id, activetime := range datas {

		if activetime <= 0 || device_id == "" {
			continue
		}

		tags := map[string]string{
			":device_id": device_id,
		}

		fields := map[string]interface{}{
			"value": activetime,
		}

		batch.AddPoint("device", tags, fields, 0)
	}

	return batch.Write()
}

func (datas *DeviceActivetimeType) Read(influx *plugin.Influx, device_ids []string) (err error) {

	var (
		query_data  = []map[string]interface{}{}
		flux_string = ""
	)

	for idx, device_id := range device_ids {

		if idx > 0 {
			flux_string += "or "
		}

		flux_string += fmt.Sprintf(`r[":device_id"] == "%s" `, device_id)
	}

	cmd := fmt.Sprintf(`
		from(bucket: "event_realtime")
		|> range(start: 0)
		|> filter(fn: (r) => { return %s })
		|> drop(columns: ["_start", "_stop", "_measurement"])
		|> yield()
		`,
		flux_string)

	if query_data, err = influx.Query(cmd); err != nil {
		return
	}

	for _, item := range query_data {

		device_id, value, ok := "", int64(0), false

		if device_id, ok = item[":device_id"].(string); !ok {
			continue
		}

		if value, ok = item["_value"].(int64); !ok {
			continue
		}

		(*datas)[device_id] = value
	}

	return
}
