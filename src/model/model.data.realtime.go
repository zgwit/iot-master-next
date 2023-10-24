package zgwit_model

import (
	"encoding/json"
	"fmt"
	zgwit_plugin "local/plugin"
)

// device_id.data_id.{time, value}
type DeviceAttributeRealtime map[string]map[string]TimeValue

func DeleteDeviceAttributeRealtime(influx *zgwit_plugin.Influx, model_id string, device_id string) (err error) {

	return influx.DeleteByTag("attribute_realtime", model_id, ":device_id", device_id)
}

func DeleteModelAttributeRealtime(influx *zgwit_plugin.Influx, model_id string) (err error) {

	return influx.DeleteByMeasurement("attribute_realtime", model_id)
}

func (datas *DeviceAttributeRealtime) Write(influx *zgwit_plugin.Influx, model *Model) (err error) {

	batch := zgwit_plugin.NewInfluxBatch(influx, "attribute_realtime")

	for device_id, attriable_data := range *datas {

		tags := map[string]string{":device_id": device_id}

		fields := KeyValue{}

		for attriable_id, data := range attriable_data {

			if _, exist := model.Events[attriable_id]; !exist {
				continue
			}

			if data_byte, err := json.Marshal(&data); err == nil {
				fields[attriable_id] = string(data_byte)
			}
		}

		batch.AddPoint(model.Id, tags, fields, 0)
	}

	return batch.Write()
}

func (datas *DeviceAttributeRealtime) Read(influx *zgwit_plugin.Influx, model_id string, device_ids []string) (err error) {

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

		device_id, dataid, value, ok := "", "", TimeValue{}, false

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
			(*datas)[device_id] = map[string]TimeValue{}
		}

		(*datas)[device_id][dataid] = value
	}

	return
}

// device_id.event_id.{time, value}
type DeviceEventRealtime map[string]map[string]int64

func DeleteDeviceEventRealtime(influx *zgwit_plugin.Influx, model_id string, device_id string) (err error) {

	return influx.DeleteByTag("event_realtime", model_id, ":device_id", device_id)
}

func DeleteModelEventRealtime(influx *zgwit_plugin.Influx, model_id string) (err error) {

	return influx.DeleteByMeasurement("event_realtime", model_id)
}

func (datas *DeviceEventRealtime) Write(influx *zgwit_plugin.Influx, model *Model) (err error) {

	batch := zgwit_plugin.NewInfluxBatch(influx, "event_realtime")

	for device_id, attriable_data := range *datas {

		tags := map[string]string{":device_id": device_id}
		fields := KeyValue{}

		for attriable_id, data := range attriable_data {

			if _, exist := model.Events[attriable_id]; !exist {
				continue
			}

			fields[attriable_id] = data
		}

		batch.AddPoint(model.Id, tags, fields, 0)
	}

	return batch.Write()
}

func (datas *DeviceEventRealtime) Read(influx *zgwit_plugin.Influx, model_id string, device_ids []string) (err error) {

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