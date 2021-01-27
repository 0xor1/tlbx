package service

import (
	"io"
	"time"

	. "github.com/0xor1/tlbx/pkg/core"
	"github.com/0xor1/tlbx/pkg/store"
	"github.com/0xor1/tlbx/pkg/web/app"
)

type tlbxKey struct {
	name string
}

func Mware(name string, store store.Client) func(app.Tlbx) {
	return func(tlbx app.Tlbx) {
		tlbx.Set(tlbxKey{name}, &client{
			tlbx:  tlbx,
			name:  name,
			store: store,
		})
	}
}

func Get(tlbx app.Tlbx, name string) store.Client {
	return tlbx.Get(tlbxKey{name}).(store.Client)
}

type client struct {
	tlbx  app.Tlbx
	name  string
	store store.Client
}

func (c *client) CreateBucket(bucket, acl string) error {
	var err error
	c.do(func() {
		err = c.store.CreateBucket(bucket, acl)
	}, Strf("%s %s %s", "CREATE_BUCKET", bucket, acl))
	return err
}

func (c *client) MustCreateBucket(bucket, acl string) {
	PanicOn(c.CreateBucket(bucket, acl))
}

func (c *client) Copy(srcBucket, dstBucket, key string) error {
	var err error
	c.do(func() {
		err = c.store.Copy(srcBucket, dstBucket, key)
	}, Strf("%s %s %s %s", "COPY_OBJECT", srcBucket, dstBucket, key))
	return err
}

func (c *client) MustCopy(srcBucket, dstBucket, key string) {
	PanicOn(c.Copy(srcBucket, dstBucket, key))
}

func (c *client) StreamUp(bucket, key, name, mimeType string, size int64, isPublic, isAttachment bool, timeout time.Duration, content io.ReadCloser) error {
	var err error
	c.do(func() {
		err = c.store.StreamUp(bucket, key, name, mimeType, size, isPublic, isAttachment, timeout, content)
	}, Strf("%s %s %s", "STREAM_UP", bucket, key))
	return err
}

func (c *client) MustStreamUp(bucket, key, name, mimeType string, size int64, isPublic, isAttachment bool, timeout time.Duration, content io.ReadCloser) {
	PanicOn(c.StreamUp(bucket, key, name, mimeType, size, isPublic, isAttachment, timeout, content))
}

func (c *client) Put(bucket, key string, name, mimeType string, size int64, isPublic, isAttachment bool, content io.ReadSeeker) error {
	var err error
	c.do(func() {
		err = c.store.Put(bucket, key, name, mimeType, size, isPublic, isAttachment, content)
	}, Strf("%s %s %s", "PUT", bucket, key))
	return err
}

func (c *client) MustPut(bucket, key string, name, mimeType string, size int64, isPublic, isAttachment bool, content io.ReadSeeker) {
	PanicOn(c.Put(bucket, key, name, mimeType, size, isPublic, isAttachment, content))
}

func (c *client) PresignedPutUrl(bucket, key string, name, mimeType string, size int64) (string, error) {
	var url string
	var err error
	c.do(func() {
		url, err = c.store.PresignedPutUrl(bucket, key, name, mimeType, size)
	}, Strf("%s %s %s", "PUT_PRESIGNED_URL", bucket, key))
	return url, err
}

func (s *client) MustPresignedPutUrl(bucket, key string, name, mimeType string, size int64) string {
	url, err := s.PresignedPutUrl(bucket, key, name, mimeType, size)
	PanicOn(err)
	return url
}

func (c *client) Get(bucket, key string) (string, string, int64, io.ReadCloser, error) {
	var name string
	var mimeType string
	var size int64
	var content io.ReadCloser
	var err error
	c.do(func() {
		name, mimeType, size, content, err = c.store.Get(bucket, key)
	}, Strf("%s %s %s", "GET", bucket, key))
	return name, mimeType, size, content, err
}

func (c *client) MustGet(bucket, key string) (string, string, int64, io.ReadCloser) {
	name, mimeType, size, content, err := c.Get(bucket, key)
	PanicOn(err)
	return name, mimeType, size, content
}

func (c *client) PresignedGetUrl(bucket, key string, name string, isAttachment bool) (string, error) {
	var url string
	var err error
	c.do(func() {
		url, err = c.store.PresignedGetUrl(bucket, key, name, isAttachment)
	}, Strf("%s %s %s", "GET_PRESIGNED_URL", bucket, key))
	return url, err
}

func (c *client) MustPresignedGetUrl(bucket, key string, name string, isAttachment bool) string {
	url, err := c.PresignedGetUrl(bucket, key, name, isAttachment)
	PanicOn(err)
	return url
}

func (c *client) Delete(bucket, key string) error {
	var err error
	c.do(func() {
		err = c.store.Delete(bucket, key)
	}, Strf("%s %s %s", "DELETE", bucket, key))
	return err
}

func (c *client) MustDelete(bucket, key string) {
	PanicOn(c.Delete(bucket, key))
}

func (c *client) DeletePrefix(bucket, prefix string) error {
	var err error
	c.do(func() {
		err = c.store.DeletePrefix(bucket, prefix)
	}, Strf("%s %s %s", "DELETE_PREFIX", bucket, prefix))
	return err
}

func (c *client) MustDeletePrefix(bucket, prefix string) {
	PanicOn(c.DeletePrefix(bucket, prefix))
}

func (c *client) do(do func(), action string) {
	start := NowUnixMilli()
	do()
	c.tlbx.LogActionStats(&app.ActionStats{
		Milli:  NowUnixMilli() - start,
		Type:   "STORE",
		Name:   c.name,
		Action: action,
	})
}
