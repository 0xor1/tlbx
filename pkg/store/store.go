package store

import (
	"io"
	"os"
	"path/filepath"
	"sync"

	. "github.com/0xor1/tlbx/pkg/core"
	"github.com/0xor1/tlbx/pkg/json"
	"github.com/0xor1/tlbx/pkg/web/app"
)

const (
	localStoreObjInfo = "__local_store_obj_info__.json"
	localPrePutPath   = "/store/put"
	localPreGetPath   = "/store/get"
)

type Client interface {
	Put(bucket, prefix string, id ID, name, mimeType string, size int64, content io.ReadCloser) error
	MustPut(bucket, prefix string, id ID, name, mimeType string, size int64, content io.ReadCloser)
	Get(bucket, prefix string, id ID) (string, string, int64, io.ReadCloser, error)
	MustGet(bucket, prefix string, id ID) (string, string, int64, io.ReadCloser)
	Delete(bucket, prefix string, id ID) error
	MustDelete(bucket, prefix string, id ID)
	PresignedPutUrl(bucket, prefix string, id ID, name, mimeType string, size int64) (string, error)
	MustPresignedPutUrl(bucket, prefix string, id ID, name, mimeType string, size int64) string
	PresignedGetUrl(bucket, prefix string, id ID, isDownload bool) (string, error)
	MustPresignedGetUrl(bucket, prefix string, id ID, isDownload bool) string
}

type LocalClient interface {
	Client
	Endpoints() []*app.Endpoint
	DeleteStore() error
	MustDeleteStore()
}

func NewLocalClient(preBaseUrl, preBucket, prePrefix, dir string) LocalClient {
	PanicIf(dir == "" || dir == ".", "dir must be a named directory")
	dir, err := filepath.Abs(dir)
	PanicOn(err)
	PanicOn(os.MkdirAll(dir, os.ModePerm))
	info := map[string]objInfo{}
	path := filepath.Join(dir, localStoreObjInfo)
	f, err := os.Open(path)
	if os.IsNotExist(err) {
		json.MustNew().ToFile(path, os.ModePerm)
	} else {
		PanicOn(err)
		json.MustUnmarshalReader(f, &info)
	}
	return &localClient{
		dir:        dir,
		objMtx:     &sync.Mutex{},
		objInfo:    info,
		preMtx:     &sync.Mutex{},
		preInfo:    map[string]objInfo{},
		preBaseUrl: preBaseUrl + app.ApiPathPrefix,
		preBucket:  preBucket,
		prePrefix:  prePrefix,
	}
}

type objInfo struct {
	Name     string
	MimeType string
	Size     int64
}

type localClient struct {
	dir        string
	objMtx     *sync.Mutex
	objInfo    map[string]objInfo
	preMtx     *sync.Mutex
	preInfo    map[string]objInfo
	preBaseUrl string
	preBucket  string
	prePrefix  string
}

func (c *localClient) Put(bucket, prefix string, id ID, name, mimeType string, size int64, content io.ReadCloser) error {
	c.objMtx.Lock()
	defer c.objMtx.Unlock()
	defer content.Close()

	err := os.MkdirAll(filepath.Join(c.dir, bucket, prefix), os.ModePerm)
	if err != nil {
		return ToError(err)
	}

	fullID := filepath.Join(bucket, prefix, id.String())
	// check for duplicate
	_, exists := c.objInfo[fullID]
	if exists {
		return Errorf("object with fullID: %s, already exists", fullID)
	}

	// write obj file
	objFile, err := os.Create(filepath.Join(c.dir, fullID))
	if err != nil {
		return ToError(err)
	}
	defer objFile.Close()

	_, err = io.Copy(objFile, content)
	if err != nil {
		return ToError(err)
	}

	c.objInfo[fullID] = objInfo{
		Name:     name,
		MimeType: mimeType,
		Size:     size,
	}

	// write info file
	path := filepath.Join(c.dir, localStoreObjInfo)
	err = os.Remove(path)
	if err != nil {
		return ToError(err)
	}

	infoFile, err := os.Create(path)
	if err != nil {
		return ToError(err)
	}
	defer infoFile.Close()

	bs, err := json.Marshal(c.objInfo)
	if err != nil {
		return ToError(err)
	}

	_, err = infoFile.Write(bs)
	return ToError(err)
}

func (c *localClient) MustPut(bucket, prefix string, id ID, name, mimeType string, size int64, content io.ReadCloser) {
	PanicOn(c.Put(bucket, prefix, id, name, mimeType, size, content))
}

func (c *localClient) Get(bucket, prefix string, id ID) (string, string, int64, io.ReadCloser, error) {
	c.objMtx.Lock()
	defer c.objMtx.Unlock()

	fullID := filepath.Join(bucket, prefix, id.String())
	info, exists := c.objInfo[fullID]
	if !exists {
		return "", "", 0, nil, Errorf("object with fullID: %s, does not exist", fullID)
	}

	objFile, err := os.Open(filepath.Join(c.dir, fullID))
	if err != nil {
		return "", "", 0, nil, err
	}

	return info.Name, info.MimeType, info.Size, objFile, nil
}

func (c *localClient) MustGet(bucket, prefix string, id ID) (string, string, int64, io.ReadCloser) {
	name, mimeType, size, content, err := c.Get(bucket, prefix, id)
	PanicOn(err)
	return name, mimeType, size, content
}

func (c *localClient) Delete(bucket, prefix string, id ID) error {
	c.objMtx.Lock()
	defer c.objMtx.Unlock()

	idStr := id.String()
	_, exists := c.objInfo[idStr]
	if !exists {
		return Errorf("object with id: %s, does not exist", idStr)
	}

	if err := os.Remove(filepath.Join(c.dir, idStr)); err != nil {
		return ToError(err)
	}

	delete(c.objInfo, idStr)
	return nil
}

func (c *localClient) MustDelete(bucket, prefix string, id ID) {
	PanicOn(c.Delete(bucket, prefix, id))
}

func (c *localClient) PresignedPutUrl(bucket, prefix string, id ID, name, mimeType string, size int64) (string, error) {
	c.preMtx.Lock()
	defer c.preMtx.Unlock()
	idStr := id.String()
	_, preExists := c.preInfo[idStr]
	_, objExists := c.objInfo[idStr]
	if preExists || objExists {
		return "", Errorf("id already in use")
	}
	c.preInfo[idStr] = objInfo{
		Name:     name,
		MimeType: mimeType,
		Size:     size,
	}
	return Sprintf("%s%s?id=%s", c.preBaseUrl, localPrePutPath, idStr), nil
}

func (c *localClient) MustPresignedPutUrl(bucket, prefix string, id ID, name, mimeType string, size int64) string {
	str, err := c.PresignedPutUrl(bucket, prefix, id, name, mimeType, size)
	PanicOn(err)
	return str
}

func (c *localClient) PresignedGetUrl(bucket, prefix string, id ID, isDownload bool) (string, error) {
	idStr := id.String()
	if _, exists := c.objInfo[idStr]; !exists {
		return "", Errorf("no such resource")
	}
	return Sprintf(`%s%s?id=%s&isDownload=%v`, c.preBaseUrl, localPreGetPath, idStr, isDownload), nil
}

func (c *localClient) MustPresignedGetUrl(bucket, prefix string, id ID, isDownload bool) string {
	str, err := c.PresignedGetUrl(bucket, prefix, id, isDownload)
	PanicOn(err)
	return str
}

func (c *localClient) DeleteStore() error {
	c.objMtx.Lock()
	defer c.objMtx.Unlock()

	return os.RemoveAll(c.dir)
}

func (c *localClient) MustDeleteStore() {
	PanicOn(c.DeleteStore())
}

func (c *localClient) Endpoints() []*app.Endpoint {
	return []*app.Endpoint{
		{
			Description:  "put an object in the store from a presigned request",
			Path:         localPrePutPath,
			Timeout:      5000,
			MaxBodyBytes: app.GB,
			IsPrivate:    false,
			GetDefaultArgs: func() interface{} {
				return &app.Stream{}
			},
			GetExampleArgs: func() interface{} {
				return &app.Stream{}
			},
			GetExampleResponse: func() interface{} {
				return nil
			},
			Handler: func(tlbx app.Tlbx, a interface{}) interface{} {
				args := a.(*app.Stream)
				query := tlbx.Req().URL.Query()
				idStr := query.Get("id")
				c.preMtx.Lock()
				defer c.preMtx.Unlock()
				info, exists := c.preInfo[idStr]
				tlbx.BadReqIf(!exists, "unknown presigned id")
				tlbx.BadReqIf(info.Name != args.Name, "name mismatch")
				tlbx.BadReqIf(info.MimeType != args.Type, "mimeType mismatch")
				tlbx.BadReqIf(info.Size != args.Size, "size mismatch")
				defer args.Content.Close()
				delete(c.preInfo, idStr)
				c.MustPut(c.preBucket, c.prePrefix, MustParseID(idStr), info.Name, info.MimeType, info.Size, args.Content)
				return nil
			},
		},
		{
			Description:  "get an object from the store from a presigned request",
			Path:         localPreGetPath,
			Timeout:      5000,
			MaxBodyBytes: app.GB,
			IsPrivate:    false,
			GetDefaultArgs: func() interface{} {
				return nil
			},
			GetExampleArgs: func() interface{} {
				return nil
			},
			GetExampleResponse: func() interface{} {
				return &app.Stream{}
			},
			Handler: func(tlbx app.Tlbx, a interface{}) interface{} {
				query := tlbx.Req().URL.Query()
				idStr := query.Get("id")
				id := MustParseID(idStr)
				isDownload := query.Get("isDownload") == "true"
				name, mimeType, size, content := c.MustGet(c.preBucket, c.prePrefix, id)
				return &app.Stream{
					ID:         id,
					Name:       name,
					Type:       mimeType,
					Size:       size,
					Content:    content,
					IsDownload: isDownload,
				}
			},
		},
	}
}
