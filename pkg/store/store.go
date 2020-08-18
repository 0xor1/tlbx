package store

import (
	"io"
	"strings"
	"time"

	. "github.com/0xor1/tlbx/pkg/core"
	"github.com/0xor1/tlbx/pkg/ptr"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/s3"
)

const (
	attachmentPrefix = "attachment; filename="
	defaultMimeType  = "application/octet-stream"
)

type Client interface {
	CreateBucket(bucket, acl string) error
	MustCreateBucket(bucket, acl string)
	Put(bucket, prefix string, id ID, name, mimeType string, size int64, isPublic, isAttachment bool, content io.ReadSeeker) error
	MustPut(bucket, prefix string, id ID, name, mimeType string, size int64, isPublic, isAttachment bool, content io.ReadSeeker)
	PresignedPutUrl(bucket, prefix string, id ID, name, mimeType string, size int64) (string, error)
	MustPresignedPutUrl(bucket, prefix string, id ID, name, mimeType string, size int64) string
	Get(bucket, prefix string, id ID) (string, string, int64, io.ReadCloser, error)
	MustGet(bucket, prefix string, id ID) (string, string, int64, io.ReadCloser)
	PresignedGetUrl(bucket, prefix string, id ID, name string, isAttachment bool) (string, error)
	MustPresignedGetUrl(bucket, prefix string, id ID, name string, isAttachment bool) string
	Delete(bucket, prefix string, id ID) error
	MustDelete(bucket, prefix string, id ID)
	DeletePrefix(bucket, prefix string) error
	MustDeletePrefix(bucket, prefix string)
}

func New(s3 *s3.S3) Client {
	PanicIf(s3 == nil, "s3 is required")
	return &client{
		s3: s3,
	}
}

type client struct {
	s3 *s3.S3
}

func (c *client) CreateBucket(bucket, acl string) error {
	_, err := c.s3.CreateBucket(&s3.CreateBucketInput{
		Bucket: ptr.String(bucket),
		ACL:    ptr.String(acl),
	})
	if aerr, ok := err.(awserr.Error); ok && aerr.Code() == s3.ErrCodeBucketAlreadyOwnedByYou {
		err = nil
	}
	return ToError(err)
}

func (c *client) MustCreateBucket(bucket, acl string) {
	PanicOn(c.CreateBucket(bucket, acl))
}

func (c *client) putReq(bucket, prefix string, id ID, name, mimeType string, size int64, isPublic, isAttachment bool, content io.ReadSeeker) (*request.Request, *s3.PutObjectOutput) {
	return c.s3.PutObjectRequest(&s3.PutObjectInput{
		Bucket:             ptr.String(bucket),
		Key:                Key(prefix, id),
		ACL:                acl(isPublic),
		Body:               content,
		ContentDisposition: contentDisposition(name, isAttachment),
		ContentLength:      contentLength(size),
		ContentType:        contentType(mimeType),
	})
}

func (c *client) Put(bucket, prefix string, id ID, name, mimeType string, size int64, isPublic, isAttachment bool, content io.ReadSeeker) error {
	req, _ := c.putReq(bucket, prefix, id, name, mimeType, size, isPublic, isAttachment, content)
	return ToError(req.Send())
}

func (c *client) MustPut(bucket, prefix string, id ID, name, mimeType string, size int64, isPublic, isAttachment bool, content io.ReadSeeker) {
	PanicOn(c.Put(bucket, prefix, id, name, mimeType, size, isPublic, isAttachment, content))
}

func (c *client) PresignedPutUrl(bucket, prefix string, id ID, name, mimeType string, size int64) (string, error) {
	req, _ := c.putReq(bucket, prefix, id, name, mimeType, size, false, true, nil)
	url, err := req.Presign(10 * time.Minute)
	return url, ToError(err)
}

func (c *client) MustPresignedPutUrl(bucket, prefix string, id ID, name, mimeType string, size int64) string {
	str, err := c.PresignedPutUrl(bucket, prefix, id, name, mimeType, size)
	PanicOn(err)
	return str
}

func (c *client) getReq(bucket, prefix string, id ID, name string, isAttachment bool) (*request.Request, *s3.GetObjectOutput) {
	return c.s3.GetObjectRequest(&s3.GetObjectInput{
		Bucket:                     ptr.String(bucket),
		Key:                        Key(prefix, id),
		ResponseContentDisposition: contentDisposition(name, isAttachment),
	})
}

func (c *client) Get(bucket, prefix string, id ID) (string, string, int64, io.ReadCloser, error) {
	req, res := c.getReq(bucket, prefix, id, "", false)
	err := req.Send()
	return getName(res.ContentDisposition), ptr.StringOr(res.ContentType, defaultMimeType), ptr.Int64Or(res.ContentLength, 0), res.Body, ToError(err)
}

func (c *client) MustGet(bucket, prefix string, id ID) (string, string, int64, io.ReadCloser) {
	name, mimeType, size, content, err := c.Get(bucket, prefix, id)
	PanicOn(err)
	return name, mimeType, size, content
}

func (c *client) PresignedGetUrl(bucket, prefix string, id ID, name string, isAttachment bool) (string, error) {
	req, _ := c.getReq(bucket, prefix, id, name, isAttachment)
	url, err := req.Presign(10 * time.Minute)
	return url, ToError(err)
}

func (c *client) MustPresignedGetUrl(bucket, prefix string, id ID, name string, isAttachment bool) string {
	str, err := c.PresignedGetUrl(bucket, prefix, id, name, isAttachment)
	PanicOn(err)
	return str
}

func (c *client) Delete(bucket, prefix string, id ID) error {
	_, err := c.s3.DeleteObject(&s3.DeleteObjectInput{
		Bucket: ptr.String(bucket),
		Key:    Key(prefix, id),
	})
	return ToError(err)
}

func (c *client) MustDelete(bucket, prefix string, id ID) {
	PanicOn(c.Delete(bucket, prefix, id))
}

func (c *client) DeletePrefix(bucket, prefix string) error {
	if !strings.HasSuffix(prefix, "/") {
		prefix = prefix + "/"
	}
	for {
		list, err := c.s3.ListObjectsV2(&s3.ListObjectsV2Input{
			Bucket: ptr.String(bucket),
			Prefix: ptr.String(prefix),
		})
		if err != nil {
			return ToError(err)
		}
		if len(list.Contents) == 0 {
			return nil
		}
		objs := make([]*s3.ObjectIdentifier, 0, len(list.Contents))
		for _, content := range list.Contents {
			objs = append(objs, &s3.ObjectIdentifier{
				Key: content.Key,
			})
		}

		_, err = c.s3.DeleteObjects(&s3.DeleteObjectsInput{
			Bucket: ptr.String(bucket),
			Delete: &s3.Delete{
				Objects: objs,
			},
		})
		return ToError(err)
	}
}

func (c *client) MustDeletePrefix(bucket, prefix string) {
	PanicOn(c.DeletePrefix(bucket, prefix))
}

func Key(prefix string, id ID) *string {
	key := id.String()
	if prefix != "" {
		key = Sprintf("%s/%s", prefix, key)
	}
	return ptr.String(key)
}

func contentDisposition(name string, isAttachment bool) *string {
	contentDisposition := "inline"
	if isAttachment {
		contentDisposition = Sprintf("%s%s", attachmentPrefix, name)
	}
	return ptr.String(contentDisposition)
}

func contentLength(size int64) *int64 {
	if size == 0 {
		return nil
	}
	return ptr.Int64(size)
}

func contentType(mimeType string) *string {
	if mimeType == "" {
		return ptr.String(defaultMimeType)
	}
	return ptr.String(mimeType)
}

func getName(contentDisposition *string) string {
	name := ptr.StringOr(contentDisposition, "")
	if name == "inline" {
		return ""
	}
	return strings.TrimPrefix(name, attachmentPrefix)
}

func acl(isPublic bool) *string {
	acl := "private"
	if isPublic {
		acl = "public-read"
	}
	return ptr.String(acl)
}
