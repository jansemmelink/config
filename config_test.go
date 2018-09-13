package config_test

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"sync"
	"testing"
	"time"

	"github.com/jansemmelink/config"
)

//use that file as configuration
type C struct {
	Name string `json:"name"`
}

var validateCalled = 0
var changedCalled = 0
var invalid = false

func (c C) Validate() error {
	validateCalled++
	if c.Name == "" {
		invalid = true
		return fmt.Errorf("Missing name")
	}
	return nil
} //C.Validate()

func (c C) Changed() {
	changedCalled++
	fmt.Fprintf(os.Stderr, "New Config: %+v\n", c)
} //C.Changed()

func (c C) Copy() interface{} {
	return c
}

func (c *C) Restore(data interface{}) {
	fmt.Printf("Restoring from %T=%+v\n", data, data)
	*c = data.(C)
}

func TestLoadAndValidateCall(t *testing.T) {
	filename, name, _ := writeConfig("./conf/Test1.json", C{Name: "123"})

	//=====[ test program ]=====
	c := C{}
	if err := config.Configure(filename, &c); err != nil {
		panic(fmt.Sprintf("Failed to load config: %v", err))
	}
	if c.Name != name {
		panic(fmt.Sprintf("Loaded name=\"%s\" != \"%s\"", c.Name, name))
	}
	if validateCalled != 1 {
		panic(fmt.Sprintf("validateCalled=%d, expected 1", validateCalled))
	}
	fmt.Fprintf(os.Stderr, "Configured: %+v\n", c)
} //Test1()

func TestFailValidation(t *testing.T) {
	filename, _, _ := writeConfig("./conf/Test2.json", C{Name: ""})
	//=====[ test program ]=====
	c := C{}
	if err := config.Configure(filename, &c); err == nil {
		panic(fmt.Sprintf("Config loaded but should have failed on empty name"))
	}
	if validateCalled != 1 {
		panic(fmt.Sprintf("validateCalled=%d, expected 1", validateCalled))
	}
	if !invalid {
		panic(fmt.Sprintf("invalid=%v, expected true", invalid))
	}
	fmt.Fprintf(os.Stderr, "Invalid config detected: %+v\n", c)
} //TestFailValidation()

func TestRuntimeChange(t *testing.T) {
	filename, _, _ := writeConfig("./conf/Test3.json", C{Name: "123"})

	//change filename to config/load/filename - we don't write to actual file
	//but to load file that is monitored
	loadFilename := path.Dir(filename) + "/load/" + path.Base(filename)

	//=====[ test program ]=====
	c := C{}
	if err := config.Configure(filename, &c); err != nil {
		panic(fmt.Sprintf("Failed to load config: %v", err))
	}

	//log the config every second during the test
	//this is the config that will be used at that time
	//if a new context is started...
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		for i := 0; i < 10; i++ {
			fmt.Fprintf(os.Stderr, "%2d: Config: %+v\n", i, c)
			if c.Name == "" {
				panic(fmt.Sprintf("USING Invalid Config %+v", c))
			}
			time.Sleep(time.Second)
		}
		wg.Done()
	}()

	//wait 1.5s while above is running, that change config to be invalid
	time.Sleep(time.Millisecond * 1500)
	_, _, _ = writeBadJSON(loadFilename, C{Name: ""})

	//wait another 1.5s then fix the config to be valid
	time.Sleep(time.Millisecond * 1500)
	_, _, _ = writeConfig(loadFilename, C{Name: "456"})

	//wait for loop to terminate
	wg.Wait()
	fmt.Fprintf(os.Stderr, "End: %+v\n", c)
} //TestRuntimeChange()

func writeConfig(filename string, c C) (string, string, int) {
	if err := config.Mkdir(path.Dir(filename), 0770); err != nil {
		panic(fmt.Sprintf("Configuration %s: failed to make load dir %s: %v", filename, path.Dir(filename), err))
	}
	f, err := os.Create(filename)
	if err != nil {
		panic(fmt.Sprintf("Failed to open file %s: %v", filename, err))
	}
	json, _ := json.Marshal(c)
	f.Write(json)
	f.Close()
	fmt.Fprintf(os.Stderr, "Wrote config %s: %+v\n", filename, c)
	validateCalled = 0
	invalid = false
	return filename, c.Name, -1
}

func writeBadJSON(filename string, c C) (string, string, int) {
	if err := config.Mkdir(path.Dir(filename), 0770); err != nil {
		panic(fmt.Sprintf("Configuration %s: failed to make load dir %s: %v", filename, path.Dir(filename), err))
	}
	f, err := os.Create(filename)
	if err != nil {
		panic(fmt.Sprintf("Failed to open file %s: %v", filename, err))
	}
	json, _ := json.Marshal(c)
	f.Write([]byte("...garbage..."))
	f.Write(json)
	f.Close()
	fmt.Fprintf(os.Stderr, "Wrote config %s: %+v\n", filename, c)
	validateCalled = 0
	invalid = false
	return filename, c.Name, -1
}
