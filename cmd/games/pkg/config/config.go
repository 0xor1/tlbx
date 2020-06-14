package config

import (
	"github.com/0xor1/tlbx/pkg/web/app/config"
)

func Get(file ...string) *config.Config {
	c := config.GetBase(file...)
	c.SetDefault("data.primary", "data_games:C0-Mm-0n-Da-Ta@tcp(localhost:3306)/data_games?parseTime=true&loc=UTC&multiStatements=true")
	return config.GetProcessed(c)
}
