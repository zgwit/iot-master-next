package utils

import (
	"bytes"
	"encoding/json"
)

func ToJson(object interface{}) (result string, err error) {

	var (
		strByte []byte
	)

	if strByte, err = json.Marshal(object); err != nil {
		return
	}

	result = string(strByte)

	return
}

func ToJson2(object interface{}) (result string) {

	var (
		err     error
		strByte []byte
	)

	if strByte, err = json.Marshal(object); err != nil {
		return
	}

	return string(strByte)
}

func ToJson3(object interface{}) (result string) {

	var (
		err        error
		strByte    []byte
		strOutByte bytes.Buffer
	)

	if strByte, err = json.Marshal(object); err != nil {
		return
	}

	if err = json.Indent(&strOutByte, strByte, "", "    "); err != nil {
		result = string(strByte)
	} else {
		result = strOutByte.String()
	}

	return result
}

func GetNumberBool(value bool) float64 {
	if value {
		return 1
	} else {
		return 0
	}
}
