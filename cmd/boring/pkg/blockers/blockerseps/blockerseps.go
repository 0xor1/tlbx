package blockerseps

import (
	"github.com/0xor1/wtf/cmd/boring/pkg/blockers"
	"github.com/0xor1/wtf/cmd/boring/pkg/game"
	. "github.com/0xor1/wtf/pkg/core"
	"github.com/0xor1/wtf/pkg/web/app"
)

const (
	gameType       = "blockers"
	boardDims      = uint8(20)
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
				game.Start(tlbx, args.RandomizePlayerOrder, gameType, g, nil)
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
				game.TakeTurn(tlbx, gameType, g, func(a game.Game) {
					g := a.(*blockers.Game)
					turnIdx := g.Base.TurnIdx
					pieceSetIdx := uint8(turnIdx % uint32(pieceSetsCount))
					if args.End {
						if !(pieceSetIdx == 3 && len(g.Players) == 3) {
							// end this players set except for last color in 3 player game
							g.PieceSetsEnded[pieceSetIdx] = 1
						}
					} else {
						// validate pieceIdx is in valid range
						tlbx.BadReqIf(
							args.PieceIdx >= blockers.PiecesCount(),
							"invalid pieceIdx value: %d, must be less than: %d", args.PieceIdx, blockers.PiecesCount())

						// validate piece is still available
						tlbx.BadReqIf(
							g.PieceSets[args.PieceIdx*pieceSetsCount+pieceSetIdx] == 0,
							"invalid pieceIdx, that piece has already been used")

						// get piece must return a copy so we arent updating the original values
						// when flipping/rotating
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
						x, y := iToXY(args.Position, boardDims, boardDims)
						tlbx.BadReqIf(
							x+piece.BB[0] > uint8(boardDims) || y+piece.BB[1] > uint8(boardDims),
							"piece/position/rotation combination is not contained on the board")

						// validate placement con(straints) met, firstCorner, diagonalTouch, sideTouch
						// firstCornerCon only needs to be met on first turns of each piece set
						firstCornerConMet := turnIdx >= uint32(pieceSetsCount)
						// diagonalTouchCon doesnt need to be met on first turn of each piece set
						diagnonalTouchConMet := turnIdx < uint32(pieceSetsCount)
						// board cell indexes to be inserted into by this placement
						insertIdxs := make([]uint16, 0, 5) // 5 because that's the largest piece by active cell count
						posX, posY := iToXY(args.Position, boardDims, boardDims)
						for pieceY := uint8(0); pieceY < piece.BB[1]; pieceY++ {
							for pieceX := uint8(0); pieceX < piece.BB[0]; pieceX++ {
								if piece.Shape[(pieceY*piece.BB[0])+pieceX] == 1 {
									cellX := posX + pieceX
									cellY := posY + pieceY
									cellI := xyToI(cellX, cellY, boardDims)

									tlbx.BadReqIf(g.Board[cellI] != 4, "cell already occupied")
									insertIdxs = append(insertIdxs, cellI)

									// check if this cell meets first corner constraint
									// 0 → 1
									// ↑   ↓
									// 3 ← 2
									firstCornerConMet = firstCornerConMet ||
										(pieceSetIdx == 0 && cellX == 0 && cellY == 0) ||
										(pieceSetIdx == 1 && cellX == boardDims-1 && cellY == 0) ||
										(pieceSetIdx == 2 && cellX == boardDims-1 && cellY == boardDims-1) ||
										(pieceSetIdx == 3 && cellX == 0 && cellY == boardDims-1)

									// loop through surrounding cells to check for diagonal and side touches
									for offsetY := -1; offsetY < 2; offsetY++ {
										for offsetX := -1; offsetX < 2; offsetX++ {
											if offsetX == 0 && offsetY == 0 {
												// it's the center of the loop i.e. the cell we're inserting into
												continue
											}
											loopBoardX := int(cellX) + offsetX
											loopBoardY := int(cellY) + offsetY
											// check coord is actually on the board
											if loopBoardX >= 0 && loopBoardY >= 0 && loopBoardX < int(boardDims) && loopBoardY < int(boardDims) {
												diagnonalTouchConMet = diagnonalTouchConMet ||
													((offsetX != 0 || offsetY != 0) &&
														g.Board[xyToI(uint8(loopBoardX), uint8(loopBoardY), boardDims)] == blockers.Pbit(pieceSetIdx))
												tlbx.BadReqIf((offsetX == 0 || offsetY == 0) &&
													g.Board[xyToI(uint8(loopBoardX), uint8(loopBoardY), boardDims)] == blockers.Pbit(pieceSetIdx),
													"face to face constraint not met")
											}
										}
									}
								}
							}
						}
						tlbx.BadReqIf(!firstCornerConMet, "first corner constraint not met")
						tlbx.BadReqIf(!diagnonalTouchConMet, "diagonal touch constraint not met")

						// update the board with the new piece cells on it
						for _, i := range insertIdxs {
							g.Board[i] = blockers.Pbit(pieceSetIdx)
						}

						// set this piece from this set as having been used.
						g.PieceSets[args.PieceIdx*pieceSetsCount+pieceSetIdx] = 0
					}
					// final section to check for finished game state and
					// auto increment turnIdx passed any given up piece sets,
					// remember game.TakeTurn() will increment turnIdx again after this also.
					pieceSetIdxsStillActive := make([]uint8, 0, pieceSetsCount)
					for i := uint8(0); i < pieceSetsCount; i++ {
						// dont consider last pieceSet in a 3 player game
						if i == 3 && len(g.Players) == 3 {
							continue
						}
						if g.PieceSetsEnded[i] == 0 {
							for j := uint8(0); j < blockers.PiecesCount(); j++ {
								if g.PieceSets[j*pieceSetsCount+i] == 1 {
									pieceSetIdxsStillActive = append(pieceSetIdxsStillActive, i)
									break
								}
							}
						}
					}
					if len(pieceSetIdxsStillActive) == 0 {
						g.State = 2
					} else {
						// increment game.TurnIdx pass any ended piece sets
						for i := uint8(1); i <= pieceSetsCount; i++ {
							if g.PieceSetsEnded[(pieceSetIdx+i)%pieceSetsCount] == 0 {
								break
							}
							g.TurnIdx++
						}
					}
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
	pieceSetsEnded := make(Bits, 0, pieceSetsCount)
	for len(pieceSetsEnded) < cap(pieceSetsEnded) {
		pieceSetsEnded = append(pieceSetsEnded, Bit(0))
	}
	pieceSets := make(Bits, 0, blockers.PiecesCount()*pieceSetsCount)
	for len(pieceSets) < cap(pieceSets) {
		pieceSets = append(pieceSets, Bit(1))
	}
	board := make(blockers.Pbits, 0, uint16(boardDims)*uint16(boardDims))
	for len(pieceSets) < cap(pieceSets) {
		board = append(board, blockers.Pbit(pieceSetsCount))
	}
	return &blockers.Game{
		Base: game.Base{
			MinPlayers: 2,
			MaxPlayers: pieceSetsCount,
			TurnIdx:    0,
		},
		PieceSetsEnded: pieceSetsEnded,
		PieceSets:      pieceSets,
		Board:          board,
	}
}

func iToXY(i uint16, xDim, yDim uint8) (x uint8, y uint8) {
	x = uint8(i % uint16(xDim))
	y = uint8(i / uint16(yDim))
	return
}

func xyToI(x, y, xDim uint8) uint16 {
	return uint16(xDim)*uint16(y) + uint16(x)
}
