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
	Put(id ID, name, mimeType string, size int64, content io.ReadCloser) error
	MustPut(id ID, name, mimeType string, size int64, content io.ReadCloser)
	Get(id ID) (string, string, int64, io.ReadCloser, error)
	MustGet(id ID) (string, string, int64, io.ReadCloser)
	Delete(id ID) error
	MustDelete(id ID)
	PresignedPutUrl(id ID, name, mimeType string, size int64) (string, error)
	MustPresignedPutUrl(id ID, name, mimeType string, size int64) string
	PresignedGetUrl(id ID, isDownload bool) (string, error)
	MustPresignedGetUrl(id ID, isDownload bool) string
}

type LocalClient interface {
	Client
	Endpoints() []*app.Endpoint
	DeleteStore() error
	MustDeleteStore()
}

func NewLocalClient(preBaseUrl, dir string) LocalClient {
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
}

func (c *localClient) Put(id ID, name, mimeType string, size int64, content io.ReadCloser) error {
	c.objMtx.Lock()
	defer c.objMtx.Unlock()
	defer content.Close()

	idStr := id.String()
	// check for duplicate
	_, exists := c.objInfo[id.String()]
	if exists {
		return Errorf("object with id: %s, already exists", idStr)
	}

	// write obj file
	objFile, err := os.Create(filepath.Join(c.dir, idStr))
	if err != nil {
		return err
	}
	defer objFile.Close()

	_, err = io.Copy(objFile, content)
	if err != nil {
		return err
	}

	c.objInfo[idStr] = objInfo{
		Name:     name,
		MimeType: mimeType,
		Size:     size,
	}

	// write info file
	path := filepath.Join(c.dir, localStoreObjInfo)
	err = os.Remove(path)
	if err != nil {
		return err
	}

	infoFile, err := os.Create(path)
	if err != nil {
		return err
	}
	defer infoFile.Close()

	bs, err := json.Marshal(c.objInfo)
	if err != nil {
		return err
	}

	_, err = infoFile.Write(bs)
	return err
}

func (c *localClient) MustPut(id ID, name, mimeType string, size int64, content io.ReadCloser) {
	PanicOn(c.Put(id, name, mimeType, size, content))
}

func (c *localClient) Get(id ID) (string, string, int64, io.ReadCloser, error) {
	c.objMtx.Lock()
	defer c.objMtx.Unlock()

	idStr := id.String()
	info, exists := c.objInfo[idStr]
	if !exists {
		return "", "", 0, nil, Errorf("object with id: %s, does not exist", idStr)
	}

	objFile, err := os.Open(filepath.Join(c.dir, idStr))
	if err != nil {
		return "", "", 0, nil, err
	}

	return info.Name, info.MimeType, info.Size, objFile, nil
}

func (c *localClient) MustGet(id ID) (string, string, int64, io.ReadCloser) {
	name, mimeType, size, content, err := c.Get(id)
	PanicOn(err)
	return name, mimeType, size, content
}

func (c *localClient) Delete(id ID) error {
	c.objMtx.Lock()
	defer c.objMtx.Unlock()

	idStr := id.String()
	_, exists := c.objInfo[idStr]
	if !exists {
		return Errorf("object with id: %s, does not exist", idStr)
	}

	if err := os.Remove(filepath.Join(c.dir, idStr)); err != nil {
		return err
	}

	delete(c.objInfo, idStr)
	return nil
}

func (c *localClient) MustDelete(id ID) {
	PanicOn(c.Delete(id))
}

func (c *localClient) PresignedPutUrl(id ID, name, mimeType string, size int64) (string, error) {
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

func (c *localClient) MustPresignedPutUrl(id ID, name, mimeType string, size int64) string {
	str, err := c.PresignedPutUrl(id, name, mimeType, size)
	PanicOn(err)
	return str
}

func (c *localClient) PresignedGetUrl(id ID, isDownload bool) (string, error) {
	idStr := id.String()
	if _, exists := c.objInfo[idStr]; !exists {
		return "", Errorf("no such resource")
	}
	return Sprintf(`%s%s?id=%s&isDownload=%v`, c.preBaseUrl, localPreGetPath, idStr, isDownload), nil
}

func (c *localClient) MustPresignedGetUrl(id ID, isDownload bool) string {
	str, err := c.PresignedGetUrl(id, isDownload)
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
				c.MustPut(MustParseID(idStr), info.Name, info.MimeType, info.Size, args.Content)
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
				name, mimeType, size, content := c.MustGet(id)
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
