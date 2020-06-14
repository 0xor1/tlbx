package email

import (
	. "github.com/0xor1/tlbx/pkg/core"
	"github.com/0xor1/tlbx/pkg/log"
	sp "github.com/SparkPost/gosparkpost"
)

type Client interface {
	Send(sendTo []string, from, subject, html, text string) error
	MustSend(sendTo []string, from, subject, html, text string)
}

func NewLocalClient(l log.Log) Client {
	return &localClient{
		l: l,
	}
}

type localClient struct {
	l log.Log
}

func (c *localClient) Send(sendTo []string, from, subject, html, text string) error {
	c.l.Info("send email:\nto: %s\nfrom: %s\nsubject: %s\nhtml: %s\ntext: %s\n", sendTo, from, subject, html, text)
	return nil
}

func (c *localClient) MustSend(sendTo []string, from, subject, html, text string) {
	PanicOn(c.Send(sendTo, from, subject, html, text))
}

func NewSparkPostClient(spClient *sp.Client) Client {
	return &sparkPostClient{
		spClient: spClient,
	}
}

type sparkPostClient struct {
	spClient *sp.Client
}

func (c *sparkPostClient) Send(sendTo []string, from, subject, html, text string) error {
	f := false
	_, _, err := c.spClient.Send(&sp.Transmission{
		Options: &sp.TxOptions{
			TmplOptions: sp.TmplOptions{
				OpenTracking:  &f,
				ClickTracking: &f,
			},
		},
		Recipients: sendTo,
		Content: sp.Content{
			From:    from,
			Subject: subject,
			HTML:    html,
			Text:    text,
		},
	})
	return err
}

func (c *sparkPostClient) MustSend(sendTo []string, from, subject, html, text string) {
	PanicOn(c.Send(sendTo, from, subject, html, text))
}
