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
					ID:        app.ExampleID(),
					CreatedOn: app.ExampleTime(),
					UpdatedOn: app.ExampleTime(),
					Started:   false,
					Players:   []ID{app.ExampleID(), app.ExampleID()},
					PieceSets: initPieceSets(),
					TurnIdx:   0,
					Board:     initBoard(),
				}
			},
			Handler: func(tlbx app.Toolbox, _ interface{}) interface{} {
				me := tlbx.Me()
				//srv := service.Get(tlbx)
				id := tlbx.NewID()
				now := NowMilli()
				game := &blockers.Game{
					ID:        id,
					CreatedOn: now,
					UpdatedOn: now,
					Started:   false,
					Players:   []ID{me, me, me, me},
					PieceSets: initPieceSets(),
					TurnIdx:   0,
					Board:     initBoard(),
				}

				return game
			},
		},
	}
)

func initPieceSets() Bits {
	return Bits{
		1, 1, 1, 1, // piece 0  -> color 0, 1, 2, 3
		1, 1, 1, 1, // piece 1  -> color 0, 1, 2, 3
		1, 1, 1, 1, // piece 2  -> color 0, 1, 2, 3
		1, 1, 1, 1, // piece 3  -> color 0, 1, 2, 3
		1, 1, 1, 1, // piece 4  -> color 0, 1, 2, 3
		1, 1, 1, 1, // piece 5  -> color 0, 1, 2, 3
		1, 1, 1, 1, // piece 6  -> color 0, 1, 2, 3
		1, 1, 1, 1, // piece 7  -> color 0, 1, 2, 3
		1, 1, 1, 1, // piece 8  -> color 0, 1, 2, 3
		1, 1, 1, 1, // piece 9  -> color 0, 1, 2, 3
		1, 1, 1, 1, // piece 10 -> color 0, 1, 2, 3
		1, 1, 1, 1, // piece 11 -> color 0, 1, 2, 3
		1, 1, 1, 1, // piece 12 -> color 0, 1, 2, 3
		1, 1, 1, 1, // piece 13 -> color 0, 1, 2, 3
		1, 1, 1, 1, // piece 14 -> color 0, 1, 2, 3
		1, 1, 1, 1, // piece 15 -> color 0, 1, 2, 3
		1, 1, 1, 1, // piece 16 -> color 0, 1, 2, 3
		1, 1, 1, 1, // piece 17 -> color 0, 1, 2, 3
		1, 1, 1, 1, // piece 18 -> color 0, 1, 2, 3
		1, 1, 1, 1, // piece 19 -> color 0, 1, 2, 3
		1, 1, 1, 1, // piece 20 -> color 0, 1, 2, 3
	}
}

func initBoard() blockers.Pbits {
	return blockers.Pbits{
		4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4,
		4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4,
		4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4,
		4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4,
		4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4,
		4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4,
		4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4,
		4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4,
		4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4,
		4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4,
		4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4,
		4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4,
		4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4,
		4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4,
		4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4,
		4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4,
		4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4,
		4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4,
		4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4,
		4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4,
	}
}
