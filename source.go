package config

import (
	"reflect"
	"sync"
)

//RegisterSource is called in init() of user packages that implements ISource
func RegisterSource(name string, constructor ISourceConstructor) {
	if len(name) == 0 {
		panic("config.ISourceConstructor[] requires a name")
	}
	constructorsLock.Lock()
	defer constructorsLock.Unlock()
	if _, ok := constructors[name]; ok {
		panic("config.ISourceConstructor[" + name + "] already registered.")
	}
	constructors[name] = constructor
} //RegisterSource()

//ISourceConstructor can be registered to create your config sources
type ISourceConstructor interface {
	New(address string) ISource
}

var (
	constructorsLock sync.Mutex
	constructors     = make(map[string]ISourceConstructor)
)

//ISource is a source of config
type ISource interface {
	//Name that describes the source - for logging and error messages
	Name() string

	//Add a named piece of config to the source
	//it must return:
	//  (nil,nil)     when the named config is not defined at this source
	//  (nil,error)   when the source is not available or the config is defined but invalid
	//	(IConfig,nil) when the config was successfully loaded
	Add(name string, configValidatorType reflect.Type) (IConfig, error)
}
