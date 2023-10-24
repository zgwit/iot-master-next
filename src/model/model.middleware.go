package zgwit_model

import (
	"errors"
	"fmt"
	zgwit_plugin "local/plugin"
	zgwit_utils "local/utils"
	"time"
)

type Middleware struct {
	Name    string `form:"name" bson:"name" json:"name"`
	Enable  bool   `form:"enable" bson:"enable" json:"enable"`
	Command string `form:"command" bson:"command" json:"command"`

	program zgwit_plugin.Program
}

func ReadMiddlewareConfig(middlewares *[]Middleware) (err error) {

	var (
		labels []string
	)

	if labels, err = zgwit_utils.GetDirsNames2("./middleware"); err != nil {
		return
	}

	for _, label := range labels {

		middleware := Middleware{}

		if err = middleware.loading(label); err != nil {
			return
		}

		*middlewares = append(*middlewares, middleware)
	}

	return
}

func MiddlewareReady(middlewares *[]Middleware) {

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

func MiddlewareMaxNameLength(middlewares *[]Middleware, max_name_length *int) {

	for index := range *middlewares {
		if len((*middlewares)[index].Name) > *max_name_length {
			*max_name_length = len((*middlewares)[index].Name)
		}
	}
}

func MiddlewareInfo(middlewares *[]Middleware, name_max_length int) (info string) {

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

func (middleware *Middleware) Run() {

	middleware.program.Runing(
		"./middleware/"+middleware.Name,
		middleware.Command,
		"./log/middleware."+middleware.Name+".txt",
	)
}

func (middleware *Middleware) Update(middleware_param Middleware) (err error) {

	if err = zgwit_utils.WriteObjectToFile2("./middleware/"+middleware.Name+"/middleware.config.txt", &middleware_param); err != nil {
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

func (middleware *Middleware) Start() {

	middleware.program.Switch = true
}

func (middleware *Middleware) Stop() {

	middleware.program.Switch = false
}

func (middleware *Middleware) Status() map[string]interface{} {

	return middleware.program.GetStatus()
}

func (middleware *Middleware) Log(row int) (contents []string, err error) {

	return middleware.program.GetLog("./log/middleware."+middleware.Name+".txt", row)
}

func (middleware *Middleware) CheckFormat() (err_str string) {

	if len(middleware.Command) == 0 {

		err_str = "运行指令不能为空"
	}

	return
}

func (middleware *Middleware) loading(name string) (err error) {

	if err = zgwit_utils.ReadFileToObject("./middleware/"+name+"/middleware.config.txt", &middleware); err != nil {
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
