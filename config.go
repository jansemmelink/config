package config

//IConfig represents a piece of configuration, validated and ready to be used in this process
type IConfig interface {
	//Name of the configuration (no file extension or path, just the name, e.g. "log" for log config)
	Name() string

	//Current returns the current configuration values using the struct type that was registered
	//Use type assertion to get it in the correct struct
	Current() IValidator
}

// //ChangeNotifier interface for config data
// type ChangeNotifier interface {
// 	Loaded(config interface{})
// 	Released(config interface{})
// }

// //Config that has been loaded
// type Config struct {
// 	//reference to the set in which this config was created
// 	set *Set

// 	//description of this config
// 	configName   string
// 	configPath   string
// 	fullFileName string
// 	extension    string
// 	configType   reflect.Type
// 	changed      []ChangeNotifier
// }

// //config represents a single configuration file
// type configFile struct {
// 	configName string
// 	config     interface{} // current valid config data
// }

// // Initialise the exported Config struct
// func (config *Config) init(configName string, configPath string, fullFileName string, extension string,
// 	configType reflect.Type) {

// 	config.configName = configName
// 	config.configPath = configPath
// 	config.fullFileName = fullFileName
// 	config.extension = extension
// 	config.configType = configType

// } // Config.init()

// //AddChangeNotifier ...
// func (config *Config) AddChangeNotifier(changed ChangeNotifier) error {

// 	const method = "AddChangeNotifier"

// 	if config == nil || changed == nil {

// 		return errors.Errorf("invalid parameters %p.%s (%p)",
// 			config,
// 			method,
// 			changed)

// 	} // if invalid params

// 	config.changed = append(
// 		config.changed,
// 		changed)

// 	return nil

// } // Config.AddLoaded()

// // initialise the config struct
// func (conf *configFile) init(configName string, userConfig interface{}) {

// 	conf.configName = configName
// 	conf.config = userConfig // new valid config

// } // config.init()

// // Watch the config for changes
// func (config *Config) watchConfig() error {

// 	watcherMutex.Lock()
// 	defer watcherMutex.Unlock()

// 	if watcher == nil {

// 		var err error

// 		if watcher, err = fsnotify.NewWatcher(); err != nil {

// 			return errors.Wrapf(err,
// 				"Failed to create watcher")

// 		} // if failed to create watcher

// 		go func() {

// 			defer watcher.Close()

// 			for {

// 				select {

// 				case event := <-watcher.Events:

// 					if event.Op&fsnotify.Write == fsnotify.Write {

// 						func() {

// 							watcherMutex.RLock()
// 							defer watcherMutex.RUnlock()

// 							if config, ok := watchedConfigs[filepath.Clean(event.Name)]; ok {

// 								logger.Debugf("Config file changed. Event [%+v]",
// 									event)

// 								if err := config.loadConfig(
// 									false); err != nil {

// 									logger.Errorf("%+v", errors.Wrap(err,
// 										"Failed to load config"))

// 								} // if failed to load config

// 							} // if got config

// 						}()

// 					} // if write

// 				case err := <-watcher.Errors:
// 					logger.Errorf("%+v", errors.Wrapf(err,
// 						"Error watching file"))

// 				} // select

// 			} //for ever

// 		}() // go

// 	} // if watcher not initialised

// 	if err := watcher.Add(config.configPath); err != nil {
// 		return errors.Wrapf(err,
// 			"Failed to add path [%s] to watcher",
// 			config.configPath)
// 	} // if failed to add

// 	watchedConfigs[config.fullFileName] = config

// 	return nil

// } // Config.watchConfig()
