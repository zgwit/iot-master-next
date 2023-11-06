package model

import (
	"encoding/json"
	"fmt"
	"strings"

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

	if script_result, err = javascript.Function("handle", script, attribute_write.Points); err != nil {
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

const (
	BUCKET_ATTRIBUTE_REALTIME = "attribute_realtime"
	BUCKET_ATTRIBUTE_HISTORY  = "attribute_history"
	BUCKET_EVENT_REALTIME     = "event_realtime"
	BUCKET_EVENT_HISTORY      = "event_history"
	BUCKET_ACTIVETIME         = "activetime"

	MEASUREMENT_MODEL = "_measurement"

	TAG_DEVICE = ":device"

	FIELD_ATTRIBUTE = "_field"
	FIELD_EVENT     = "_field"

	FIELD_TIME  = "time"
	FIELD_VALUE = "_value"
)

// model_id.device_id.[attribute_id, event_id]
type DataFilterType map[string]map[string]map[string]bool

// device_id.data_id.{time, value}
type DeviceAttributeRealtimeType map[string]map[string]TimeValueType

func DeleteDeviceAttributeRealtime(influx *plugin.Influx, model_id string, device_id string) (err error) {

	return influx.DeleteByTag(BUCKET_ATTRIBUTE_REALTIME, model_id, TAG_DEVICE, device_id)
}

func DeleteModelAttributeRealtime(influx *plugin.Influx, model_id string) (err error) {

	return influx.DeleteByMeasurement(BUCKET_ATTRIBUTE_REALTIME, model_id)
}

func (datas DeviceAttributeRealtimeType) Write(influx *plugin.Influx, model *DeviceModelType) (err error) {

	defer utils.ErrorRecover("DeviceAttributeRealtimeType")

	batch := plugin.NewInfluxBatch(influx, BUCKET_ATTRIBUTE_REALTIME)

	for device_id, attriable_data := range datas {

		tags := map[string]string{
			TAG_DEVICE: device_id,
		}

		fields := KeyValueType{}

		for attribute_id, data := range attriable_data {

			exist := false

			for index := range model.Attributes {
				if model.Attributes[index].Id == attribute_id {
					exist = true
					break
				}
			}

			if exist {
				if data_byte, err := json.Marshal(&data); err == nil {
					fields[attribute_id] = string(data_byte)
				}
			}
		}

		batch.AddPoint(model.Id, tags, fields, 0)
	}

	err = batch.Write()

	return
}

func (datas *DeviceAttributeRealtimeType) Read(influx *plugin.Influx, filter DataFilterType) (err error) {

	var (
		query_data  = []map[string]interface{}{}
		flux_string = "false "
	)

	for model_id, filter_device := range filter {

		flux_device := ""

		for device_id, filter_attribute := range filter_device {

			flux_attribute := ""

			for attribute_id, enable := range filter_attribute {

				if !enable {
					continue
				}

				if strings.TrimSpace(attribute_id) == "" {
					flux_attribute += fmt.Sprintf(`or ( false %s ) `, "or true")
				} else {
					flux_attribute += fmt.Sprintf(`or ( r["%s"] == "%s" ) `, FIELD_ATTRIBUTE, attribute_id)
				}
			}

			if strings.TrimSpace(device_id) == "" {
				flux_device += fmt.Sprintf(`or ( false %s ) `, flux_attribute)
			} else {
				flux_device += fmt.Sprintf(`or ( r["%s"] == "%s" and ( false %s) ) `, TAG_DEVICE, device_id, flux_attribute)
			}
		}

		if strings.TrimSpace(model_id) == "" {
			flux_string += fmt.Sprintf(`or ( false %s ) `, flux_device)
		} else {
			flux_string += fmt.Sprintf(`or ( r["%s"] == "%s" and ( false %s) ) `, MEASUREMENT_MODEL, model_id, flux_device)
		}
	}

	cmd := fmt.Sprintf(`
		from(bucket: "%s")
		|> range(start: 0)
		|> filter(fn: (r) => { return %s })
		|> drop(columns: ["_start", "_stop", "_measurement"])
		|> yield()
		`,
		BUCKET_ATTRIBUTE_REALTIME,
		flux_string,
	)

	if query_data, err = influx.Query(cmd); err != nil {
		return
	}

	for _, item := range query_data {

		device_id, dataid, value, ok := "", "", TimeValueType{}, false

		if device_id, ok = item[TAG_DEVICE].(string); !ok {
			continue
		}

		if dataid, ok = item[FIELD_ATTRIBUTE].(string); !ok {
			continue
		}

		if _value, ok := item[FIELD_VALUE].(string); !ok {
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

// device_id.time.data_id.value
type DeviceAttributeHistoryType map[string]map[int64]KeyValueType

func DeleteDeviceAttributeHistory(influx *plugin.Influx, model_id string, device_id string) (err error) {

	return influx.DeleteByTag(BUCKET_ATTRIBUTE_HISTORY, model_id, TAG_DEVICE, device_id)
}

func DeleteModelAttributeHistory(influx *plugin.Influx, model_id string) (err error) {

	return influx.DeleteByMeasurement(BUCKET_ATTRIBUTE_HISTORY, model_id)
}

func (datas DeviceAttributeHistoryType) Write(influx *plugin.Influx, model *DeviceModelType) (err error) {

	defer utils.ErrorRecover("DeviceAttributeHistoryType")

	batch := plugin.NewInfluxBatch(influx, BUCKET_ATTRIBUTE_HISTORY)

	for device_id, datas := range datas {

		for time, data := range datas {

			tags := map[string]string{
				TAG_DEVICE: device_id,
			}

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

func (datas *DeviceAttributeHistoryType) Read(influx *plugin.Influx, filter DataFilterType, page PageType) (err error) {

	var (
		query_data  = []map[string]interface{}{}
		flux_string = "false "
	)

	for model_id, filter_device := range filter {

		flux_device := ""

		for device_id, filter_attribute := range filter_device {

			flux_attribute := ""

			for attribute_id, enable := range filter_attribute {

				if !enable {
					continue
				}

				if strings.TrimSpace(attribute_id) == "" {
					flux_attribute += fmt.Sprintf(`or ( false %s ) `, "or true")
				} else {
					flux_attribute += fmt.Sprintf(`or ( r["%s"] == "%s" ) `, FIELD_ATTRIBUTE, attribute_id)
				}
			}

			if strings.TrimSpace(device_id) == "" {
				flux_device += fmt.Sprintf(`or ( false %s ) `, flux_attribute)
			} else {
				flux_device += fmt.Sprintf(`or ( r["%s"] == "%s" and ( false %s) ) `, TAG_DEVICE, device_id, flux_attribute)
			}
		}

		if strings.TrimSpace(model_id) == "" {
			flux_string += fmt.Sprintf(`or ( false %s ) `, flux_device)
		} else {
			flux_string += fmt.Sprintf(`or ( r["%s"] == "%s" and ( false %s) ) `, MEASUREMENT_MODEL, model_id, flux_device)
		}
	}

	if page.Stop == 0 && page.Start == 0 {
		page.Stop = 66666666666
	}

	cmd := fmt.Sprintf(`
		from(bucket: "%s")
		|> range(start: %d, stop: %d)
		|> filter(fn: (r) => { return %s })
		|> sort(columns: ["_time"], desc: %t)
		|> limit(n: %d, offset: %d)
		|> drop(columns: ["_start", "_stop", "_measurement"])
		|> yield()
		`,
		BUCKET_ATTRIBUTE_HISTORY,
		page.Start, page.Stop,
		flux_string,
		page.Desc,
		page.Size, page.Offset-1,
	)

	if query_data, err = influx.Query(cmd); err != nil {
		return
	}

	for _, item := range query_data {

		device_id, data_id, value, time, ok := "", "", interface{}(nil), int64(0), false

		if device_id, ok = item[TAG_DEVICE].(string); !ok {
			continue
		}

		if data_id, ok = item[FIELD_ATTRIBUTE].(string); !ok {
			continue
		}

		if time, ok = item[FIELD_TIME].(int64); !ok {
			continue
		}

		if value, ok = item[FIELD_VALUE]; !ok {
			continue
		}

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

// device_id.event_id.{time, value}
type DeviceEventRealtimeType map[string]map[string]int64

func DeleteDeviceEventRealtime(influx *plugin.Influx, model_id string, device_id string) (err error) {

	return influx.DeleteByTag(BUCKET_EVENT_REALTIME, model_id, TAG_DEVICE, device_id)
}

func DeleteModelEventRealtime(influx *plugin.Influx, model_id string) (err error) {

	return influx.DeleteByMeasurement(BUCKET_EVENT_REALTIME, model_id)
}

func (datas DeviceEventRealtimeType) Write(influx *plugin.Influx, model *DeviceModelType) (err error) {

	defer utils.ErrorRecover("DeviceEventRealtimeType")

	batch := plugin.NewInfluxBatch(influx, BUCKET_EVENT_REALTIME)

	for device_id, attriable_data := range datas {

		tags := map[string]string{
			TAG_DEVICE: device_id,
		}

		fields := KeyValueType{}

		for event_id, data := range attriable_data {

			exist := false

			for index := range model.Events {
				if model.Events[index].Id == event_id {
					exist = true
					break
				}
			}

			if exist {
				fields[event_id] = data
			}
		}

		batch.AddPoint(model.Id, tags, fields, 0)
	}

	return batch.Write()
}

func (datas *DeviceEventRealtimeType) Read(influx *plugin.Influx, filter DataFilterType) (err error) {

	var (
		query_data  = []map[string]interface{}{}
		flux_string = "false "
	)

	for model_id, filter_device := range filter {

		flux_device := ""

		for device_id, filter_event := range filter_device {

			flux_event := ""

			for event_id, enable := range filter_event {

				if !enable {
					continue
				}

				if strings.TrimSpace(event_id) == "" {
					flux_event += fmt.Sprintf(`or ( false %s ) `, "or true")
				} else {
					flux_event += fmt.Sprintf(`or ( r["%s"] == "%s" ) `, FIELD_EVENT, event_id)
				}
			}

			if strings.TrimSpace(device_id) == "" {
				flux_device += fmt.Sprintf(`or ( false %s ) `, flux_event)
			} else {
				flux_device += fmt.Sprintf(`or ( r["%s"] == "%s" and ( false %s) ) `, TAG_DEVICE, device_id, flux_event)
			}
		}

		if strings.TrimSpace(model_id) == "" {
			flux_string += fmt.Sprintf(`or ( false %s ) `, flux_device)
		} else {
			flux_string += fmt.Sprintf(`or ( r["%s"] == "%s" and ( false %s) ) `, MEASUREMENT_MODEL, model_id, flux_device)
		}
	}

	cmd := fmt.Sprintf(`
		from(bucket: "%s")
		|> range(start: 0)
		|> filter(fn: (r) => { return r["%s"] != 0 and ( %s ) })
		|> drop(columns: ["_start", "_stop", "_measurement"])
		|> yield()
		`,
		BUCKET_EVENT_REALTIME,
		FIELD_VALUE, flux_string)

	if query_data, err = influx.Query(cmd); err != nil {
		return
	}

	for _, item := range query_data {

		device_id, event_id, value, ok := "", "", int64(0), false

		if device_id, ok = item[TAG_DEVICE].(string); !ok {
			continue
		}

		if event_id, ok = item[FIELD_EVENT].(string); !ok {
			continue
		}

		if value, ok = item[FIELD_VALUE].(int64); !ok {
			continue
		}

		if _, ok = (*datas)[device_id]; !ok {
			(*datas)[device_id] = map[string]int64{}
		}

		(*datas)[device_id][event_id] = value
	}

	return
}

// device_id.begin_time.event_id.end_time
type DeviceEventHistoryType map[string]map[int64]KeyValueNumberType

func DeleteDeviceEventHistory(influx *plugin.Influx, model_id string, device_id string) (err error) {

	return influx.DeleteByTag(BUCKET_EVENT_HISTORY, model_id, TAG_DEVICE, device_id)
}

func DeleteModelEventHistory(influx *plugin.Influx, model_id string) (err error) {

	return influx.DeleteByMeasurement(BUCKET_EVENT_HISTORY, model_id)
}

func (datas DeviceEventHistoryType) Write(influx *plugin.Influx, model *DeviceModelType) (err error) {

	defer utils.ErrorRecover("DeviceEventHistoryType")

	batch := plugin.NewInfluxBatch(influx, BUCKET_EVENT_HISTORY)

	for device_id, datas := range datas {

		for begin_time, data := range datas {

			tags := map[string]string{
				TAG_DEVICE: device_id,
			}

			fields := KeyValueType{}

			for event_id, end_time := range data {

				exist := false

				for index := range model.Events {
					if model.Events[index].Id == event_id {
						exist = true
						break
					}
				}

				if exist {
					fields[event_id] = end_time
				}
			}

			batch.AddPoint(model.Id, tags, fields, begin_time)
		}
	}

	return batch.Write()
}

func (datas *DeviceEventHistoryType) Read(influx *plugin.Influx, filter DataFilterType, page PageType) (err error) {

	var (
		query_data  = []map[string]interface{}{}
		flux_string = "false "
	)

	for model_id, filter_device := range filter {

		flux_device := ""

		for device_id, filter_event := range filter_device {

			flux_event := ""

			for event_id, enable := range filter_event {

				if !enable {
					continue
				}

				if strings.TrimSpace(event_id) == "" {
					flux_event += fmt.Sprintf(`or ( false %s ) `, "or true")
				} else {
					flux_event += fmt.Sprintf(`or ( r["%s"] == "%s" ) `, FIELD_EVENT, event_id)
				}
			}

			if strings.TrimSpace(device_id) == "" {
				flux_device += fmt.Sprintf(`or ( false %s ) `, flux_event)
			} else {
				flux_device += fmt.Sprintf(`or ( r["%s"] == "%s" and ( false %s) ) `, TAG_DEVICE, device_id, flux_event)
			}
		}

		if strings.TrimSpace(model_id) == "" {
			flux_string += fmt.Sprintf(`or ( false %s ) `, flux_device)
		} else {
			flux_string += fmt.Sprintf(`or ( r["%s"] == "%s" and ( false %s) ) `, MEASUREMENT_MODEL, model_id, flux_device)
		}
	}

	if page.Stop == 0 && page.Start == 0 {
		page.Stop = 66666666666
	}

	cmd := fmt.Sprintf(`
		from(bucket: "%s")
		|> range(start: %d, stop: %d)
		|> filter(fn: (r) => { return %s })
		|> sort(columns: ["_time"], desc: %t)
		|> limit(n: %d, offset: %d)
		|> drop(columns: ["_start", "_stop", "_measurement"])
		|> yield()
		`,
		BUCKET_EVENT_HISTORY,
		page.Start, page.Stop,
		flux_string,
		page.Desc,
		page.Size, page.Offset-1,
	)

	if query_data, err = influx.Query(cmd); err != nil {
		return
	}

	for _, item := range query_data {

		device_id, event_id, end_time, begin_time, ok := "", "", int64(0), int64(0), false

		if device_id, ok = item[TAG_DEVICE].(string); !ok {
			continue
		}

		if event_id, ok = item[FIELD_EVENT].(string); !ok {
			continue
		}

		if begin_time, ok = item[FIELD_TIME].(int64); !ok {
			continue
		}

		if end_time, ok = item[FIELD_VALUE].(int64); !ok {
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

// device_id.time
type DeviceActivetimeType map[string]int64

func (datas DeviceActivetimeType) Write(influx *plugin.Influx) (err error) {

	defer utils.ErrorRecover("DeviceActivetimeType")

	batch := plugin.NewInfluxBatch(influx, BUCKET_ACTIVETIME)

	for device_id, activetime := range datas {

		if activetime <= 0 || device_id == "" {
			continue
		}

		tags := map[string]string{
			TAG_DEVICE: device_id,
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

		flux_string += fmt.Sprintf(`r["%s"] == "%s" `, TAG_DEVICE, device_id)
	}

	if len(device_ids) > 0 {
		flux_string = `and (` + flux_string + `)`
	}

	cmd := fmt.Sprintf(`
		from(bucket: "%s")
		|> range(start: 0)
		|> filter(fn: (r) => { return r["_measurement"] == "device" %s })
		|> drop(columns: ["_start", "_stop", "_measurement"])
		|> yield()
		`,
		BUCKET_ACTIVETIME,
		flux_string)

	if query_data, err = influx.Query(cmd); err != nil {
		return
	}

	for _, item := range query_data {

		device_id, value, ok := "", int64(0), false

		if device_id, ok = item[TAG_DEVICE].(string); !ok {
			continue
		}

		if value, ok = item[FIELD_VALUE].(int64); !ok {
			continue
		}

		(*datas)[device_id] = value
	}

	return
}
