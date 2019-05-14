package config

import (
	"sync"
)

//Getter ...
func Getter(g IGetter) {
	getterMutex.Lock()
	defer getterMutex.Unlock()
	getters = append(getters, g)
}

//IGetter ...
type IGetter interface {
	//Get name uses dot-notation: "a" or "a.b" or "a.b.c" ...
	//
	//Get returns:
	//		nil,  nil when config does not exist
	//		nil,  err when config exists but is invalid
	//		data, nil when config exists and is valid
	Get(name string) (interface{}, error)
}

var (
	getterMutex = sync.Mutex{}
	getters     = make([]IGetter, 0)
)
