package file

import (
	"time"

	. "github.com/0xor1/tlbx/pkg/core"
	"github.com/0xor1/tlbx/pkg/web/app"
)

type File struct {
	Task      ID        `json:"task"`
	ID        ID        `json:"id"`
	CreatedBy ID        `json:"createdBy"`
	CreatedOn time.Time `json:"createdOn"`
	Name      string    `json:"name"`
	MimeType  string    `json:"mimeType"`
	Size      uint64    `json:"size"`
}

type GetPresignedPutUrl struct {
	Host     ID     `json:"host"`
	Project  ID     `json:"project"`
	Task     ID     `json:"task"`
	Name     string `json:"name,omitempty"`
	MimeType string `json:"mimeType,omitempty"`
	Size     uint64 `json:"size,omitempty"`
}

type GetPresignedPutUrlRes struct {
	URL string `json:"url"`
	ID  ID     `json:"id"`
}

func (_ *GetPresignedPutUrl) Path() string {
	return "/file/getPresignedPutUrl"
}

func (a *GetPresignedPutUrl) Do(c *app.Client) (*GetPresignedPutUrlRes, error) {
	res := &GetPresignedPutUrlRes{}
	err := app.Call(c, a.Path(), a, &res)
	return res, err
}

func (a *GetPresignedPutUrl) MustDo(c *app.Client) *GetPresignedPutUrlRes {
	res, err := a.Do(c)
	PanicOn(err)
	return res
}

type Finalize struct {
	Host    ID `json:"host"`
	Project ID `json:"project"`
	Task    ID `json:"task"`
	ID      ID `json:"id"`
}

func (_ *Finalize) Path() string {
	return "/file/finalize"
}

func (a *Finalize) Do(c *app.Client) (*File, error) {
	res := &File{}
	err := app.Call(c, a.Path(), a, &res)
	return res, err
}

func (a *Finalize) MustDo(c *app.Client) *File {
	res, err := a.Do(c)
	PanicOn(err)
	return res
}

type GetPresignedGetUrl struct {
	Host    ID `json:"host"`
	Project ID `json:"project"`
	Task    ID `json:"task"`
	ID      ID `json:"id"`
}

func (_ *GetPresignedGetUrl) Path() string {
	return "/file/getPresignedGetUrl"
}

func (a *GetPresignedGetUrl) Do(c *app.Client) (string, error) {
	var res string
	err := app.Call(c, a.Path(), a, &res)
	return res, err
}

func (a *GetPresignedGetUrl) MustDo(c *app.Client) string {
	res, err := a.Do(c)
	PanicOn(err)
	return res
}

type Get struct {
	Host         ID         `json:"host"`
	Project      ID         `json:"project"`
	Task         *ID        `json:"task,omitempty"`
	IDs          IDs        `json:"ids,omitempty"`
	CreatedOnMin *time.Time `json:"createdOnMin,omitempty"`
	CreatedOnMax *time.Time `json:"createdOnMax,omitempty"`
	CreatedBy    *ID        `json:"createdBy,omitempty"`
	Asc          *bool      `json:"asc,omitempty"`
	Limit        uint16     `json:"limit,omitempty"`
	After        *ID        `json:"after,omitempty"`
}

type GetRes struct {
	Set  []*File `json:"set"`
	More bool    `json:"more"`
}

func (_ *Get) Path() string {
	return "/file/get"
}

func (a *Get) Do(c *app.Client) (*GetRes, error) {
	res := &GetRes{}
	err := app.Call(c, a.Path(), a, &res)
	return res, err
}

func (a *Get) MustDo(c *app.Client) *GetRes {
	res, err := a.Do(c)
	PanicOn(err)
	return res
}

type Delete struct {
	Host    ID `json:"host"`
	Project ID `json:"project"`
	Task    ID `json:"task"`
	ID      ID `json:"id"`
}

func (_ *Delete) Path() string {
	return "/file/delete"
}

func (a *Delete) Do(c *app.Client) error {
	err := app.Call(c, a.Path(), a, nil)
	return err
}

func (a *Delete) MustDo(c *app.Client) {
	PanicOn(a.Do(c))
}
