package config

import (
	"github.com/0xor1/tlbx/pkg/web/app/config"
)

func Get(file ...string) *config.Config {
	c := config.GetBase(file...)
	c.SetDefault("sql.data.primary", "games_data:C0-Mm-0n-Da-Ta@tcp(localhost:3306)/games_data?parseTime=true&loc=UTC&multiStatements=true")
	return config.GetProcessed(c)
}
