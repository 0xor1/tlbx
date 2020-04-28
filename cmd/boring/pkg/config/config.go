package config

import (
	"github.com/0xor1/wtf/pkg/web/app/common/config"
)

func Get(file ...string) *config.Config {
	c := config.GetBase(file...)
	c.SetDefault("data.primary", "data_boring:C0-Mm-0n-Da-Ta@tcp(localhost:3306)/data_boring?parseTime=true&loc=UTC&multiStatements=true")
	return config.GetProcessed(c)
}
