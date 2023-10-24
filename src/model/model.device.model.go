package zgwit_model

import (
	zgwit_utils "local/utils"
	"strings"
)

type Model struct {
	Name string `form:"name" bson:"name" json:"name"`
	Id   string `form:"id" bson:"id" json:"id"`

	Attributes map[string]ModelAttribute `form:"attributes" bson:"attributes" json:"attributes"`
	Events     map[string]ModelEvent     `form:"events" bson:"events" json:"events"`
	Actions    map[string]ModelAction    `form:"actions" bson:"actions" json:"actions"`

	KeepAlive int64  `form:"keep_alive" bson:"keep_alive" json:"keep_alive"`
	Script    string `form:"script" bson:"script" json:"script"`

	Drive       string      `form:"drive" bson:"drive" json:"drive"`
	DriveConfig interface{} `form:"drive_config" bson:"drive_config" json:"drive_config"`
}

type ModelAttribute struct {
	Name        string `form:"name" bson:"name" json:"name"`
	DataType    string `form:"data_type" bson:"data_type" json:"data_type"`
	LastestOnly bool   `form:"lastest_only" bson:"lastest_only" json:"lastest_only"`
}

type ModelEvent struct {
	Name string `form:"name" bson:"name" json:"name"`
}

type ModelAction struct {
	Name string `form:"name" bson:"name" json:"name"`
}

func ModelFetchList() (models []Model, err error) {

	var (
		labels []string
	)

	models = []Model{}

	if labels, err = zgwit_utils.GetDirFileNames2("./config/device.model"); err != nil {
		return
	}

	for _, label := range labels {

		label := strings.Replace(label, ".txt", "", -1)

		model := Model{}

		if err = zgwit_utils.ReadFileToObject("./config/device.model/"+label+".txt", &model); err != nil {
			continue
		}

		models = append(models, model)
	}

	return
}

func ModelFind(model_id string, model *Model) (err error) {

	return zgwit_utils.ReadFileToObject("./config/device.model/"+model_id+".txt", model)
}

func ModelExist(model_id string) (result bool) {

	return zgwit_utils.FileExist("./config/device.model/" + model_id + ".txt")
}

func (model Model) Create() (err error) {

	if err = zgwit_utils.WriteObjectToFile2("./config/device.model/"+model.Id+".txt", &model); err != nil {
		return
	}

	return
}
