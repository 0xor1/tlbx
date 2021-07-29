package filetest

import (
	"bytes"
	"io/ioutil"
	"testing"
	"time"

	"github.com/0xor1/tlbx/cmd/trees/pkg/cnsts"
	"github.com/0xor1/tlbx/cmd/trees/pkg/config"
	"github.com/0xor1/tlbx/cmd/trees/pkg/file"
	"github.com/0xor1/tlbx/cmd/trees/pkg/file/fileeps"
	"github.com/0xor1/tlbx/cmd/trees/pkg/project"
	"github.com/0xor1/tlbx/cmd/trees/pkg/project/projecteps"
	"github.com/0xor1/tlbx/cmd/trees/pkg/task"
	"github.com/0xor1/tlbx/cmd/trees/pkg/task/taskeps"
	"github.com/0xor1/tlbx/cmd/trees/pkg/testutil"
	. "github.com/0xor1/tlbx/pkg/core"
	"github.com/0xor1/tlbx/pkg/ptr"
	"github.com/0xor1/tlbx/pkg/web/app/test"
	"github.com/stretchr/testify/assert"
)

func Everything(t *testing.T) {
	var (
		tree task.GetTreeRes
		pID  ID
	)

	defer func() {
		testutil.PrintFullTree(pID, tree)
	}()

	a := assert.New(t)
	r := test.NewMeRig(
		config.Get(),
		append(append(projecteps.Eps, taskeps.Eps...), fileeps.Eps...),
		nil,
		nil,
		nil,
		projecteps.OnDelete,
		projecteps.OnSetSocials,
		projecteps.ValidateFCMTopic,
		true,
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

	// currently no way of setting this through api as it requires a paywall
	// which hasnt been implemented yet
	_, err := r.Data().Primary().Exec(`UPDATE projects SET fileLimit=5000 WHERE id=?`, p.ID)
	PanicOn(err)

	content1 := []byte("1")
	create := &file.Create{
		Args: &file.CreateArgs{
			Host:    r.Ali().ID(),
			Project: p.ID,
			Task:    t1p0.ID,
		},
	}
	create.Name = "one"
	create.Type = "text/plain"
	create.Size = int64(len(content1))
	create.Content = ioutil.NopCloser(bytes.NewBuffer(content1))
	createRes := create.MustDo(ac)
	a.True(createRes.Task.ID.Equal(t1p0.ID))
	a.Equal(createRes.Task.FileN, uint64(1))
	a.Equal(createRes.Task.FileSize, uint64(1))
	f1 := createRes.File
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
	a.Equal(create.Name, f1Get.Name)
	a.Equal(create.Size, f1Get.Size)
	a.Equal(create.Type, f1Get.Type)

	content2 := []byte("2")
	create.Content = ioutil.NopCloser(bytes.NewBuffer(content2))
	create.Name = "two"
	f2 := create.MustDo(ac).File
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

	t1p0Updated := (&file.Delete{
		Host:    r.Ali().ID(),
		Project: p.ID,
		Task:    t1p0.ID,
		ID:      f1.ID,
	}).MustDo(ac)
	a.True(t1p0Updated.ID.Equal(t1p0.ID))

	pID = p.ID
	tree = (&task.GetTree{
		Host:    r.Ali().ID(),
		Project: p.ID,
		ID:      p.ID,
	}).MustDo(ac)
}
