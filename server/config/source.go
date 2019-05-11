package config

import (
	"github.com/jansemmelink/log"
)

//ISourceSettings ...
type ISourceSettings interface {
	Validate() error
}

//ISourceConstructor is registered to construct a named source
type ISourceConstructor interface {
	NewSettings() interface{} // ISourceSettings
	New(settings ISourceSettings) (ISource, error)
}

var registeredSource = make(map[string]ISourceConstructor)

//RegisterSource ...
func RegisterSource(name string, settings ISourceSettings, constructor ISourceConstructor) {
	log.DebugOn()
	if !ValidName(name) {
		panic(log.Wrapf(nil, "Invalid source name \"%s\"", name))
	}
	if _, ok := registeredSource[name]; ok {
		panic(log.Wrapf(nil, "Source name \"%s\" already registered", name))
	}
	registeredSource[name] = constructor
	log.Debugf("Registered config source \"%s\"", name)
}

//Source gets a named source constructor
func Source(name string) ISourceConstructor {
	s, ok := registeredSource[name]
	if !ok {
		return nil
	}
	return s
}

//AddSource ...
func AddSource(ISource) {
	panic("nyi")
}

//Get config from any source
// func Get(name string) (IConfig, error) {
// 	return nil, log.Wrapf(nil, "NYI")
// }

//ISource is a place where config can be read from
type ISource interface {
	Get(name string) (interface{}, error)
}
