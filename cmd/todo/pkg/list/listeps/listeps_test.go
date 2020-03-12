package listeps

import (
	"testing"

	"github.com/0xor1/wtf/cmd/todo/pkg/list"
	. "github.com/0xor1/wtf/pkg/core"
	"github.com/0xor1/wtf/pkg/web/app/common/test"
	"github.com/stretchr/testify/assert"
)

func TestEverything(t *testing.T) {
	a := assert.New(t)
	r := test.NewRig(Eps, OnDelete)
	defer r.CleanUp()

	name := "Test list"
	testList := (&list.Create{
		Name: name,
	}).MustDo(r.Ali().Client())
	a.Equal(name, testList.Name)

	testListAgain := (&list.Get{
		ID: testList.ID,
	}).MustDo(r.Ali().Client())
	a.Equal(testList, testListAgain)

	(&list.Delete{
		IDs: IDs{testList.ID},
	}).MustDo(r.Ali().Client())

}
