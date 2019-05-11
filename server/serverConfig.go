package main

import (
	"os"
	"encoding/json"
	"github.com/jansemmelink/log"
	"github.com/jansemmelink/config/server/config"
)

var serverConfig = serverConfigStruct{
	Sources: make(map[string]sourceConfig),
}
var serverConfigFilename = "./conf/server.json"

func init() {
	jsonFile, err := os.Open(serverConfigFilename)
	if err != nil {
		panic(log.Wrapf(err, "Failed to load server config from %s", serverConfigFilename))
	}
	defer jsonFile.Close()

	err = json.NewDecoder(jsonFile).Decode(&serverConfig)
	if err != nil {
		panic(log.Wrapf(err, "Failed to read JSON data from file %s", serverConfigFilename))
	}

	log.Debugf("Successfully loaded config file %s.", serverConfigFilename)

	//add all sources
	for _,sc := range serverConfig.Sources {		
		//load the settings
		if err := addSource(sc); err != nil {
			log.Errorf("Failed to add configured source: %+v", sc)
		}
	}
}

type serverConfigStruct struct {
	Sources map[string]sourceConfig `json:"sources"`
}

type sourceConfig struct {
	Name string `json:"name"`
	Type string `json:"type"`
	Settings map[string]interface{} `json:"settings"`
}

func (sc sourceConfig) Validate() error {
	if !ValidName(sc.Name) {
		return log.Wrapf(nil, "Invalid source name")
	}
	if s := config.Source(sc.Type); s == nil {
		return log.Wrapf(nil, "Unknown Type=%s", sc.Type)
	}
	return nil
}

func updateServerConfig() {
	f,err := os.Create(serverConfigFilename)
	if err != nil {
		panic(log.Wrapf(err, "Failed to open file %s", serverConfigFilename))
	}
	defer f.Close()
	jsonConfig,_ := json.Marshal(serverConfig)
	f.Write(jsonConfig)
}