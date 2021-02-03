package fcm

import (
	"context"

	"firebase.google.com/go/messaging"
	. "github.com/0xor1/tlbx/pkg/core"
	"github.com/0xor1/tlbx/pkg/fcm"
	"github.com/0xor1/tlbx/pkg/web/app"
)

type tlbxKey struct {
	name string
}

func Mware(name string, fcm fcm.Client) func(app.Tlbx) {
	return func(tlbx app.Tlbx) {
		tlbx.Set(tlbxKey{name}, &client{
			tlbx: tlbx,
			name: name,
			fcm:  fcm,
		})
	}
}

func Get(tlbx app.Tlbx, name string) fcm.Client {
	return tlbx.Get(tlbxKey{name}).(fcm.Client)
}

type client struct {
	tlbx app.Tlbx
	name string
	fcm  fcm.Client
}

func (c *client) Send(ctx context.Context, m *messaging.MulticastMessage) (*messaging.BatchResponse, error) {
	var err error
	var res *messaging.BatchResponse
	n := len(m.Tokens)
	c.do(func() {
		res, err = c.fcm.Send(ctx, m)
	}, Strf("%s %d", "SEND", n))
	return res, err
}

func (c *client) MustSend(ctx context.Context, m *messaging.MulticastMessage) *messaging.BatchResponse {
	res, err := c.Send(ctx, m)
	PanicOn(err)
	return res
}

func (c *client) do(do func(), action string) {
	start := NowUnixMilli()
	do()
	c.tlbx.LogActionStats(&app.ActionStats{
		Milli:  NowUnixMilli() - start,
		Type:   "FCM",
		Name:   c.name,
		Action: action,
	})
}
