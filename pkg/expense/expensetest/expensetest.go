package expensetest

import (
	"testing"
	"time"

	"github.com/0xor1/tlbx/cmd/trees/pkg/cnsts"
	"github.com/0xor1/tlbx/cmd/trees/pkg/config"
	"github.com/0xor1/tlbx/cmd/trees/pkg/expense"
	"github.com/0xor1/tlbx/cmd/trees/pkg/expense/expenseeps"
	"github.com/0xor1/tlbx/cmd/trees/pkg/project"
	"github.com/0xor1/tlbx/cmd/trees/pkg/project/projecteps"
	"github.com/0xor1/tlbx/cmd/trees/pkg/task"
	"github.com/0xor1/tlbx/cmd/trees/pkg/task/taskeps"
	"github.com/0xor1/tlbx/cmd/trees/pkg/testutil"
	. "github.com/0xor1/tlbx/pkg/core"
	"github.com/0xor1/tlbx/pkg/field"
	"github.com/0xor1/tlbx/pkg/ptr"
	"github.com/0xor1/tlbx/pkg/web/app/test"
	"github.com/stretchr/testify/assert"
)

func Everything(t *testing.T) {
	var (
		tree map[string]*task.Task
		pID  ID
	)

	defer func() {
		testutil.PrintFullTree(pID, tree)
	}()

	a := assert.New(t)
	r := test.NewRig(
		config.Get(),
		append(append(projecteps.Eps, taskeps.Eps...), expenseeps.Eps...),
		true,
		nil,
		projecteps.OnDelete,
		true,
		projecteps.OnSetSocials,
		cnsts.TempFileBucket,
		cnsts.FileBucket)
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
		Host:             r.Ali().ID(),
		Project:          p.ID,
		Parent:           p.ID,
		PreviousSibling:  nil,
		Name:             "1.0",
		Description:      "",
		IsParallel:       true,
		User:             ptr.ID(r.Bob().ID()),
		EstimatedExpense: 100,
	}).MustDo(ac)
	a.NotNil(t1p0)

	e1 := (&expense.Create{
		Host:    r.Ali().ID(),
		Project: p.ID,
		Task:    t1p0.ID,
		Value:   77,
		Note:    "yolo",
	}).MustDo(ac)
	a.NotNil(e1)

	e1 = (&expense.Update{
		Host:    r.Ali().ID(),
		Project: p.ID,
		Task:    t1p0.ID,
		ID:      e1.ID,
		Value:   &field.UInt64{V: 33},
		Note:    &field.String{V: "polo"},
	}).MustDo(ac)
	a.NotNil(e1)

	e1 = (&expense.Update{
		Host:    r.Ali().ID(),
		Project: p.ID,
		Task:    t1p0.ID,
		ID:      e1.ID,
		Value:   &field.UInt64{V: 44},
		Note:    &field.String{V: "polo"},
	}).MustDo(ac)
	a.NotNil(e1)

	// nil
	eNil := (&expense.Update{
		Host:    r.Ali().ID(),
		Project: p.ID,
		Task:    t1p0.ID,
		ID:      e1.ID,
	}).MustDo(ac)
	a.Nil(eNil)

	es := (&expense.Get{
		Host:         r.Ali().ID(),
		Project:      p.ID,
		Task:         &t1p0.ID,
		CreatedBy:    ptr.ID(r.Ali().ID()),
		CreatedOnMin: ptr.Time(Now().Add(-1 * time.Hour)),
		CreatedOnMax: ptr.Time(Now()),
	}).MustDo(ac)
	a.Equal(e1, es.Set[0])
	a.False(es.More)

	es = (&expense.Get{
		Host:    r.Ali().ID(),
		Project: p.ID,
		IDs:     IDs{e1.ID},
	}).MustDo(ac)
	a.Equal(e1, es.Set[0])
	a.False(es.More)

	e2 := (&expense.Create{
		Host:    r.Ali().ID(),
		Project: p.ID,
		Task:    t1p0.ID,
		Value:   77,
		Note:    "solo",
	}).MustDo(ac)
	a.NotNil(e2)

	es = (&expense.Get{
		Host:         r.Ali().ID(),
		Project:      p.ID,
		Task:         &t1p0.ID,
		CreatedBy:    ptr.ID(r.Ali().ID()),
		CreatedOnMin: ptr.Time(Now().Add(-1 * time.Hour)),
		CreatedOnMax: ptr.Time(Now()),
	}).MustDo(ac)
	a.Equal(e2, es.Set[0])
	a.Equal(e1, es.Set[1])
	a.False(es.More)

	es = (&expense.Get{
		Host:         r.Ali().ID(),
		Project:      p.ID,
		Task:         &t1p0.ID,
		CreatedBy:    ptr.ID(r.Ali().ID()),
		CreatedOnMin: ptr.Time(Now().Add(-1 * time.Hour)),
		CreatedOnMax: ptr.Time(Now()),
		Limit:        1,
	}).MustDo(ac)
	a.Equal(e2, es.Set[0])
	a.True(es.More)

	es = (&expense.Get{
		Host:         r.Ali().ID(),
		Project:      p.ID,
		Task:         &t1p0.ID,
		CreatedBy:    ptr.ID(r.Ali().ID()),
		CreatedOnMin: ptr.Time(Now().Add(-1 * time.Hour)),
		CreatedOnMax: ptr.Time(Now()),
		After:        ptr.ID(e2.ID),
		Limit:        1,
	}).MustDo(ac)
	a.Equal(e1, es.Set[0])
	a.False(es.More)

	(&expense.Delete{
		Host:    r.Ali().ID(),
		Project: p.ID,
		Task:    t1p0.ID,
		ID:      e1.ID,
	}).MustDo(ac)

	pID = p.ID
	tree = testutil.GrabFullTree(r, r.Ali().ID(), p.ID)
}
