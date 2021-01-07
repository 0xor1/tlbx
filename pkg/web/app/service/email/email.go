package email

import (
	. "github.com/0xor1/tlbx/pkg/core"
	"github.com/0xor1/tlbx/pkg/email"
	"github.com/0xor1/tlbx/pkg/web/app"
)

type tlbxKey struct {
	name string
}

func Mware(name string, email email.Client) func(app.Tlbx) {
	return func(tlbx app.Tlbx) {
		tlbx.Set(tlbxKey{name}, &client{tlbx: tlbx, name: name, email: email})
	}
}

func Get(tlbx app.Tlbx, name string) email.Client {
	return tlbx.Get(tlbxKey{name}).(email.Client)
}

type client struct {
	tlbx  app.Tlbx
	name  string
	email email.Client
}

func (c *client) Send(sendTo []string, from, subject, html, text string) error {
	var err error
	c.do(func() {
		err = c.email.Send(sendTo, from, subject, html, text)
	}, "SEND")
	return err
}

func (c *client) MustSend(sendTo []string, from, subject, html, text string) {
	PanicOn(c.Send(sendTo, from, subject, html, text))
}

func (c *client) do(do func(), action string) {
	start := NowUnixMilli()
	do()
	c.tlbx.LogActionStats(&app.ActionStats{
		Milli:  NowUnixMilli() - start,
		Type:   "EMAIL",
		Name:   c.name,
		Action: action,
	})
}
