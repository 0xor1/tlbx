package tasktest

import (
	"testing"

	"github.com/0xor1/tlbx/cmd/trees/pkg/cnsts"
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
		Description:     ptr.String(""),
		IsParallel:      true,
		User:            ptr.ID(r.Bob().ID()),
		EstimatedTime:   100,
	}).MustDo(ac)
	a.NotNil(t1p0)

	t1p1 := (&task.Create{
		Host:            r.Ali().ID(),
		Project:         p.ID,
		Parent:          p.ID,
		PreviousSibling: ptr.ID(t1p0.ID),
		Name:            "1.1",
		Description:     ptr.String("1.1"),
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

	tp := (&task.Update{
		Host:             r.Ali().ID(),
		Project:          p.ID,
		ID:               p.ID,
		IsParallel:       &field.Bool{V: false},
		EstimatedTime:    &field.UInt64{V: 50},
		EstimatedExpense: &field.UInt64{V: 50},
	}).MustDo(ac)
	a.NotNil(tp)

	tNil := (&task.Update{
		Host:    r.Ali().ID(),
		Project: p.ID,
		ID:      p.ID,
	}).MustDo(ac)
	a.Nil(tNil)

	// try to move 2.0 to be a child of itself
	tNil, err := (&task.Update{
		Host:    r.Ali().ID(),
		Project: p.ID,
		ID:      t2p0.ID,
		Parent:  &field.ID{V: t2p0.ID},
	}).Do(ac)
	a.Nil(tNil)
	a.Error(err, "ancestor loop detected, invalid parent value")

	// try to move 2.0 to be a descendant of itself
	tNil, err = (&task.Update{
		Host:    r.Ali().ID(),
		Project: p.ID,
		ID:      t2p0.ID,
		Parent:  &field.ID{V: t4p0.ID},
	}).Do(ac)
	a.Nil(tNil)
	a.Error(err, "ancestor loop detected, invalid parent value")

	t4p0 = (&task.Update{
		Host:            r.Ali().ID(),
		Project:         p.ID,
		ID:              t4p0.ID,
		Parent:          &field.ID{V: t2p0.ID},
		PreviousSibling: &field.IDPtr{V: ptr.ID(t1p1.ID)},
	}).MustDo(ac)
	a.NotNil(t4p0)

	t4p0 = (&task.Update{
		Host:            r.Ali().ID(),
		Project:         p.ID,
		ID:              t4p0.ID,
		Parent:          &field.ID{V: p.ID},
		PreviousSibling: &field.IDPtr{V: ptr.ID(t1p2.ID)},
	}).MustDo(ac)
	a.NotNil(t4p0)

	// illegal horizontal move
	tNil, err = (&task.Update{
		Host:            r.Ali().ID(),
		Project:         p.ID,
		ID:              t4p0.ID,
		Parent:          &field.ID{V: p.ID},
		PreviousSibling: &field.IDPtr{V: ptr.ID(t4p0.ID)},
	}).Do(ac)
	a.Nil(tNil)
	a.Error(err, "sibling loop detected, invalid previousSibling value")

	// horizontal move
	t4p0 = (&task.Update{
		Host:            r.Ali().ID(),
		Project:         p.ID,
		ID:              t4p0.ID,
		Parent:          &field.ID{V: p.ID},
		PreviousSibling: &field.IDPtr{V: ptr.ID(t1p0.ID)},
	}).MustDo(ac)
	a.NotNil(t4p0)

	// horizontal move to first position
	t4p0 = (&task.Update{
		Host:            r.Ali().ID(),
		Project:         p.ID,
		ID:              t4p0.ID,
		Parent:          &field.ID{V: p.ID},
		PreviousSibling: &field.IDPtr{V: nil},
		User:            &field.IDPtr{V: ptr.ID(r.Bob().ID())},
	}).MustDo(ac)
	a.NotNil(t4p0)

	t1p0 = (&task.Update{
		Host:            r.Ali().ID(),
		Project:         p.ID,
		ID:              t1p0.ID,
		PreviousSibling: &field.IDPtr{V: &t1p2.ID},
		Description:     &field.StringPtr{V: ptr.String("")},
	}).MustDo(ac)
	a.NotNil(t1p0)

	// repeat call should not change anything
	t1p0 = (&task.Update{
		Host:            r.Ali().ID(),
		Project:         p.ID,
		ID:              t1p0.ID,
		PreviousSibling: &field.IDPtr{V: &t1p2.ID},
	}).MustDo(ac)
	a.NotNil(t1p0)

	(&task.Delete{
		Host:    r.Ali().ID(),
		Project: p.ID,
		ID:      t3p0.ID,
	}).MustDo(ac)

	t1p0Get := (&task.Get{
		Host:    r.Ali().ID(),
		Project: p.ID,
		ID:      t1p0.ID,
	}).MustDo(ac)
	a.Equal(*t1p0, *t1p0Get)

	t1p1Ancestors := (&task.GetAncestors{
		Host:    r.Ali().ID(),
		Project: p.ID,
		ID:      t1p1.ID,
	}).MustDo(ac)
	a.Equal(t2p0.ID, t1p1Ancestors.Set[0].ID)
	a.Equal(t1p0.ID, t1p1Ancestors.Set[1].ID)
	a.Equal(p.ID, t1p1Ancestors.Set[2].ID)
	a.False(t1p1Ancestors.More)

	t1p1Ancestors = (&task.GetAncestors{
		Host:    r.Ali().ID(),
		Project: p.ID,
		ID:      t1p1.ID,
		Limit:   1,
	}).MustDo(ac)
	a.Equal(t2p0.ID, t1p1Ancestors.Set[0].ID)
	a.True(t1p1Ancestors.More)

	pChildren := (&task.GetChildren{
		Host:    r.Ali().ID(),
		Project: p.ID,
		ID:      p.ID,
	}).MustDo(ac)
	a.Equal(t4p0.ID, pChildren.Set[0].ID)
	a.Equal(t1p2.ID, pChildren.Set[1].ID)
	a.Equal(t1p0.ID, pChildren.Set[2].ID)
	a.False(pChildren.More)

	pChildren = (&task.GetChildren{
		Host:    r.Ali().ID(),
		Project: p.ID,
		ID:      p.ID,
		After:   ptr.ID(t4p0.ID),
	}).MustDo(ac)
	a.Equal(t1p2.ID, pChildren.Set[0].ID)
	a.Equal(t1p0.ID, pChildren.Set[1].ID)
	a.False(pChildren.More)

	pChildren = (&task.GetChildren{
		Host:    r.Ali().ID(),
		Project: p.ID,
		ID:      p.ID,
		After:   ptr.ID(t4p0.ID),
		Limit:   1,
	}).MustDo(ac)
	a.Equal(t1p2.ID, pChildren.Set[0].ID)
	a.True(pChildren.More)

	(&task.Delete{
		Host:    r.Ali().ID(),
		Project: p.ID,
		ID:      t4p0.ID,
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
