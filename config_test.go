package config_test

import (
	"testing"
	"github.com/jansemmelink/config"
	"github.com/jansemmelink/config/mem"
	//"github.com/jansemmelink/config/file/json"
)


func Test1(t *testing.T) {
	//set some initial config in memory
	memConfig := mem.New().Set(
		map[string]interface{
			"test":map[string]interface{
				"value":123,
			},
		},
	)

	//load file: JSON/CSV/properties/YML/...

	//load all files in this directory

	//load from mongo

	//load from mysql


	//default to reading conf/test.json
	//or conf/<image>.json ".test"
	cfg := config.With("test", testConfig{})
	tc := cfg.Get().(testConfig)

	tc := testConfig{
		IConfigurable: config.Static("test"),
	}
	err := tc.Get("test")

	if err != nil {
		t.Fatalf("Failed to get test config: %v", err)
	}

	if tc.Value != 123 {
		t.Fatalf("test config not defaulted: value=%d", tc.Value)
	}
}

type testConfig struct {
	config.IConfigurable
	Value int `json:"value"`
}
