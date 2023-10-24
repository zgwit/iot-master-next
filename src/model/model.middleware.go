package model

import (
	"errors"
	"fmt"
	"time"

	"github.com/zgwit/iot-master-next/src/plugin"

	"github.com/zgwit/iot-master-next/src/utils"
)

type MiddlewareType struct {
	Name    string `form:"name" bson:"name" json:"name"`
	Enable  bool   `form:"enable" bson:"enable" json:"enable"`
	Command string `form:"command" bson:"command" json:"command"`

	program plugin.Program
}

func ReadMiddlewareConfig(middlewares *[]MiddlewareType) (err error) {

	var (
		labels []string
	)

	if labels, err = utils.GetDirsNames2("./middleware"); err != nil {
		return
	}

	for _, label := range labels {

		middleware := MiddlewareType{}

		if err = middleware.loading(label); err != nil {
			return
		}

		*middlewares = append(*middlewares, middleware)
	}

	return
}

func MiddlewareReady(middlewares *[]MiddlewareType) {

	for {
		for index := range *middlewares {
			if (*middlewares)[index].Enable && !(*middlewares)[index].program.Isrun {
				goto WAIT
			}
		}

		break

	WAIT:
		time.Sleep(time.Second)
	}
}

func MiddlewareMaxNameLength(middlewares *[]MiddlewareType, max_name_length *int) {

	for index := range *middlewares {
		if len((*middlewares)[index].Name) > *max_name_length {
			*max_name_length = len((*middlewares)[index].Name)
		}
	}
}

func MiddlewareInfo(middlewares *[]MiddlewareType, name_max_length int) (info string) {

	for index := range *middlewares {

		info += fmt.Sprintf("%-*s%-*t%-*s%-*d \n",
			name_max_length+6, (*middlewares)[index].Name,
			name_max_length, (*middlewares)[index].Enable,
			name_max_length, (*middlewares)[index].program.GetStatename(),
			name_max_length, (*middlewares)[index].program.GetRestartCount(),
		)
	}

	return
}

func (middleware *MiddlewareType) Run() {

	middleware.program.Runing(
		"./middleware/"+middleware.Name,
		middleware.Command,
		"./log/middleware."+middleware.Name+".txt",
	)
}

func (middleware *MiddlewareType) Update(middleware_param MiddlewareType) (err error) {

	if err = utils.WriteObjectToFile2("./middleware/"+middleware.Name+"/middleware.config.txt", &middleware_param); err != nil {
		return
	}

	if middleware.Enable && !middleware_param.Enable {

		if !middleware.program.Exiting(30) {
			return errors.New("等待超时")
		}
	}

	if err = middleware.loading(middleware.Name); err != nil {
		return
	}

	if middleware_param.Enable {
		middleware.Run()
	}

	return
}

func (middleware *MiddlewareType) Start() {

	middleware.program.Switch = true
}

func (middleware *MiddlewareType) Stop() {

	middleware.program.Switch = false
}

func (middleware *MiddlewareType) Status() map[string]interface{} {

	return middleware.program.GetStatus()
}

func (middleware *MiddlewareType) Log(row int) (contents []string, err error) {

	return middleware.program.GetLog("./log/middleware."+middleware.Name+".txt", row)
}

func (middleware *MiddlewareType) CheckFormat() (err_str string) {

	if len(middleware.Command) == 0 {

		err_str = "运行指令不能为空"
	}

	return
}

func (middleware *MiddlewareType) loading(name string) (err error) {

	if err = utils.ReadFileToObject("./middleware/"+name+"/middleware.config.txt", &middleware); err != nil {
		return
	}

	if name == "" {
		return errors.New("./middleware/" + name + "/middleware.config.txt: name format is wroing")
	}

	if len(middleware.Command) == 0 {
		return errors.New("./middleware/" + name + "/middleware.config.txt: command format is wroing")
	}

	return
}
