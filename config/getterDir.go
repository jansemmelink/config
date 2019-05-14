package config

import (
	"encoding/json"
	"os"
	"regexp"
	"strings"

	"github.com/jansemmelink/log"
)

//Dir makes a getter to read files in a directory
func Dir(path string) IGetter {
	return dirGetter{
		path: path,
	}
}

type dirGetter struct {
	//IGetter
	path string
}

func (dg dirGetter) Get(name string) (interface{}, error) {
	//-----------------------------------------------------------------------
	//when dir is "./conf" and name is "a.b.c",
	//get the config from any of the following:
	//		./conf/a.json then return JSON item .b.c
	//		./conf/a/b.json then return JSON item .c
	//		./conf/a/b/c.json then return the whole JSON item in the file
	//
	//when dir is "./conf" and name is "a.b[1].c[2]",
	//get the config from any of the following:
	//		./conf/a.json then return JSON item .b[1].c[2]
	//		./conf/a/b_1.json then return JSON item .c[2]
	//		./conf/a/b_1/c_2.json then return the whole JSON item in the file
	//-----------------------------------------------------------------------
	//to do that, we split the name is dot-parts:
	//		"a.b.c" -> "a" "b" "c"
	//		"a.b[1].c[2]" -> "a" "b[1]" "c[2]"
	//-----------------------------------------------------------------------
	return dg.get(dg.path, name)
}

func (dg dirGetter) get(dir, name string) (interface{}, error) {
	//log.Debugf("%T.Get(name=\"%s\")", dg, name)

	//skip over optional leading '.' then split name on the first '.':
	for len(name) > 0 && name[0:1] == "." {
		name = name[1:]
	}
	nameParts := strings.SplitN(name, ".", 2)
	//log.Debugf("dir=\"%s\", name=\"%s\": %d parts: %+v", dir, name, len(nameParts), nameParts)

	//-----------------------------------------------------
	//try to load JSON file into memory
	//if nested, start with deepest file first
	//	a b c   tries to load conf/a/b/c.json, or conf/a/b.json, or conf/a.json
	//-----------------------------------------------------
	for depth := len(nameParts); depth > 0; depth-- {
		//log.Debugf("Try depth %d: %v", depth, nameParts[0:depth])
		filename := dir
		for _, sub := range nameParts[0:depth] {
			filename += "/" + sub //todo: indexing... +"/" + arrayIndexesAsFilename(nameParts[0]) + ".json"
		}
		filename += ".json"

		itemName := ""
		for _, sub := range nameParts[depth:] {
			itemName += "." + sub //todo: indexing... +"/" + arrayIndexesAsFilename(nameParts[0]) + ".json"
		}
		//log.Debugf("filename=%s item=%s", filename, itemName)

		jsonFile, err := os.Open(filename)
		if err != nil {
			log.Debugf("Cannot open JSON file %s", filename)
			continue
		}
		defer jsonFile.Close()

		fileData := map[string]interface{}{}
		err = json.NewDecoder(jsonFile).Decode(&fileData)
		if err != nil {
			log.Debugf("Failed to read JSON data from file %s: %v", filename, err)
			continue
		}

		log.Debugf("Loaded JSON file %s", filename)
		configData := NewData(fileData)
		if value := configData.Get(itemName); value != nil {
			log.Debugf("%s(%s):%v", filename, itemName, value)
			return value, nil
		}

		log.Debugf("File %s Item %s not found", filename, itemName)
	} //for each level of depth
	return nil, log.Wrapf(nil, "Dir(%s).Item(%s) not found", dg.path, name)
}

//arrayIndexesAsFilename() replaces any array indexes with a leading-underscores
//using a regular expression:
//so: "b[1]" -> "b_1"
//    "b[1][2]" -> "b_1_2"
func arrayIndexesAsFilename(name string) string {
	return indexPattern.ReplaceAllString(name, "_$1")
}

var (
	indexPattern *regexp.Regexp
)

func init() {
	indexPattern = regexp.MustCompile(`\[([0-9]+)\]`)
}
