package config

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	//"bitbucket.org/vservices/etl/lib-etl/binary"

	"github.com/jansemmelink/log"
)

//Worker is a container to do consecutive operations on the same source data
//Worker reduces code effort by avoiding errors checks after each operation
//After the first failure, subsequent steps are skipped
//At the end of your function, you only check the error at the end
type Worker struct {
	data interface{}
	ref  string
	root *Item
	err  error
}

//WorkWith starts working with a piece of decoded (JSON) data which must be either a object or a list
func WorkWith(data interface{}, ref string) *Worker {
	w := &Worker{
		data: data,
		ref:  ref,
		err:  nil,
	}

	if data == nil {
		w.err = log.Wrapf(nil, "WorkWith(%s:data==nil)", ref)
		return w
	}

	//make root item
	w.root, w.err = newItem(w, nil, "", &w.data)
	return w
} //WorkWith()

//Root returns the root data item specified in WorkWith converted to a json.Data item
func (w Worker) Root() *Item {
	return w.root
} //Worker.Root()

//Result return nil on success or error on failure
func (w Worker) Result() error {
	return w.err
} //Worker.Result()

//Errorf adds an error to the worker error stack
func (w *Worker) Error(err error) {
	// first := false
	// if w.err == nil {
	// first = true
	// }
	w.err = log.Wrapf(w.err, "%v", err)
	//if first {
	// 	log.Errorf("%s: FIRST ERROR: %v", w.ref, w.err)
	// }
} //Worker.Errorf()

//ItemType indicate the type of JSON item
type ItemType int

//unique identifiers for JSON data types
const (
	typeNull ItemType = iota
	typeBool
	typeString
	typeNumber
	typeObject
	typeList
)

func (it ItemType) String() string {
	return [...]string{"Null", "Bool", "String", "Number", "Object", "List"}[it]
} //ItemType.String()

//Item is a named reference to a point in the JSON item being used in the worker
type Item struct {
	//all items in a worker has this reference back to the worker
	//so that it can set the error if an operation fails, and do nothing
	//if the worker already failed!
	worker *Worker

	//parent is nil only for the root item in the worker
	parent *Item

	//name is "" for root, and relative to parent using dot-notation
	//so get full name by printing parent.Name()+"."+name
	name string

	//data pointer to this item somewhere in the worker.root
	dataPtr *interface{}

	//JSON type identified for this piece of data
	itemType ItemType

	//count is nr of entries when this Item is a list, else 0
	count int
}

//newItem makes a new item to refer to a point in the data
func newItem(w *Worker, p *Item, n string, d *interface{}) (*Item, error) {
	i := Item{
		worker:  w,
		parent:  p,
		name:    n,
		dataPtr: d,
	}

	i.setItemType()
	if err := i.setItemType(); err != nil {
		return nil, err
	}

	//log.Debugf("CREATED %p.%v", i.parent, &i)
	return &i, nil
} //newItem()

func (i *Item) setItemType() error {
	if i == nil || i.dataPtr == nil {
		return nil
	}
	//oldType := i.itemType
	switch (*i.dataPtr).(type) {
	case nil:
		i.itemType = typeNull
	case bool:
		i.itemType = typeBool
	case int, float64:
		i.itemType = typeNumber
	case string:
		i.itemType = typeString
	case map[string]interface{}:
		i.itemType = typeObject
	case []interface{}:
		i.itemType = typeList
		listValue := (*i.dataPtr).([]interface{})
		i.count = len(listValue)
		//strip "[]" from end of name
		if i.name[len(i.name)-2:] == "[]" {
			i.name = i.name[:len(i.name)-2]
		}
	default:
		return log.Wrapf(nil, "%T cannot be used as a JSON item", *i.dataPtr)
	}
	/*if oldType != i.itemType {
		log.Debugf("%v: Updated ItemType from %v to %v", i, oldType, i.itemType)
	}*/
	return nil
} //Item.setItemType()

//Type returns the JSON ItemType
func (i *Item) Type() ItemType {
	if i == nil {
		return typeNull
	}
	return i.itemType
}

//Value returns the items' dereferenced value
func (i *Item) Value() interface{} {
	if i == nil || i.dataPtr == nil {
		return nil
	}
	return *i.dataPtr
} //Item.Value()

//FullName append own name to parent's name
func (i *Item) FullName() string {
	if i == nil {
		return "nil"
	}
	if i.parent == nil {
		return i.name
	}
	return i.parent.FullName() + "." + i.name
} //Item.FullName()

//Errorf adds an error to the item's worker error stack
func (i *Item) Errorf(format string, args ...interface{}) {
	if i != nil {
		i.worker.err = log.Wrapf(i.worker.err, format, args...)
	}
}

//AsObject returns the value as an object
func (i *Item) AsObject() map[string]interface{} {
	if i == nil {
		return nil
	}

	if i.itemType != typeObject {
		i.worker.Error(log.Wrapf(nil, "%s type %s is not an object", i.itemType, i.FullName()))
		return nil
	}
	if i.dataPtr == nil {
		return nil
	}
	return (*(i.dataPtr)).(map[string]interface{})
} //Item.AsObject()

//String returns the value as a string and converts to string if not one
//but will fail if called for object/list
func (i *Item) String() string {
	if i == nil {
		return "nil"
	}
	if i.itemType == typeObject || i.itemType == typeList {
		i.worker.Error(log.Wrapf(nil, "%s (%s) cannot convert to string", i.itemType, i.FullName()))
		return "N/A"
	}

	if i.dataPtr == nil {
		return ""
	}
	v := *(i.dataPtr)
	if s, ok := v.(string); ok {
		return s
	}
	return fmt.Sprintf("%v", v)
} //Item.String()

//OptInt returns pointer to int if present and valid
func (i *Item) OptInt() *int {
	if i == nil || i.dataPtr == nil {
		return nil
	}
	if i.itemType == typeObject || i.itemType == typeList {
		i.worker.Error(log.Wrapf(nil, "%v %s cannot convert to int", i.itemType, i.FullName()))
		return nil
	}
	v := *(i.dataPtr)
	stringValue := ""
	switch v.(type) {
	case int:
		intValue := v.(int)
		return &intValue
	case float64:
		intValue := int(v.(float64))
		return &intValue
	default:
		//not already number: try to convert
		stringValue = fmt.Sprintf("%v", v)
	} //switch(value type)

	//try to convert decimal string, e.g. "123" -> 123
	if intValue, err := strconv.Atoi(stringValue); err != nil {
		//log.Debugf("Cannot convert \"%s\" from dec to int: %v", stringValue, err)
	} else {
		return &intValue
	}

	//cannot convert to int
	i.worker.Error(log.Wrapf(nil, "%v %s=%s cannot convert to int", i.itemType, i.FullName(), stringValue))
	return nil
}

//Int returns the value as an int and converts to int if not one
//but will fail if called for object/list or non-integer item value
func (i *Item) Int() int {
	if i == nil || i.dataPtr == nil {
		return 0
	}
	if i.itemType == typeObject || i.itemType == typeList {
		i.worker.Error(log.Wrapf(nil, "%v %s cannot convert to int", i.itemType, i.FullName()))
		return 0
	}
	v := *(i.dataPtr)
	stringValue := ""
	switch v.(type) {
	case int:
		return v.(int)
	case float64:
		return int(v.(float64))
	default:
		//not already number: try to convert
		stringValue = fmt.Sprintf("%v", v)
	} //switch(value type)

	//try to convert decimal string, e.g. "123" -> 123
	if intValue, err := strconv.Atoi(stringValue); err != nil {
		//log.Debugf("Cannot convert \"%s\" from dec to int: %v", stringValue, err)
	} else {
		return intValue
	}

	//cannot convert to int
	i.worker.Error(log.Wrapf(nil, "%v %s=%s cannot convert to int", i.itemType, i.FullName(), stringValue))
	return 0
} //Item.Int()

//OptFloat64 returns the value as a float64 and converts to float64 if not one
//and return ok if valid, else false if cannot determine float value
func (i *Item) OptFloat64() (float64, bool) {
	if i == nil || i.dataPtr == nil {
		return 0, false
	}
	if i.itemType == typeObject || i.itemType == typeList {
		i.worker.Error(log.Wrapf(nil, "%v: %v %s cannot convert to float64", i, i.itemType, i.FullName()))
		return 0, false
	}

	v := *(i.dataPtr)
	switch v.(type) {
	case int:
		return float64(v.(int)), true
	case float64:
		return v.(float64), true
	case string:
		if v.(string) == "" {
			return 0.0, false
		}
		f, err := strconv.ParseFloat(v.(string), 64)
		if err != nil {
			i.worker.Error(log.Wrapf(nil, "%v %s=(%T)%v cannot convert to float64", i.itemType, i.FullName(), v, v))
		} else {
			return f, true
		}
	default:
		i.worker.Error(log.Wrapf(nil, "%v %s=(%T)%v cannot convert to float64", i.itemType, i.FullName(), v, v))
	}
	return 0, false
}

//Float64 returns the value as an float64 and converts to float64 if not one
//but will fail if called for object/list or non-number item value
func (i *Item) Float64() float64 {
	if i == nil || i.dataPtr == nil {
		return 0
	}
	if i.itemType == typeObject || i.itemType == typeList {
		i.worker.Error(log.Wrapf(nil, "%v: %v %s cannot convert to float64", i, i.itemType, i.FullName()))
		return 0
	}

	v := *(i.dataPtr)
	switch v.(type) {
	case int:
		return float64(v.(int))
	case float64:
		return v.(float64)
	case string:
		if v.(string) == "" {
			return 0.0
		}
		f, err := strconv.ParseFloat(v.(string), 64)
		if err != nil {
			i.worker.Error(log.Wrapf(nil, "%v %s=(%T)%v cannot convert to float64", i.itemType, i.FullName(), v, v))
		} else {
			return f
		}
	default:
		i.worker.Error(log.Wrapf(nil, "%v %s=(%T)%v cannot convert to float64", i.itemType, i.FullName(), v, v))
	}
	return 0
} //Item.Float64()

//TimePtr calls Time and if nil, return nil so that JSON can omit empty values
func (i *Item) TimePtr(parseFormat ...string) *time.Time {
	t := i.Time(parseFormat...)
	null := time.Time{}
	if t == null {
		return nil
	}
	return &t
}

//OptTimestamp calls Time and if nil, return nil so that JSON can omit empty values
func (i *Item) OptTimestamp(parseFormat ...string) *Timestamp {
	t := i.Time(parseFormat...)
	null := time.Time{}
	if t == null {
		return nil
	}
	return &Timestamp{Time: t}
}

//OptDate calls Time and if nil, return nil so that JSON can omit empty values
func (i *Item) OptDate(parseFormat ...string) *Date {
	t := i.Time(parseFormat...)
	null := time.Time{}
	if t == null {
		return nil
	}
	return &Date{Time: t}
}

//Time returns the value as a time.Time value, converting to time if not already one
//but will fail if called for object/list or non-integer item value
//parseFormat can be "" to use default format
func (i *Item) Time(parseFormat ...string) time.Time {
	if i == nil || i.dataPtr == nil {
		return time.Time{}
	}
	if i.itemType == typeObject || i.itemType == typeList {
		i.worker.Error(log.Wrapf(nil, "%s %s cannot convert to time", i.itemType, i.FullName()))
		return time.Time{}
	}

	v := *(i.dataPtr)
	if timeValue, ok := v.(time.Time); ok {
		return timeValue
	}

	//not already time: try to convert
	stringValue := fmt.Sprintf("%v", v)
	if len(parseFormat) == 0 {
		parseFormat = []string{"2006-01-02T15:04:05-0700"}
	}

	//try all specified formats:
	errors := ""
	for _, format := range parseFormat {
		timeValue, err := time.Parse(format, stringValue)
		if err == nil {
			return timeValue
		}
		errors += fmt.Sprintf("\ncannot parse as time(%s)", format)
	}
	i.worker.Error(log.Wrapf(nil, "%s %s=%s parsing failed: %s", i.itemType, i.FullName(), stringValue, errors))
	return time.Time{}
} //Item.Time()

//Bool returns the value as a bool and converts to true|false if not one
//but will fail if called for anything not true|false|"true"|"false"|1|0
func (i *Item) Bool() bool {
	if i == nil {
		return false
	}
	if i.itemType == typeObject || i.itemType == typeList {
		i.worker.Error(log.Wrapf(nil, "%s (%s) cannot convert to bool", i.itemType, i.FullName()))
		return false
	}

	if i.dataPtr == nil {
		return false
	}
	v := *(i.dataPtr)
	switch v.(type) {
	case bool:
		return v.(bool)
	case int:
		i := v.(int)
		if i != 0 {
			return true
		}
	case string:
		s := v.(string)
		if s == "true" || s == "1" {
			return true
		}
	default:
		i.worker.Error(log.Wrapf(nil, "%v %s=(%T)%v cannot convert to bool", i.itemType, i.FullName(), v, v))
	}
	return false
} //Item.Bool()

func (i *Item) subItem(n string, d *interface{}) *Item {
	if i == nil {
		return nil
	}
	ni, err := newItem(i.worker, i, n, d)
	if err != nil {
		i.worker.Error(log.Wrapf(err, "Failed to create new sub item"))
		return nil
	}
	return ni
} //Item.subItem()

//Get specified sub item and fail if absent
func (i *Item) Get(name string) *Item {
	if i == nil {
		return nil
	}
	return i.get(name, true)
}

//Has gets specified sub item and return nil when absent
func (i *Item) Has(name ...string) *Item {
	if i == nil {
		return nil
	}
	for _, n := range name {
		found := i.get(n, false)
		if found != nil {
			return found
		}
	}
	return nil
}

//Get this (with name="") or named sub item, fail if required and absent, else return nil
//if not able to get it, return nilItem (not go value nil) else for x := range item()
func (i *Item) get(name string, required bool) *Item {
	if i == nil || i.worker.err != nil {
		return nil
	}

	if name == "" {
		return i //no relative name = return this item
	}

	//skip over optional leading '.'
	if name[0:1] == "." {
		name = name[1:]
	}

	//when return, append error to worker's error stack (only if required)
	var err error
	defer func() {
		if err != nil {
			if required {
				names := make([]string, 0)
				for n := range (*i.dataPtr).(map[string]interface{}) {
					names = append(names, n)
				}
				i.worker.Error(log.Wrapf(err, "%v: Failed to get required %s from %v", i, name, names))
			} /* else {
				log.Debugf("%s is absent", name)
			}*/
		}
	}()

	names := strings.SplitN(name, ".", 2)
	if len(names) < 1 {
		err = log.Wrapf(nil, "Bad name \"%s\"", name)
		return nil
	}

	if i.dataPtr == nil {
		err = log.Wrapf(nil, "No data for %s", name)
		return nil
	}

	obj, ok := (*(i.dataPtr)).(map[string]interface{})
	if !ok {
		//not an object
		err = log.Wrapf(nil, "Cannot get(%s) from %v %s which is not an object: %T", name, i.itemType, i.FullName(), *(i.dataPtr))
		return nil
	}

	//if name is "*", and this is an object with just one field, take that field
	if names[0] == "*" && i.itemType == typeObject {
		if len(obj) != 1 {
			//obj has >1 fields, format nice error message:
			fieldNames := make([]string, 0)
			for n := range obj {
				fieldNames = append(fieldNames, n)
			}
			err = log.Wrapf(nil, "Field %s cannot be used in obj with %d fields: %v", name, len(obj), fieldNames)
			return nil
		} //if more than one field

		//use the one and only field
		for n := range obj {
			names[0] = n
			log.Debugf("Replacing '*' with %s", names[0])
		}
	} //if used wildcard field '*' on object

	//see if requested a list of items or a specific list item
	// (e.g. "abc[1]") or simple name (e.g. "abc")
	f := func(c rune) bool {
		return c == '[' || c == ']' //!unicode.IsLetter(c) && !unicode.IsNumber(c)
	}
	index := strings.FieldsFunc(names[0], f)

	//cater for case when name ends in [], above won't detect it, so add manually
	{
		offset := 0
		count := 0
		for foundAt := strings.Index(names[0][offset:], "[]"); foundAt >= 0; foundAt = strings.Index(names[0][offset:], "[]") {
			index = append(index, "*")
			offset = foundAt + 1
			count++
		}
		if count > 1 {
			err = log.Wrapf(nil, "list directive '[]' used multiple times in name=%s", name)
			return nil
		}
		if count > 0 && len(names) > 1 {
			err = log.Wrapf(nil, "list directive '[]' used before last element in %s", name)
			return nil
		}
	}

	if len(index) < 1 {
		err = log.Wrapf(nil, "Invalid name \"%s\" in %s", names[0], name)
		return nil
	}

	sub, ok := obj[index[0]]
	if !ok {
		names := make([]string, 0)
		for n := range obj {
			names = append(names, n)
		}
		err = log.Wrapf(nil, "%s.map[%s] is undefined. It has %+v", i.FullName(), index[0], names)
		return nil
	}

	//step into array lists
	if len(index) > 1 {
		listName := index[0]
		for len(index) > 1 {
			//log.Debugf("Getting %s[%s]", index[0], index[1])
			list, ok := sub.([]interface{})
			if !ok {
				//not a list: handle * because specified '[]'
				if index[1] == "*" {
					subObj, ok := sub.(map[string]interface{})
					if !ok {
						//neither an object nor a list, return item with this value, ignoring '[]'
						return i.subItem(name, &sub)
					}
					//object: return all object field values because specified '[]'
					values := make([]interface{}, 0)
					for _, v := range subObj {
						values = append(values, v)
					}
					var v interface{}
					v = values
					return i.subItem(name, &v)
				}

				//name has index but item is not a list and does not support [] either:
				err = log.Wrapf(nil, "Not a list: %s", name)
				return nil
			} //if list

			if index[1] == "*" {
				//return all list items (when specified "name[]")
				return i.subItem(name, &sub)
			} //if all list items

			//not '*', expect specific index value
			//todo: not yet supporting ranges e.g. [0:1]
			indexInt, err := strconv.Atoi(index[1])
			if err != nil {
				err = log.Wrapf(err, "non-integer index=\"%s\"", index[1])
				return nil
			}
			if indexInt < 0 {
				err = log.Wrapf(nil, "negative index=\"%s\"", index[1])
				return nil
			}
			if indexInt >= len(list) {
				err = log.Wrapf(nil, "index=\"%s\" out of range 0..%d", index[1], len(list)-1)
				return nil
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
		err = nil
		return i.subItem(name, &sub)
	}

	//need to go deeper
	return i.subItem(names[0], &sub).get(names[1], required)
} //Item.get()

//Count returns the nr of items in a list or fail if not a list
func (i *Item) Count() int {
	if i == nil || i.worker.err != nil {
		return 0
	}
	if i.itemType != typeList {
		i.worker.Error(log.Wrapf(nil, "Cannot use Count() on %s %s (not a list)", i.itemType, i.FullName()))
		return 0
	}
	return i.count
} //Item.Count()

//ListItem returns the indexed=0,1,2,...,Count()-1 list item
func (i *Item) ListItem(index int) *Item {
	if i == nil || i.worker.err != nil {
		return nil
	}
	if i.itemType != typeList {
		i.worker.Error(log.Wrapf(nil, "Cannot use ListItem() on %s %s (not a list)", i.itemType, i.FullName()))
		return nil
	}
	if index < 0 || index >= i.count {
		i.worker.Error(log.Wrapf(nil, "%s.ListItem(index=%d) out of range 0..%d", i.FullName(), index, i.count-1))
		return nil
	}
	listItems := (*(i.dataPtr)).([]interface{})
	return i.subItem(fmt.Sprintf("[%d]", index), &listItems[index])
} //Item.ListItem()

//Select an object where the named field has the specified value
func (i *Item) Select(name string, value interface{}, optional bool) *Item {
	if i == nil || i.worker.err != nil {
		return nil
	}
	if i.itemType != typeList {
		i.worker.Error(log.Wrapf(nil, "Cannot select from %s %s (not a list)", i.itemType, i.FullName()))
		return nil
	}

	//skip over optional leading '.'
	if name[0:1] == "." {
		name = name[1:]
	}

	countKeyFieldPresent := 0
	//log.Debugf("%v.Select(name=\"%s\")...", i, name)
	listOfValues := (*(i.dataPtr)).([]interface{})
	for index, listEntry := range listOfValues {
		//log.Debugf("%v[%d]:%+v", i, index, listEntry)
		obj, ok := listEntry.(map[string]interface{})
		if !ok {
			//list entry is not a JSON object - will not be able to get key field in non-object
			continue
		}

		//list entry is a JSON object: get the key field
		entryKeyFieldValue, err := GetObjField(obj, name)
		if err != nil {
			//key field not found
			continue
		}

		//got the key field from the list entry
		countKeyFieldPresent++

		//compare key value
		if reflect.DeepEqual(value, entryKeyFieldValue) {
			//found match - return the list entry
			si := i.subItem(fmt.Sprintf("[%d]", index), &listEntry)
			//log.Debugf("SELECTED %v.(%s==%s) -> %v", i, name, value, si)
			return si
		}
	} //for each list entry

	//if optional - no error is raised, but nil is returned
	if !optional {
		//fail to select: give a descriptive error message
		if len(listOfValues) == 0 {
			i.worker.Error(log.Wrapf(nil, "List %s is empty", i.FullName()))
		} else {
			if countKeyFieldPresent == 0 {
				i.worker.Error(log.Wrapf(nil, "Key %s not found in any list entry %s", name, i.FullName()))
			} else {
				i.worker.Error(log.Wrapf(nil, "Not found %s.%s==%v", i.FullName(), name, value))
			}
		}
	}
	return nil
} //Item.Select()

//Format controls how an item is logged using Printf()
func (i *Item) Format(state fmt.State, c rune) {
	if i != nil {
		state.Write([]byte(fmt.Sprintf("Item(%p:%s:%s)", i, i.itemType, i.name)))
	} else {
		state.Write([]byte("nil"))
	}
} //Item.Format()

//ListOfItems is returns to iterate over items in a list with the range keyword
type ListOfItems []*Item

//Items returns array of items to iterate over
func (i *Item) Items() ListOfItems {
	if i == nil || i.worker.err != nil {
		return nil
	}
	if i.itemType != typeList {
		i.worker.Error(log.Wrapf(nil, "Cannot use Items() on %s %s (not a list)", i.itemType, i.FullName()))
		return nil
	}

	result := make([]*Item, 0)
	for n := 0; n < i.Count(); n++ {
		result = append(result, i.ListItem(n))
	}
	return result
} //Item.Items()

//Get a list containing the named sub item present in all list items
//(error if absent in some, use Has() if may be absent)
func (loi ListOfItems) Get(name string) ListOfItems {
	result := make([]*Item, 0)
	if loi != nil {
		for _, item := range loi {
			if item.worker.err != nil {
				break
			}
			subItem := item.Get(name)
			result = append(result, subItem)
		}
	}
	return result
} //ListOfItems.Get()

//Has is similar to Get(), to return a list containing
//the named sub item, but it skips over list items that does not
//have the specified sub-item and does not report an error when absent
//(no error if absent, use Get() if must be present in all list items)
func (loi ListOfItems) Has(name string) ListOfItems {
	result := make([]*Item, 0)
	if loi != nil {
		for _, item := range loi {
			if item.worker.err != nil {
				break
			}
			subItem := item.Has(name)
			if subItem != nil {
				result = append(result, subItem)
			}
		}
	}
	return result
} //ListOfItems.Has()

//Value returns a list of values from the list of items
func (loi ListOfItems) Value() []interface{} {
	result := make([]interface{}, 0)
	for _, item := range loi {
		subValue := item.Value()
		if item.worker.err != nil {
			break
		}
		result = append(result, subValue)
	}
	return result
} //ListOfItems.Value()

func dotBase(n string) string {
	names := strings.Split(n, ".")
	l := len(names)
	if l == 0 {
		return ""
	}
	return names[l-1] //dotBase()
}
