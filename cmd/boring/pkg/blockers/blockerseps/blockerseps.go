package blockerseps

import (
	"github.com/0xor1/wtf/cmd/boring/pkg/blockers"
	"github.com/0xor1/wtf/cmd/boring/pkg/game"
	. "github.com/0xor1/wtf/pkg/core"
	"github.com/0xor1/wtf/pkg/web/app"
)

const (
	gameType       = "blockers"
	boardDims      = uint16(20)
	pieceSetsCount = uint8(4)
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
					if args.Pass {
						return
					}

					b := g.(*blockers.Game)
					turnIdx := b.Base.TurnIdx

					pieceSetIdx := uint8(turnIdx % uint32(pieceSetsCount))
					tlbx.BadReqIf(
						args.PieceIdx >= blockers.PiecesCount(),
						"invalid pieceIdx value: %d, must be less than: %d", args.PieceIdx, blockers.PiecesCount())

					tlbx.BadReqIf(
						b.PieceSets[args.PieceIdx*pieceSetsCount+pieceSetIdx] == 0,
						"invalid pieceIdx, you have already used that piece")

					// get piece must return a copy so we arent updating the original values
					piece := blockers.GetPiece(args.PieceIdx)

					// flip the piece if directed to, can think of this as reversing each row
					//
					//	■■□   □■■
					//  □■■ → ■■□
					//	□□■   ■□□
					//
					if args.Flip.Bool() {
						flippedShape := make(Bits, len(piece.Shape))
						for y := uint8(0); y < piece.BB[1]; y++ {
							for x := uint8(0); x < piece.BB[0]; x++ {
								flippedShape[(y*piece.BB[0])+x] = piece.Shape[(y*piece.BB[0])+piece.BB[0]-1-x]
							}
						}
						piece.Shape = flippedShape
					}

					// rotate clockwise 90 degrees * args.Rotation
					//
					// ■■□ → □■
					// □■■   ■■
					//       ■□
					//
					args.Rotation = args.Rotation % 4
					for i := uint8(0); i < args.Rotation; i++ {
						rotatedShape := make(Bits, len(piece.Shape))
						for y := uint8(0); y < piece.BB[1]; y++ {
							for x := uint8(0); x < piece.BB[0]; x++ {
								rotatedShape[(x*piece.BB[1])+(piece.BB[1]-1-y)] = piece.Shape[(y*piece.BB[0])+x]
							}
						}
						piece.Shape = rotatedShape
						bb0 := piece.BB[0]
						piece.BB[0] = piece.BB[1]
						piece.BB[1] = bb0
					}

					// validate piece is contained by board
					x, y := iToXY(args.Position)
					tlbx.BadReqIf(
						x+piece.BB[0] > uint8(boardDims) || y+piece.BB[1] > uint8(boardDims),
						"piece/position/rotation combination is not contained on the board")

					// if first piece, validate it is in the correct corner
					// 0 → 1
					// ↑   ↓
					// 3 ← 2
					if turnIdx < 4 {

					}

					// validate

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
	pieceSets := make(Bits, 0, blockers.PiecesCount()*pieceSetsCount)
	for len(pieceSets) < cap(pieceSets) {
		pieceSets = append(pieceSets, Bit(1))
	}
	board := make(blockers.Pbits, 0, boardDims*boardDims)
	for len(pieceSets) < cap(pieceSets) {
		board = append(board, blockers.Pbit(4))
	}
	return &blockers.Game{
		Base: game.Base{
			Type:       gameType,
			MinPlayers: 2,
			MaxPlayers: 4,
			TurnIdx:    0,
		},
		PieceSets: pieceSets,
		Board:     board,
	}
}

func iToXY(i uint16) (x uint8, y uint8) {
	x = uint8(i % boardDims)
	y = uint8(i / boardDims)
	return
}

func xyToI(x, y uint8) uint16 {
	return boardDims*uint16(y) + uint16(x)
}
