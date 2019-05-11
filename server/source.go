package main

// //ISource where config data can be fetched from
// type ISource interface {
// 	Source(relPath string) ISource
// 	//History(limit int) []IEvent
// 	Latest() interface{}
// 	Rev(rev string) interface{}
// }

import (
	"fmt"
	"net/http"
	"encoding/json"
	"strings"
	"reflect"
	"github.com/jansemmelink/log"
	"github.com/jansemmelink/config/server/config"
)

//HTTP PUT /source/{name}
//to add a new config source to the server
func configSourcePutHandler(res http.ResponseWriter, req *http.Request) {
	name := ""
	if len(req.URL.Path) >= 8 && req.URL.Path[0:8] == "/source/" {
		name = strings.Replace(req.URL.Path[8:], "/", ".", -1)
	}
	if _,ok := source[name]; ok {
		http.Error(res, fmt.Sprintf("Source %s already exists",name), http.StatusBadRequest)
		return
	}


	log.Debugf("HTTP %s %s -> %s", req.Method, req.URL.Path, name)
	sc := sourceConfig{}
	if err := json.NewDecoder(req.Body).Decode(&sc); err != nil {
		log.Errorf("Cannot parse body as JSON source config: %v", err)
		http.Error(res, "Failed to parse JSON body", http.StatusBadRequest)
		return
	}

	if len(sc.Name) > 0 && sc.Name != name {
		log.Errorf("Name in body mismath name in URL")
		http.Error(res, "Invalid source config", http.StatusBadRequest)
		return
	}
	sc.Name = name

	if err := addSource(sc); err != nil {
		log.Errorf("Failed to add source: %v", err)
		http.Error(res, fmt.Sprintf("ERROR: %s", err), http.StatusBadRequest)
		return
	}
}

var source = make(map[string]config.ISource)

func addSource(sc sourceConfig) error {
	if err := sc.Validate(); err != nil {
		return log.Wrapf(err, "Invalid source config")
	}
	sourceConstructor := config.Source(sc.Type)
	sourceSettingsData := reflect.New(reflect.TypeOf(sourceConstructor.NewSettings())).Interface()
	{
		jsonSettings,_ := json.Marshal(sc.Settings)
		if err := json.Unmarshal(jsonSettings, sourceSettingsData); err != nil {
			return log.Wrapf(err, "Failed to encode/decode source settings: %+v", sc)
		}
	}
	sourceSettings, ok := sourceSettingsData.(config.ISourceSettings);
	if ok {
		if err := sourceSettings.Validate(); err != nil {
			return log.Wrapf(err, "Invalid source settings")
		}
	}

	newSource,err := sourceConstructor.New(sourceSettings)
	if err != nil {
		return log.Wrapf(err, "Failed to construct new source")
	}

	//add to config to load when server start again
	serverConfig.Sources[sc.Name] = sc
	updateServerConfig()

	source[sc.Name] = newSource
	log.Infof("Added source %+v", sc)
	return nil
}
