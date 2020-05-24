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
	IsMyTurn(app.Toolbox) bool
}

type Base struct {
	Type       string    `json:"type"`
	ID         ID        `json:"id"`
	UpdatedOn  time.Time `json:"updatedOn"`
	State      uint8     `json:"state"` // 0 not started, 1 started, 2 finished, 3 abandoned
	MinPlayers uint8     `json:"minPlayers"`
	MaxPlayers uint8     `json:"maxPlayers"`
	Players    []ID      `json:"players"`
	TurnIdx    uint32    `json:"turnIdx"`
}

func (b *Base) GetBase() *Base {
	return b
}

func (b *Base) IsMyTurn(tlbx app.Toolbox) bool {
	return b.Started() &&
		b.Players[(int(b.TurnIdx)%len(b.Players))].Equal(tlbx.Me())
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

func New(tlbx app.Toolbox, game Game) {
	base := game.GetBase()
	PanicIf(base.Type == "", "game type must be set")
	PanicIf(StrLen(base.Type) > gameTypeMaxLen, "game type len must be < %d", gameTypeMaxLen)
	PanicIf(base.MinPlayers <= 0 || base.MinPlayers > base.MaxPlayers, "invalid min/max player values")
	base.UpdatedOn = NowMilli()
	srv := service.Get(tlbx)
	tx := srv.Data().Begin()
	defer tx.Rollback()
	validateUserIsntInAnActiveGame(tlbx, tx, "create")
	id := tlbx.NewID()
	// assign new session id for a new game so no clashes with old finished games
	tlbx.Session().Login(id)
	base.ID = id
	base.Players = []ID{id}
	serialized := json.MustMarshal(game)
	tx.Exec(`INSERT INTO games (id, updatedOn, serialized) VALUES (?, ?, ?)`, base.ID, base.UpdatedOn, serialized)
	tx.Exec(`INSERT INTO players (id, game) VALUES (?, ?)`, id, id)
	tx.Commit()
	cacheSerializedGame(tlbx, id, serialized)
}

func Join(tlbx app.Toolbox, gameType string, game ID, dst Game) {
	srv := service.Get(tlbx)
	tx := srv.Data().Begin()
	defer tx.Rollback()
	read(tlbx, tx, true, gameType, game, dst)
	b := dst.GetBase()
	tlbx.BadReqIf(!b.NotStarted(), "can't join a game that has already been started")
	tlbx.BadReqIf(len(b.Players) >= int(b.MaxPlayers), "game is already at max player limit: %d", b.MaxPlayers)
	validateUserIsntInAnActiveGame(tlbx, tx, "join")
	// assign new session id for a new game so no clashes with old finished games
	newUserID := tlbx.NewID()
	b.Players = append(b.Players, newUserID)
	tlbx.Session().Login(newUserID)
	tx.Exec(`INSERT INTO players (id, game) VALUES (?, ?)`, newUserID, b.ID)
	update(tlbx, tx, dst)
	tx.Commit()
}

func Start(tlbx app.Toolbox, randomizePlayerOrder bool, gameType string, dst Game, customSetup func(game Game)) {
	srv := service.Get(tlbx)
	tx := srv.Data().Begin()
	defer tx.Rollback()
	g := getUsersActiveGame(tlbx, tx, true, gameType, dst)
	b := g.GetBase()
	tlbx.BadReqIf(!b.NotStarted(), "can't start a game that has already been started")
	tlbx.BadReqIf(len(b.Players) < int(b.MinPlayers), "game hasn't met minimum player count requirement: %d", b.MinPlayers)
	tlbx.BadReqIf(!b.ID.Equal(tlbx.Me()), "only the creator can start the game")
	if customSetup != nil {
		customSetup(dst)
	}
	b.State = 1
	if randomizePlayerOrder {
		reorderedPlayers := make([]ID, 0, len(b.Players))
		for i := int32(len(b.Players)); i >= 0; i-- {
			j := rand.Int31n(i)
			reorderedPlayers = append(reorderedPlayers, b.Players[j])
			b.Players[j] = b.Players[len(b.Players)-1]
			b.Players = b.Players[:len(b.Players)-1]
		}
		b.Players = reorderedPlayers
	}
	update(tlbx, tx, dst)
	tx.Commit()
}

func TakeTurn(tlbx app.Toolbox, gameType string, dst Game, takeTurn func(game Game)) {
	tx := service.Get(tlbx).Data().Begin()
	defer tx.Rollback()
	g := getUsersActiveGame(tlbx, tx, true, gameType, dst)
	tlbx.BadReqIf(g == nil, "you are not in an active game")
	b := g.GetBase()
	tlbx.BadReqIf(b.Type != gameType, "types do not match, your active game: %s, expected game: %s", g.GetBase().Type, gameType)
	tlbx.BadReqIf(!b.Started(), "game isn't started")
	tlbx.BadReqIf(!g.IsMyTurn(tlbx), "it's not you're turn")
	takeTurn(g)
	b.TurnIdx++
	update(tlbx, tx, g)
}

func Abandon(tlbx app.Toolbox, gameType string, dst Game) {
	srv := service.Get(tlbx)
	tx := srv.Data().Begin()
	defer tx.Rollback()
	g := getUsersActiveGame(tlbx, tx, true, gameType, dst)
	if g != nil {
		g.GetBase().State = 3
		update(tlbx, tx, dst)
		tx.Commit()
	}
}

func Get(tlbx app.Toolbox, gameType string, game ID, dst Game) {
	read(tlbx, nil, false, gameType, game, dst)
}

func read(tlbx app.Toolbox, tx service.Tx, forUpdate bool, gameType string, game ID, dst Game) {
	PanicIf(forUpdate && tx == nil, "tx required forUpdate get call")
	PanicIf(!forUpdate && tx != nil, "tx must be nil if it is a not forUpdate get call")
	serialized := make([]byte, 0, 5*app.KB)
	if !forUpdate {
		serialized = getSerializedGameFromCache(tlbx, game)
	}
	if len(serialized) == 0 {
		query := `SELECT serialized FROM games WHERE id=?`
		var row isql.Row
		if forUpdate {
			query += ` FOR UPDATE`
			row = tx.QueryRow(query, game)
		} else {
			row = service.Get(tlbx).Data().QueryRow(query, game)
		}
		comsql.ReturnNotFoundOrPanicOn(row.Scan(&serialized))
	}
	json.MustUnmarshal(serialized, dst)
	tlbx.BadReqIf(dst.GetBase().Type != gameType, "types do not match, got: %s, expected: %s", dst.GetBase().Type, gameType)
}

func update(tlbx app.Toolbox, tx service.Tx, game Game) {
	base := game.GetBase()
	base.UpdatedOn = NowMilli()
	serialized := json.MustMarshal(game)
	tx.Exec(`UPDATE games Set updatedOn=? AND serialized=? WHERE id=?`, base.UpdatedOn, serialized, base.ID)
	cacheSerializedGame(tlbx, base.ID, serialized)
}

func Delete(tlbx app.Toolbox, game ID) {
	// relies on foreign key ON DELETE CASCADE to delete players rows
	service.Get(tlbx).Data().Exec(`DELETE FROM games WHERE id=?`, game)
}

func DeleteOutdated(tlbx app.Toolbox) {
	lastDeleteOutdatedCalledOnMtx.RLock()
	lastCalledOn := lastDeleteOutdatedCalledOn
	lastDeleteOutdatedCalledOnMtx.RUnlock()
	if !lastCalledOn.IsZero() && !lastCalledOn.Before(Now().Add(-1*time.Hour)) {
		return
	}
	// relies on foreign key ON DELETE CASCADE to delete players rows
	service.Get(tlbx).Data().Exec(`DELETE FROM games WHERE updatedOn<?`, NowMilli().Add(-24*time.Hour))
	lastDeleteOutdatedCalledOnMtx.Lock()
	defer lastDeleteOutdatedCalledOnMtx.Unlock()
	lastDeleteOutdatedCalledOn = Now()
}

func getUsersActiveGame(tlbx app.Toolbox, tx service.Tx, forUpdate bool, gameType string, dst Game) Game {
	PanicIf(forUpdate && tx == nil, "tx required forUpdate get call")
	PanicIf(!forUpdate && tx != nil, "tx must be nil if it is a not forUpdate get call")
	buf := make([]byte, 0, 5*app.KB)
	ses := tlbx.Session()
	if ses.IsAuthed() {
		query := `SELECT g.serialized FROM games g INNER JOIN players p ON p.game=g.id WHERE p.id=?`
		var row isql.Row
		if forUpdate {
			query += ` FOR UPDATE`
			row = tx.QueryRow(query, ses.Me())
		} else {
			row = service.Get(tlbx).Data().QueryRow(query, ses.Me())
		}
		err := row.Scan(&buf)
		if err != nil && err != sql.ErrNoRows {
			PanicOn(err)
		}
		if len(buf) > 0 {
			json.MustUnmarshal(buf, dst)
			tlbx.BadReqIf(forUpdate && dst.GetBase().Type != gameType, "types do not match, your active game: %s, expected game: %s", dst.GetBase().Type, gameType)
			if dst.GetBase().IsActive() {
				return dst
			}
		}
	}
	return nil
}

func validateUserIsntInAnActiveGame(tlbx app.Toolbox, tx service.Tx, verb string) {
	g := getUsersActiveGame(tlbx, tx, false, "", &Base{})
	tlbx.BadReqIf(
		g != nil,
		"can not %s a new game while you are still participating in an active game, id: %s, type: %s",
		verb,
		g.GetBase().ID,
		g.GetBase().Type)
}

func cacheSerializedGame(tlbx app.Toolbox, id ID, serialized []byte) {
	cnn := service.Get(tlbx).Cache().Get()
	defer cnn.Close()
	_, err := cnn.Do("SETEX", id.String(), 3600, serialized)
	tlbx.Log().ErrorOn(err)
}

func getSerializedGameFromCache(tlbx app.Toolbox, id ID) []byte {
	cnn := service.Get(tlbx).Cache().Get()
	defer cnn.Close()
	serialized, err := redis.Bytes(cnn.Do("GET", id.String()))
	tlbx.Log().ErrorOn(err)
	return serialized
}
