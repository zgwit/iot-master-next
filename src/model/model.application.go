package model

import (
	"errors"
	"fmt"

	"github.com/zgwit/iot-master-next/src/plugin"

	"github.com/zgwit/iot-master-next/src/utils"
)

type ApplicationType struct {
	Name    string `form:"name" bson:"name" json:"name"`
	Enable  bool   `form:"enable" bson:"enable" json:"enable"`
	Type    string `form:"type" bson:"type" json:"type"`
	Command string `form:"command" bson:"command" json:"command"`

	program plugin.Program
}

func ReadApplicationConfig(applications *[]ApplicationType) (err error) {

	var (
		labels []string
	)

	if labels, err = utils.GetDirsNames2("./application"); err != nil {
		return
	}

	for _, label := range labels {

		application := ApplicationType{}

		if err = application.loading(label); err != nil {
			return
		}

		*applications = append(*applications, application)
	}

	return
}

func ApplicationMaxNameLength(applications *[]ApplicationType, max_name_length *int) {

	for index := range *applications {
		if len((*applications)[index].Name) > *max_name_length {
			*max_name_length = len((*applications)[index].Name)
		}
	}
}

func ApplicationInfo(applications *[]ApplicationType, name_max_length int) (info string) {

	for index := range *applications {

		info += fmt.Sprintf("%-*s%-*t%-*s%-*d \n",
			name_max_length+6, (*applications)[index].Name,
			name_max_length, (*applications)[index].Enable,
			name_max_length, (*applications)[index].program.GetStatename(),
			name_max_length, (*applications)[index].program.GetRestartCount(),
		)
	}

	return
}

func (application *ApplicationType) Run() {

	application.program.Runing(
		"./application/"+application.Name,
		application.Command,
		"./log/application."+application.Name+".txt",
	)
}

func (application *ApplicationType) Update(application_param ApplicationType) (err error) {

	if err = utils.WriteObjectToFile2("./application/"+application.Name+"/application.config.txt", &application_param); err != nil {
		return
	}

	if application.Enable && !application_param.Enable {

		if !application.program.Exiting(30) {
			return errors.New("等待超时")
		}
	}

	if err = application.loading(application.Name); err != nil {
		return
	}

	application.Run()

	return
}

func (application *ApplicationType) Start() {

	application.program.Switch = true
}

func (application *ApplicationType) Stop() {

	application.program.Switch = false
}

func (application *ApplicationType) Status() map[string]interface{} {

	return application.program.GetStatus()
}

func (application *ApplicationType) Log(row int) (contents []string, err error) {

	return application.program.GetLog("./log/application."+application.Name+".txt", row)
}

func (application *ApplicationType) CheckFormat() (err_str string) {

	if len(application.Command) == 0 {

		err_str = "运行指令不能为空"
	}

	return
}

func (application *ApplicationType) loading(name string) (err error) {

	if err = utils.ReadFileToObject("./application/"+name+"/application.config.txt", &application); err != nil {
		return
	}

	if name == "" {
		return errors.New("./application/" + name + "/application.config.txt: name format is wroing")
	}

	if len(application.Command) == 0 {
		return errors.New("./application/" + name + "/application.config.txt: command format is wroing")
	}

	return
}
