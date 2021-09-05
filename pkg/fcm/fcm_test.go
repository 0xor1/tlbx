package fcm_test

import (
	"context"
	"testing"

	firebase "firebase.google.com/go"
	"firebase.google.com/go/messaging"
	. "github.com/0xor1/tlbx/pkg/core"
	"github.com/0xor1/tlbx/pkg/fcm"
	"github.com/stretchr/testify/assert"
	"google.golang.org/api/option"
)

func TestClient(t *testing.T) {
	a := assert.New(t)
	opt := option.WithoutAuthentication()
	app, _ := firebase.NewApp(context.Background(), nil, opt)
	msgr, _ := app.Messaging(context.Background())
	c := fcm.NewClient(msgr)
	defer Recover(func(i interface{}) {
		// TODO try to test better, this is just failing before attempting to send
		a.Contains(i.(Error).Message(), "runtime error:")
	})
	c.MustSend(context.Background(), &messaging.MulticastMessage{
		Tokens: []string{"asd"},
		Data: map[string]string{
			"yolo": "nolo",
		},
	})
}
