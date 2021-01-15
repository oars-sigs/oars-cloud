package admin

import (
	"encoding/json"
	"errors"
	"reflect"
	"regexp"
)

//unmarshalArgs 解析参数
func unmarshalArgs(in, out interface{}) error {
	if in == nil {
		return nil
	}
	inType := reflect.TypeOf(in)
	outType := reflect.TypeOf(out)
	if inType == outType {
		reflect.ValueOf(out).Elem().Set(reflect.ValueOf(in).Elem())
		return nil
	}
	inpType := reflect.TypeOf(&in)
	if inpType == outType {
		out = &in
		return nil
	}
	if inType.String() == "float64" {
		inv := in.(float64)
		switch outType.String() {
		case "int":
			v := int(inv)
			out = &v
		case "int32":
			v := int32(inv)
			out = &v
		case "int64":
			v := int64(inv)
			out = &v
		default:
			return errors.New("float64 can not change to " + outType.String())
		}
		return nil
	}

	if inType.String() == "int64" {
		inv := in.(int64)
		switch outType.String() {
		case "int":
			v := int(inv)
			out = &v
		case "int32":
			v := int32(inv)
			out = &v
		default:
			return errors.New("int64 can not change to " + outType.String())
		}
		return nil
	}
	data, err := json.Marshal(in)
	if err != nil {
		return err
	}
	err = json.Unmarshal(data, out)
	return err
}

const (
	nameRegexString = `^[a-zA-Z]([a-zA-Z0-9\-]+)*[a-zA-Z0-9]$`
)

var (
	nameRegex = regexp.MustCompile(nameRegexString)
)
