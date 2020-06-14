package store

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"

	. "github.com/0xor1/tlbx/pkg/core"
	"github.com/0xor1/tlbx/pkg/json"
)

const (
	localStoreObjInfo = "__local_store_obj_info__.json"
)

var (
	localErr = Errorf("presigned urls not valid in local store setting")
)

type Client interface {
	Put(key, mimeType string, size int64, content io.ReadCloser) error
	MustPut(key, mimeType string, size int64, content io.ReadCloser)
	Get(key string) (string, int64, io.ReadCloser, error)
	MustGet(key string) (string, int64, io.ReadCloser)
	Delete(key string) error
	MustDelete(key string)
	PresignedPutUrl(key, mimeType string, size int64) (string, error)
	MustPresignedPutUrl(key, mimeType string, size int64) string
	PresignedGetUrl(key string) (string, error)
	MustPresignedGetUrl(key string) string
}

type LocalClient interface {
	Client
	DeleteStore() error
	MustDeleteStore()
}

func NewLocalClient(dir string) LocalClient {
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
		dir:     dir,
		mtx:     &sync.Mutex{},
		objInfo: info,
	}
}

type objInfo struct {
	MimeType string
	Size     int64
}

type localClient struct {
	dir                 string
	mtx                 *sync.Mutex
	objInfo             map[string]objInfo
	presignedPutBaseUrl string
	presignedGetBaseUrl string
}

func (c *localClient) Put(key string, mimeType string, size int64, content io.ReadCloser) error {
	if err := checkKey(key); err != nil {
		return err
	}
	c.mtx.Lock()
	defer c.mtx.Unlock()
	defer content.Close()

	// check for duplicate
	_, exists := c.objInfo[key]
	if exists {
		return Errorf("object with key: %s, already exists", key)
	}

	// write obj file
	objFile, err := os.Create(filepath.Join(c.dir, key))
	if err != nil {
		return err
	}
	defer objFile.Close()

	_, err = io.Copy(objFile, content)
	if err != nil {
		return err
	}

	c.objInfo[key] = objInfo{
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

func (c *localClient) MustPut(key string, mimeType string, size int64, content io.ReadCloser) {
	PanicOn(c.Put(key, mimeType, size, content))
}

func (c *localClient) Get(key string) (string, int64, io.ReadCloser, error) {
	if err := checkKey(key); err != nil {
		return "", 0, nil, err
	}
	c.mtx.Lock()
	defer c.mtx.Unlock()

	info, exists := c.objInfo[key]
	if !exists {
		return "", 0, nil, Errorf("object with key: %s, does not exist", key)
	}

	objFile, err := os.Open(filepath.Join(c.dir, key))
	if err != nil {
		return "", 0, nil, err
	}

	return info.MimeType, info.Size, objFile, nil
}

func (c *localClient) MustGet(key string) (string, int64, io.ReadCloser) {
	mimeType, size, content, err := c.Get(key)
	PanicOn(err)
	return mimeType, size, content
}

func (c *localClient) Delete(key string) error {
	if err := checkKey(key); err != nil {
		return err
	}
	c.mtx.Lock()
	defer c.mtx.Unlock()

	_, exists := c.objInfo[key]
	if !exists {
		return Errorf("object with key: %s, does not exist", key)
	}

	if err := os.Remove(filepath.Join(c.dir, key)); err != nil {
		return err
	}

	delete(c.objInfo, key)
	return nil
}

func (c *localClient) MustDelete(key string) {
	PanicOn(c.Delete(key))
}

func (c *localClient) PresignedPutUrl(key string, mimeType string, size int64) (string, error) {
	return "", localErr
}

func (c *localClient) MustPresignedPutUrl(key string, mimeType string, size int64) string {
	str, err := c.PresignedPutUrl(key, mimeType, size)
	PanicOn(err)
	return str
}

func (c *localClient) PresignedGetUrl(key string) (string, error) {
	return "", localErr
}

func (c *localClient) MustPresignedGetUrl(key string) string {
	str, err := c.PresignedGetUrl(key)
	PanicOn(err)
	return str
}

func (c *localClient) DeleteStore() error {
	c.mtx.Lock()
	defer c.mtx.Unlock()

	return os.RemoveAll(c.dir)
}

func (c *localClient) MustDeleteStore() {
	PanicOn(c.DeleteStore())
}

func checkKey(key string) error {
	if strings.ToLower(key) == localStoreObjInfo {
		return Errorf("invalid key, may not use %s", localStoreObjInfo)
	}
	return nil
}
