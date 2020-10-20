package tasktest

import (
	"testing"

	"github.com/0xor1/tlbx/cmd/trees/pkg/config"
	"github.com/0xor1/tlbx/cmd/trees/pkg/project"
	"github.com/0xor1/tlbx/cmd/trees/pkg/project/projecteps"
	"github.com/0xor1/tlbx/cmd/trees/pkg/task"
	"github.com/0xor1/tlbx/cmd/trees/pkg/task/taskeps"
	. "github.com/0xor1/tlbx/pkg/core"
	"github.com/0xor1/tlbx/pkg/field"
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

	t4p0 := (&task.Create{
		Host:            r.Ali().ID(),
		Project:         p.ID,
		Parent:          t1p1.ID,
		PreviousSibling: nil,
		Name:            "4.0",
		Description:     nil,
		IsParallel:      true,
		User:            ptr.ID(r.Ali().ID()),
		EstimatedTime:   100,
	}).MustDo(ac)
	a.NotNil(t4p0)

	t1p2 := (&task.Create{
		Host:            r.Ali().ID(),
		Project:         p.ID,
		Parent:          p.ID,
		PreviousSibling: ptr.ID(t1p0.ID),
		Name:            "1.2",
		Description:     nil,
		IsParallel:      true,
		User:            ptr.ID(r.Ali().ID()),
		EstimatedTime:   100,
	}).MustDo(ac)
	a.NotNil(t1p2)

	t1p1 = (&task.Update{
		Host:             r.Ali().ID(),
		Project:          p.ID,
		ID:               t1p1.ID,
		Parent:           &field.ID{V: t2p0.ID},
		PreviousSibling:  nil,
		Name:             &field.String{V: "1.1 - updated"},
		Description:      &field.StringPtr{V: ptr.String("an actual description")},
		IsParallel:       &field.Bool{V: false},
		User:             &field.IDPtr{V: nil},
		EstimatedTime:    &field.UInt64{V: 50},
		EstimatedExpense: &field.UInt64{V: 50},
	}).MustDo(ac)
	a.NotNil(t1p1)

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

	var print func(t *task.Task, as []*task.Task)
	print = func(t *task.Task, as []*task.Task) {
		p := 0
		if t.IsParallel {
			p = 1
		}
		v := Strf(`[n: %s, p: %d, m: %d, e: %d, es: %d]`, t.Name, p, t.MinimumTime, t.EstimatedTime, t.EstimatedSubTime)
		if len(as) > 0 {
			pre := ``
			for _, a := range as[1:] {
				if a.NextSibling != nil {
					pre += `|    `
				} else {
					pre += `     `
				}
			}
			Println(Strf(`%s|`, pre))
			Println(Strf(`%s|`, pre))
			Println(Strf(`%s|____%s`, pre, v))
		} else {
			Println(v)
		}
		if t.FirstChild != nil {
			print(ts[t.FirstChild.String()], append(as, t))
		}
		if t.NextSibling != nil {
			print(ts[t.NextSibling.String()], as)
		}
	}
	println("n: name")
	println("p: isParallel")
	println("m: minimumTime")
	println("e: estimatedTime")
	println("es: estimatedSubTime")
	println()
	print(ts[project.String()], nil)
}
