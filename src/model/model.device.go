package zgwit_model

import (
	zgwit_utils "local/utils"
	"strings"
)

type Device struct {
	Name    string `form:"name" bson:"name" json:"name"`
	Id      string `form:"id" bson:"id" json:"id"`
	ModelId string `form:"model_id" bson:"model_id" json:"model_id"`

	Drive       string      `form:"drive" bson:"drive" json:"drive"`
	DriveConfig interface{} `form:"drive_config" bson:"drive_config" json:"drive_config"`
}

func DeviceFetchList(model_id *string) (devices []Device, err error) {

	var (
		labels []string
	)

	devices = []Device{}

	if labels, err = zgwit_utils.GetDirFileNames2("./config/device.model"); err != nil {
		return
	}

	for _, label := range labels {

		label := strings.Replace(label, ".txt", "", -1)

		device := Device{}

		if err = zgwit_utils.ReadFileToObject("./config/device/"+label+".txt", &device); err != nil {
			continue
		}

		if device.ModelId == *model_id {
			continue
		}

		devices = append(devices, device)
	}

	return
}
