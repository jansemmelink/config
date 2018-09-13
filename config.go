package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"sync"

	"github.com/fsnotify/fsnotify"
	"github.com/pkg/errors"
)

var (
	watcher *fsnotify.Watcher
)

//Configurable must be implemented by config structs
type Configurable interface {
	Validate() error
	Changed()
	Copy() interface{}
	Restore(interface{})
}

//list of configs to monitor
type configuration struct {
	fileName     string
	loadFileName string
	config       Configurable
}

var (
	list  = make(map[string]*configuration, 0)
	mutex = sync.Mutex{}
)

func init() {
	var err error
	watcher, err = fsnotify.NewWatcher()
	if err != nil {
		panic("Failed to create new file watcher for configuration files.")
	}

	//run background task that will block until configuration file
	//events are detects, then call the relevant handlers
	//only one such routine, means that changes will be done one at a time
	//and the mutex is only used when adding to the list of configurations
	go func() {
		for {
			//block until we detect the next file event
			select {
			case event := <-watcher.Events:
				{
					fmt.Fprintf(os.Stderr, "EVENT! %+v\n", event)
					switch event.Op {
					case fsnotify.Write:
						{
							//the file we watch is the loadFileName and used to index the configuration map
							configuration, ok := list[event.Name]
							if !ok {
								fmt.Fprintf(os.Stderr, "Configuration \"%s\" does not exist for file event %s\n", configuration.fileName, event.String())
							} else {
								oldConfig := configuration.config.Copy()
								if err := loadConfig(event.Name, configuration.config); err != nil {
									//log error, but in memory we keep last good data...
									fmt.Fprintf(os.Stderr, "Configuration \"%s\" failed to load changes from file: %v\n", configuration.fileName, err)
								} else {
									fmt.Fprintf(os.Stderr, "LOADED: old=%T=%+v\n", configuration.config, configuration.config)

									/*var temp Configurable
									temp = oldConfig.(Configurable)
									fmt.Fprintf(os.Stderr, "temp=%T=%+v\n", temp, temp)*/

									if err := configuration.config.Validate(); err != nil {
										fmt.Fprintf(os.Stderr, "Configuration \"%s\" invalid changes in file: %+v: %v\n", configuration.fileName, configuration.config, err)
										//restore config
										configuration.config.Restore(oldConfig)
									} else {
										//copy the changes loaded to the actual config file so that when this
										//process is restarted, it will load the latest software
										if err := FileCopy(event.Name, configuration.fileName, true); err != nil {
											fmt.Fprintf(os.Stderr, "Configuration \"%s\" changes cannot be copied to : %+v: %v\n", configuration.fileName, configuration.config, err)
											configuration.config.Restore(oldConfig)
										} else {
											//loaded, validated and copied
											//we accept the changes and notify the user
											configuration.config.Changed()
										}
									} //if valid
								} //if loaded
							} //if found config
						} //if file write event
					default:
						{
							fmt.Fprintf(os.Stderr, "Configuration \"%s\" ignore event %s\n", event.Name, event.String())
						}
					} //switch(operation)
				} //file event
			// watch for errors
			case err := <-watcher.Errors:
				fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
			} //case watcher error
		}
	}()
} //init()

//Configure loads config from the file into a struct
//  'config' must be a pointer to the configurable struct
func Configure(fileName string, config Configurable) error {
	fileName = filepath.Clean(fileName)
	if err := loadConfig(fileName, config); err != nil {
		return errors.Wrapf(err, "Configuration failed on %s: %v", fileName, err)
	}
	if err := config.Validate(); err != nil {
		return errors.Wrapf(err, "Configuration invalid in %s", fileName)
	}

	//add to the list before we start watching, cause the fsnotify
	//handler started above in init(), will look for the file in the
	//list when it changes
	loadDir := path.Dir(fileName) + "/load/"
	loadFileName := loadDir + path.Base(fileName)

	mutex.Lock()
	defer mutex.Unlock()
	if _, ok := list[loadFileName]; ok {
		return fmt.Errorf("Configuration file %s cannot load again", fileName)
	}
	c := &configuration{
		fileName:     fileName,
		loadFileName: loadFileName,
		config:       config,
	}
	list[loadFileName] = c

	//start watching the configuration file for changes
	//but don't watch the actual file, rather watch a copy in a /load/-sub-folder
	if err := Mkdir(loadDir, 0770); err != nil {
		return errors.Wrapf(err, "Configuration %s: failed to make load dir %s", fileName, loadDir)
	}
	if err := FileCopy(fileName, loadFileName, true); err != nil {
		return errors.Wrapf(err, "Configuration %s failed to copy to load file %s", fileName, loadFileName)
	}
	if err := watcher.Add(loadFileName); err != nil {
		return errors.Wrapf(err, "Configuration %s failed to start watching load file ", fileName, loadFileName)
	}
	return nil
} //Configure()

//MustConfigure calls Configure() and panic on error
func MustConfigure(fileName string, config Configurable) {
	if err := Configure(fileName, config); err != nil {
		panic(err.Error())
	}
}

func loadConfig(fileName string, config Configurable) error {
	fileData, err := ioutil.ReadFile(fileName)
	if err != nil {
		return err
	}
	err = json.Unmarshal(fileData, config)
	return err
}
