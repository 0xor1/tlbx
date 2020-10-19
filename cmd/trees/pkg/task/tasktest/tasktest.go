package tasktest

import (
	"strings"
	"testing"

	"github.com/0xor1/tlbx/cmd/trees/pkg/config"
	"github.com/0xor1/tlbx/cmd/trees/pkg/project"
	"github.com/0xor1/tlbx/cmd/trees/pkg/project/projecteps"
	"github.com/0xor1/tlbx/cmd/trees/pkg/task"
	"github.com/0xor1/tlbx/cmd/trees/pkg/task/taskeps"
	. "github.com/0xor1/tlbx/pkg/core"
	"github.com/0xor1/tlbx/pkg/ptr"
	"github.com/0xor1/tlbx/pkg/web/app/test"
	"github.com/stretchr/testify/assert"
)

func Everything(t *testing.T) {
	a := assert.New(t)
	r := test.NewRig(
		config.Get(),
		append(projecteps.Eps, taskeps.Eps...),
		true,
		nil,
		projecteps.OnDelete,
		true,
		projecteps.OnSetSocials)
	defer r.CleanUp()

	ac := r.Ali().Client()

	p := (&project.Create{
		Name: "0",
	}).MustDo(ac)
	a.NotNil(p)

	defer func() {
		(&project.Delete{p.ID}).MustDo(ac)
	}()

	t1p0 := (&task.Create{
		Host:            r.Ali().ID(),
		Project:         p.ID,
		Parent:          p.ID,
		PreviousSibling: nil,
		Name:            "1.0",
		Description:     nil,
		IsParallel:      true,
		User:            ptr.ID(r.Ali().ID()),
		EstimatedTime:   100,
	}).MustDo(ac)
	a.NotNil(t1p0)

	t1p1 := (&task.Create{
		Host:            r.Ali().ID(),
		Project:         p.ID,
		Parent:          p.ID,
		PreviousSibling: ptr.ID(t1p0.ID),
		Name:            "1.1",
		Description:     nil,
		IsParallel:      true,
		User:            ptr.ID(r.Ali().ID()),
		EstimatedTime:   100,
	}).MustDo(ac)
	a.NotNil(t1p1)

	t2p0 := (&task.Create{
		Host:            r.Ali().ID(),
		Project:         p.ID,
		Parent:          t1p0.ID,
		PreviousSibling: nil,
		Name:            "2.0",
		Description:     nil,
		IsParallel:      true,
		User:            ptr.ID(r.Ali().ID()),
		EstimatedTime:   100,
	}).MustDo(ac)
	a.NotNil(t2p0)

	t3p0 := (&task.Create{
		Host:            r.Ali().ID(),
		Project:         p.ID,
		Parent:          t2p0.ID,
		PreviousSibling: nil,
		Name:            "3.0",
		Description:     nil,
		IsParallel:      true,
		User:            ptr.ID(r.Ali().ID()),
		EstimatedTime:   100,
	}).MustDo(ac)
	a.NotNil(t3p0)

	// t1p1 = (&task.Update{
	// 	Host:             r.Ali().ID(),
	// 	Project:          p.ID,
	// 	ID:               t1p1.ID,
	// 	Parent:           &field.ID{V: t2p0.ID},
	// 	PreviousSibling:  nil,
	// 	Name:             &field.String{V: "1.1 - updated"},
	// 	Description:      &field.StringPtr{V: ptr.String("an actual description")},
	// 	IsParallel:       &field.Bool{V: false},
	// 	User:             &field.IDPtr{V: nil},
	// 	EstimatedTime:    &field.UInt64{V: 50},
	// 	EstimatedExpense: &field.UInt64{V: 50},
	// }).MustDo(ac)
	// a.NotNil(t1p1)

	printFullTree(r, r.Ali().ID(), p.ID)
}

// only suitable for small test trees for visual validation
// whilst writing/debugging unit tests
func printFullTree(r test.Rig, host, project ID) {
	rows, err := r.Data().Primary().Query(Strf(`SELECT %s FROM tasks t WHERE t.host=? AND t.Project=?`, taskeps.Sql_task_columns_prefixed), host, project)
	if rows != nil {
		defer rows.Close()
	}
	PanicOn(err)
	ts := map[string]*task.Task{}
	for rows.Next() {
		t, err := taskeps.Scan(rows)
		PanicOn(err)
		ts[t.ID.String()] = t
	}

	var print func(t *task.Task, hs []int)
	print = func(t *task.Task, hs []int) {
		currentH := 0
		if len(hs) > 0 {
			currentH = hs[len(hs)-1]
			pre := ``
			for i, h := range hs {
				useH := h
				if i > 0 {
					useH = h - hs[i-1]
				}
				useH++
				pre += Strf(`%s|`, strings.Repeat(` `, (useH-1)*4))
			}
			Println(Strf(`%s`, pre))
			Println(Strf(`%s____%s`, pre, t.Name))
		} else {
			Println(t.Name)
		}
		if t.FirstChild != nil {
			print(ts[t.FirstChild.String()], append(hs, currentH+1))
		}
		if t.NextSibling != nil {
			print(ts[t.NextSibling.String()], hs)
		}
	}
	print(ts[project.String()], []int{0})
}
