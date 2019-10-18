package config

import (
	"os"
	"strings"
	"testing"
	"time"

	. "github.com/0xor1/wtf/pkg/core"
	"github.com/0xor1/wtf/pkg/json"
	"github.com/stretchr/testify/assert"
)

var (
	str = "test"
)

func TestNew(t *testing.T) {
	a := assert.New(t)

	c := New("")
	a.NotNil(c)

	fileName := `tmp_test.json`
	json.MustFromString(`{"test":true}`).MustToFile(fileName, os.ModePerm)

	c = New(fileName)
	a.True(c.GetBool(str))

	PanicOn(os.Remove(fileName))
}

func TestConfig_SetDefault(t *testing.T) {
	a := assert.New(t)

	c := New("")
	a.NotNil(c)

	a.True(c.SetDefault(str, true).GetBool(str))
}

func TestConfig_GetFromEnvVar(t *testing.T) {
	a := assert.New(t)

	c := New("")
	a.NotNil(c)
	PanicOn(os.Unsetenv(strings.ToUpper(str)))
	PanicOn(os.Setenv(strings.ToUpper(str), "true"))

	a.True(c.SetDefault(str, false).GetBool(str))
	PanicOn(os.Unsetenv(strings.ToUpper(str)))
}

func TestConfig_GetString(t *testing.T) {
	a := assert.New(t)

	a.Equal(New("").SetDefault(str, str).GetString(str), str)
}

func TestConfig_GetStringSlice(t *testing.T) {
	a := assert.New(t)

	a.Equal(New("").SetDefault(str, []string{str}).GetStringSlice(str), []string{str})
}

func TestConfig_GetMap(t *testing.T) {
	a := assert.New(t)

	a.Equal(New("").SetDefault(str, map[string]interface{}{str: str}).GetMap(str), map[string]interface{}{str: str})
}

func TestConfig_GetMapString(t *testing.T) {
	a := assert.New(t)

	a.Equal(New("").SetDefault(str, map[string]string{str: str}).GetMapString(str), map[string]string{str: str})
}

func TestConfig_GetInt(t *testing.T) {
	a := assert.New(t)

	a.Equal(New("").SetDefault(str, 1).GetInt(str), 1)
}

func TestConfig_GetInt64(t *testing.T) {
	a := assert.New(t)

	a.Equal(New("").SetDefault(str, int64(1)).GetInt64(str), int64(1))
}

func TestConfig_GetBool(t *testing.T) {
	a := assert.New(t)

	a.Equal(New("").SetDefault(str, true).GetBool(str), true)
}

func TestConfig_GetTime(t *testing.T) {
	a := assert.New(t)
	now := Now()
	a.Equal(New("").SetDefault(str, now).GetTime(str), now)
}

func TestConfig_GetDuration(t *testing.T) {
	a := assert.New(t)

	a.Equal(New("").SetDefault(str, time.Second).GetDuration(str), time.Second)
}
