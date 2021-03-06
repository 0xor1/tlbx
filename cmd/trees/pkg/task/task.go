package task

import (
	"time"

	. "github.com/0xor1/tlbx/pkg/core"
	"github.com/0xor1/tlbx/pkg/field"
	"github.com/0xor1/tlbx/pkg/web/app"
)

type Task struct {
	ID          ID        `json:"id"`
	Parent      *ID       `json:"parent"`
	FirstChild  *ID       `json:"firstChild"`
	NextSib     *ID       `json:"nextSib"`
	User        *ID       `json:"user"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedBy   ID        `json:"createdBy"`
	CreatedOn   time.Time `json:"createdOn"`
	TimeSubMin  uint64    `json:"timeSubMin"`
	TimeEst     uint64    `json:"timeEst"`
	TimeInc     uint64    `json:"timeInc"`
	TimeSubEst  uint64    `json:"timeSubEst"`
	TimeSubInc  uint64    `json:"timeSubInc"`
	CostEst     uint64    `json:"costEst"`
	CostInc     uint64    `json:"costInc"`
	CostSubEst  uint64    `json:"costSubEst"`
	CostSubInc  uint64    `json:"costSubInc"`
	FileN       uint64    `json:"fileN"`
	FileSize    uint64    `json:"fileSize"`
	FileSubN    uint64    `json:"fileSubN"`
	FileSubSize uint64    `json:"fileSubSize"`
	ChildN      uint64    `json:"childN"`
	DescN       uint64    `json:"descN"`
	IsParallel  bool      `json:"isParallel"`
}

type CreateRes struct {
	Parent *Task `json:"parent,omitempty"`
	Task   *Task `json:"task"`
}

type Create struct {
	Host        ID     `json:"host"`
	Project     ID     `json:"project"`
	Parent      ID     `json:"parent"`
	PrevSib     *ID    `json:"prevSib,omitempty"`
	Name        string `json:"name"`
	Description string `json:"description"`
	IsParallel  bool   `json:"isParallel"`
	User        *ID    `json:"user,omitempty"`
	TimeEst     uint64 `json:"timeEst"`
	CostEst     uint64 `json:"costEst"`
}

func (_ *Create) Path() string {
	return "/task/create"
}

func (a *Create) Do(c *app.Client) (*CreateRes, error) {
	res := &CreateRes{}
	err := app.Call(c, a.Path(), a, &res)
	return res, err
}

func (a *Create) MustDo(c *app.Client) *CreateRes {
	res, err := a.Do(c)
	PanicOn(err)
	return res
}

type Update struct {
	Host        ID            `json:"host"`
	Project     ID            `json:"project"`
	ID          ID            `json:"id"`
	Parent      *field.ID     `json:"parent,omitempty"`
	PrevSib     *field.IDPtr  `json:"prevSib,omitempty"`
	Name        *field.String `json:"name,omitempty"`
	Description *field.String `json:"description,omitempty"`
	IsParallel  *field.Bool   `json:"isParallel,omitempty"`
	User        *field.IDPtr  `json:"user,omitempty"`
	TimeEst     *field.UInt64 `json:"timeEst,omitempty"`
	CostEst     *field.UInt64 `json:"costEst,omitempty"`
}

type UpdateRes struct {
	OldParent *Task `json:"oldParent,omitempty"`
	NewParent *Task `json:"newParent,omitempty"`
	Task      *Task `json:"task"`
}

func (_ *Update) Path() string {
	return "/task/update"
}

func (a *Update) Do(c *app.Client) (*UpdateRes, error) {
	res := &UpdateRes{}
	err := app.Call(c, a.Path(), a, &res)
	return res, err
}

func (a *Update) MustDo(c *app.Client) *UpdateRes {
	res, err := a.Do(c)
	PanicOn(err)
	return res
}

type Get struct {
	Host    ID `json:"host"`
	Project ID `json:"project"`
	ID      ID `json:"id"`
}

func (_ *Get) Path() string {
	return "/task/get"
}

func (a *Get) Do(c *app.Client) (*Task, error) {
	res := &Task{}
	err := app.Call(c, a.Path(), a, &res)
	return res, err
}

func (a *Get) MustDo(c *app.Client) *Task {
	res, err := a.Do(c)
	PanicOn(err)
	return res
}

type Delete Get

func (_ *Delete) Path() string {
	return "/task/delete"
}

func (a *Delete) Do(c *app.Client) (*Task, error) {
	res := &Task{}
	err := app.Call(c, a.Path(), a, &res)
	return res, err
}

func (a *Delete) MustDo(c *app.Client) *Task {
	res, err := a.Do(c)
	PanicOn(err)
	return res
}

type GetSetRes struct {
	Set  []*Task `json:"set"`
	More bool    `json:"more"`
}

type GetAncestors struct {
	Host    ID     `json:"host"`
	Project ID     `json:"project"`
	ID      ID     `json:"id"`
	Limit   uint16 `json:"limit,omitempty"`
}

func (_ *GetAncestors) Path() string {
	return "/task/getAncestors"
}

func (a *GetAncestors) Do(c *app.Client) (*GetSetRes, error) {
	res := &GetSetRes{}
	err := app.Call(c, a.Path(), a, res)
	return res, err
}

func (a *GetAncestors) MustDo(c *app.Client) *GetSetRes {
	res, err := a.Do(c)
	PanicOn(err)
	return res
}

type GetChildren struct {
	Host    ID     `json:"host"`
	Project ID     `json:"project"`
	ID      ID     `json:"id"`
	After   *ID    `json:"after,omitempty"`
	Limit   uint16 `json:"limit,omitempty"`
}

func (_ *GetChildren) Path() string {
	return "/task/getChildren"
}

func (a *GetChildren) Do(c *app.Client) (*GetSetRes, error) {
	res := &GetSetRes{}
	err := app.Call(c, a.Path(), a, res)
	return res, err
}

func (a *GetChildren) MustDo(c *app.Client) *GetSetRes {
	res, err := a.Do(c)
	PanicOn(err)
	return res
}

// can only be called on a node with <= 1000 descN
type GetTree struct {
	Host    ID `json:"host"`
	Project ID `json:"project"`
	ID      ID `json:"id"`
}

type GetTreeRes map[ID]*Task

func (_ *GetTree) Path() string {
	return "/task/getTree"
}

func (a *GetTree) Do(c *app.Client) (GetTreeRes, error) {
	res := &GetTreeRes{}
	err := app.Call(c, a.Path(), a, res)
	if res == nil {
		return nil, err
	}
	return *res, err
}

func (a *GetTree) MustDo(c *app.Client) GetTreeRes {
	res, err := a.Do(c)
	PanicOn(err)
	return res
}
