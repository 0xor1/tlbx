package fcm

import (
	"context"
	"time"

	"firebase.google.com/go/messaging"
	. "github.com/0xor1/tlbx/pkg/core"
	"github.com/0xor1/tlbx/pkg/fcm"
	"github.com/0xor1/tlbx/pkg/web/app"
)

type tlbxKey struct {
	name string
}

type Client interface {
	fcm.Client
	AsyncSend(tokens []string, data map[string]string, timeout time.Duration)
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

func Get(tlbx app.Tlbx, name string) Client {
	return tlbx.Get(tlbxKey{name}).(Client)
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

var clientHeaderName = "X-Fcm-Client"

func (c *client) AsyncSend(tokens []string, data map[string]string, timeout time.Duration) {
	if len(tokens) == 0 {
		return
	}
	if timeout <= 0 {
		timeout = 3 * time.Second
	}
	log := c.tlbx.Log()
	_, clientExists := data[clientHeaderName]
	PanicIf(clientExists, clientHeaderName+" is a reserved fcm push property for internal api use")
	client := c.tlbx.Req().Header.Get(clientHeaderName)
	if client != "" {
		data[clientHeaderName] = client
	}
	Go(func() {
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()
		log.Info("doing async call to fcm service with %d tokens", len(tokens))
		res := c.MustSend(ctx, &messaging.MulticastMessage{
			Tokens: tokens,
			Data:   data,
		})
		log.Info("FCM success: %d, fail: %d", res.SuccessCount, res.FailureCount)
	}, log.ErrorOn)
}
