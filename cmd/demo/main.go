package main

import (
	"github.com/0xor1/wtf/pkg/iredis"
	"github.com/0xor1/wtf/pkg/web/app"
)

func main() {
	app.Run(func(c *app.Config) {
		c.RateLimiterPool = iredis.CreatePool("localhost:6379")
	})
}
