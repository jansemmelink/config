package main

import (
	"encoding/json"
	"net/http"
	"github.com/jansemmelink/log"
)

type info struct {
	Sources map[string]sourceConfig `json:"source"`
	Configs map[string]infoConfig `json:"config"`
}

type infoConfig struct {
	Name string `json:"name"`
}

func infoHandler (res http.ResponseWriter, req *http.Request) {
	log.Debugf("INFO")


	info := info{
		Sources:make(map[string]sourceConfig),
		Configs:make(map[string]infoConfig),
	}
	for name,sc := range serverConfig.Sources {
		info.Sources[name] = sc
	}
	for name := range root.Subs() {
		info.Configs[name] = infoConfig{}
	}
	//todo: document the api?

	jsonInfo,_ := json.Marshal(info)
	res.Write(jsonInfo)
}

