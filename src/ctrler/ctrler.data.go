package zgwit_ctrler

import (
	"encoding/json"
	zgwit_model "local/model"
	zgwit_plugin "local/plugin"
	zgwit_utils "local/utils"
	"time"
)

type DataCtrler struct {
	Influx     *zgwit_plugin.Influx
	NsqClient  *zgwit_plugin.NsqClient
	Javascript zgwit_plugin.Javascript

	lock    bool
	history map[string]zgwit_model.DeviceAttributeHistory
}

func (ctrler *DataCtrler) Init(influx *zgwit_plugin.Influx, nsq_server *zgwit_plugin.NsqServer, nsq_client *zgwit_plugin.NsqClient) (err error) {

	ctrler.Influx = influx
	ctrler.NsqClient = nsq_client
	ctrler.Javascript = zgwit_plugin.NewJavascript()

	ctrler.lock = false
	ctrler.history = map[string]zgwit_model.DeviceAttributeHistory{}

	if err = nsq_server.Subscribe("attribute.write", ctrler.AttributeRead); err != nil {
		return
	}

	if err = nsq_server.Subscribe("event.read", ctrler.EventRead); err != nil {
		return
	}

	go ctrler.AttriableWrite()

	return
}

func (ctrler *DataCtrler) Lock() {

	timer := int64(0)

	for ctrler.lock {

		if timer > 60*1000 {
			break
		}

		time.Sleep(1 * time.Millisecond)
		timer++
	}

	ctrler.lock = true
}

func (ctrler *DataCtrler) Unlock() {
	ctrler.lock = false
}

func (ctrler *DataCtrler) AttributeRead(data_str string) {

	var (
		err   error
		exist bool

		model           zgwit_model.Model
		attribute_write = zgwit_model.AttributeWritePackage{}

		attriables zgwit_model.KeyValue
		events     zgwit_model.KeyValueBool
	)

	if err = json.Unmarshal([]byte(data_str), &attribute_write); err != nil {
		return
	}

	if attribute_write.ModelId == "" || attribute_write.DeviceId == "" {
		return
	}

	if attribute_write.Time == 0 {
		attribute_write.Time = time.Now().Unix()
	}

	if err := zgwit_model.ModelFind(attribute_write.ModelId, &model); err != nil {
		return
	}

	if attriables, events, err = attribute_write.Scripting(&ctrler.Javascript, model.Script); err != nil {
		return
	}

	go ctrler.EventWrite(attribute_write.Time, &model, attribute_write.DeviceId, events)

	ctrler.Lock()

	if _, exist = ctrler.history[attribute_write.ModelId]; !exist {
		ctrler.history[attribute_write.ModelId] = zgwit_model.DeviceAttributeHistory{}
	}

	if _, exist = ctrler.history[attribute_write.ModelId][attribute_write.DeviceId]; !exist {
		ctrler.history[attribute_write.ModelId][attribute_write.DeviceId] = map[int64]zgwit_model.KeyValue{}
	}

	if _, exist = ctrler.history[attribute_write.ModelId][attribute_write.DeviceId][attribute_write.Time]; !exist {
		ctrler.history[attribute_write.ModelId][attribute_write.DeviceId][attribute_write.Time] = zgwit_model.KeyValue{}
	}

	ctrler.history[attribute_write.ModelId][attribute_write.DeviceId][attribute_write.Time] = attriables

	ctrler.Unlock()
}

func (ctrler *DataCtrler) AttriableWrite() {

	var (
		ticker = time.NewTicker(200 * time.Millisecond)

		model zgwit_model.Model
	)

	for range ticker.C {

		ctrler.Lock()

		for model_id, model_data := range ctrler.history {

			if err := zgwit_model.ModelFind(model_id, &model); err != nil {
				return
			}

			realtime := zgwit_model.DeviceAttributeRealtime{}

			for device_id, device_data := range model_data {

				if _, exist := realtime[device_id]; !exist {
					realtime[device_id] = map[string]zgwit_model.TimeValue{}
				}

				for time, datas := range device_data {

					for data_id, value := range datas {

						if _, exist := realtime[device_id][data_id]; !exist {
							realtime[device_id][data_id] = zgwit_model.TimeValue{Time: time, Value: value}
						}

						if time > realtime[device_id][data_id].Time {
							realtime[device_id][data_id] = zgwit_model.TimeValue{Time: time, Value: value}
						}
					}
				}

				go realtime.Write(ctrler.Influx, &model)
				go model_data.Write(ctrler.Influx, &model)

				ctrler.history[model_id] = zgwit_model.DeviceAttributeHistory{}
			}
		}

		ctrler.Unlock()
	}
}

func (ctrler *DataCtrler) EventWrite(time int64, model *zgwit_model.Model, device_id string, events zgwit_model.KeyValueBool) {

	event_realtime_datas := zgwit_model.DeviceEventRealtime{device_id: {}}
	// event_history_datas := zgwit_model.DeviceEventHistory{device_id: {time: {}}}

	for event_id, value := range events {

		if _, exist := model.Events[event_id]; !exist {
			delete(events, event_id)
			continue
		}

		if value {
			event_realtime_datas[device_id][event_id] = time
		} else {
			event_realtime_datas[device_id][event_id] = 0
		}
	}

	event_datas_cache := zgwit_model.DeviceEventRealtime{}

	if err := event_datas_cache.Read(ctrler.Influx, model.Id, []string{device_id}); err != nil {
		return
	}

	if _, exist := event_datas_cache[device_id]; !exist {
		event_datas_cache[device_id] = map[string]int64{}
	}

	for event_id, value := range events {

		value_cache, exist := event_datas_cache[device_id][event_id]

		event_write_pacakge := zgwit_model.EventWritePackage{
			ModelId: model.Id, DeviceId: device_id, EventId: event_id,
		}

		switch {

		case !value && (!exist || value_cache == 0):
		case !value && (exist && value_cache == 0):

		case value && (!exist || value_cache == 0):
			event_write_pacakge.Type = "begin"
			event_write_pacakge.BeginTime = time

		case value && (exist && value_cache != 0):
			event_write_pacakge.Type = "continue"
			event_write_pacakge.BeginTime = value_cache
			event_write_pacakge.EndTime = time

		case !value && (exist && value_cache != 0):
			event_write_pacakge.Type = "end"
			event_write_pacakge.BeginTime = value_cache
			event_write_pacakge.EndTime = time
		}

		if event_write_pacakge.Type == "" {
			delete(event_realtime_datas[device_id], event_id)
			continue
		}

		if err := ctrler.NsqClient.Publish("event.read", zgwit_utils.ToJson2(event_write_pacakge)); err != nil {
		}

		if event_write_pacakge.Type == "continue" {
			delete(event_realtime_datas[device_id], event_id)
		}
	}

	go event_realtime_datas.Write(ctrler.Influx, model)
}

func (ctrler *DataCtrler) EventRead(message string) {

	event_write_pacakge := zgwit_model.EventWritePackage{}
	model := zgwit_model.Model{}

	if err := json.Unmarshal([]byte(message), &event_write_pacakge); err != nil {
		return
	}

	if event_write_pacakge.Type != "end" {
		return
	}

	if err := zgwit_model.ModelFind(event_write_pacakge.ModelId, &model); err != nil {
		return
	}

	event_history := zgwit_model.DeviceEventHistory{
		event_write_pacakge.DeviceId: {
			event_write_pacakge.BeginTime: {
				event_write_pacakge.EventId: event_write_pacakge.EndTime,
			},
		},
	}

	go event_history.Write(ctrler.Influx, &model)
}
