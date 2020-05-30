package game

import (
	"database/sql"
	"math/rand"
	"sync"
	"time"

	. "github.com/0xor1/wtf/pkg/core"
	"github.com/0xor1/wtf/pkg/isql"
	"github.com/0xor1/wtf/pkg/json"
	"github.com/0xor1/wtf/pkg/web/app"
	"github.com/0xor1/wtf/pkg/web/app/common/service"
	comsql "github.com/0xor1/wtf/pkg/web/app/common/sql"
	"github.com/gomodule/redigo/redis"
)

const (
	gameTypeMaxLen = 20
)

var (
	lastDeleteOutdatedCalledOn    = time.Time{}
	lastDeleteOutdatedCalledOnMtx = sync.RWMutex{}
)

type Game interface {
	GetBase() *Base
	IsMyTurn() bool
}

type Base struct {
	ID        ID        `json:"id"`
	UpdatedOn time.Time `json:"updatedOn"`
	State     uint8     `json:"state"` // 0 not started, 1 started, 2 finished, 3 abandoned
	MyID      *ID       `json:"myId,omitempty"`
	Players   []ID      `json:"players"`
	Turn      uint32    `json:"turn"`
}

func (b *Base) GetBase() *Base {
	return b
}

func (b *Base) IsMyTurn() bool {
	return b.Players[(int(b.Turn) % len(b.Players))].Equal(*b.MyID)
}

func (b *Base) IsActive() bool {
	return b.NotStarted() || b.Started()
}

func (b *Base) NotStarted() bool {
	return b.State == 0
}

func (b *Base) Started() bool {
	return b.State == 1
}

func (b *Base) Finished() bool {
	return b.State == 2
}

func (b *Base) Abandoned() bool {
	return b.State == 3
}

func (b *Base) setMyID(tlbx app.Toolbox) {
	// only set myId on active games
	if b.IsActive() && tlbx.Session().IsAuthed() {
		me := tlbx.Me()
		// loop through players only setting myId
		// if they're an active player in this game
		for _, p := range b.Players {
			if p.Equal(me) {
				b.MyID = &me
				return
			}
		}
	}
	// otherwise set it to nil
	b.MyID = nil
}

func New(tlbx app.Toolbox, gameType string, game Game) {
	b := game.GetBase()
	PanicIf(gameType == "", "gameType must be set")
	PanicIf(StrLen(gameType) > gameTypeMaxLen, "gameType len must be < %d", gameTypeMaxLen)
	validateUserIsntInAnActiveGame(tlbx, "create")
	b.UpdatedOn = NowMilli()
	srv := service.Get(tlbx)
	tx := srv.Data().Begin()
	defer tx.Rollback()
	id := tlbx.NewID()
	// assign new session id for a new game so no clashes with old finished games
	tlbx.Session().Login(id)
	b.ID = id
	b.Players = []ID{id}
	serialized := json.MustMarshal(game)
	tx.Exec(`INSERT INTO games (id, type, updatedOn, serialized) VALUES (?, ?, ?, ?)`, b.ID, gameType, b.UpdatedOn, serialized)
	tx.Exec(`INSERT INTO players (id, game) VALUES (?, ?)`, id, id)
	tx.Commit()
	cacheSerializedGame(tlbx, gameType, id, serialized)
	b.setMyID(tlbx)
}

func Join(tlbx app.Toolbox, maxPlayers uint8, gameType string, game ID, dst Game) {
	validateUserIsntInAnActiveGame(tlbx, "join")
	srv := service.Get(tlbx)
	tx := srv.Data().Begin()
	defer tx.Rollback()
	read(tlbx, tx, true, gameType, game, dst)
	b := dst.GetBase()
	tlbx.BadReqIf(!b.NotStarted(), "can't join a game that has already been started")
	tlbx.BadReqIf(len(b.Players) >= int(maxPlayers), "game is already at max player limit: %d", maxPlayers)
	// assign new session id for a new game so no clashes with old finished games
	newUserID := tlbx.NewID()
	b.Players = append(b.Players, newUserID)
	tlbx.Session().Login(newUserID)
	tx.Exec(`INSERT INTO players (id, game) VALUES (?, ?)`, newUserID, b.ID)
	update(tlbx, tx, gameType, dst)
	tx.Commit()
	b.setMyID(tlbx)
}

func Start(tlbx app.Toolbox, minPlayers uint8, randomizePlayerOrder bool, gameType string, dst Game, customSetup func(game Game)) {
	srv := service.Get(tlbx)
	tx := srv.Data().Begin()
	defer tx.Rollback()
	g, _ := getUsersActiveGame(tlbx, tx, true, gameType, dst)
	b := g.GetBase()
	tlbx.BadReqIf(!b.NotStarted(), "can't start a game that has already been started")
	tlbx.BadReqIf(len(b.Players) < int(minPlayers), "game hasn't met minimum player count requirement: %d", minPlayers)
	tlbx.BadReqIf(!b.ID.Equal(tlbx.Me()), "only the creator can start the game")
	if customSetup != nil {
		customSetup(dst)
	}
	b.State = 1
	if randomizePlayerOrder {
		reorderedPlayers := make([]ID, 0, len(b.Players))
		for i := int32(len(b.Players)); i > 0; i-- {
			j := rand.Int31n(i)
			reorderedPlayers = append(reorderedPlayers, b.Players[j])
			b.Players[j] = b.Players[len(b.Players)-1]
			b.Players = b.Players[:len(b.Players)-1]
		}
		b.Players = reorderedPlayers
	}
	update(tlbx, tx, gameType, dst)
	tx.Commit()
	b.setMyID(tlbx)
}

func TakeTurn(tlbx app.Toolbox, gameType string, dst Game, takeTurn func(game Game)) {
	tx := service.Get(tlbx).Data().Begin()
	defer tx.Rollback()
	g, _ := getUsersActiveGame(tlbx, tx, true, gameType, dst)
	tlbx.BadReqIf(g == nil, "you are not in an active game")
	b := g.GetBase()
	tlbx.BadReqIf(!b.Started(), "game isn't started")
	tlbx.BadReqIf(!g.IsMyTurn(), "it's not your turn")
	takeTurn(g)
	b.Turn++
	update(tlbx, tx, gameType, g)
	tx.Commit()
	b.setMyID(tlbx)
}

func Abandon(tlbx app.Toolbox, gameType string, dst Game) {
	srv := service.Get(tlbx)
	tx := srv.Data().Begin()
	defer tx.Rollback()
	g, _ := getUsersActiveGame(tlbx, tx, true, gameType, dst)
	if g != nil && g.GetBase().IsActive() {
		g.GetBase().State = 3
		update(tlbx, tx, gameType, g)
		tx.Commit()
	}
	g.GetBase().setMyID(tlbx)
}

func Get(tlbx app.Toolbox, gameType string, game ID, dst Game) {
	read(tlbx, nil, false, gameType, game, dst)
}

func read(tlbx app.Toolbox, tx service.Tx, forUpdate bool, gameType string, game ID, dst Game) {
	PanicIf(forUpdate && tx == nil, "tx required forUpdate get call")
	PanicIf(!forUpdate && tx != nil, "tx must be nil if it is a not forUpdate get call")
	serialized := make([]byte, 0, 5*app.KB)
	if !forUpdate {
		serialized = getSerializedGameFromCache(tlbx, gameType, game)
	}
	gotType := ""
	if len(serialized) == 0 {
		query := `SELECT type, serialized FROM games WHERE id=?`
		var row isql.Row
		if forUpdate {
			query += ` FOR UPDATE`
			row = tx.QueryRow(query, game)
		} else {
			row = service.Get(tlbx).Data().QueryRow(query, game)
		}
		comsql.ReturnNotFoundOrPanicOn(row.Scan(&gotType, &serialized))
	} else {
		// cache key was successful which includes gameType
		gotType = gameType
	}
	json.MustUnmarshal(serialized, dst)
	tlbx.BadReqIf(gotType != gameType, "types do not match, got: %s, expected: %s", gotType, gameType)
	dst.GetBase().setMyID(tlbx)
}

func update(tlbx app.Toolbox, tx service.Tx, gameType string, game Game) {
	base := game.GetBase()
	base.UpdatedOn = NowMilli()
	base.MyID = nil
	serialized := json.MustMarshal(game)
	tx.Exec(`UPDATE games Set updatedOn=?, serialized=? WHERE id=? AND type=?`, base.UpdatedOn, serialized, base.ID, gameType)
	cacheSerializedGame(tlbx, gameType, base.ID, serialized)
}

func DeleteOutdated(exec func(query string, args ...interface{}), delay time.Duration, expire time.Duration) {
	lastDeleteOutdatedCalledOnMtx.RLock()
	lastCalledOn := lastDeleteOutdatedCalledOn
	lastDeleteOutdatedCalledOnMtx.RUnlock()
	if !lastCalledOn.IsZero() && !lastCalledOn.Before(Now().Add(-1*delay)) {
		return
	}
	// relies on foreign key ON DELETE CASCADE to delete players rows
	exec(`DELETE FROM games WHERE updatedOn<?`, NowMilli().Add(-1*expire))
	lastDeleteOutdatedCalledOnMtx.Lock()
	defer lastDeleteOutdatedCalledOnMtx.Unlock()
	lastDeleteOutdatedCalledOn = Now()
}

func getUsersActiveGame(tlbx app.Toolbox, tx service.Tx, forUpdate bool, gameType string, dst Game) (Game, string) {
	PanicIf(forUpdate && tx == nil, "tx required forUpdate get call")
	PanicIf(!forUpdate && tx != nil, "tx must be nil if it is a not forUpdate get call")
	buf := make([]byte, 0, 5*app.KB)
	ses := tlbx.Session()
	if ses.IsAuthed() {
		query := `SELECT g.type, g.serialized FROM games g INNER JOIN players p ON p.game=g.id WHERE p.id=?`
		var row isql.Row
		if forUpdate {
			query += ` FOR UPDATE`
			row = tx.QueryRow(query, ses.Me())
		} else {
			row = service.Get(tlbx).Data().QueryRow(query, ses.Me())
		}
		gotType := ""
		err := row.Scan(&gotType, &buf)
		if err != nil && err != sql.ErrNoRows {
			PanicOn(err)
		}
		// if we dont care what type we want
		// and it's not for update, ignore
		// the type check. This is only for
		// validating if a user is in an active game
		if gameType == "" && !forUpdate {
			gameType = gotType
		}
		if len(buf) > 0 {
			json.MustUnmarshal(buf, dst)
			tlbx.BadReqIf(forUpdate && gotType != gameType, "types do not match, your active game: %s, expected game: %s", gotType, gameType)
			dst.GetBase().setMyID(tlbx)
			if dst.GetBase().IsActive() {
				return dst, gotType
			}
		}
	}
	return nil, ""
}

func validateUserIsntInAnActiveGame(tlbx app.Toolbox, verb string) {
	g, gameType := getUsersActiveGame(tlbx, nil, false, "", &Base{})
	if g == nil {
		return
	}
	tlbx.BadReqIf(
		true,
		"can not %s a new game while you are still participating in an active game, id: %s, type: %s",
		verb,
		g.GetBase().ID,
		gameType)
}

func cacheSerializedGame(tlbx app.Toolbox, gameType string, id ID, serialized []byte) {
	cnn := service.Get(tlbx).Cache().Get()
	defer cnn.Close()
	_, err := cnn.Do("SETEX", gameType+id.String(), 3600, serialized)
	tlbx.Log().ErrorOn(err)
}

func getSerializedGameFromCache(tlbx app.Toolbox, gameType string, id ID) []byte {
	cnn := service.Get(tlbx).Cache().Get()
	defer cnn.Close()
	serialized, err := redis.Bytes(cnn.Do("GET", gameType+id.String()))
	tlbx.Log().ErrorOn(err)
	return serialized
}

type Active struct{}
type ActiveInfo struct {
	Type string `json:"type"`
	ID   ID     `json:"id"`
}

func (_ *Active) Path() string {
	return "/game/active"
}

func (a *Active) Do(c *app.Client) (*ActiveInfo, error) {
	res := &ActiveInfo{}
	err := app.Call(c, a.Path(), nil, &res)
	return res, err
}

func (a *Active) MustDo(c *app.Client) *ActiveInfo {
	res, err := a.Do(c)
	PanicOn(err)
	return res
}

var (
	Eps = []*app.Endpoint{
		{
			Description:  "Get your active game info",
			Path:         (&Active{}).Path(),
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
				return &ActiveInfo{
					Type: "a_game_type",
					ID:   app.ExampleID(),
				}
			},
			Handler: func(tlbx app.Toolbox, _ interface{}) interface{} {
				g, gameType := getUsersActiveGame(tlbx, nil, false, "", &Base{})
				if g == nil {
					return nil
				}
				return &ActiveInfo{
					Type: gameType,
					ID:   g.GetBase().ID,
				}
			},
		},
	}
)
