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

func New(tlbx app.Toolbox, gameType string, game Game) {
	base := game.GetBase()
	PanicIf(gameType == "", "gameType must be set")
	PanicIf(StrLen(gameType) > gameTypeMaxLen, "gameType len must be < %d", gameTypeMaxLen)
	validateUserIsntInAnActiveGame(tlbx, "create")
	base.UpdatedOn = NowMilli()
	srv := service.Get(tlbx)
	tx := srv.Data().Begin()
	defer tx.Rollback()
	id := tlbx.NewID()
	// assign new session id for a new game so no clashes with old finished games
	tlbx.Session().Login(id)
	base.ID = id
	base.Players = []ID{id}
	serialized := json.MustMarshal(game)
	tx.Exec(`INSERT INTO games (id, type, updatedOn, serialized) VALUES (?, ?, ?, ?)`, base.ID, gameType, base.UpdatedOn, serialized)
	tx.Exec(`INSERT INTO players (id, game) VALUES (?, ?)`, id, id)
	tx.Commit()
	base.MyID = &id
	cacheSerializedGame(tlbx, gameType, id, serialized)
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
	b.MyID = &newUserID
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
	me := tlbx.Me()
	b.MyID = &me
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
	me := tlbx.Me()
	b.MyID = &me
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
	me := tlbx.Me()
	g.GetBase().MyID = &me
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
	if tlbx.Session().IsAuthed() {
		me := tlbx.Me()
		dst.GetBase().MyID = &me
	} else {
		dst.GetBase().MyID = nil
	}
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
			gotType = ""
		}
		if len(buf) > 0 {
			json.MustUnmarshal(buf, dst)
			tlbx.BadReqIf(forUpdate && gotType != gameType, "types do not match, your active game: %s, expected game: %s", gotType, gameType)
			me := tlbx.Me()
			dst.GetBase().MyID = &me
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
