package tasktest

import (
	"bytes"
	"io/ioutil"
	"testing"

	"github.com/0xor1/tlbx/cmd/trees/pkg/cnsts"
	"github.com/0xor1/tlbx/cmd/trees/pkg/file"
	"github.com/0xor1/tlbx/cmd/trees/pkg/project"
	"github.com/0xor1/tlbx/cmd/trees/pkg/task"
	. "github.com/0xor1/tlbx/pkg/core"
	"github.com/0xor1/tlbx/pkg/field"
	"github.com/0xor1/tlbx/pkg/ptr"
	"github.com/0xor1/tlbx/pkg/web/app/test"
	"github.com/0xor1/trees/pkg/config"
	"github.com/0xor1/trees/pkg/file/fileeps"
	"github.com/0xor1/trees/pkg/project/projecteps"
	"github.com/0xor1/trees/pkg/task/taskeps"
	"github.com/0xor1/trees/pkg/testutil"
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
		append(append(projecteps.Eps, taskeps.Eps...), fileeps.Eps...),
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

	createRes := (&task.Create{
		Host:       r.Ali().ID(),
		Project:    p.ID,
		Parent:     p.ID,
		PrevSib:    nil,
		Name:       "1.0",
		IsParallel: true,
		User:       ptr.ID(r.Bob().ID()),
		TimeEst:    100,
	}).MustDo(ac)
	a.NotNil(createRes)
	a.True(p.ID.Equal(createRes.Parent.ID))

	t1p0 := createRes.Task

	t1p1 := (&task.Create{
		Host:        r.Ali().ID(),
		Project:     p.ID,
		Parent:      p.ID,
		PrevSib:     ptr.ID(t1p0.ID),
		Name:        "1.1",
		Description: "1.1",
		IsParallel:  true,
		User:        ptr.ID(r.Ali().ID()),
		TimeEst:     100,
	}).MustDo(ac).Task
	a.NotNil(createRes)

	// move first child to a different position
	t1p0 = (&task.Update{
		Host:    r.Ali().ID(),
		Project: p.ID,
		ID:      t1p0.ID,
		PrevSib: &field.IDPtr{V: &t1p1.ID},
	}).MustDo(ac).Task

	t2p0 := (&task.Create{
		Host:        r.Ali().ID(),
		Project:     p.ID,
		Parent:      t1p0.ID,
		PrevSib:     nil,
		Name:        "2.0",
		Description: "",
		IsParallel:  true,
		User:        ptr.ID(r.Ali().ID()),
		TimeEst:     100,
	}).MustDo(ac).Task
	a.NotNil(t2p0)

	t3p0 := (&task.Create{
		Host:        r.Ali().ID(),
		Project:     p.ID,
		Parent:      t2p0.ID,
		PrevSib:     nil,
		Name:        "3.0",
		Description: "",
		IsParallel:  true,
		User:        ptr.ID(r.Ali().ID()),
		TimeEst:     100,
	}).MustDo(ac).Task
	a.NotNil(t3p0)

	t4p0 := (&task.Create{
		Host:        r.Ali().ID(),
		Project:     p.ID,
		Parent:      t1p1.ID,
		PrevSib:     nil,
		Name:        "4.0",
		Description: "",
		IsParallel:  true,
		User:        ptr.ID(r.Ali().ID()),
		TimeEst:     100,
	}).MustDo(ac).Task
	a.NotNil(t4p0)

	t1p2 := (&task.Create{
		Host:        r.Ali().ID(),
		Project:     p.ID,
		Parent:      p.ID,
		PrevSib:     ptr.ID(t1p0.ID),
		Name:        "1.2",
		Description: "",
		IsParallel:  true,
		User:        ptr.ID(r.Ali().ID()),
		TimeEst:     100,
	}).MustDo(ac).Task
	a.NotNil(t1p2)

	uRes := (&task.Update{
		Host:        r.Ali().ID(),
		Project:     p.ID,
		ID:          t1p1.ID,
		Parent:      &field.ID{V: t2p0.ID},
		PrevSib:     nil,
		Name:        &field.String{V: "1.1 - updated"},
		Description: &field.String{V: "an actual description"},
		IsParallel:  &field.Bool{V: false},
		User:        &field.IDPtr{V: nil},
		TimeEst:     &field.UInt64{V: 50},
		CostEst:     &field.UInt64{V: 50},
	}).MustDo(ac)
	a.NotNil(uRes)
	a.True(t1p1.ID.Equal(uRes.Task.ID))
	a.True(p.ID.Equal(uRes.OldParent.ID))
	a.True(t2p0.ID.Equal(uRes.NewParent.ID))

	tp := (&task.Update{
		Host:       r.Ali().ID(),
		Project:    p.ID,
		ID:         p.ID,
		IsParallel: &field.Bool{V: false},
		TimeEst:    &field.UInt64{V: 50},
		CostEst:    &field.UInt64{V: 50},
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
		Host:    r.Ali().ID(),
		Project: p.ID,
		ID:      t4p0.ID,
		Parent:  &field.ID{V: t2p0.ID},
		PrevSib: &field.IDPtr{V: ptr.ID(t1p1.ID)},
	}).MustDo(ac).Task
	a.NotNil(t4p0)

	// create new sub tree to test move special case
	t5p0 := (&task.Create{
		Host:    r.Ali().ID(),
		Project: p.ID,
		Parent:  t4p0.ID,
		Name:    "5.0",
	}).MustDo(ac).Task

	t6p0 := (&task.Create{
		Host:    r.Ali().ID(),
		Project: p.ID,
		Parent:  t5p0.ID,
		Name:    "6.0",
	}).MustDo(ac).Task

	t6p1 := (&task.Create{
		Host:    r.Ali().ID(),
		Project: p.ID,
		Parent:  t5p0.ID,
		PrevSib: ptr.ID(t6p0.ID),
		Name:    "6.1",
	}).MustDo(ac).Task

	//now move 6.1 under 6.0
	uRes = (&task.Update{
		Host:    r.Ali().ID(),
		Project: p.ID,
		ID:      t6p1.ID,
		Parent:  &field.ID{V: t6p0.ID},
	}).MustDo(ac)
	a.True(uRes.OldParent.ID.Equal(t5p0.ID))
	a.True(uRes.NewParent.ID.Equal(t6p0.ID))
	a.True(uRes.NewParent.NextSib == nil)
	a.True(uRes.NewParent.FirstChild.Equal(uRes.Task.ID))
	a.True(uRes.Task.ID.Equal(t6p1.ID))

	t4p0 = (&task.Update{
		Host:    r.Ali().ID(),
		Project: p.ID,
		ID:      t4p0.ID,
		Parent:  &field.ID{V: p.ID},
		PrevSib: &field.IDPtr{V: ptr.ID(t1p2.ID)},
	}).MustDo(ac).Task
	a.NotNil(t4p0)

	// illegal horizontal move
	tNil, err = (&task.Update{
		Host:    r.Ali().ID(),
		Project: p.ID,
		ID:      t4p0.ID,
		Parent:  &field.ID{V: p.ID},
		PrevSib: &field.IDPtr{V: ptr.ID(t4p0.ID)},
	}).Do(ac)
	a.Nil(tNil)
	a.Error(err, "sib loop detected, invalid prevSib value")

	// horizontal move
	t4p0 = (&task.Update{
		Host:    r.Ali().ID(),
		Project: p.ID,
		ID:      t4p0.ID,
		Parent:  &field.ID{V: p.ID},
		PrevSib: &field.IDPtr{V: ptr.ID(t1p0.ID)},
	}).MustDo(ac).Task
	a.NotNil(t4p0)

	// horizontal move to first position
	t4p0 = (&task.Update{
		Host:    r.Ali().ID(),
		Project: p.ID,
		ID:      t4p0.ID,
		Parent:  &field.ID{V: p.ID},
		PrevSib: &field.IDPtr{V: nil},
		User:    &field.IDPtr{V: ptr.ID(r.Bob().ID())},
	}).MustDo(ac).Task
	a.NotNil(t4p0)

	t1p0 = (&task.Update{
		Host:        r.Ali().ID(),
		Project:     p.ID,
		ID:          t1p0.ID,
		PrevSib:     &field.IDPtr{V: &t1p2.ID},
		Description: &field.String{V: ""},
	}).MustDo(ac).Task
	a.NotNil(t1p0)

	// repeat call should not change anything
	t1p0 = (&task.Update{
		Host:    r.Ali().ID(),
		Project: p.ID,
		ID:      t1p0.ID,
		PrevSib: &field.IDPtr{V: &t1p2.ID},
	}).MustDo(ac).Task
	a.NotNil(t1p0)

	content1 := []byte("yolo")
	put := &file.Put{
		Args: &file.PutArgs{
			Host:    r.Ali().ID(),
			Project: p.ID,
			Task:    t1p0.ID,
		},
	}
	put.Name = "yolo.test.txt"
	put.Type = "text/plain"
	put.Size = int64(len(content1))
	put.Content = ioutil.NopCloser(bytes.NewBuffer(content1))
	f := put.MustDo(ac).File
	a.NotNil(f)

	t2p0 = (&task.Get{
		Host:    r.Ali().ID(),
		Project: p.ID,
		ID:      t2p0.ID,
	}).MustDo(ac)

	deleteParent := (&task.Delete{
		Host:    r.Ali().ID(),
		Project: p.ID,
		ID:      t3p0.ID,
	}).MustDo(ac)
	a.True(t2p0.ID.Equal(deleteParent.ID))
	a.Equal(t2p0.ChildN-1, deleteParent.ChildN)

	t1p0Get := (&task.Get{
		Host:    r.Ali().ID(),
		Project: p.ID,
		ID:      t1p0.ID,
	}).MustDo(ac)
	a.True(t1p0.ID.Equal(t1p0Get.ID))

	t1p1Ancestors := (&task.GetAncestors{
		Host:    r.Ali().ID(),
		Project: p.ID,
		ID:      p.ID,
	}).MustDo(ac)
	a.Equal(len(t1p1Ancestors.Set), 0)
	a.False(t1p1Ancestors.More)

	t1p1Ancestors = (&task.GetAncestors{
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

	pID = p.ID
	tree = testutil.GrabFullTree(r, r.Ali().ID(), p.ID)
}
