package fcm_test

import (
	"context"
	"testing"

	firebase "firebase.google.com/go"
	"firebase.google.com/go/messaging"
	"github.com/0xor1/tlbx/pkg/fcm"
	"github.com/0xor1/tlbx/pkg/log"
	"google.golang.org/api/option"
)

func TestClient(t *testing.T) {
	opt := option.WithoutAuthentication()
	app, _ := firebase.NewApp(context.Background(), &firebase.Config{ProjectID: "123"}, opt)
	msgr, _ := app.Messaging(context.Background())
	c := fcm.NewClient(msgr)
	ts := make([]string, 501)
	for i := range ts {
		ts[i] = "a"
	}
	c.MustSend(context.Background(), &messaging.MulticastMessage{
		Tokens: ts,
		Data: map[string]string{
			"yolo": "nolo",
		},
	})
}

func TestNopClient(t *testing.T) {
	l := log.New()
	c := fcm.NewNopClient(l)
	c.MustSend(context.Background(), &messaging.MulticastMessage{
		Tokens: []string{"asd"},
		Data: map[string]string{
			"yolo": "nolo",
		},
	})
}
