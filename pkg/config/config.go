package config

import (
	"os"
	"strings"
	"time"

	. "github.com/0xor1/wtf/pkg/core"
	"github.com/0xor1/wtf/pkg/json"
)

type config struct {
	defaults   *json.Json
	fileValues *json.Json
}

func New(file string) *config {
	ret := &config{
		defaults:   json.MustNew(),
		fileValues: json.MustNew(),
	}
	if file != "" {
		ret.fileValues = json.MustFromFile(file)
	}
	return ret
}

func (c *config) SetDefault(path string, val interface{}) *config {
	c.defaults.MustSet(append(json.MustSplitPath(path), val)...)
	return c
}

func (c *config) GetString(path string) string {
	return c.mustGet(path, func(js *json.Json, path ...interface{}) (interface{}, error) {
		return js.String(path...)
	}).(string)
}

func (c *config) GetStringSlice(path string) []string {
	return c.mustGet(path, func(js *json.Json, path ...interface{}) (interface{}, error) {
		return js.StringSlice(path...)
	}).([]string)
}

func (c *config) GetMap(path string) map[string]interface{} {
	return c.mustGet(path, func(js *json.Json, path ...interface{}) (interface{}, error) {
		return js.Map(path...)
	}).(map[string]interface{})
}

func (c *config) GetMapString(path string) map[string]string {
	return c.mustGet(path, func(js *json.Json, path ...interface{}) (interface{}, error) {
		return js.MapString(path...)
	}).(map[string]string)
}

func (c *config) GetInt(path string) int {
	return c.mustGet(path, func(js *json.Json, path ...interface{}) (interface{}, error) {
		return js.Int(path...)
	}).(int)
}

func (c *config) GetInt64(path string) int64 {
	return c.mustGet(path, func(js *json.Json, path ...interface{}) (interface{}, error) {
		return js.Int64(path...)
	}).(int64)
}

func (c *config) GetBool(path string) bool {
	return c.mustGet(path, func(js *json.Json, path ...interface{}) (interface{}, error) {
		return js.Bool(path...)
	}).(bool)
}

func (c *config) GetTime(path string) time.Time {
	return c.mustGet(path, func(js *json.Json, path ...interface{}) (interface{}, error) {
		return js.Time(path...)
	}).(time.Time)
}

func (c *config) GetDuration(path string) time.Duration {
	return c.mustGet(path, func(js *json.Json, path ...interface{}) (interface{}, error) {
		return js.Duration(path...)
	}).(time.Duration)
}

func (c *config) mustGet(path string, jsonGetter func(js *json.Json, path ...interface{}) (interface{}, error)) interface{} {
	var val interface{}
	var err error
	var envVal string
	envValExists := false

	envName := strings.ToUpper(strings.Replace(path, ".", "_", -1))
	if envName != "" {
		envVal, envValExists = os.LookupEnv(envName)
	}

	if envValExists {
		val, err = jsonGetter(json.MustFromString(envVal))
	} else {
		jsonPath := json.MustSplitPath(path)
		if val, err = jsonGetter(c.fileValues, jsonPath...); err != nil {
			val, err = jsonGetter(c.defaults, jsonPath...)
		}
	}
	PanicOn(err)
	return val
}
