//Package main is a HTTP REST config server
//reading data from various source in future, but files for now
package main

import (
	"time"
	"encoding/json"
	"net/http"
	"strings"
	"github.com/gorilla/pat"
	"github.com/jansemmelink/log"
	_ "github.com/jansemmelink/config/server/config/jsonfiles"
)

func main() {
	log.DebugOn()
	addr := "0.0.0.0:12345"

	//load and watch all configured sources

	//serve with HTTP REST
	if err := http.ListenAndServe(addr, app()); err != nil {
		log.Errorf("HTTP server failed: %v", err)
	}
}

func app() http.Handler {
	r := pat.New()
	r.Get("/config", configGetHandler)
	r.Put("/config/{name}", configPutHandler)
	r.Delete("/config/{name}", configDel)

	r.Get("/status", configStatus)

	r.Put("/source/{name}", configSourcePutHandler)

	r.Get("/", infoHandler)
	return r
}

func configGetHandler (res http.ResponseWriter, req *http.Request) {
	var value interface{}
	value = nil

	name := ""
	if len(req.URL.Path) >= 8 && req.URL.Path[0:8] == "/config/" {
		name = strings.Replace(req.URL.Path[8:], "/", ".", -1)
	}

	log.Debugf("HTTP %s %s -> %s", req.Method, req.URL.Path, name)
	item := root.Find(name)
	if item != nil {
		value = item.Value()
		log.Debugf("FOUND %s: (%T)=%+v", name, value, value)
	} else {
		//not found in memory
		//see if available in one of the sources
		for n,s := range source {
			log.Debugf("Looking for %s in source %s", 
				name, n)
			var err error 
			value,err = s.Get(name)
			if err != nil {
				log.Errorf("%s not avail in source %s: %v", name, n, err)
			} else {
				log.Debugf("Found %s in source %s, value:%+v", name, n, value)
				break
			}
		}
		if value == nil {
			log.Debugf("NOT FOUND %s", name)
		} else {
			log.Debugf("FOUND: %+v", value)
		}
	}

	jsonValue,_ := json.Marshal(value)
	res.Header().Set("Content-Type", "application/json")
	res.Write(jsonValue)
}

func configPutHandler (res http.ResponseWriter, req *http.Request) {
	//parse the body value
	var value interface{}
	if err := json.NewDecoder(req.Body).Decode(&value); err != nil {
		log.Errorf("Cannot parse body as JSON value")
		http.Error(res, "Failed to parse JSON body", http.StatusBadRequest)
		return
	}

	log.Debugf("Parsed value(%T): %+v", value, value)

	name := ""
	if len(req.URL.Path) >= 8 && req.URL.Path[0:8] == "/config/" {
		name = strings.Replace(req.URL.Path[8:], "/", ".", -1)
	}

	log.Debugf("HTTP %s %s -> %s", req.Method, req.URL.Path, name)
	item := root.FindOrMake(name, true)
	if item == nil || item == root {
		http.Error(res, "No item selected in URL", http.StatusBadRequest)
		return
	}

	item.SetValue(value)
}

func configDel (res http.ResponseWriter, req *http.Request) {
	http.Error(res, "Not yet implemented", http.StatusInternalServerError)
}

type cs struct {
	Name string `json:"name"`
	FullName string `json:"fullName"`
	Timestamp time.Time `json:"timestamp"`
}

func configStatus (res http.ResponseWriter, req *http.Request) {
	s := cs{}

	name := ""
	if len(req.URL.Path) >= 8 && req.URL.Path[0:8] == "/status/" {
		name = strings.Replace(req.URL.Path[8:], "/", ".", -1)
	}

	log.Debugf("HTTP %s %s -> %s", req.Method, req.URL.Path, name)
	item := root.Find(name)
	if item == nil {
		http.Error(res, "Not Found", http.StatusNotFound)
		return
	}

	s.Name = item.Name()
	s.FullName = item.FullName()
	s.Timestamp = item.ValueTime()

	jsonValue,_ := json.Marshal(s)
	res.Header().Set("Content-Type", "application/json")
	res.Write(jsonValue)
}//configStatus()
