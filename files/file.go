package files

import (
	"encoding/json"
	"io/ioutil"
	"reflect"

	"github.com/magiconair/properties"
	"gopkg.in/yaml.v2"

	"github.com/jansemmelink/config"
	"github.com/jansemmelink/log"
)

//readFileIntoStruct ...
func readFileIntoStruct(fileName string, ext string, userStructPtr interface{}) error {
	fileContentsByteArray, err := ioutil.ReadFile(fileName)
	if err != nil {
		return log.Wrapf(err, "Failed to read file [%s]", fileName)
	}

	switch ext {
	case "yaml", "yml":
		err = yaml.Unmarshal(fileContentsByteArray, userStructPtr)
	case "json":
		err = json.Unmarshal(fileContentsByteArray, userStructPtr)
	case "properties", "props", "prop":
		p := properties.NewProperties()
		err = p.Load(fileContentsByteArray, properties.UTF8)
		if err == nil {
			err = p.Decode(userStructPtr)
		}
	default:
		err = log.Wrapf(nil, "File [%s] extension [%s] is not supported as a config file", fileName, ext)
	} //switch(ext)

	if err != nil {
		return log.Wrapf(err, "Failed to load file [%s] into [%T]", fileName, userStructPtr)
	}
	return nil
} //readFileIntoStruct()

//newFile loads the file contents into memory and validate it
//and then creates an config.IConfig that represents the file
//for subsequent data access and to control the reload mechanism
func newFile(configName string, fileName string, ext string, configType reflect.Type) (*file, error) {
	newConfigPtrValue := reflect.New(configType)
	userStructPtr := newConfigPtrValue.Interface()
	if err := readFileIntoStruct(fileName, ext, userStructPtr); err != nil {
		return nil, log.Wrapf(err, "Failed to read config from file [%s] into [%T]", fileName, userStructPtr)
	}

	userConfig, ok := userStructPtr.(config.IValidator)
	if !ok {
		return nil, log.Wrapf(nil, "config file [%s] loaded into [%T] is not config.IValidator", fileName, userStructPtr)
	}

	if err := userConfig.Validate(); err != nil {
		return nil, log.Wrapf(err, "config file [%s] has invalid contents for [%T]", fileName, userStructPtr)
	}

	//successfully loaded and vlaidated into user struct
	//now we start managing the config
	configFile := &file{
		configName:    configName,
		fileName:      fileName,
		ext:           ext,
		currentStruct: newConfigPtrValue.Elem().Interface(), //dereference the config struct pointer to just a struct that implements config.IValidator
	}
	return configFile, nil
} //newFile()

//file implements config.IConfig for a configuration file loaded from disk
type file struct {
	configName    string
	fileName      string
	ext           string
	currentStruct interface{}

	//file watcher
	// watcher        *fsnotify.Watcher
	// watcherMutex   sync.RWMutex
	// watchedConfigs map[string]*Config
}

func (f file) Name() string {
	return f.configName
}

func (f file) Current() config.IValidator {
	return f.currentStruct.(config.IValidator)
}
