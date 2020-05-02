package blockerseps

import (
	"github.com/0xor1/wtf/cmd/boring/pkg/blockers"
	. "github.com/0xor1/wtf/pkg/core"
	"github.com/0xor1/wtf/pkg/web/app"
)

var (
	Eps = []*app.Endpoint{
		{
			Description:  "Create a new game",
			Path:         (&blockers.New{}).Path(),
			Timeout:      500,
			MaxBodyBytes: app.KB,
			IsPrivate:    false,
			GetDefaultArgs: func() interface{} {
				return nil
			},
			GetExampleArgs: func() interface{} {
				return nil
			},
			GetExampleResponse: func() interface{} {
				return &blockers.Game{
					ID:            app.ExampleID(),
					CreatedOn:     app.ExampleTime(),
					UpdatedOn:     app.ExampleTime(),
					Started:       false,
					Players:       []ID{app.ExampleID(), app.ExampleID()},
					PieceSetsIdxs: initPieceSetsIdxs(),
					TurnIdx:       0,
					Board:         make([]uint8, 400),
				}
			},
			Handler: func(tlbx app.Toolbox, _ interface{}) interface{} {
				return map[string]string{"yolo": "nolo"}
			},
		},
	}
)

func initPieceSetsIdxs() []map[uint8]Bit {
	pieceSets := make([]map[uint8]Bit, 0, 4)
	for i := 0; i < 4; i++ {
		pieceSets = append(pieceSets, map[uint8]Bit{
			1:  1,
			2:  1,
			3:  1,
			4:  1,
			5:  1,
			6:  1,
			7:  1,
			8:  1,
			9:  1,
			10: 1,
			11: 1,
			12: 1,
			13: 1,
			14: 1,
			15: 1,
			16: 1,
			17: 1,
			18: 1,
			19: 1,
			20: 1,
			21: 1,
		})
	}

	return pieceSets
}
