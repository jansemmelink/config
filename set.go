package config

import (
	"reflect"
	"sync"

	"github.com/jansemmelink/log"
)

//NewSet creates a new config set
func NewSet() ISet {
	return &set{
		sources: make([]ISource, 0),
		configs: make(map[string]IConfig),
	}
}

//ISet represents a set of valid configuration used in this process
type ISet interface {
	//WithSource adds the source and returns the set to add more
	WithSource(s ISource) ISet

	//MustSource constructs the named source and add it or panic
	MustSource(name string, address string) ISet

	Add(name string, tmpl IValidator) (IConfig, error)
	MustAdd(name string, tmpl IValidator) IConfig

	Get(name string) (IConfig, error)
	MustGet(name string) IConfig
}

//IValidator must be implemented by all configuration data structures
//because after loading, the config is validated before given to the caller
type IValidator interface {
	Validate() error
}

//set implements ISet
type set struct {
	lock    sync.RWMutex
	sources []ISource
	configs map[string]IConfig
}

func (set *set) WithSource(s ISource) ISet {
	if set == nil || s == nil {
		panic(log.Wrapf(nil, "(%T=%p).WithSource(%p)", set, set, s))
	}
	set.lock.Lock()
	defer set.lock.Unlock()
	set.sources = append(set.sources, s)
	return set
} //set.WithSource()

func (set *set) MustSource(name string, address string) ISet {
	constructor, ok := constructors[name]
	if !ok {
		panic(name + " not registered as a config source")
	}

	source := constructor.New(address)
	if source == nil {
		panic(name + "(" + address + ") creation returned nil")
	}

	return set.WithSource(source)
} //set.MustSource()

//old func: func Add(configPath string, configName string, configType reflect.Type) (*Config, error) {
func (set *set) Add(name string, tmpl IValidator) (IConfig, error) {
	if set == nil || len(name) == 0 {
		return nil, log.Wrapf(nil, "(%T=%p).Add(%s,%T)", set, set, name, tmpl)
	}

	set.lock.Lock()
	defer set.lock.Unlock()
	if _, ok := set.configs[name]; ok {
		return nil, log.Wrapf(nil, "config[%s] already added", name)
	}

	//try to load using any of the sources added to this set
	//they are tried in the order they were added
	var config IConfig
	notAvailableFromSources := ""
	for _, source := range set.sources {
		//this function returns an error only if it found the named config
		//but cannot load it for some reason, e.g. its invalid, or if the
		//source cannot be opened (e.g. failed to connect etc).
		//but no error is returned if the config is not available in this
		//loader, because another source may be able to provide it
		var err error
		config, err = source.Add(name, reflect.TypeOf(tmpl))
		if err != nil {
			return nil, log.Wrapf(err, "config source[%s] failed to load config[%s]", source.Name(), name)
		}
		if config != nil {
			break
		}
		notAvailableFromSources += "|" + source.Name()
	} //for each source

	if config == nil {
		if len(set.sources) == 0 || len(notAvailableFromSources) == 0 {
			return nil, log.Wrapf(nil, "config[%s] is not available from any of %d sources", name, len(set.sources))
		}
		return nil, log.Wrapf(nil, "config[%s] is not available from any of %s", name, notAvailableFromSources[1:])
	}

	set.configs[name] = config
	return config, nil
} //set.Add()

func (set *set) MustAdd(name string, tmpl IValidator) IConfig {
	config, err := set.Add(name, tmpl)
	if err != nil {
		panic(log.Wrapf(err, "Failed to add config [%s,(%T)]", name, tmpl))
	}
	return config
} //set.MustAdd()

func (set *set) Get(name string) (IConfig, error) {
	if set == nil || len(name) == 0 {
		return nil, log.Wrapf(nil, "(%T=%p).Get(%s)", set, set, name)
	}
	if existing, ok := set.configs[name]; ok {
		return existing, nil
	}
	return nil, log.Wrapf(nil, "config[%s] not yet added to config set", name)
} //set.Get()

func (set *set) MustGet(name string) IConfig {
	config, err := set.Get(name)
	if err != nil {
		panic(log.Wrapf(err, "Failed to get config [%s]", name))
	}
	return config
} //set.MustGet()
