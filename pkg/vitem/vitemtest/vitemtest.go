package vitemtest

import (
	"testing"
	time_ "time"

	"github.com/0xor1/tlbx/cmd/trees/pkg/cnsts"
	"github.com/0xor1/tlbx/cmd/trees/pkg/project"
	"github.com/0xor1/tlbx/cmd/trees/pkg/task"
	"github.com/0xor1/tlbx/cmd/trees/pkg/vitem"
	. "github.com/0xor1/tlbx/pkg/core"
	"github.com/0xor1/tlbx/pkg/field"
	"github.com/0xor1/tlbx/pkg/ptr"
	"github.com/0xor1/tlbx/pkg/web/app/test"
	"github.com/0xor1/trees/pkg/config"
	"github.com/0xor1/trees/pkg/project/projecteps"
	"github.com/0xor1/trees/pkg/task/taskeps"
	"github.com/0xor1/trees/pkg/testutil"
	"github.com/0xor1/trees/pkg/vitem/vitemeps"
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
		append(append(projecteps.Eps, taskeps.Eps...), vitemeps.Eps...),
		true,
		nil,
		projecteps.OnDelete,
		true,
		projecteps.OnSetSocials,
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
		Host:        r.Ali().ID(),
		Project:     p.ID,
		Parent:      p.ID,
		PrevSib:     nil,
		Name:        "1.0",
		Description: "",
		IsParallel:  true,
		User:        ptr.ID(r.Bob().ID()),
		TimeEst:     100,
	}).MustDo(ac).Task
	a.NotNil(t1p0)

	t1Res := (&vitem.Create{
		Host:    r.Ali().ID(),
		Project: p.ID,
		Task:    t1p0.ID,
		Type:    vitem.TypeTime,
		Est:     ptr.Uint64(45),
		Inc:     77,
		Note:    "yolo",
	}).MustDo(ac)
	a.NotNil(t1Res)
	a.True(t1Res.Task.ID.Equal(t1p0.ID))
	a.True(t1Res.Task.TimeEst == 45)
	t1 := t1Res.Item
	a.NotNil(t1)

	t1Res = (&vitem.Update{
		Host:    r.Ali().ID(),
		Project: p.ID,
		Task:    t1p0.ID,
		Type:    vitem.TypeTime,
		ID:      t1.ID,
		Inc:     &field.UInt64{V: 33},
		Note:    &field.String{V: "polo"},
	}).MustDo(ac)
	a.NotNil(t1Res)
	a.True(t1Res.Task.TimeInc == 33)
	a.True(t1Res.Item.Inc == 33)

	t1 = (&vitem.Update{
		Host:    r.Ali().ID(),
		Project: p.ID,
		Task:    t1p0.ID,
		Type:    vitem.TypeTime,
		ID:      t1.ID,
		Inc:     &field.UInt64{V: 44},
		Note:    &field.String{V: "polo"},
	}).MustDo(ac).Item
	a.NotNil(t1)

	tNil := (&vitem.Update{
		Host:    r.Ali().ID(),
		Project: p.ID,
		Task:    t1p0.ID,
		Type:    vitem.TypeTime,
		ID:      t1.ID,
	}).MustDo(ac)
	a.Nil(tNil)

	ts := (&vitem.Get{
		Host:         r.Ali().ID(),
		Project:      p.ID,
		Task:         &t1p0.ID,
		Type:         vitem.TypeTime,
		CreatedBy:    ptr.ID(r.Ali().ID()),
		CreatedOnMin: ptr.Time(Now().Add(-1 * time_.Hour)),
		CreatedOnMax: ptr.Time(Now()),
	}).MustDo(ac)
	a.Equal(t1, ts.Set[0])
	a.False(ts.More)

	ts = (&vitem.Get{
		Host:    r.Ali().ID(),
		Project: p.ID,
		Type:    vitem.TypeTime,
		IDs:     IDs{t1.ID},
	}).MustDo(ac)
	a.Equal(t1, ts.Set[0])
	a.False(ts.More)

	t2 := (&vitem.Create{
		Host:    r.Ali().ID(),
		Project: p.ID,
		Task:    t1p0.ID,
		Type:    vitem.TypeTime,
		Inc:     77,
		Note:    "solo",
	}).MustDo(ac).Item
	a.NotNil(t2)

	ts = (&vitem.Get{
		Host:         r.Ali().ID(),
		Project:      p.ID,
		Type:         vitem.TypeTime,
		Task:         &t1p0.ID,
		CreatedBy:    ptr.ID(r.Ali().ID()),
		CreatedOnMin: ptr.Time(Now().Add(-1 * time_.Hour)),
		CreatedOnMax: ptr.Time(Now()),
	}).MustDo(ac)
	a.Equal(t2, ts.Set[0])
	a.Equal(t1, ts.Set[1])
	a.False(ts.More)

	ts = (&vitem.Get{
		Host:         r.Ali().ID(),
		Project:      p.ID,
		Type:         vitem.TypeTime,
		Task:         &t1p0.ID,
		CreatedBy:    ptr.ID(r.Ali().ID()),
		CreatedOnMin: ptr.Time(Now().Add(-1 * time_.Hour)),
		CreatedOnMax: ptr.Time(Now()),
		Limit:        1,
	}).MustDo(ac)
	a.Equal(t2, ts.Set[0])
	a.True(ts.More)

	ts = (&vitem.Get{
		Host:         r.Ali().ID(),
		Project:      p.ID,
		Type:         vitem.TypeTime,
		Task:         &t1p0.ID,
		CreatedBy:    ptr.ID(r.Ali().ID()),
		CreatedOnMin: ptr.Time(Now().Add(-1 * time_.Hour)),
		CreatedOnMax: ptr.Time(Now()),
		After:        ptr.ID(t2.ID),
		Limit:        1,
	}).MustDo(ac)
	a.Equal(t1, ts.Set[0])
	a.False(ts.More)

	t1p0 = (&vitem.Delete{
		Host:    r.Ali().ID(),
		Project: p.ID,
		Task:    t1p0.ID,
		Type:    vitem.TypeTime,
		ID:      t1.ID,
	}).MustDo(ac)
	a.True(t1p0.TimeInc == 77)
	a.True(t1p0.CostInc == 0)

	// costs section

	t1Res = (&vitem.Create{
		Host:    r.Ali().ID(),
		Project: p.ID,
		Task:    t1p0.ID,
		Type:    vitem.TypeCost,
		Est:     ptr.Uint64(45),
		Inc:     77,
		Note:    "yolo",
	}).MustDo(ac)
	a.NotNil(t1Res)
	a.True(t1Res.Task.ID.Equal(t1p0.ID))
	a.True(t1Res.Task.CostEst == 45)
	t1 = t1Res.Item
	a.NotNil(t1)

	t1Res = (&vitem.Update{
		Host:    r.Ali().ID(),
		Project: p.ID,
		Task:    t1p0.ID,
		Type:    vitem.TypeCost,
		ID:      t1.ID,
		Inc:     &field.UInt64{V: 33},
		Note:    &field.String{V: "polo"},
	}).MustDo(ac)
	a.NotNil(t1Res)
	a.True(t1Res.Task.CostInc == 33)
	a.True(t1Res.Item.Inc == 33)

	t1 = (&vitem.Update{
		Host:    r.Ali().ID(),
		Project: p.ID,
		Task:    t1p0.ID,
		Type:    vitem.TypeCost,
		ID:      t1.ID,
		Inc:     &field.UInt64{V: 44},
		Note:    &field.String{V: "polo"},
	}).MustDo(ac).Item
	a.NotNil(t1)

	tNil = (&vitem.Update{
		Host:    r.Ali().ID(),
		Project: p.ID,
		Task:    t1p0.ID,
		Type:    vitem.TypeCost,
		ID:      t1.ID,
	}).MustDo(ac)
	a.Nil(tNil)

	ts = (&vitem.Get{
		Host:         r.Ali().ID(),
		Project:      p.ID,
		Task:         &t1p0.ID,
		Type:         vitem.TypeCost,
		CreatedBy:    ptr.ID(r.Ali().ID()),
		CreatedOnMin: ptr.Time(Now().Add(-1 * time_.Hour)),
		CreatedOnMax: ptr.Time(Now()),
	}).MustDo(ac)
	a.Equal(t1, ts.Set[0])
	a.False(ts.More)

	ts = (&vitem.Get{
		Host:    r.Ali().ID(),
		Project: p.ID,
		Type:    vitem.TypeCost,
		IDs:     IDs{t1.ID},
	}).MustDo(ac)
	a.Equal(t1, ts.Set[0])
	a.False(ts.More)

	t2 = (&vitem.Create{
		Host:    r.Ali().ID(),
		Project: p.ID,
		Task:    t1p0.ID,
		Type:    vitem.TypeCost,
		Inc:     77,
		Note:    "solo",
	}).MustDo(ac).Item
	a.NotNil(t2)

	ts = (&vitem.Get{
		Host:         r.Ali().ID(),
		Project:      p.ID,
		Type:         vitem.TypeCost,
		Task:         &t1p0.ID,
		CreatedBy:    ptr.ID(r.Ali().ID()),
		CreatedOnMin: ptr.Time(Now().Add(-1 * time_.Hour)),
		CreatedOnMax: ptr.Time(Now()),
	}).MustDo(ac)
	a.Equal(t2, ts.Set[0])
	a.Equal(t1, ts.Set[1])
	a.False(ts.More)

	ts = (&vitem.Get{
		Host:         r.Ali().ID(),
		Project:      p.ID,
		Type:         vitem.TypeCost,
		Task:         &t1p0.ID,
		CreatedBy:    ptr.ID(r.Ali().ID()),
		CreatedOnMin: ptr.Time(Now().Add(-1 * time_.Hour)),
		CreatedOnMax: ptr.Time(Now()),
		Limit:        1,
	}).MustDo(ac)
	a.Equal(t2, ts.Set[0])
	a.True(ts.More)

	ts = (&vitem.Get{
		Host:         r.Ali().ID(),
		Project:      p.ID,
		Type:         vitem.TypeCost,
		Task:         &t1p0.ID,
		CreatedBy:    ptr.ID(r.Ali().ID()),
		CreatedOnMin: ptr.Time(Now().Add(-1 * time_.Hour)),
		CreatedOnMax: ptr.Time(Now()),
		After:        ptr.ID(t2.ID),
		Limit:        1,
	}).MustDo(ac)
	a.Equal(t1, ts.Set[0])
	a.False(ts.More)

	t1p0 = (&vitem.Delete{
		Host:    r.Ali().ID(),
		Project: p.ID,
		Task:    t1p0.ID,
		Type:    vitem.TypeCost,
		ID:      t1.ID,
	}).MustDo(ac)
	a.True(t1p0.CostInc == 77)

	pID = p.ID
	tree = testutil.GrabFullTree(r, r.Ali().ID(), p.ID)
}
