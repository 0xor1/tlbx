package config

import (
	"github.com/0xor1/tlbx/pkg/web/app/config"
)

func Get(file ...string) *config.Config {
	c := config.GetBase(file...)
	c.SetDefault("user.primary", "trees_users:C0-Mm-0n-U5-3r5@tcp(localhost:3306)/trees_users?parseTime=true&loc=UTC&multiStatements=true")
	c.SetDefault("pwd.primary", "trees_pwds:C0-Mm-0n-Pwd5@tcp(localhost:3306)/trees_pwds?parseTime=true&loc=UTC&multiStatements=true")
	c.SetDefault("data.primary", "trees_data:C0-Mm-0n-Da-Ta@tcp(localhost:3306)/trees_data?parseTime=true&loc=UTC&multiStatements=true")
	return config.GetProcessed(c)
}
