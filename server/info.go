package main

import (
	"encoding/json"
	"net/http"
	"github.com/jansemmelink/log"
)

type info struct {
	Sources map[string]sourceConfig `json:"source"`
	Configs map[string]configInfo `json:"config"`
}

type configInfo struct {
	Source map[string]interface{} `json:"source"`
	//todo: revision, list of users, ... history, ..
}

func infoHandler (res http.ResponseWriter, req *http.Request) {
	log.Debugf("INFO")

	info := info{
		Sources:make(map[string]sourceConfig),
		Configs:make(map[string]configInfo),
	}
	for name,sc := range serverConfig.Sources {
		info.Sources[name] = sc
	}
	for name := range root.Subs() {
		info.Configs[name] = configInfo{
			Source: map[string]interface{} {
				"Name": "sourcename",//todo - get from item
				"Type": "sourcetype",
			},
		}
	}
	//todo: document the api?

	jsonInfo,_ := json.Marshal(info)
	res.Write(jsonInfo)
}

