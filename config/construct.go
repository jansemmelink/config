//Package config ...
//This module implements construction of configured items
//Say you implement the same "server" interface for "rest", "redis" and "nats",
//then register them each as "server.rest", "server.redis" and "server.nats"
//then you can start a process to construct your server and it will check in
//the config which one is configured, validate it and call the constructor
//so your code does not need to know which server is used and can be started
//with configuration to say which one you want to use.
package config

import (
	"encoding/json"
	"reflect"
	"strings"
	"sync"

	"github.com/jansemmelink/log"
)

//IConstructor ...
type IConstructor interface {
	IConfig
	Construct() (IConstructed, error)
}

//AddConstructor e.g. for "server.nats"
//The dot-notated name must have at least two parts
//and all parts must be defined, i.e.:
//		".server.nats" is invalid
//	but "server.nats" is valid
//  and "car.toyota.hilux" is also valid and will
//  be used to construct a "car.toyota" which will
//  then be a "hilux".
//because it will be called to construct name="server"
//then server.* will be tried...
func AddConstructor(name string, constructorPtr IConstructor) {
	if len(name) < 1 {
		panic("AddConstructor(name=\"\") requires a unique name")
	}
	nameParts := strings.Split(name, ".")
	if len(nameParts) < 2 {
		panic(log.Wrapf(nil, "Invalid constructor name=\"%s\" with fewer than 2 parts.", name))
	}
	for _, np := range nameParts {
		if len(np) < 1 {
			panic(log.Wrapf(nil, "Invalid constructor name=\"%s\" with empty parts.", name))
		}
	}
	if _, ok := constructors[name]; ok {
		panic(log.Wrapf(nil, "Duplicate name in AddConstructor(name=\"%s\")", name))
	}
	if constructorPtr == nil {
		panic(log.Wrapf(nil, "AddConstructor(name=\"%s\",nil) requires a !nil constructor", name))
	}
	if reflect.TypeOf(constructorPtr).Kind() != reflect.Ptr {
		panic(log.Wrapf(nil, "AddConstructor(name=\"%s\",%T) requires &%T", name, constructorPtr, constructorPtr))
	}

	constructorsMutex.Lock()
	defer constructorsMutex.Unlock()
	constructors[name] = constructorPtr
	log.Debugf("Added config.Constructor(%s)", name)
	return
}

var (
	constructorsMutex = sync.Mutex{}
	constructors      = make(map[string]IConstructor)
)

//Construct configured named item
//return the constructed name (e.g. "server.nats"), the constructed item, error
func Construct(name string) (string, IConstructed, error) {
	constructorsMutex.Lock()
	defer constructorsMutex.Unlock()

	registeredNames := []string{}
	{
		l := len(name) + 1
		for n := range constructors {
			if n[0:l] == name+"." {
				registeredNames = append(registeredNames, n[l:])
			} else {
				log.Debugf("constructor(\"%s\") is not a %s", n, name)
			}
		}
	}
	//log.Debugf("%d total: %d %s constructors: %s", len(constructors), len(registeredNames), name, registeredNames)
	if len(registeredNames) < 1 {
		panic(log.Wrapf(nil, "No constructors registered for %s (did you forget the import?)", name))
	}

	//when constructing "server", look for all configured "server.*" items
	namedItems := namedConfig(make(map[string]interface{}))
	err := Get(name, &namedItems)
	if err != nil {
		return "", nil, log.Wrapf(err, "Cannot get config for %s.* (expect one of %s)", name, registeredNames)
	}
	log.Debugf("Got %s: %+v", name, namedItems)

	//see which constructors have config - expect only 1
	found := ""
	count := 0
	var foundName string
	var foundValue interface{}
	var foundConstructor IConstructor
	for n, v := range namedItems {
		fn := name + "." + n
		if fc, ok := constructors[fn]; ok {
			found += "," + n
			count++
			foundValue = v
			foundConstructor = fc
			foundName = fn
			log.Debugf("FOUND C=%T N=%v v=%v", foundConstructor, foundName, foundValue)
		}
	}
	if count != 1 {
		panic(log.Wrapf(nil, "Construct(%s): expect 1 but found %d%s", name, count, found))
	}

	//parse the named config data into the constructor
	foundConfig, _ := json.Marshal(foundValue)
	if err := json.Unmarshal(foundConfig, foundConstructor); err != nil {
		panic(log.Wrapf(err, "Cannot unmarshal into %T", foundConstructor))
	}

	//validate
	if err := foundConstructor.Validate(); err != nil {
		panic(log.Wrapf(err, "Invalid %s config", foundName))
	}

	//call the found constructor
	log.Debugf("Constructing %s: %+v", foundName, foundConstructor)
	constructed, err := foundConstructor.Construct()
	if err != nil {
		panic(log.Wrapf(err, "Failed to construct %s", foundName))
	}

	log.Debugf("Constructed %s", foundName)
	return foundName, constructed, nil
} //Construct()

//IConstructed ...
type IConstructed interface{}

type namedConfig map[string]interface{}

func (nc namedConfig) Validate() error {
	/*if len(nc) != 1 {
		names := ""
		for n := range nc {
			names += "," + n
		}
		if len(names) > 0 {
			names = names[1:]
		}
		return log.Wrapf(nil, "%d named items (%s) instead of 1", len(nc), names)
	}*/
	return nil
}
