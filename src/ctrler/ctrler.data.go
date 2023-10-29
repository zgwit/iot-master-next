package ctrler

import (
	"encoding/json"
	"time"

	"github.com/zgwit/iot-master-next/src/model"
	"github.com/zgwit/iot-master-next/src/plugin"
	"github.com/zgwit/iot-master-next/src/utils"
)

type DataCtrler struct {
	Influx     *plugin.Influx
	NsqClient  *plugin.NsqClient
	Javascript plugin.Javascript

	lock    bool
	history map[string]model.DeviceAttributeHistoryType
}

func (ctrler *DataCtrler) Init(influx *plugin.Influx, nsq_server *plugin.NsqServer, nsq_client *plugin.NsqClient) (err error) {

	ctrler.Influx = influx
	ctrler.NsqClient = nsq_client
	ctrler.Javascript = plugin.NewJavascript()

	ctrler.lock = false
	ctrler.history = map[string]model.DeviceAttributeHistoryType{}

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

		device_model    model.DeviceModelType
		attribute_write = model.AttributeWriteType{}

		attriables model.KeyValueType
		events     model.KeyValueBoolType
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

	if err := device_model_find(attribute_write.ModelId, &device_model); err != nil {
		return
	}

	if attriables, events, err = attribute_write.Scripting(&ctrler.Javascript, device_model.Script); err != nil {
		return
	}

	go ctrler.EventWrite(attribute_write.Time, &device_model, attribute_write.DeviceId, events)

	ctrler.Lock()

	if _, exist = ctrler.history[attribute_write.ModelId]; !exist {
		ctrler.history[attribute_write.ModelId] = model.DeviceAttributeHistoryType{}
	}

	if _, exist = ctrler.history[attribute_write.ModelId][attribute_write.DeviceId]; !exist {
		ctrler.history[attribute_write.ModelId][attribute_write.DeviceId] = map[int64]model.KeyValueType{}
	}

	if _, exist = ctrler.history[attribute_write.ModelId][attribute_write.DeviceId][attribute_write.Time]; !exist {
		ctrler.history[attribute_write.ModelId][attribute_write.DeviceId][attribute_write.Time] = model.KeyValueType{}
	}

	ctrler.history[attribute_write.ModelId][attribute_write.DeviceId][attribute_write.Time] = attriables

	ctrler.Unlock()
}

func (ctrler *DataCtrler) AttriableWrite() {

	var (
		ticker = time.NewTicker(200 * time.Millisecond)

		device_model model.DeviceModelType
	)

	for range ticker.C {

		ctrler.Lock()

		for model_id, model_data := range ctrler.history {

			if err := device_model_find(model_id, &device_model); err != nil {
				return
			}

			realtime := model.DeviceAttributeRealtimeType{}

			for device_id, device_data := range model_data {

				if _, exist := realtime[device_id]; !exist {
					realtime[device_id] = map[string]model.TimeValueType{}
				}

				for time, datas := range device_data {

					for data_id, value := range datas {

						if _, exist := realtime[device_id][data_id]; !exist {
							realtime[device_id][data_id] = model.TimeValueType{Time: time, Value: value}
						}

						if time > realtime[device_id][data_id].Time {
							realtime[device_id][data_id] = model.TimeValueType{Time: time, Value: value}
						}
					}
				}

				go realtime.Write(ctrler.Influx, &device_model)
				go model_data.Write(ctrler.Influx, &device_model)

				ctrler.history[model_id] = model.DeviceAttributeHistoryType{}
			}
		}

		ctrler.Unlock()
	}
}

func (ctrler *DataCtrler) EventWrite(time int64, device_model *model.DeviceModelType, device_id string, events model.KeyValueBoolType) {

	event_realtime_datas := model.DeviceEventRealtimeType{device_id: {}}

	for event_id, value := range events {

		exist := false

		for index := range device_model.Events {
			if device_model.Events[index].Id == event_id {
				exist = true
				break
			}
		}

		if !exist {
			delete(events, event_id)
			continue
		}

		if value {
			event_realtime_datas[device_id][event_id] = time
		} else {
			event_realtime_datas[device_id][event_id] = 0
		}
	}

	event_datas_cache := model.DeviceEventRealtimeType{}

	if err := event_datas_cache.Read(ctrler.Influx, device_model.Id, []string{device_id}); err != nil {
		return
	}

	if _, exist := event_datas_cache[device_id]; !exist {
		event_datas_cache[device_id] = map[string]int64{}
	}

	for event_id, value := range events {

		value_cache, exist := event_datas_cache[device_id][event_id]

		event_write_pacakge := model.EventWriteType{
			ModelId: device_model.Id, DeviceId: device_id, EventId: event_id,
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

		ctrler.NsqClient.Publish("event.read", utils.ToJson2(event_write_pacakge))

		if event_write_pacakge.Type == "continue" {
			delete(event_realtime_datas[device_id], event_id)
		}
	}

	go event_realtime_datas.Write(ctrler.Influx, device_model)
}

func (ctrler *DataCtrler) EventRead(message string) {

	event_write_pacakge := model.EventWriteType{}
	device_model := model.DeviceModelType{}

	if err := json.Unmarshal([]byte(message), &event_write_pacakge); err != nil {
		return
	}

	if event_write_pacakge.Type != "end" {
		return
	}

	if err := device_model_find(event_write_pacakge.ModelId, &device_model); err != nil {
		return
	}

	event_history := model.DeviceEventHistoryType{
		event_write_pacakge.DeviceId: {
			event_write_pacakge.BeginTime: {
				event_write_pacakge.EventId: event_write_pacakge.EndTime,
			},
		},
	}

	go event_history.Write(ctrler.Influx, &device_model)
}
