package model

import (
	"strings"

	"github.com/zgwit/iot-master-next/src/utils"
)

type DeviceModelType struct {
	Name string `form:"name" bson:"name" json:"name"`
	Id   string `form:"id" bson:"id" json:"id"`

	Attributes map[string]DeviceModelAttributeType `form:"attributes" bson:"attributes" json:"attributes"`
	Events     map[string]DeviceModelEventType     `form:"events" bson:"events" json:"events"`
	Actions    map[string]DeviceModelActionType    `form:"actions" bson:"actions" json:"actions"`

	KeepAlive int64  `form:"keep_alive" bson:"keep_alive" json:"keep_alive"`
	Script    string `form:"script" bson:"script" json:"script"`

	Drive       string      `form:"drive" bson:"drive" json:"drive"`
	DriveConfig interface{} `form:"drive_config" bson:"drive_config" json:"drive_config"`
}

type DeviceModelAttributeType struct {
	Name        string `form:"name" bson:"name" json:"name"`
	DataType    string `form:"data_type" bson:"data_type" json:"data_type"`
	LastestOnly bool   `form:"lastest_only" bson:"lastest_only" json:"lastest_only"`
}

type DeviceModelEventType struct {
	Name string `form:"name" bson:"name" json:"name"`
}

type DeviceModelActionType struct {
	Name string `form:"name" bson:"name" json:"name"`
}

func ModelFetchList() (models []DeviceModelType, err error) {

	var (
		labels []string
	)

	models = []DeviceModelType{}

	if labels, err = utils.GetDirFileNames2("./config/device.model"); err != nil {
		return
	}

	for _, label := range labels {

		label := strings.Replace(label, ".txt", "", -1)

		model := DeviceModelType{}

		if err = utils.ReadFileToObject("./config/device.model/"+label+".txt", &model); err != nil {
			continue
		}

		models = append(models, model)
	}

	return
}

func ModelFind(model_id string, model *DeviceModelType) (err error) {

	return utils.ReadFileToObject("./config/device.model/"+model_id+".txt", model)
}

func ModelExist(model_id string) (result bool) {

	return utils.FileExist("./config/device.model/" + model_id + ".txt")
}

func (model DeviceModelType) Create() (err error) {

	if err = utils.WriteObjectToFile2("./config/device.model/"+model.Id+".txt", &model); err != nil {
		return
	}

	return
}
