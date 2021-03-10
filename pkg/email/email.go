package email

import (
	. "github.com/0xor1/tlbx/pkg/core"
	"github.com/0xor1/tlbx/pkg/log"
	"github.com/0xor1/tlbx/pkg/ptr"
	sp "github.com/SparkPost/gosparkpost"
	"github.com/aws/aws-sdk-go/service/ses"
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
	return ToError(err)
}

func (c *sparkPostClient) MustSend(sendTo []string, from, subject, html, text string) {
	PanicOn(c.Send(sendTo, from, subject, html, text))
}

func NewSESClient(ses *ses.SES) Client {
	return &sesClient{
		ses: ses,
	}
}

type sesClient struct {
	ses *ses.SES
}

func (c *sesClient) Send(sendTo []string, from, subject, html, text string) error {
	sendToPtrs := make([]*string, 0, len(sendTo))
	for _, a := range sendTo {
		sendToPtrs = append(sendToPtrs, ptr.String(a))
	}
	utf8 := ptr.String("UTF-8")
	_, err := c.ses.SendEmail(&ses.SendEmailInput{
		Destination: &ses.Destination{
			ToAddresses: sendToPtrs,
		},
		Message: &ses.Message{
			Body: &ses.Body{
				Html: &ses.Content{
					Charset: utf8,
					Data:    ptr.String(html),
				},
				Text: &ses.Content{
					Charset: utf8,
					Data:    ptr.String(text),
				},
			},
			Subject: &ses.Content{
				Charset: utf8,
				Data:    ptr.String(subject),
			},
		},
		Source: ptr.String(from),
	})
	return ToError(err)
}

func (c *sesClient) MustSend(sendTo []string, from, subject, html, text string) {
	PanicOn(c.Send(sendTo, from, subject, html, text))
}
