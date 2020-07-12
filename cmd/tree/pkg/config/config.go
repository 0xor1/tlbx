package config

import (
	"github.com/0xor1/tlbx/pkg/web/app/config"
)

func Get(file ...string) *config.Config {
	c := config.GetBase(file...)
	c.SetDefault("user.primary", "users_tree:C0-Mm-0n-U5-3r5@tcp(localhost:3306)/users_tree?parseTime=true&loc=UTC&multiStatements=true")
	c.SetDefault("pwd.primary", "pwds_tree:C0-Mm-0n-Pwd5@tcp(localhost:3306)/pwds_tree?parseTime=true&loc=UTC&multiStatements=true")
	c.SetDefault("data.primary", "data_tree:C0-Mm-0n-Da-Ta@tcp(localhost:3306)/data_tree?parseTime=true&loc=UTC&multiStatements=true")
	return config.GetProcessed(c)
}
