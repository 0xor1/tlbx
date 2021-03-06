package cnsts

import (
	"strconv"

	. "github.com/0xor1/tlbx/pkg/core"
	"github.com/0xor1/tlbx/pkg/web/app"
)

const (
	FileBucket = `files`
)

type Sort string

const (
	SortName      Sort = "name"
	SortCreatedOn Sort = "createdon"
	SortStartOn   Sort = "starton"
	SortEndOn     Sort = "endon"
)

func (s *Sort) Validate() {
	app.BadReqIf(s != nil && !(*s == SortName || *s == SortCreatedOn || *s == SortStartOn || *s == SortEndOn), "invalid sort")
}

func (s *Sort) String() string {
	return string(*s)
}

func (s *Sort) UnmarshalJSON(raw []byte) error {
	val := StrTrim(StrLower(string(raw)), `"`)
	*s = Sort(val)
	s.Validate()
	return nil
}

type Type string

const (
	TypeProject Type = "project"
	TypeUser    Type = "user"
	TypeTask    Type = "task"
	TypeVitem   Type = "vitem"
	TypeFile    Type = "file"
	TypeComment Type = "comment"
)

func (t *Type) Validate() {
	app.BadReqIf(t != nil && !(*t == TypeProject || *t == TypeUser || *t == TypeTask || *t == TypeVitem || *t == TypeFile || *t == TypeComment), "invalid type")
}

func (t *Type) String() string {
	return string(*t)
}

func (t *Type) UnmarshalJSON(raw []byte) error {
	val := StrTrim(StrLower(string(raw)), `"`)
	*t = Type(val)
	t.Validate()
	return nil
}

type Action string

const (
	ActionCreated Action = "created"
	ActionUpdated Action = "updated"
	ActionDeleted Action = "deleted"
)

func (a *Action) Validate() {
	app.BadReqIf(a != nil && !(*a == ActionCreated || *a == ActionUpdated || *a == ActionDeleted), "invalid action")
}

func (a *Action) String() string {
	return string(*a)
}

func (a *Action) UnmarshalJSON(raw []byte) error {
	val := StrTrim(StrLower(string(raw)), `"`)
	*a = Action(val)
	a.Validate()
	return nil
}

type Role uint8

const (
	RoleAdmin  Role = 0
	RoleWriter Role = 1
	RoleReader Role = 2
)

func (r *Role) Validate() {
	app.BadReqIf(r != nil && !(*r == RoleAdmin || *r == RoleWriter || *r == RoleReader), "invalid role")
}

func (r *Role) String() string {
	if r == nil {
		return ""
	}
	return strconv.Itoa(int(*r))
}

func (r *Role) UnmarshalJSON(raw []byte) error {
	val, err := strconv.ParseUint(string(raw), 10, 8)
	PanicOn(err)
	*r = Role(val)
	r.Validate()
	return nil
}
