package jsonfiles

import (
	"encoding/json"
	"os"
	"strings"

	"github.com/jansemmelink/config/server/config"
	"github.com/jansemmelink/log"
)

func init() {
	config.RegisterSource("jsonfiles", settings{}, constructor{})
}

type settings struct {
	Dir string `json:"dir"`
}

func (s settings) Validate() error {
	if len(s.Dir) == 0 {
		return log.Wrapf(nil, "Missing required source settings:{dir:...}")
	}
	if info, err := os.Stat(s.Dir); err != nil || !info.IsDir() {
		return log.Wrapf(nil, "%s is not a directory for json files", s.Dir)
	}
	return nil
}

type constructor struct{}

func (c constructor) NewSettings() interface{} { //config.ISourceSettings {
	return settings{}
}

func (c constructor) New(s config.ISourceSettings) (config.ISource, error) {
	ss, ok := s.(*settings)
	if !ok {
		return nil, log.Wrapf(nil, "%T is not *settings", s)
	}

	return jsonFiles{
		//ISource:  NewSource(),
		settings: *ss,
	}, nil
}

//jsonFiles implements config server.ISource
type jsonFiles struct {
	//config.ISource
	settings settings
}

func (jf jsonFiles) Get(name string) (interface{}, error) {
	//the first part of the name should be the file name in the directory
	nameParts := strings.Split(name, ".")
	for len(nameParts) > 0 && len(nameParts[0]) == 0 {
		nameParts = nameParts[1:]
	}
	log.Debugf("Getting %+v", nameParts)

	//load the whole JSON file into memory
	filename := jf.settings.Dir + "/" + nameParts[0] + ".json"
	jsonFile, err := os.Open(filename)
	if err != nil {
		return nil, log.Wrapf(err, "Cannot open JSON file %s", filename)
	}
	defer jsonFile.Close()

	var jsonData jsonConfig
	err = json.NewDecoder(jsonFile).Decode(&jsonData)
	if err != nil {
		return nil, log.Wrapf(err, "Failed to read JSON data from file %s", filename)
	}
	log.Debugf("Successfully loaded JSON file %s", filename)

	//get deeper into the file if more names remain
	subName := ""
	if len(nameParts) > 1 {
		for _, n := range nameParts[1:] {
			subName += "." + n
		}
		subName = subName[1:]
	}
	return jsonData.Get(subName), nil
}

type jsonConfig map[string]interface{}

func (jc jsonConfig) Validate() error {
	return nil
}

func (jc jsonConfig) Get(name string) interface{} {
	log.Debugf("Getting name=\"%s\" from %+v", name, jc)
	if name == "" {
		return jc
	}

	v, err := config.GetObjField(jc, name)
	if err != nil {
		log.Errorf("Failed to get %s: %v", name, err)
	}
	return v
}
