package zgwit_model

import (
	"fmt"
	zgwit_plugin "local/plugin"
)

type AttributeWritePackage struct {
	Time int64 `form:"time" bson:"time" json:"time"`

	ModelId  string `form:"model_id" bson:"model_id" json:"model_id"`
	DeviceId string `form:"device_id" bson:"device_id" json:"device_id"`

	Points KeyValue `form:"points" bson:"points" json:"points"`
}

func (pack *AttributeWritePackage) Scripting(javascript *zgwit_plugin.Javascript, script string) (attributes KeyValue, events KeyValueBool, err error) {

	var (
		ok bool

		script_result zgwit_plugin.JavaScriptResult
		result        KeyValue
	)

	if script_result, err = javascript.Function("handle",
		`function handle(points) { let attributes = {}, events = {};`+script+`;return { attributes, events } }`,
		pack.Points); err != nil {
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

	events = KeyValueBool{}

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

type EventWritePackage struct {
	Type string `form:"type" bson:"type" json:"type"`

	ModelId  string `form:"model_id" bson:"model_id" json:"model_id"`
	DeviceId string `form:"device_id" bson:"device_id" json:"device_id"`
	EventId  string `form:"event_id" bson:"event_id" json:"event_id"`

	BeginTime int64 `form:"begin_time" bson:"begin_time" json:"begin_time"`
	EndTime   int64 `form:"end_time" bson:"end_time" json:"end_time"`
}
