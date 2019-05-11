package config

//IConfig is configuration data
type IConfig interface {
	Validate() error
	Get(name string) interface{}
}
