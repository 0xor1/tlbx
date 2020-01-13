package main

import (
	"net/http"
	"strings"

	. "github.com/0xor1/wtf/pkg/core"
	"github.com/0xor1/wtf/pkg/log"
	"github.com/0xor1/wtf/pkg/web/close"
	"github.com/0xor1/wtf/pkg/web/endpoint"
	logger "github.com/0xor1/wtf/pkg/web/log"
	"github.com/0xor1/wtf/pkg/web/lpath"
	"github.com/0xor1/wtf/pkg/web/mdo"
	"github.com/0xor1/wtf/pkg/web/options"
	"github.com/0xor1/wtf/pkg/web/redis"
	"github.com/0xor1/wtf/pkg/web/returnerror"
	"github.com/0xor1/wtf/pkg/web/server"
	"github.com/0xor1/wtf/pkg/web/sql"
	"github.com/0xor1/wtf/pkg/web/static"
	"github.com/0xor1/wtf/pkg/web/stats"
	"github.com/0xor1/wtf/pkg/web/toolbox"
)

type testArgs struct {
	Test string `json:"test"`
}

type testResp struct {
	Res int `json:"res"`
}

func main() {
	log := log.New()
	log.ErrorOn(server.Run(func(c *server.Config) {
		c.Log = log
		eps := &endpoint.Endpoints{
			Name:        "tmp test",
			Description: "a quick test of the web code",
			Endpoints: []*endpoint.Endpoint{
				{
					Path:         "/api/test",
					Description:  "a quick test endpoint",
					Timeout:      100,
					MaxBodyBytes: 1000000,
					IsPrivate:    false,
					GetDefaultArgs: func() interface{} {
						return &testArgs{
							Test: "you didnt pass a value",
						}
					},
					GetExampleArgs: func() interface{} {
						return &testArgs{
							Test: "I want to print this",
						}
					},
					GetExampleResponse: func() interface{} {
						return &testResp{
							Res: 1,
						}
					},
					Handler: func(r *http.Request, args interface{}) interface{} {
						Println(args)
						return &testResp{3}
					},
				},
			},
		}
		var root http.HandlerFunc
		c.Handler = func(w http.ResponseWriter, r *http.Request) {
			root(w, r)
		}
		root = close.Mware(
			toolbox.Mware(
				logger.Mware(c.Log,
					stats.Mware(
						returnerror.Mware(
							options.Mware(
								static.Mware(".", func(r *http.Request) bool { return !strings.HasPrefix(r.URL.Path, "/api/") },
									redis.Mware("redis:6379",
										sql.Mware("db:3306", "db:3306", "db:3306",
											lpath.Mware(
												mdo.Mware(func(r *http.Request) bool { return r.URL.Path == "/api/mdo" }, c.Handler,
													endpoint.Serve(func(r *http.Request) bool { return r.URL.Path == "/api/docs" }, eps))))))))))))
	}))
}
