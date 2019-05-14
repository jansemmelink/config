package config

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/jansemmelink/log"
)

//NewList creates a new JSON list, which is stored as []interface{}
func NewList() []interface{} {
	return make([]interface{}, 0)
}

//GetObjField retrieves a named field, using dot-notation for sub fields
//name must be either:
//		simple name without any dots/indices, e.g. "abc"
//      dot notation for sub object fields, e.g. "abc.def"
//		indexed for single list item, e.g. "abc[1]"  (index=0,1,2,...N-1)
//      multi-level index also supported for list item, e.g. abc[0][1]
//      wild card can be used if object has only one field (e.g. in union), e.g. "abc.*.def"
//      list output abc[] for all items in abc can be used on lists/objects
//				(when used on object, order of values must not be assumed, although seems to be sorted by field name)
func GetObjField(obj map[string]interface{}, name string) (interface{}, error) {
	//skip over optional leading '.'
	if len(name) > 0 && name[0:1] == "." {
		name = name[1:]
	}

	names := strings.SplitN(name, ".", 2)
	//log.Debugf("GetField(%s) -> %d: %+v", name, len(names), names)
	if len(names) < 1 {
		return nil, log.Wrapf(nil, "Bad name \"%s\"", name)
	}

	//if name is "*", and this is an object with just one field, take that field
	if names[0] == "*" {
		if len(obj) != 1 {
			fieldNames := make([]string, 0)
			for n := range obj {
				fieldNames = append(fieldNames, n)
			}
			return nil, log.Wrapf(nil, "Field %s cannot be used in obj with %d fields: %v",
				name, len(obj), fieldNames)
		}
		//use the first field
		for n := range obj {
			names[0] = n
			log.Debugf("Replacing '*' with %s", names[0])
		}
	} //if '*'

	//see if array (e.g. "abc[1]") or simple name (e.g. "abc")
	f := func(c rune) bool {
		return c == '[' || c == ']' //!unicode.IsLetter(c) && !unicode.IsNumber(c)
	}
	index := strings.FieldsFunc(names[0], f)
	//cater for if name ends in [], above won't detect it, so add manually
	{
		offset := 0
		count := 0
		for foundAt := strings.Index(names[0][offset:], "[]"); foundAt >= 0; foundAt = strings.Index(names[0][offset:], "[]") {
			index = append(index, "*")
			offset = foundAt + 1
			count++
		}
		if count > 1 {
			return nil, log.Wrapf(nil, "list directive '[]' used multiple times in name=%s", name)
		}
		if count > 0 && len(names) > 1 {
			return nil, log.Wrapf(nil, "list directive '[]' used before last element in %s", name)
		}
	}

	if len(index) < 1 {
		return nil, log.Wrapf(nil, "Invalid name \"%s\" in %s", names[0], name)
	}
	/*if len(index) == 1 {
		log.Debugf("Simple name: \"%s\"", names[0])
	} else {
		log.Debugf("Array (parts=%d) name: \"%s\" index=\"%s\"", len(index), index[0], index[1])
	}*/

	parent := ""
	sub, ok := obj[index[0]]
	if !ok {
		return nil, log.Wrapf(nil, "%s%s undefined", parent, index[0])
	}

	//step into array lists
	if len(index) > 1 {
		listName := index[0]
		for len(index) > 1 {
			log.Debugf("Getting %s[%s]", index[0], index[1])
			list, ok := sub.([]interface{})
			if !ok {
				//not a list: handle * because specified '[]'
				if index[1] == "*" {
					subObj, ok := sub.(map[string]interface{})
					if !ok {
						//neither an object nor a list, return value ignoring '[]'
						return sub, nil
					}
					//object: return all object field values because specified '[]'
					values := make([]interface{}, 0)
					for _, v := range subObj {
						values = append(values, v)
					}
					return values, nil
				}
				//name has index but item is not a list and does not support [] either:
				return nil, log.Wrapf(nil, "Not a list: %s", name)
			}

			if index[1] == "*" {
				//return all list items (when specified "name[]")
				return sub, nil
			} //if all list items

			//not '*', expect specific index value
			//todo: not yet supporting ranges e.g. [0:1]
			indexInt, err := strconv.Atoi(index[1])
			if err != nil {
				return nil, log.Wrapf(err, "non-integer index=\"%s\"", index[1])
			}
			if indexInt < 0 {
				return nil, log.Wrapf(nil, "negative index=\"%s\"", index[1])
			}
			if indexInt >= len(list) {
				return nil, log.Wrapf(nil, "index=\"%s\" out of range 0..%d", index[1], len(list)-1)
			}
			sub = list[indexInt]

			//append the index to the list name for subsequent logging
			listName = fmt.Sprintf("%s[%d]", listName, indexInt)

			//skip over it
			index = append(index[0:1], index[2:]...)
			log.Debugf("Now index=%+v", index)
		}
	}

	if len(names) == 1 {
		return sub, nil
	}

	//need to go deeper
	log.Debugf("Go Deeper: %+v", names[1])
	switch sub.(type) {
	case map[string]interface{}:
		var err error
		sub, err = GetObjField(sub.(map[string]interface{}), names[1])
		return sub, err
	default:
		return nil, log.Wrapf(nil, "Cannot go deeper into %T for %+v", sub, names[1])
	}
} //Object.GetField()

//GetObjTime gets the named field(s) and parse string to time using the specified format(s)
func GetObjTime(obj map[string]interface{}, names []string, formats []string) (time.Time, error) {
	//look for any of the specified names, using the first found
	var value interface{}
	value = nil
	var name string
	for _, name = range names {
		var err error
		value, err = GetObjField(obj, name)
		if err == nil && value != nil {
			break
		}
	}
	if value == nil {
		return time.Time{}, log.Wrapf(nil, "Not found any of %v", names)
	}

	//return time value if available, or convert to string so we can parse a time value
	var stringValue string
	switch value.(type) {
	case time.Time:
		return value.(time.Time), nil
	case string:
		stringValue = value.(string)
	default:
		stringValue = fmt.Sprintf("%v", value)
	}

	//parse timestamp in any of the expected formats
	var valueTime time.Time
	for _, f := range formats {
		var err error
		valueTime, err = time.Parse(f, stringValue)
		if err == nil {
			return valueTime, nil
		}
	}
	return time.Time{}, log.Wrapf(nil, "Time \"%s\" cannot be parsed with formats=%v", stringValue, formats)
} //Object.GetObjTime()
