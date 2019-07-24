//Package files implements a config.ISource that loads from configDir in a directory
package files

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync"

	"github.com/jansemmelink/config"
	"github.com/jansemmelink/log"
)

//register as a source of config
func init() {
	config.RegisterSource("files", constructor{})
}

//Constructor of this config
type constructor struct{}

//New creates a source with the address being a directory name
func (constructor constructor) New(address string) config.ISource {
	return New(address)
}

//New creates a new configDir source to load config from the specified directory
func New(dir string) config.ISource {
	return &configDir{
		dir:   filepath.Clean(dir),
		files: make(map[string]*file),
	}
}

//configDir implements config.ISource to load and watch config from a directory of configDir
type configDir struct {
	dir string

	//files are indexed by configName (not full file name)
	lock  sync.RWMutex
	files map[string]*file
}

func (configDir *configDir) Name() string {
	return "configDir(" + configDir.dir + ")"
}

//Add tries to load the specified configuration from a file in this directory
//Arguments:
//		configName is the simple name given to the config, e.g. "domain"
//		configType is the type of struct to create from the data in the file, which implements config.IConfig
func (configDir *configDir) Add(configName string, configType reflect.Type) (config.IConfig, error) {
	//quick check without lock... will check again before add
	//because not holding onto the lock while finding the file in the file system
	if _, ok := configDir.files[configName]; ok {
		return nil, log.Wrapf(nil, "dir[%s].config[%s] already exists", configDir.dir, configName)
	}

	//walkFileName := ""
	fileName := ""
	found := false
	if err := filepath.Walk(
		configDir.dir,
		func(walkFileName string, info os.FileInfo, err error) error {
			//optimization: stop walking if found
			if found {
				return filepath.SkipDir
			}

			//match any file in this directory (not sub-dirs)
			//with name starting with config.Name
			fileDir := filepath.Dir(walkFileName)
			if fileDir == configDir.dir &&
				info.Mode().IsRegular() &&
				strings.HasPrefix(info.Name(), configName) {
				fileName = walkFileName
				found = true
			} //if found
			return nil
		}); err != nil {
		return nil, log.Wrapf(err, "Failed to walk config dir[%s]", configDir.dir)
	} //if failed to walk

	if !found {
		return nil, log.Wrapf(nil, "Config file [%s] not found in file dir[%s]", configName, configDir.dir)
	}

	//determine file ext (skipping the ".") to expect "json" or "yml" or ...
	ext := filepath.Ext(fileName)
	if len(ext) > 0 {
		ext = ext[1:]
	}

	//load the current initial contents and validate it to create a file handle
	//that implements config.IConfig
	file, err := newFile(configName, fileName, ext, configType)
	if err != nil {
		return nil, log.Wrapf(err, "Failed to load config[%s] from file[%s]", configName, fileName)
	}

	//loaded: add to this set of configDir to be monitored for changes
	if err := configDir.add(configName, file); err != nil {
		return nil, log.Wrapf(err, "Failed to add config[%s]", configName)
	}

	//  /*
	//   * Notify listeners of the config changes
	//   */
	//  for _, changed := range config.changed {

	// 	 if oldConfig != nil {
	// 		 changed.Released(oldConfig.config)
	// 	 }

	// 	 changed.Loaded(newConfig.config)

	//  } // for each load

	//  /*
	//   * Done
	//   */
	//  logger.Debugf("Successfully loaded config [%s]",
	// 	 config.configName)
	//  return nil

	// /*
	//  * Notify us of config file changes
	//  */
	// if err := config.watchConfig(); err != nil {

	// 	return nil, log.Wrapf(err,
	// 		"Failed to watch config")

	// } // if failed to watch

	return file, nil
} //configDir.Add()

//add the file to configDir
func (configDir *configDir) add(configName string, f *file) error {
	configDir.lock.Lock()
	defer configDir.lock.Unlock()

	if _, ok := configDir.files[configName]; ok {
		return log.Wrapf(nil, "dir[%s].config[%s] already exists", configDir.dir, configName)
	}

	//JS: In the context of a global config set, I do not care about this
	//when the config is given to the transaction context, it can keep its own map
	//  /*
	//   * Create a new map to hold the configs. The reason we create a new map is
	//   * because the map holds a snapshot of the current configs. As a new config
	//   * has been loaded, this snapshot has changed. Each transaction obtains a
	//   * copy of the map, thus, the existing transactions with a handle to the
	//   * map should not see the new configs. This is more efficient than creating
	//   * a new map with the snapshot for each new transaction.
	//   */
	//  oldConfigs := currentConfigs

	//  currentConfigs = make(map[string]*configFile, len(oldConfigs)+1)

	//  for _, c := range oldConfigs {
	// 	 if c.configName != config.configName {
	// 		 currentConfigs[c.configName] = c
	// 	 } // if not new config
	//  } // for each config

	configDir.files[configName] = f
	return nil
} //configDir.add(*file)
