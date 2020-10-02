package tasktest

import (
	"testing"

	"github.com/0xor1/tlbx/cmd/trees/pkg/config"
	"github.com/0xor1/tlbx/cmd/trees/pkg/project"
	"github.com/0xor1/tlbx/cmd/trees/pkg/project/projecteps"
	"github.com/0xor1/tlbx/cmd/trees/pkg/task"
	"github.com/0xor1/tlbx/cmd/trees/pkg/task/taskeps"
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

	p = (&project.One{
		Host: r.Ali().ID(),
		ID:   p.ID,
	}).MustDo(ac)
	a.NotNil(p)
}
