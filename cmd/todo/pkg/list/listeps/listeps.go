package listeps

import (
	"bytes"
	"net/http"

	"github.com/0xor1/wtf/cmd/todo/pkg/list"
	. "github.com/0xor1/wtf/pkg/core"
	"github.com/0xor1/wtf/pkg/isql"
	"github.com/0xor1/wtf/pkg/ptr"
	"github.com/0xor1/wtf/pkg/web/app"
	"github.com/0xor1/wtf/pkg/web/app/common/service"
	"github.com/0xor1/wtf/pkg/web/app/common/sql"
	"github.com/0xor1/wtf/pkg/web/app/common/validate"
)

var (
	Eps = []*app.Endpoint{
		{
			Description:  "Create a new list",
			Path:         (&list.Create{}).Path(),
			Timeout:      500,
			MaxBodyBytes: app.KB,
			IsPrivate:    false,
			GetDefaultArgs: func() interface{} {
				return &list.Create{}
			},
			GetExampleArgs: func() interface{} {
				return &list.Create{
					Name: "My List",
				}
			},
			GetExampleResponse: func() interface{} {
				return exampleList
			},
			Handler: func(tlbx app.Toolbox, a interface{}) interface{} {
				args := a.(*list.Create)
				validate.Str("name", args.Name, tlbx, nameMinLen, nameMaxLen)
				me := tlbx.Me()
				serv := service.Get(tlbx)
				res := &list.List{
					ID:        tlbx.NewID(),
					CreatedOn: Now(),
					Name:      args.Name,
					ItemCount: 0,
				}
				_, err := serv.Data().Exec(
					`INSERT INTO lists (user, id, createdOn, name, itemCount, firstItem) VALUES (?, ?, ?, ?, ?, ?)`,
					me, res.ID, res.CreatedOn, res.Name, res.ItemCount, nil)
				PanicOn(err)
				return res
			},
		},
		{
			Description:  "Get a list",
			Path:         (&list.Get{}).Path(),
			Timeout:      500,
			MaxBodyBytes: app.KB,
			IsPrivate:    false,
			GetDefaultArgs: func() interface{} {
				return &list.Get{}
			},
			GetExampleArgs: func() interface{} {
				return &list.Get{
					ID: app.ExampleID(),
				}
			},
			GetExampleResponse: func() interface{} {
				return exampleList
			},
			Handler: func(tlbx app.Toolbox, a interface{}) interface{} {
				args := a.(*list.Get)
				me := tlbx.Me()
				serv := service.Get(tlbx)
				res := &list.List{
					ID: args.ID,
				}
				row := serv.Data().QueryRow(
					`SELECT createdOn, name, itemCount FROM lists WHERE user=? AND id=?`,
					me, res.ID)
				sql.ReturnNotFoundOrPanic(row.Scan(&res.CreatedOn, &res.Name, &res.ItemCount))
				return res
			},
		},
		{
			Description:  "Get a list set",
			Path:         (&list.GetSet{}).Path(),
			Timeout:      500,
			MaxBodyBytes: app.KB,
			IsPrivate:    false,
			GetDefaultArgs: func() interface{} {
				return &list.GetSet{
					Sort:  list.SortCreatedOn,
					Asc:   true,
					Limit: 100,
				}
			},
			GetExampleArgs: func() interface{} {
				return &list.GetSet{
					NameStartsWith:  ptr.String("My L"),
					CreatedOnAfter:  ptr.Time(app.ExampleTime()),
					CreatedOnBefore: ptr.Time(app.ExampleTime()),
					ItemCountOver:   ptr.Int(2),
					ItemCountUnder:  ptr.Int(5),
					After:           ptr.ID(app.ExampleID()),
					Sort:            list.SortName,
					Asc:             true,
					Limit:           50,
				}
			},
			GetExampleResponse: func() interface{} {
				return &list.GetSetRes{
					Lists: []*list.List{
						exampleList,
					},
					More: true,
				}
			},
			Handler: func(tlbx app.Toolbox, a interface{}) interface{} {
				args := a.(*list.GetSet)
				tlbx.ReturnMsgIf(
					args.CreatedOnAfter != nil &&
						args.CreatedOnBefore != nil &&
						args.CreatedOnAfter.After(*args.CreatedOnBefore),
					http.StatusBadRequest, "createdOnAfter must be before createdOnBefore")
				tlbx.ReturnMsgIf(
					args.ItemCountOver != nil &&
						args.ItemCountUnder != nil &&
						*args.ItemCountOver >= *args.ItemCountUnder,
					http.StatusBadRequest, "itemCountOver must not be greater than or equal to itemCountUnder")
				args.Limit = sql.Limit(args.Limit, 100)
				me := tlbx.Me()
				serv := service.Get(tlbx)
				res := &list.GetSetRes{
					Lists: make([]*list.List, 0, args.Limit),
				}
				query := bytes.NewBufferString(`SELECT id, createdOn, name, itemCount FROM lists WHERE user=?`)
				queryArgs := make([]interface{}, 0, 10)
				queryArgs = append(queryArgs, me)
				if args.NameStartsWith != nil && *args.NameStartsWith != "" {
					query.WriteString(` AND name LIKE ?`)
					queryArgs = append(queryArgs, Sprintf(`%s%%`, *args.NameStartsWith))
				}
				if args.CreatedOnAfter != nil {
					query.WriteString(` AND createdOn > ?`)
					queryArgs = append(queryArgs, *args.CreatedOnAfter)
				}
				if args.CreatedOnBefore != nil {
					query.WriteString(` AND createdOn < ?`)
					queryArgs = append(queryArgs, *args.CreatedOnBefore)
				}
				if args.ItemCountOver != nil {
					query.WriteString(` AND itemCount > ?`)
					queryArgs = append(queryArgs, *args.ItemCountOver)
				}
				if args.ItemCountUnder != nil {
					query.WriteString(` AND itemCount < ?`)
					queryArgs = append(queryArgs, *args.ItemCountUnder)
				}
				if args.After != nil {
					query.WriteString(Sprintf(` AND %s %s= (SELECT %s FROM projects WHERE user=? AND id=?) AND id > ?`, args.Sort, sql.GtLtSymbol(args.Asc), args.Sort))
					queryArgs = append(queryArgs, me, *args.After, *args.After)
				}
				query.WriteString(Sprintf(` ORDER BY %s %s, id LIMIT %d`, args.Sort, sql.Asc(args.Asc), args.Limit))
				serv.Data().Query(func(rows isql.Rows) {
					for rows.Next() {
						if len(res.Lists) == args.Limit {
							res.More = true
							break
						}
						l := &list.List{}
						PanicOn(rows.Scan(&l.ID, &l.CreatedOn, &l.Name, &l.ItemCount))
						res.Lists = append(res.Lists, l)
					}
				}, query.String(), queryArgs...)
				return res
			},
		},
		{
			Description:  "Delete lists",
			Path:         (&list.Delete{}).Path(),
			Timeout:      500,
			MaxBodyBytes: app.KB,
			IsPrivate:    false,
			GetDefaultArgs: func() interface{} {
				return &list.Delete{}
			},
			GetExampleArgs: func() interface{} {
				return &list.Delete{
					IDs: []ID{app.ExampleID()},
				}
			},
			GetExampleResponse: func() interface{} {
				return exampleList
			},
			Handler: func(tlbx app.Toolbox, a interface{}) interface{} {
				args := a.(*list.Delete)
				if len(args.IDs) == 0 {
					return nil
				}
				tlbx.ReturnMsgIf(len(args.IDs) > 100, http.StatusBadRequest, "may not delete attempt to delete over 100 lists at a time")
				me := tlbx.Me()
				serv := service.Get(tlbx)
				tx, err := serv.Data().Base().Primary().Begin()
				if tx != nil {
					defer tx.Rollback()
				}
				PanicOn(err)
				query := bytes.NewBufferString(`DELETE FROM %s WHERE user=? AND %s IN (?`)
				queryArgs := make([]interface{}, 0, len(args.IDs)+1)
				queryArgs = append(queryArgs, me, args.IDs[0])
				for _, id := range args.IDs[1:] {
					query.WriteString(`,?`)
					queryArgs = append(queryArgs, id)
				}
				query.WriteString(`)`)
				qryStr := query.String()
				_, err = serv.Data().Exec(Sprintf(qryStr, "lists", "id"), queryArgs...)
				PanicOn(err)
				_, err = serv.Data().Exec(Sprintf(qryStr, "items", "list"), queryArgs...)
				PanicOn(err)
				PanicOn(tx.Commit())
				return nil
			},
		},
	}
	nameMinLen  = 1
	nameMaxLen  = 100
	exampleList = &list.List{
		ID:        app.ExampleID(),
		Name:      "My List",
		CreatedOn: app.ExampleTime(),
		ItemCount: 3,
	}
)

func OnDelete(tlbx app.Toolbox, me ID) {
	serv := service.Get(tlbx)
	tx, err := serv.Data().Base().Primary().Begin()
	if err != nil {
		defer tx.Rollback()
	}
	PanicOn(err)
	_, err = serv.Data().Exec(`DELETE FROM lists WHERE user=?`, me)
	PanicOn(err)
	_, err = serv.Data().Exec(`DELETE FROM items WHERE user=?`, me)
	PanicOn(err)
	PanicOn(tx.Commit())
}
