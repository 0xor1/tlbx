package store

import (
	"bytes"
	"io"
	"io/ioutil"
	"net/http"
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
	Copy(srcBucket, dstBucket, key string) error
	MustCopy(srcBucket, dstBucket, key string)
	StreamUp(bucket, key, name, mimeType string, size int64, isPublic, isAttachment bool, timeout time.Duration, content io.ReadCloser) error
	MustStreamUp(bucket, key, name, mimeType string, size int64, isPublic, isAttachment bool, timeout time.Duration, content io.ReadCloser)
	Put(bucket, key string, name, mimeType string, size int64, isPublic, isAttachment bool, content io.ReadSeeker) error
	MustPut(bucket, key string, name, mimeType string, size int64, isPublic, isAttachment bool, content io.ReadSeeker)
	PresignedPutUrl(bucket, key string, name, mimeType string, size int64) (string, error)
	MustPresignedPutUrl(bucket, key string, name, mimeType string, size int64) string
	Get(bucket, key string) (string, string, int64, io.ReadCloser, error)
	MustGet(bucket, key string) (string, string, int64, io.ReadCloser)
	PresignedGetUrl(bucket, key string, name string, isAttachment bool) (string, error)
	MustPresignedGetUrl(bucket, key string, name string, isAttachment bool) string
	Delete(bucket, key string) error
	MustDelete(bucket, key string)
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

func (c *client) Copy(srcBucket, dstBucket, key string) error {
	_, err := c.s3.CopyObject(&s3.CopyObjectInput{
		Bucket:     ptr.String(dstBucket),
		CopySource: ptr.String(srcBucket + "/" + key),
		Key:        ptr.String(key),
	})
	if err != nil {
		return ToError(err)
	}
	err = c.s3.WaitUntilObjectExists(&s3.HeadObjectInput{Bucket: ptr.String(dstBucket), Key: ptr.String(key)})
	if err != nil {
		return ToError(err)
	}
	return c.Delete(srcBucket, key)
}

func (c *client) MustCopy(srcBucket, dstBucket, key string) {
	PanicOn(c.Copy(srcBucket, dstBucket, key))
}

func (c *client) StreamUp(bucket, key, name, mimeType string, size int64, isPublic, isAttachment bool, timeout time.Duration, content io.ReadCloser) error {
	defer content.Close()
	putUrl, err := c.PresignedPutUrl(bucket, key, name, mimeType, size)
	if err != nil {
		return ToError(err)
	}
	req, err := http.NewRequest(http.MethodPut, putUrl, content)
	if err != nil {
		return ToError(err)
	}
	if isPublic {
		req.Header.Add("X-Amz-Acl", "public-read")
	} else {
		req.Header.Add("X-Amz-Acl", "private")
	}
	req.Header.Add("Content-Length", Strf(`%d`, size))
	req.Header.Add("Content-Type", mimeType)
	req.Header.Add("Content-Disposition", Strf("attachment; filename=%s", name))
	req.Header.Add("Host", req.Host)
	req.ContentLength = size
	resp, err := (&http.Client{Timeout: timeout}).Do(req)
	if resp != nil && resp.Body != nil {
		defer resp.Body.Close()
	}
	if err != err {
		return ToError(err)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return ToError(err)
	}
	if resp.StatusCode != 200 {
		return Err("resp.StatusCode: %d, resp.Body: %s", resp.StatusCode, string(body))
	}
	return nil
}

func (c *client) MustStreamUp(bucket, key, name, mimeType string, size int64, isPublic, isAttachment bool, timeout time.Duration, content io.ReadCloser) {
	PanicOn(c.StreamUp(bucket, key, name, mimeType, size, isPublic, isAttachment, timeout, content))
}

func (c *client) putReq(bucket, key string, name, mimeType string, size int64, isPublic, isAttachment bool, content io.ReadSeeker) (*request.Request, *s3.PutObjectOutput) {
	return c.s3.PutObjectRequest(&s3.PutObjectInput{
		Bucket:             ptr.String(bucket),
		Key:                ptr.String(key),
		ACL:                acl(isPublic),
		Body:               content,
		ContentDisposition: contentDisposition(name, isAttachment),
		ContentLength:      contentLength(size),
		ContentType:        contentType(mimeType),
	})
}

func (c *client) Put(bucket, key string, name, mimeType string, size int64, isPublic, isAttachment bool, content io.ReadSeeker) error {
	req, _ := c.putReq(bucket, key, name, mimeType, size, isPublic, isAttachment, content)
	return ToError(req.Send())
}

func (c *client) MustPut(bucket, key string, name, mimeType string, size int64, isPublic, isAttachment bool, content io.ReadSeeker) {
	PanicOn(c.Put(bucket, key, name, mimeType, size, isPublic, isAttachment, content))
}

func (c *client) PresignedPutUrl(bucket, key string, name, mimeType string, size int64) (string, error) {
	req, _ := c.putReq(bucket, key, name, mimeType, size, false, true, nil)
	url, err := req.Presign(10 * time.Minute)
	return url, ToError(err)
}

func (c *client) MustPresignedPutUrl(bucket, key string, name, mimeType string, size int64) string {
	str, err := c.PresignedPutUrl(bucket, key, name, mimeType, size)
	PanicOn(err)
	return str
}

func (c *client) getReq(bucket, key string, name string, isAttachment bool) (*request.Request, *s3.GetObjectOutput) {
	return c.s3.GetObjectRequest(&s3.GetObjectInput{
		Bucket:                     ptr.String(bucket),
		Key:                        ptr.String(key),
		ResponseContentDisposition: contentDisposition(name, isAttachment),
	})
}

func (c *client) Get(bucket, key string) (string, string, int64, io.ReadCloser, error) {
	req, res := c.getReq(bucket, key, "", false)
	err := req.Send()
	return getName(res.ContentDisposition), ptr.StringOr(res.ContentType, defaultMimeType), ptr.Int64Or(res.ContentLength, 0), res.Body, ToError(err)
}

func (c *client) MustGet(bucket, key string) (string, string, int64, io.ReadCloser) {
	name, mimeType, size, content, err := c.Get(bucket, key)
	PanicOn(err)
	return name, mimeType, size, content
}

func (c *client) PresignedGetUrl(bucket, key string, name string, isAttachment bool) (string, error) {
	req, _ := c.getReq(bucket, key, name, isAttachment)
	url, err := req.Presign(10 * time.Minute)
	return url, ToError(err)
}

func (c *client) MustPresignedGetUrl(bucket, key string, name string, isAttachment bool) string {
	str, err := c.PresignedGetUrl(bucket, key, name, isAttachment)
	PanicOn(err)
	return str
}

func (c *client) Delete(bucket, key string) error {
	_, err := c.s3.DeleteObject(&s3.DeleteObjectInput{
		Bucket: ptr.String(bucket),
		Key:    ptr.String(key),
	})
	return ToError(err)
}

func (c *client) MustDelete(bucket, key string) {
	PanicOn(c.Delete(bucket, key))
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

func Key(prefix string, root ID, ids ...ID) string {
	var key *bytes.Buffer
	if prefix != "" {
		key = bytes.NewBufferString(prefix + "/" + root.String())
	} else {
		key = bytes.NewBufferString(root.String())
	}
	for _, id := range ids {
		key.WriteString("/" + id.String())
	}
	return key.String()
}

func contentDisposition(name string, isAttachment bool) *string {
	contentDisposition := "inline"
	if isAttachment {
		contentDisposition = Strf("%s%s", attachmentPrefix, name)
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
