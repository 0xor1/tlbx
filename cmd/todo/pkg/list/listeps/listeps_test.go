package listeps

import (
	"testing"

	"github.com/0xor1/wtf/cmd/todo/pkg/list"
	"github.com/0xor1/wtf/pkg/web/app/common/test"
)

func TestEverything(t *testing.T) {
	r := test.NewRig(Eps, OnDelete)
	defer r.CleanUp()

	(&list.Create{
		Name: "Test list",
	}).MustDo(r.Ali().Client())
}
