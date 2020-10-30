package timetest

import (
	"testing"

	"github.com/0xor1/tlbx/cmd/trees/pkg/cnsts"
	"github.com/0xor1/tlbx/cmd/trees/pkg/config"
	"github.com/0xor1/tlbx/cmd/trees/pkg/project"
	"github.com/0xor1/tlbx/cmd/trees/pkg/project/projecteps"
	"github.com/0xor1/tlbx/cmd/trees/pkg/task"
	"github.com/0xor1/tlbx/cmd/trees/pkg/task/taskeps"
	"github.com/0xor1/tlbx/cmd/trees/pkg/time"
	. "github.com/0xor1/tlbx/pkg/core"
	"github.com/0xor1/tlbx/pkg/ptr"
	"github.com/0xor1/tlbx/pkg/web/app/test"
	"github.com/stretchr/testify/assert"
)

func Everything(t *testing.T) {
	defer printFullTree()

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

	(&project.AddUsers{
		Host:    r.Ali().ID(),
		Project: p.ID,
		Users: []*project.SendUser{
			{
				ID:   r.Bob().ID(),
				Role: cnsts.RoleWriter,
			},
		},
	}).MustDo(ac)

	t1p0 := (&task.Create{
		Host:            r.Ali().ID(),
		Project:         p.ID,
		Parent:          p.ID,
		PreviousSibling: nil,
		Name:            "1.0",
		Description:     "",
		IsParallel:      true,
		User:            ptr.ID(r.Bob().ID()),
		EstimatedTime:   100,
	}).MustDo(ac)
	a.NotNil(t1p0)

	(&time.Create{
		H
	}).MustDo(ac)

	grabFullTree(r, r.Ali().ID(), p.ID)
}

var (
	fullTree = map[string]*task.Task{}
	pID      ID
)

// only suitable for small test trees for visual validation
// whilst writing/debugging unit tests
func grabFullTree(r test.Rig, host, project ID) {
	pID = project
	rows, err := r.Data().Primary().Query(Strf(`SELECT %s FROM tasks t WHERE t.host=? AND t.Project=?`, taskeps.Sql_task_columns_prefixed), host, project)
	if rows != nil {
		defer rows.Close()
	}
	PanicOn(err)
	for rows.Next() {
		t, err := taskeps.Scan(rows)
		PanicOn(err)
		fullTree[t.ID.String()] = t
	}
}

func printFullTree() {
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
			print(fullTree[t.FirstChild.String()], append(as, t))
		}
		if t.NextSibling != nil {
			print(fullTree[t.NextSibling.String()], as)
		}
	}
	println("n: name")
	println("p: isParallel")
	println("m: minimumTime")
	println("e: estimatedTime")
	println("es: estimatedSubTime")
	println()
	print(fullTree[pID.String()], nil)
}
