package blockerseps

import (
	"github.com/0xor1/wtf/cmd/boring/pkg/blockers"
	"github.com/0xor1/wtf/cmd/boring/pkg/game"
	. "github.com/0xor1/wtf/pkg/core"
	"github.com/0xor1/wtf/pkg/web/app"
)

var (
	gameType = "blockers"
	Eps      = []*app.Endpoint{
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
				return NewGame()
			},
			Handler: func(tlbx app.Toolbox, _ interface{}) interface{} {
				g := NewGame()
				game.New(tlbx, g)
				return g
			},
		},
		{
			Description:  "Join a new game",
			Path:         (&blockers.Join{}).Path(),
			Timeout:      500,
			MaxBodyBytes: app.KB,
			IsPrivate:    false,
			GetDefaultArgs: func() interface{} {
				return &blockers.Join{}
			},
			GetExampleArgs: func() interface{} {
				return &blockers.Join{
					Game: app.ExampleID(),
				}
			},
			GetExampleResponse: func() interface{} {
				return NewGame()
			},
			Handler: func(tlbx app.Toolbox, a interface{}) interface{} {
				args := a.(*blockers.Join)
				g := &blockers.Game{}
				game.Join(tlbx, gameType, args.Game, g)
				return g
			},
		},
		{
			Description:  "Start your current game",
			Path:         (&blockers.Start{}).Path(),
			Timeout:      500,
			MaxBodyBytes: app.KB,
			IsPrivate:    false,
			GetDefaultArgs: func() interface{} {
				return &blockers.Start{
					RandomizePlayerOrder: false,
				}
			},
			GetExampleArgs: func() interface{} {
				return &blockers.Start{
					RandomizePlayerOrder: true,
				}
			},
			GetExampleResponse: func() interface{} {
				return NewGame()
			},
			Handler: func(tlbx app.Toolbox, a interface{}) interface{} {
				args := a.(*blockers.Start)
				g := &blockers.Game{}
				game.Start(tlbx, args.RandomizePlayerOrder, gameType, g)
				return g
			},
		},
		{
			Description:  "Take your turn",
			Path:         (&blockers.TakeTurn{}).Path(),
			Timeout:      500,
			MaxBodyBytes: app.KB,
			IsPrivate:    false,
			GetDefaultArgs: func() interface{} {
				return &blockers.TakeTurn{}
			},
			GetExampleArgs: func() interface{} {
				return &blockers.TakeTurn{}
			},
			GetExampleResponse: func() interface{} {
				return NewGame()
			},
			Handler: func(tlbx app.Toolbox, a interface{}) interface{} {
				args := a.(*blockers.TakeTurn)
				g := &blockers.Game{}
				game.TakeTurn(tlbx, gameType, g, func(g game.Game) {
					b := g.(*blockers.Game)
					Println(b, args)
					// TODO take turn
				})
				return g
			},
		},
		{
			Description:  "Get a game",
			Path:         (&blockers.Get{}).Path(),
			Timeout:      500,
			MaxBodyBytes: app.KB,
			IsPrivate:    false,
			GetDefaultArgs: func() interface{} {
				return &blockers.Get{}
			},
			GetExampleArgs: func() interface{} {
				return &blockers.Get{
					Game: app.ExampleID(),
				}
			},
			GetExampleResponse: func() interface{} {
				return NewGame()
			},
			Handler: func(tlbx app.Toolbox, a interface{}) interface{} {
				args := a.(*blockers.Get)
				g := &blockers.Game{}
				game.Get(tlbx, gameType, args.Game, g)
				return g
			},
		},
	}
)

func NewGame() *blockers.Game {
	return &blockers.Game{
		Base: game.Base{
			Type:         gameType,
			MinPlayers:   2,
			MaxPlayers:   4,
			TurnIdx:      0,
			TurnDuration: 0,
		},
		PieceSets: Bits{
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
		},
		Board: blockers.Pbits{
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
		},
	}
}
