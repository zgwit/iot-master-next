package plugin

import (
	"errors"

	"github.com/dop251/goja"
)

type Javascript struct {
	Client *goja.Runtime
}

type JavaScriptResult goja.Value

func NewJavascript() (js Javascript) {
	js.Client = goja.New()
	return
}

func (js *Javascript) Function(func_name, code string, func_params ...interface{}) (result JavaScriptResult, err error) {

	var (
		ok     bool
		handle goja.Callable
		params = []goja.Value{}
	)

	for _, param := range func_params {
		params = append(params, js.Client.ToValue(param))
	}

	if _, err = js.Client.RunString(code); err != nil {
		return
	}

	if handle, ok = goja.AssertFunction(js.Client.Get(func_name)); !ok {
		goto FAIL
	}

	if result, err = handle(goja.Undefined(), params...); err != nil {
		return
	}

	return
FAIL:
	err = errors.New("格式错误")
	return
}
