package commenttest

import (
	"testing"

	"github.com/0xor1/tlbx/cmd/trees/pkg/cnsts"
	"github.com/0xor1/tlbx/cmd/trees/pkg/comment"
	"github.com/0xor1/tlbx/cmd/trees/pkg/project"
	"github.com/0xor1/tlbx/cmd/trees/pkg/task"
	. "github.com/0xor1/tlbx/pkg/core"
	"github.com/0xor1/tlbx/pkg/ptr"
	"github.com/0xor1/tlbx/pkg/web/app/test"
	"github.com/0xor1/trees/pkg/comment/commenteps"
	"github.com/0xor1/trees/pkg/config"
	"github.com/0xor1/trees/pkg/project/projecteps"
	"github.com/0xor1/trees/pkg/task/taskeps"
	"github.com/0xor1/trees/pkg/testutil"
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
		append(append(projecteps.Eps, taskeps.Eps...), commenteps.Eps...),
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

	e1 := (&comment.Create{
		Host:    r.Ali().ID(),
		Project: p.ID,
		Task:    t1p0.ID,
		Body:    "yolo",
	}).MustDo(ac)
	a.NotNil(e1)

	e1 = (&comment.Update{
		Host:    r.Ali().ID(),
		Project: p.ID,
		Task:    t1p0.ID,
		ID:      e1.ID,
		Body:    "polo",
	}).MustDo(ac)
	a.NotNil(e1)

	e1 = (&comment.Update{
		Host:    r.Ali().ID(),
		Project: p.ID,
		Task:    t1p0.ID,
		ID:      e1.ID,
		Body:    "polo",
	}).MustDo(ac)
	a.NotNil(e1)

	// nil
	cNil, eNotNil := (&comment.Update{
		Host:    r.Ali().ID(),
		Project: p.ID,
		Task:    t1p0.ID,
		ID:      e1.ID,
	}).Do(ac)
	a.Nil(cNil)
	a.NotNil(eNotNil)

	es := (&comment.Get{
		Host:    r.Ali().ID(),
		Project: p.ID,
		Task:    &t1p0.ID,
	}).MustDo(ac)
	a.Equal(e1, es.Set[0])
	a.False(es.More)

	es = (&comment.Get{
		Host:    r.Ali().ID(),
		Project: p.ID,
	}).MustDo(ac)
	a.Equal(e1, es.Set[0])
	a.False(es.More)

	e2 := (&comment.Create{
		Host:    r.Ali().ID(),
		Project: p.ID,
		Task:    t1p0.ID,
		Body:    "solo",
	}).MustDo(ac)
	a.NotNil(e2)

	es = (&comment.Get{
		Host:    r.Ali().ID(),
		Project: p.ID,
		Task:    &t1p0.ID,
	}).MustDo(ac)
	a.Equal(e2, es.Set[0])
	a.Equal(e1, es.Set[1])
	a.False(es.More)

	es = (&comment.Get{
		Host:    r.Ali().ID(),
		Project: p.ID,
		Task:    &t1p0.ID,
		Limit:   1,
	}).MustDo(ac)
	a.Equal(e2, es.Set[0])
	a.True(es.More)

	es = (&comment.Get{
		Host:    r.Ali().ID(),
		Project: p.ID,
		Task:    &t1p0.ID,
		After:   ptr.ID(e2.ID),
		Limit:   1,
	}).MustDo(ac)
	a.Equal(e1, es.Set[0])
	a.False(es.More)

	(&comment.Delete{
		Host:    r.Ali().ID(),
		Project: p.ID,
		Task:    t1p0.ID,
		ID:      e1.ID,
	}).MustDo(ac)

	pID = p.ID
	tree = (&task.GetTree{
		Host:    r.Ali().ID(),
		Project: p.ID,
		ID:      p.ID,
	}).MustDo(ac)
}
