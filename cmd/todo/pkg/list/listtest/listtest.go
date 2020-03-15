package listtest

import (
	"testing"

	"github.com/0xor1/wtf/cmd/todo/pkg/list"
	"github.com/0xor1/wtf/cmd/todo/pkg/list/listeps"
	. "github.com/0xor1/wtf/pkg/core"
	"github.com/0xor1/wtf/pkg/web/app/common/test"
	"github.com/stretchr/testify/assert"
)

func Everything(t *testing.T) {
	a := assert.New(t)
	r := test.NewRig(listeps.Eps, listeps.OnDelete)
	defer r.CleanUp()

	name1 := "Test list 1"
	testList1 := (&list.Create{
		Name: name1,
	}).MustDo(r.Ali().Client())
	a.Equal(name1, testList1.Name)

	name2 := "Test list 2"
	testList2 := (&list.Create{
		Name: name2,
	}).MustDo(r.Ali().Client())
	a.Equal(name2, testList2.Name)

	get := (&list.Get{
		ID: testList1.ID,
	}).MustDo(r.Ali().Client())
	a.Equal(testList1, get)

	getSet := (&list.GetSet{}).MustDo(r.Ali().Client())
	a.Equal(testList1, getSet.Set[0])
	a.Equal(testList2, getSet.Set[1])
	a.False(getSet.More)

	getSet = (&list.GetSet{
		IDs:   IDs{testList2.ID, testList1.ID},
	}).MustDo(r.Ali().Client())
	a.Equal(testList2, getSet.Set[0])
	a.Equal(testList1, getSet.Set[1])
	a.False(getSet.More)
}
