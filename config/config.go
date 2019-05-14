package config

import (
	"encoding/json"

	"github.com/jansemmelink/log"
)

//New config...
func New() IConfig {
	return config{}
}

//IConfig is configuration data
type IConfig interface {
	//Get(name string) (IConfig, error)
	Validate() error
}

type config struct {
}

//Get named config from any of the registered getters
func Get(name string, structPtr interface{}) error {
	// if reflect.TypeOf(structPtr).Kind() != reflect.Ptr ||
	// 	reflect.TypeOf(structPtr).Elem().Kind() != reflect.Struct {
	// 	return log.Wrapf(nil, "config.Get(name=\"%s\" ->%T) is not ptr to struct", name, structPtr)
	// }

	//make sure this data implements IConfig
	configData, ok := structPtr.(IConfig)
	if !ok {
		return log.Wrapf(nil, "config.Get(name=\"%s\" ->%T) does not implement IConfig", name, structPtr)
	}

	//without any getters, default to using ./conf directory
	if len(getters) == 0 {
		Getter(Dir("./conf"))
	}

	//try each getter
	for _, g := range getters {
		value, err := g.Get(name)
		if err == nil {
			log.Debugf("Got(%s):%+v", name, value)

			//encode data into JSON then decode into the config struct
			//as this data might be a small part of bigger JSON file
			jsonData, _ := json.Marshal(value)
			if err := json.Unmarshal(jsonData, structPtr); err != nil {
				return log.Wrapf(err, "Failed to decode %s into %T", name, structPtr)
			}
			if err := configData.Validate(); err != nil {
				return log.Wrapf(err, "invalid %s configuration data", name)
			}
			return nil
		} //if no error
		log.Debugf("%T.Get(%s): %v", g, name, err)
	}
	return log.Wrapf(nil, "config(%s) not found in any getter", name)
}

func (c config) Validate() error {
	return log.Wrapf(nil, "Validate() method not implemented")
}
