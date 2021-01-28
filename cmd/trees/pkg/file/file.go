package file

import (
	"time"

	"github.com/0xor1/tlbx/cmd/trees/pkg/task"
	. "github.com/0xor1/tlbx/pkg/core"
	"github.com/0xor1/tlbx/pkg/web/app"
)

type PutRes struct {
	Task *task.Task `json:"task"`
	File *File      `json:"file"`
}

type Put struct {
	app.UpStream
	Args *PutArgs
}

type PutArgs struct {
	Host    ID `json:"host"`
	Project ID `json:"project"`
	Task    ID `json:"task"`
}

func (_ *Put) Path() string {
	return "/file/put"
}

func (a *Put) Do(c *app.Client) (*PutRes, error) {
	res := &PutRes{}
	if a.Args == nil {
		return nil, Err("PutArgs must be specified")
	}
	a.UpStream.Args = a.Args
	err := app.Call(c, a.Path(), &a.UpStream, &res)
	return res, err
}

func (a *Put) MustDo(c *app.Client) *PutRes {
	res, err := a.Do(c)
	PanicOn(err)
	return res
}

type GetContent struct {
	Host       ID   `json:"host"`
	Project    ID   `json:"project"`
	Task       ID   `json:"task"`
	ID         ID   `json:"id"`
	IsDownload bool `json:"isDownload"`
}

func (_ *GetContent) Path() string {
	return "/file/getContent"
}

func (a *GetContent) Do(c *app.Client) (*app.DownStream, error) {
	res := &app.DownStream{}
	err := app.Call(c, a.Path(), a, &res)
	return res, err
}

func (a *GetContent) MustDo(c *app.Client) *app.DownStream {
	res, err := a.Do(c)
	if err != nil && res != nil && res.Content != nil {
		defer res.Content.Close()
	}
	PanicOn(err)
	return res
}

type File struct {
	Task      ID        `json:"task"`
	ID        ID        `json:"id"`
	CreatedBy ID        `json:"createdBy"`
	CreatedOn time.Time `json:"createdOn"`
	Name      string    `json:"name"`
	Type      string    `json:"type"`
	Size      uint64    `json:"size"`
}

type Get struct {
	Host         ID         `json:"host"`
	Project      ID         `json:"project"`
	Task         *ID        `json:"task,omitempty"`
	IDs          IDs        `json:"ids,omitempty"`
	CreatedOnMin *time.Time `json:"createdOnMin,omitempty"`
	CreatedOnMax *time.Time `json:"createdOnMax,omitempty"`
	CreatedBy    *ID        `json:"createdBy,omitempty"`
	After        *ID        `json:"after,omitempty"`
	Asc          *bool      `json:"asc,omitempty"`
	Limit        uint16     `json:"limit,omitempty"`
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
