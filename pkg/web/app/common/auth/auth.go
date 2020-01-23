package common

import (
	"github.com/0xor1/wtf/pkg/web/app"
)

func AuthEndpoints() []*app.Endpoint {
	return []*app.Endpoint{
		{
			Path: "/api/user/register",
		},
		{
			Path: "/api/user/confirmEmail",
		},
		{
			Path: "/api/user/changeEmail",
		},
		{
			Path: "/api/user/unregister",
		},
		{
			Path: "/api/user/login",
		},
		{
			Path: "/api/user/logout",
		},
	}
}
