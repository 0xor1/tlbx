package timetest

import (
	"testing"
	time_ "time"

	. "github.com/0xor1/tlbx/pkg/core"
	"github.com/0xor1/tlbx/pkg/field"
	"github.com/0xor1/tlbx/pkg/ptr"
	"github.com/0xor1/tlbx/pkg/web/app/test"
	"github.com/0xor1/trees/pkg/cnsts"
	"github.com/0xor1/trees/pkg/config"
	"github.com/0xor1/trees/pkg/project"
	"github.com/0xor1/trees/pkg/project/projecteps"
	"github.com/0xor1/trees/pkg/task"
	"github.com/0xor1/trees/pkg/task/taskeps"
	"github.com/0xor1/trees/pkg/testutil"
	"github.com/0xor1/trees/pkg/time"
	"github.com/0xor1/trees/pkg/time/timeeps"
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
		append(append(projecteps.Eps, taskeps.Eps...), timeeps.Eps...),
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

	t1 := (&time.Create{
		Host:     r.Ali().ID(),
		Project:  p.ID,
		Task:     t1p0.ID,
		Duration: 77,
		Note:     "yolo",
	}).MustDo(ac)
	a.NotNil(t1)

	t1 = (&time.Update{
		Host:     r.Ali().ID(),
		Project:  p.ID,
		Task:     t1p0.ID,
		ID:       t1.ID,
		Duration: &field.UInt64{V: 33},
		Note:     &field.String{V: "polo"},
	}).MustDo(ac)
	a.NotNil(t1)

	t1 = (&time.Update{
		Host:     r.Ali().ID(),
		Project:  p.ID,
		Task:     t1p0.ID,
		ID:       t1.ID,
		Duration: &field.UInt64{V: 44},
		Note:     &field.String{V: "polo"},
	}).MustDo(ac)
	a.NotNil(t1)

	// nil
	tNil := (&time.Update{
		Host:    r.Ali().ID(),
		Project: p.ID,
		Task:    t1p0.ID,
		ID:      t1.ID,
	}).MustDo(ac)
	a.Nil(tNil)

	ts := (&time.Get{
		Host:         r.Ali().ID(),
		Project:      p.ID,
		Task:         &t1p0.ID,
		CreatedBy:    ptr.ID(r.Ali().ID()),
		CreatedOnMin: ptr.Time(Now().Add(-1 * time_.Hour)),
		CreatedOnMax: ptr.Time(Now()),
	}).MustDo(ac)
	a.Equal(t1, ts.Set[0])
	a.False(ts.More)

	ts = (&time.Get{
		Host:    r.Ali().ID(),
		Project: p.ID,
		IDs:     IDs{t1.ID},
	}).MustDo(ac)
	a.Equal(t1, ts.Set[0])
	a.False(ts.More)

	t2 := (&time.Create{
		Host:     r.Ali().ID(),
		Project:  p.ID,
		Task:     t1p0.ID,
		Duration: 77,
		Note:     "solo",
	}).MustDo(ac)
	a.NotNil(t2)

	ts = (&time.Get{
		Host:         r.Ali().ID(),
		Project:      p.ID,
		Task:         &t1p0.ID,
		CreatedBy:    ptr.ID(r.Ali().ID()),
		CreatedOnMin: ptr.Time(Now().Add(-1 * time_.Hour)),
		CreatedOnMax: ptr.Time(Now()),
	}).MustDo(ac)
	a.Equal(t2, ts.Set[0])
	a.Equal(t1, ts.Set[1])
	a.False(ts.More)

	ts = (&time.Get{
		Host:         r.Ali().ID(),
		Project:      p.ID,
		Task:         &t1p0.ID,
		CreatedBy:    ptr.ID(r.Ali().ID()),
		CreatedOnMin: ptr.Time(Now().Add(-1 * time_.Hour)),
		CreatedOnMax: ptr.Time(Now()),
		Limit:        1,
	}).MustDo(ac)
	a.Equal(t2, ts.Set[0])
	a.True(ts.More)

	ts = (&time.Get{
		Host:         r.Ali().ID(),
		Project:      p.ID,
		Task:         &t1p0.ID,
		CreatedBy:    ptr.ID(r.Ali().ID()),
		CreatedOnMin: ptr.Time(Now().Add(-1 * time_.Hour)),
		CreatedOnMax: ptr.Time(Now()),
		After:        ptr.ID(t2.ID),
		Limit:        1,
	}).MustDo(ac)
	a.Equal(t1, ts.Set[0])
	a.False(ts.More)

	(&time.Delete{
		Host:    r.Ali().ID(),
		Project: p.ID,
		Task:    t1p0.ID,
		ID:      t1.ID,
	}).MustDo(ac)

	pID = p.ID
	tree = testutil.GrabFullTree(r, r.Ali().ID(), p.ID)
}
