package zgwit_model

import (
	"fmt"
	zgwit_plugin "local/plugin"
	zgwit_utils "local/utils"
)

// device_id.time.data_id.value
type DeviceAttributeHistory map[string]map[int64]KeyValue

func DeleteDeviceAttributeHistory(influx *zgwit_plugin.Influx, model_id string, device_id string) (err error) {

	return influx.DeleteByTag("attribute_history", model_id, ":device_id", device_id)
}

func DeleteModelAttributeHistory(influx *zgwit_plugin.Influx, model_id string) (err error) {

	return influx.DeleteByMeasurement("attribute_history", model_id)
}

func (datas *DeviceAttributeHistory) Write(influx *zgwit_plugin.Influx, model *Model) (err error) {

	batch := zgwit_plugin.NewInfluxBatch(influx, "attribute_history")

	for device_id, datas := range *datas {

		for time, data := range datas {

			tags := map[string]string{":device_id": device_id}

			fields := KeyValue{}

			for data_id, value_interface := range data {

				attribute, exist := ModelAttribute{}, false

				if attribute, exist = model.Attributes[data_id]; !exist {
					continue
				}

				if attribute.LastestOnly {
					continue
				}

				switch value := value_interface.(type) {
				case float64:
					if attribute.DataType == "number" {
						fields[data_id] = value
					}
				case int64:
					if attribute.DataType == "number" {
						fields[data_id] = float64(value)
					}
				case string:
					if attribute.DataType == "string" {
						fields[data_id] = value
					}
				case bool:
					if attribute.DataType == "number" {
						fields[data_id] = zgwit_utils.GetNumberBool(value)
					}
				}
			}

			batch.AddPoint(model.Id, tags, fields, time)
		}
	}

	return batch.Write()
}

func (datas *DeviceAttributeHistory) Read(influx *zgwit_plugin.Influx, model_id string, device_ids, attribute_ids []string, page Page) (err error) {

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

		// map[string]map[int64]KeyValue

		if _, ok = (*datas)[device_id]; !ok {
			(*datas)[device_id] = map[int64]KeyValue{}
		}

		if _, ok = (*datas)[device_id][time]; !ok {
			(*datas)[device_id][time] = KeyValue{}
		}

		(*datas)[device_id][time][data_id] = value
	}

	return
}

// device_id.begin_time.event_id.end_time
type DeviceEventHistory map[string]map[int64]KeyValueNumber

func DeleteDeviceEventHistory(influx *zgwit_plugin.Influx, model_id string, device_id string) (err error) {

	return influx.DeleteByTag("event_history", model_id, ":device_id", device_id)
}

func DeleteModelEventHistory(influx *zgwit_plugin.Influx, model_id string) (err error) {

	return influx.DeleteByMeasurement("event_history", model_id)
}

func (datas *DeviceEventHistory) Write(influx *zgwit_plugin.Influx, model *Model) (err error) {

	batch := zgwit_plugin.NewInfluxBatch(influx, "event_history")

	for device_id, datas := range *datas {

		for begin_time, data := range datas {

			tags := map[string]string{":device_id": device_id}

			fields := KeyValue{}

			for event_id, end_time := range data {

				if _, exist := model.Events[event_id]; !exist {
					continue
				}

				fields[event_id] = end_time
			}

			batch.AddPoint(model.Id, tags, fields, begin_time)
		}
	}

	return batch.Write()
}

func (datas *DeviceEventHistory) Read(influx *zgwit_plugin.Influx, model_id string, device_ids, event_ids []string, page Page) (err error) {

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
			(*datas)[device_id] = map[int64]KeyValueNumber{}
		}

		if _, ok = (*datas)[device_id][begin_time]; !ok {
			(*datas)[device_id][begin_time] = KeyValueNumber{}
		}

		(*datas)[device_id][begin_time][event_id] = end_time
	}

	return
}
