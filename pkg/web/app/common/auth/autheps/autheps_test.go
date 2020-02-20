package autheps_test

import (
	"testing"

	. "github.com/0xor1/wtf/pkg/core"
	"github.com/0xor1/wtf/pkg/web/app/common/test"
)

func TestEverything(t *testing.T) {
	test.NewRig(t, nil, func(id ID) {}).CleanUp()
}
