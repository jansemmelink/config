package config

//IConfigurable is something that can get its own config
type IConfigurable interface {
	IConfig
	Get(name string) error
}

//IConfig is configuration data with defaults, that can be read from a source, be validated and documented
type IConfig interface {
	Validate() error
}
