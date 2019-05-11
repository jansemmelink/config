package main

import (
	"time"
	"regexp"
	"strings"
	"github.com/jansemmelink/log"
)

var root = &item{
	parent:nil,
	name:"",
	subs:make(map[string]IItem),
}

//IItem ...
type IItem interface {
	Parent() IItem
	Name() string
	FullName() string
	Subs() map[string]IItem
	Find(relName string)IItem
	FindOrMake(relName string, makeNew bool) IItem
	Value() interface{}
	ValueTime() time.Time
	SetNamedValue(relName string, value interface{})
	SetValue(value interface{})
}

//NewItem in the parent
func NewItem(parent IItem, name string) IItem {
	if !ValidName(name) {
		panic(log.Wrapf(nil,"NewItem(%p,%s)", parent, name))
	}
	return &item{
		parent:parent,
		name:name,
		subs:make(map[string]IItem),
		value:nil,
		valueTime:time.Now(),
	}
}

type item struct {
	parent IItem
	name string
	subs map[string]IItem
	value interface{}
	valueTime time.Time
}

func (i item) Parent() IItem {
	return i.parent
}

func (i item) Name() string {
	return i.name
}

func (i item) FullName() string {
	if i.parent == nil {
		return i.name
	}
	return i.parent.FullName()+"."+i.name
}

func (i item) Subs() map[string]IItem {
	return i.subs
}


func (i *item) Find(relName string) IItem {
	return i.FindOrMake(relName,false)
}


//find relative name e.g. "abc" or ".abc" or ".abc.def.ghi"
func (i *item) FindOrMake(relName string, makeNew bool) IItem {
	log.Debugf("Finding \"%s\" ...", relName)

	//skip over optional leading '.'
	for len(relName) > 0 && relName[0:1] == "." {
		relName = relName[1:]
	}

	if relName == "" {
		return i //no relative name = return this item
	}

	nameParts := strings.SplitN(relName, ".", 2)
	if len(nameParts) < 1 {
		return i
	}

	sub, ok := i.subs[nameParts[0]]
	if !ok {
		if !makeNew {
			log.Debugf("%s.sub(%s) does not yet exist", i.FullName(), nameParts[0])
			return nil
		}
		sub = NewItem(i, nameParts[0])
		i.subs[nameParts[0]] = sub
		log.Debugf("%s.sub(%s) created", i.FullName(), nameParts[0])
	}

	if len(nameParts) == 1 {
		return sub
	}
	return sub.FindOrMake(nameParts[1],makeNew)
}//item.Find()

//set relative name e.g. "abc" or ".abc" or ".abc.def.ghi"
//to specified value, overwriting old value
func (i *item) SetNamedValue(relName string, newValue interface{}) {
	item := i.FindOrMake(relName, true)
	item.SetValue(newValue)
}//item.SetNamedValue()

func (i *item) SetValue(newValue interface{}) {
	//delete all subs and set this item's value
	for _,sub := range i.subs {
		if len(sub.Subs()) > 0 {
			sub.SetValue(nil)
		}
	}
	i.subs = make(map[string]IItem)
	i.value = newValue
	i.valueTime = time.Now()
	log.Debugf("SET %s = %T = %+v", i.FullName(), i.value, i.value)
}//item.SetValue()

func (i item)Value() interface{} {
	return i.value
}

func (i item)ValueTime() time.Time {
	return i.valueTime
}

var (
	namePattern = regexp.MustCompile(`[a-z][a-z0-9]*`)
)

//ValidName ...
func ValidName(name string) bool {
	if !namePattern.MatchString(name) {
		return false
	}
	return true
}
