package config

import (
	"reflect"
	"strings"

	"github.com/jansemmelink/log"
)

//NewData ...
func NewData(v interface{}) IData {
	return data{
		value: v,
	}
}

//IData ...
type IData interface {
	Validate() error
	Get(name string) interface{}
}

type data struct {
	value interface{}
}

func (d data) Validate() error {
	return nil
}

func (d data) Get(name string) interface{} {
	//log.Debugf("data.Get(name=\"%s\")", name)
	if name == "" {
		return d.value
	}

	from := d.value

	//skip over optional leading '.' then split name on the first '.':
	for len(name) > 0 && name[0:1] == "." {
		name = name[1:]
	}
	names := strings.SplitN(name, ".", 2)
	/*log.Debugf("%T.value=%T.get(%s) -> %d: %+v",
	d,
	d.value,
	name,
	len(names), names)*/
	if len(names) < 1 {
		panic(log.Wrapf(nil, "%T.get(%s): Bad name?", d, name))
	}
	selector := names[0]
	restOfName := ""
	if len(names) > 1 {
		restOfName = names[1]
	}
	//log.Debugf("selector=\"%s\" rest=\"%s\"", selector, restOfName)

	tt := reflect.TypeOf(from).Elem()
	if tt.Kind() == reflect.Map && tt.Key().Kind() == reflect.String {
		//select the named item
		fromValue := reflect.ValueOf(from).Elem()
		selectedValue := fromValue.MapIndex(reflect.ValueOf(selector))
		log.Debugf("Selected map item[%s]: %v", selector, selectedValue)
		return selectedValue.Interface()
	}

	switch from.(type) {
	case map[string]interface{}, *map[string]interface{}:
		//log.Debugf("object...")
		//value is an object with named-items: select by name
		mapValue := from.(map[string]interface{})
		selected, ok := mapValue[selector]
		if !ok {
			return nil
		}
		return NewData(selected).Get(restOfName)

	case []interface{}:
		log.Debugf("list...")
		//value is a list of items: select by index

		//listValue := from.([]interface{})
		//value := listValue[names[0]]

		panic("NYI: name=\"" + name + "\"")
	default:
		log.Debugf("other %T...", from)
		//flat value, cannot select
	}
	return nil
}
