package filetest

import (
	"bytes"
	"io/ioutil"
	"testing"
	"time"

	"github.com/0xor1/tlbx/cmd/trees/pkg/cnsts"
	"github.com/0xor1/tlbx/cmd/trees/pkg/file"
	"github.com/0xor1/tlbx/cmd/trees/pkg/project"
	"github.com/0xor1/tlbx/cmd/trees/pkg/task"
	. "github.com/0xor1/tlbx/pkg/core"
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

	t1p0 := (&task.Create{
		Host:        r.Ali().ID(),
		Project:     p.ID,
		Parent:      p.ID,
		PrevSib:     nil,
		Name:        "1.0",
		Description: "",
		IsParallel:  true,
		User:        ptr.ID(r.Bob().ID()),
	}).MustDo(ac).Task
	a.NotNil(t1p0)

	content1 := []byte("1")
	put := &file.Put{
		Args: &file.PutArgs{
			Host:    r.Ali().ID(),
			Project: p.ID,
			Task:    t1p0.ID,
		},
	}
	put.Name = "one"
	put.Type = "text/plain"
	put.Size = int64(len(content1))
	put.Content = ioutil.NopCloser(bytes.NewBuffer(content1))
	putRes := put.MustDo(ac)
	a.True(putRes.Task.ID.Equal(t1p0.ID))
	a.Equal(putRes.Task.FileN, uint64(1))
	a.Equal(putRes.Task.FileSize, uint64(1))
	f1 := putRes.File
	a.NotNil(f1)

	f1Get := (&file.GetContent{
		Host:    r.Ali().ID(),
		Project: p.ID,
		Task:    t1p0.ID,
		ID:      f1.ID,
	}).MustDo(ac)
	bs, err := ioutil.ReadAll(f1Get.Content)
	a.Nil(err)
	a.Equal(content1, bs)
	a.Equal(put.Name, f1Get.Name)
	a.Equal(put.Size, f1Get.Size)
	a.Equal(put.Type, f1Get.Type)

	content2 := []byte("2")
	put.Content = ioutil.NopCloser(bytes.NewBuffer(content2))
	put.Name = "two"
	f2 := put.MustDo(ac).File
	a.NotNil(f2)

	res := (&file.Get{
		Host:    r.Ali().ID(),
		Project: p.ID,
		IDs:     IDs{f1.ID, f2.ID},
	}).MustDo(ac)
	a.Equal(2, len(res.Set))
	a.True(f1.ID.Equal(res.Set[0].ID))
	a.True(f2.ID.Equal(res.Set[1].ID))
	a.False(res.More)

	res = (&file.Get{
		Host:         r.Ali().ID(),
		Project:      p.ID,
		Task:         &t1p0.ID,
		CreatedOnMin: ptr.Time(Now().Add(-5 * time.Second)),
		CreatedOnMax: ptr.Time(Now()),
		Limit:        1,
	}).MustDo(ac)
	a.Equal(1, len(res.Set))
	a.True(f2.ID.Equal(res.Set[0].ID))
	a.True(res.More)

	res = (&file.Get{
		Host:         r.Ali().ID(),
		Project:      p.ID,
		Task:         &t1p0.ID,
		CreatedOnMin: ptr.Time(Now().Add(-5 * time.Second)),
		CreatedOnMax: ptr.Time(Now()),
		CreatedBy:    ptr.ID(r.Ali().ID()),
		Limit:        1,
		After:        &f2.ID,
	}).MustDo(ac)
	a.Equal(1, len(res.Set))
	a.True(f1.ID.Equal(res.Set[0].ID))
	a.False(res.More)

	res = (&file.Get{
		Host:         r.Ali().ID(),
		Project:      p.ID,
		Task:         &t1p0.ID,
		CreatedOnMin: ptr.Time(Now().Add(-5 * time.Second)),
		CreatedOnMax: ptr.Time(Now()),
		CreatedBy:    ptr.ID(r.Ali().ID()),
		Limit:        1,
		Asc:          ptr.Bool(true),
	}).MustDo(ac)
	a.Equal(1, len(res.Set))
	a.True(f1.ID.Equal(res.Set[0].ID))
	a.True(res.More)

	(&file.Delete{
		Host:    r.Ali().ID(),
		Project: p.ID,
		Task:    t1p0.ID,
		ID:      f1.ID,
	}).MustDo(ac)

	pID = p.ID
	tree = testutil.GrabFullTree(r, r.Ali().ID(), p.ID)
}
