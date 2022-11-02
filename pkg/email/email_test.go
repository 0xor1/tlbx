package email_test

import (
	"testing"

	. "github.com/0xor1/tlbx/pkg/core"
	"github.com/0xor1/tlbx/pkg/email"
	"github.com/0xor1/tlbx/pkg/log"
	"github.com/0xor1/tlbx/pkg/ptr"
	sp "github.com/SparkPost/gosparkpost"
	"github.com/stretchr/testify/assert"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ses"
)

func TestLocalClient(t *testing.T) {
	l := log.New()
	c := email.NewLocalClient(l)
	c.MustSend([]string{"a@a.a"}, "a@a.a", "subject", "<h1>yolo</h1>", "yolo")
}

func TestSparkPostClient(t *testing.T) {
	a := assert.New(t)
	spC := &sp.Client{}
	spC.Init(&sp.Config{
		BaseUrl:    "https://api.eu.sparkpost.com",
		ApiKey:     "123",
		ApiVersion: 1,
	})
	c := email.NewSparkPostClient(spC)
	defer Recover(func(i interface{}) {
		a.Contains(i.(Error).Message(), "Unauthorized")
	})
	c.MustSend([]string{"a@a.a"}, "a@a.a", "subject", "<h1>yolo</h1>", "yolo")
}

func TestSesClient(t *testing.T) {
	a := assert.New(t)
	c := email.NewSESClient(
		ses.New(
			session.New(
				&aws.Config{
					Region:      ptr.String("eu-west-1"),
					Credentials: credentials.NewStaticCredentials("abc", "123", ""),
				})))
	defer Recover(func(i interface{}) {
		a.Contains(i.(Error).Message(), "InvalidClientTokenId")
	})
	c.MustSend([]string{"a@a.a"}, "a@a.a", "subject", "<h1>yolo</h1>", "yolo")
}
