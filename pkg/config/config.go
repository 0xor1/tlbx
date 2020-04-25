package config

import (
	"os"
	"strings"
	"time"

	. "github.com/0xor1/wtf/pkg/core"
	"github.com/0xor1/wtf/pkg/json"
)

type Config struct {
	defaults   *json.Json
	fileValues *json.Json
}

func New(file ...string) *Config {
	ret := &Config{
		defaults:   json.MustNew(),
		fileValues: json.MustNew(),
	}
	if len(file) > 0 && file[0] != "" {
		ret.fileValues = json.MustFromFile(file[0])
	}
	return ret
}

func (c *Config) SetDefault(path string, val interface{}) *Config {
	c.defaults.MustSet(append(json.MustSplitPath(path), val)...)
	return c
}

func (c *Config) GetString(path string) string {
	return c.mustGet(path, func(js *json.Json, path ...interface{}) (interface{}, error) {
		return js.String(path...)
	}).(string)
}

func (c *Config) GetStringSlice(path string) []string {
	return c.mustGet(path, func(js *json.Json, path ...interface{}) (interface{}, error) {
		return js.StringSlice(path...)
	}).([]string)
}

func (c *Config) GetMap(path string) map[string]interface{} {
	return c.mustGet(path, func(js *json.Json, path ...interface{}) (interface{}, error) {
		return js.Map(path...)
	}).(map[string]interface{})
}

func (c *Config) GetMapString(path string) map[string]string {
	return c.mustGet(path, func(js *json.Json, path ...interface{}) (interface{}, error) {
		return js.MapString(path...)
	}).(map[string]string)
}

func (c *Config) GetInt(path string) int {
	return c.mustGet(path, func(js *json.Json, path ...interface{}) (interface{}, error) {
		return js.Int(path...)
	}).(int)
}

func (c *Config) GetInt64(path string) int64 {
	return c.mustGet(path, func(js *json.Json, path ...interface{}) (interface{}, error) {
		return js.Int64(path...)
	}).(int64)
}

func (c *Config) GetBool(path string) bool {
	return c.mustGet(path, func(js *json.Json, path ...interface{}) (interface{}, error) {
		return js.Bool(path...)
	}).(bool)
}

func (c *Config) GetTime(path string) time.Time {
	return c.mustGet(path, func(js *json.Json, path ...interface{}) (interface{}, error) {
		return js.Time(path...)
	}).(time.Time)
}

func (c *Config) GetDuration(path string) time.Duration {
	return c.mustGet(path, func(js *json.Json, path ...interface{}) (interface{}, error) {
		return js.Duration(path...)
	}).(time.Duration)
}

func (c *Config) mustGet(path string, jsonGetter func(js *json.Json, path ...interface{}) (interface{}, error)) interface{} {
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
