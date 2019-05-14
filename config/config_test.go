package config_test

import (
	"os"
	"path"
	"testing"
	"github.com/jansemmelink/config/config"
	"github.com/jansemmelink/log"
)

func init() {
	log.DebugOn()
}

func TestGetterDir(t *testing.T) {
	type testCase struct {
		fn string
		get string
		jsonText string
		valid bool
		expectedValue interface{}
	}
	testCases := []testCase{
		//flat:
		{fn:"c", get:"c", jsonText: "{\"s\":\"abc\"}", valid: true, expectedValue: "abc"},
		{fn:"c", get:"c", jsonText: "{\"S\":\"def\"}", valid: true, expectedValue: "def"},
		{fn:"c", get:"c", jsonText: "{\"X\":\"def\"}", valid: true, expectedValue: "---n/a---"},
		{fn:"c", get:"c", jsonText: "{\"s\":\"\"}", valid: false, expectedValue: "---n/a---"},
		//nested:
		{fn:"c/d", get:"c.d", jsonText: "{\"s\":\"abc\"}", valid: true, expectedValue: "abc"},
		{fn:"c/d", get:"c.e", jsonText: "{\"s\":\"abc\"}", valid: false, expectedValue: "---n/a---"},
	}

	//specify getter from file
	config.Getter(config.Dir("./conf"))

	for _,tc := range testCases {
		//create the file used below:
		makeTextFile(t, "./conf/"+tc.fn+".json", tc.jsonText)

		cc := c{S:"---n/a---"}
		err := config.Get(tc.get, &cc)
		if tc.valid {
			if err != nil {
				t.Fatalf("Failed to get %s from ./conf: %v", tc.get, err)
			}
			log.Debugf("Got %s: %+v", tc.get, cc)
			if cc.S != tc.expectedValue {
				t.Fatalf("Loaded c.s=\"%s\", expected \"%s\"", cc.S, tc.expectedValue)
			}
		}
		if !tc.valid && err == nil {
			t.Fatalf("Got invalid %s from ./conf/%s.json", tc.get, tc.fn)
		}

		delTextFile(t, "./conf/"+tc.fn+".json")
	}
}

type c struct {
	S string `json:"s"`
}

func (c c) Validate() error {
	if len(c.S) < 1 {
		return log.Wrapf (nil,"Missing value for s=\"%s\"", c.S)
	}
	return nil
}

func makeTextFile(t *testing.T, fn string, text string) {
	os.Mkdir(path.Dir(fn), 0770)
	f,err := os.Create(fn)
	if err != nil {
		t.Fatalf("failed to create text file %s: %v", fn, err)
	}
	defer f.Close()
	f.Write([]byte(text))
	return
}

func delTextFile(t *testing.T, fn string) {
	os.Remove(fn)
}
