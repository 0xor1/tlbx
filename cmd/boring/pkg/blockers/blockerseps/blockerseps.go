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
					turn := g.Base.Turn
					pieceSet := uint8(turn % uint32(pieceSetsCount))
					if args.End.Bool() {
						if !(pieceSet == 3 && len(g.Players) == 3) {
							// end this players set except for last color in 3 player game
							g.PieceSetsEnded[pieceSet] = 1
						}
					} else {
						// validate piece is in valid range
						tlbx.BadReqIf(
							args.Piece >= blockers.PiecesCount(),
							"invalid piece value: %d, must be less than: %d", args.Piece, blockers.PiecesCount())

						// validate piece is still available
						tlbx.BadReqIf(
							g.PieceSets[pieceSet*blockers.PiecesCount()+args.Piece] == 0,
							"invalid piece, that piece has already been used")

						// get piece must return a copy so we arent updating the original values
						// when flipping/rotating
						piece := blockers.GetPiece(args.Piece)

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
						firstCornerConMet := turn >= uint32(pieceSetsCount)
						// diagonalTouchCon doesnt need to be met on first turn of each piece set
						diagnonalTouchConMet := turn < uint32(pieceSetsCount)
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
										(pieceSet == 0 && cellX == 0 && cellY == 0) ||
										(pieceSet == 1 && cellX == boardDims-1 && cellY == 0) ||
										(pieceSet == 2 && cellX == boardDims-1 && cellY == boardDims-1) ||
										(pieceSet == 3 && cellX == 0 && cellY == boardDims-1)

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
														g.Board[xyToI(uint8(loopBoardX), uint8(loopBoardY), boardDims)] == blockers.Pbit(pieceSet))
												tlbx.BadReqIf((offsetX == 0 || offsetY == 0) &&
													g.Board[xyToI(uint8(loopBoardX), uint8(loopBoardY), boardDims)] == blockers.Pbit(pieceSet),
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
							g.Board[i] = blockers.Pbit(pieceSet)
						}

						// set this piece from this set as having been used.
						g.PieceSets[pieceSet*blockers.PiecesCount()+args.Piece] = 0
					}
					// final section to check for finished game state and
					// auto increment turnIdx passed any given up piece sets,
					// remember game.TakeTurn() will increment turnIdx again after this also.
					pieceSetsStillActive := make([]uint8, 0, pieceSetsCount)
					for j := uint8(0); j < pieceSetsCount; j++ {
						// dont consider last pieceSet in a 3 player game
						if j == 3 && len(g.Players) == 3 {
							continue
						}
						if g.PieceSetsEnded[j] == 0 {
							for i := uint8(0); i < blockers.PiecesCount(); i++ {
								if g.PieceSets[blockers.PiecesCount()*j+i] == 1 {
									pieceSetsStillActive = append(pieceSetsStillActive, j)
									break
								}
								if i+1 == blockers.PiecesCount() {
									// if we've processed the last piece of this set
									// the whole set has been placed so this set is ended
									g.PieceSetsEnded[j] = 1
								}
							}
						}
					}
					if len(pieceSetsStillActive) == 0 {
						g.State = 2
					} else {
						// increment game.TurnIdx pass any ended piece sets
						for i := uint8(1); i <= pieceSetsCount; i++ {
							if g.PieceSetsEnded[(pieceSet+i)%pieceSetsCount] == 0 {
								break
							}
							g.Turn++
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
	for len(board) < cap(board) {
		board = append(board, blockers.Pbit(pieceSetsCount))
	}
	return &blockers.Game{
		Base: game.Base{
			Type:       gameType,
			MinPlayers: 2,
			MaxPlayers: pieceSetsCount,
			Turn:       0,
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
