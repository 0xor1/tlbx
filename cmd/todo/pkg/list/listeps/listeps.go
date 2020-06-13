package listeps

import (
	"bytes"
	"net/http"

	"github.com/0xor1/wtf/cmd/todo/pkg/list"
	. "github.com/0xor1/wtf/pkg/core"
	"github.com/0xor1/wtf/pkg/field"
	"github.com/0xor1/wtf/pkg/isql"
	"github.com/0xor1/wtf/pkg/ptr"
	"github.com/0xor1/wtf/pkg/web/app"
	"github.com/0xor1/wtf/pkg/web/app/service"
	"github.com/0xor1/wtf/pkg/web/app/sql"
	"github.com/0xor1/wtf/pkg/web/app/validate"
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
				srv := service.Get(tlbx)
				res := &list.List{
					ID:                 tlbx.NewID(),
					CreatedOn:          NowMilli(),
					Name:               args.Name,
					TodoItemCount:      0,
					CompletedItemCount: 0,
				}
				_, err := srv.Data().Exec(
					`INSERT INTO lists (user, id, createdOn, name, todoItemCount, completedItemCount) VALUES (?, ?, ?, ?, ?, ?)`,
					me, res.ID, res.CreatedOn, res.Name, res.TodoItemCount, res.CompletedItemCount)
				PanicOn(err)
				return res
			},
		},
		{
			Description:  "Get a list set",
			Path:         (&list.Get{}).Path(),
			Timeout:      500,
			MaxBodyBytes: app.KB,
			IsPrivate:    false,
			GetDefaultArgs: func() interface{} {
				return &list.Get{
					Sort:  list.SortCreatedOn,
					Asc:   ptr.Bool(true),
					Limit: ptr.Int(100),
				}
			},
			GetExampleArgs: func() interface{} {
				return &list.Get{
					NameStartsWith:        ptr.String("My L"),
					CreatedOnMin:          ptr.Time(app.ExampleTime()),
					CreatedOnMax:          ptr.Time(app.ExampleTime()),
					TodoItemCountMin:      ptr.Int(2),
					TodoItemCountMax:      ptr.Int(5),
					CompletedItemCountMin: ptr.Int(3),
					CompletedItemCountMax: ptr.Int(4),
					After:                 ptr.ID(app.ExampleID()),
					Sort:                  list.SortName,
					Asc:                   ptr.Bool(true),
					Limit:                 ptr.Int(50),
				}
			},
			GetExampleResponse: func() interface{} {
				return &list.GetRes{
					Set: []*list.List{
						exampleList,
					},
					More: true,
				}
			},
			Handler: func(tlbx app.Toolbox, a interface{}) interface{} {
				return getSet(tlbx, a.(*list.Get))
			},
		},
		{
			Description:  "Update a list",
			Path:         (&list.Update{}).Path(),
			Timeout:      500,
			MaxBodyBytes: app.KB,
			IsPrivate:    false,
			GetDefaultArgs: func() interface{} {
				return &list.Update{}
			},
			GetExampleArgs: func() interface{} {
				return &list.Update{
					ID:   app.ExampleID(),
					Name: field.String{V: "New List Name"},
				}
			},
			GetExampleResponse: func() interface{} {
				return exampleList
			},
			Handler: func(tlbx app.Toolbox, a interface{}) interface{} {
				args := a.(*list.Update)
				validate.Str("name", args.Name.V, tlbx, nameMinLen, nameMaxLen)
				getSetRes := getSet(tlbx, &list.Get{
					IDs:   IDs{args.ID},
					Limit: ptr.Int(1),
				})
				tlbx.ExitIf(len(getSetRes.Set) == 0, http.StatusNotFound, "no list with that id")
				list := getSetRes.Set[0]
				list.Name = args.Name.V
				srv := service.Get(tlbx)
				_, err := srv.Data().Exec(`UPDATE lists SET name=? WHERE user=? AND id=?`, list.Name, tlbx.Me(), list.ID)
				PanicOn(err)
				return list
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
				return nil
			},
			Handler: func(tlbx app.Toolbox, a interface{}) interface{} {
				args := a.(*list.Delete)
				idsLen := len(args.IDs)
				if idsLen == 0 {
					return nil
				}
				validate.MaxIDs(tlbx, "ids", args.IDs, 100)
				me := tlbx.Me()
				srv := service.Get(tlbx)
				queryArgs := make([]interface{}, 0, idsLen+1)
				queryArgs = append(queryArgs, me)
				queryArgs = append(queryArgs, args.IDs.ToIs()...)
				tx := srv.Data().Begin()
				defer tx.Rollback()
				_, err := tx.Exec(`DELETE FROM lists WHERE user=?`+sql.InCondition(true, "id", idsLen), queryArgs...)
				PanicOn(err)
				_, err = tx.Exec(`DELETE FROM items WHERE user=?`+sql.InCondition(true, "list", idsLen), queryArgs...)
				PanicOn(err)
				tx.Commit()
				return nil
			},
		},
	}
	nameMinLen  = 1
	nameMaxLen  = 250
	exampleList = &list.List{
		ID:                 app.ExampleID(),
		Name:               "My List",
		CreatedOn:          app.ExampleTime(),
		TodoItemCount:      3,
		CompletedItemCount: 4,
	}
)

func OnDelete(tlbx app.Toolbox, me ID) {
	srv := service.Get(tlbx)
	tx := srv.Data().Begin()
	defer tx.Rollback()
	_, err := tx.Exec(`DELETE FROM lists WHERE user=?`, me)
	PanicOn(err)
	_, err = tx.Exec(`DELETE FROM items WHERE user=?`, me)
	PanicOn(err)
	tx.Commit()
}

func getSet(tlbx app.Toolbox, args *list.Get) *list.GetRes {
	validate.MaxIDs(tlbx, "ids", args.IDs, 100)
	tlbx.BadReqIf(
		args.CreatedOnMin != nil &&
			args.CreatedOnMax != nil &&
			args.CreatedOnMin.After(*args.CreatedOnMax),
		"createdOnMin must be before createdOnMax")
	tlbx.BadReqIf(
		args.TodoItemCountMin != nil &&
			args.TodoItemCountMax != nil &&
			*args.TodoItemCountMin > *args.TodoItemCountMax,
		"todoItemCountMin must not be greater than todoItemCountMax")
	tlbx.BadReqIf(
		args.CompletedItemCountMin != nil &&
			args.CompletedItemCountMax != nil &&
			*args.CompletedItemCountMin > *args.CompletedItemCountMax,
		"completedItemCountMin must not be greater than completedItemCountMax")
	limit := sql.Limit(*args.Limit, 100)
	me := tlbx.Me()
	srv := service.Get(tlbx)
	res := &list.GetRes{
		Set: make([]*list.List, 0, limit),
	}
	query := bytes.NewBufferString(`SELECT id, createdOn, name, todoItemCount, completedItemCount FROM lists WHERE user=?`)
	queryArgs := make([]interface{}, 0, 10)
	queryArgs = append(queryArgs, me)
	idsLen := len(args.IDs)
	if idsLen > 0 {
		query.WriteString(sql.InCondition(true, `id`, idsLen))
		query.WriteString(sql.OrderByField(`id`, idsLen))
		queryArgs = append(queryArgs, args.IDs.ToIs()...)
		queryArgs = append(queryArgs, args.IDs.ToIs()...)
	} else {
		if ptr.StringOr(args.NameStartsWith, "") != "" {
			query.WriteString(` AND name LIKE ?`)
			queryArgs = append(queryArgs, Sprintf(`%s%%`, *args.NameStartsWith))
		}
		if args.CreatedOnMin != nil {
			query.WriteString(` AND createdOn >= ?`)
			queryArgs = append(queryArgs, *args.CreatedOnMin)
		}
		if args.CreatedOnMax != nil {
			query.WriteString(` AND createdOn <= ?`)
			queryArgs = append(queryArgs, *args.CreatedOnMax)
		}
		if args.TodoItemCountMin != nil {
			query.WriteString(` AND todoItemCount >= ?`)
			queryArgs = append(queryArgs, *args.TodoItemCountMin)
		}
		if args.TodoItemCountMax != nil {
			query.WriteString(` AND todoItemCount <= ?`)
			queryArgs = append(queryArgs, *args.TodoItemCountMax)
		}
		if args.CompletedItemCountMin != nil {
			query.WriteString(` AND completedItemCount >= ?`)
			queryArgs = append(queryArgs, *args.CompletedItemCountMin)
		}
		if args.CompletedItemCountMax != nil {
			query.WriteString(` AND completedItemCount <= ?`)
			queryArgs = append(queryArgs, *args.CompletedItemCountMax)
		}
		if args.After != nil {
			query.WriteString(Sprintf(` AND %s %s= (SELECT %s FROM lists WHERE user=? AND id=?) AND id <> ?`, args.Sort, sql.GtLtSymbol(*args.Asc), args.Sort))
			queryArgs = append(queryArgs, me, *args.After, *args.After)
			if args.Sort != list.SortCreatedOn {
				query.WriteString(Sprintf(` AND createdOn %s (SELECT createdOn FROM lists WHERE user=? AND id=?)`, sql.GtLtSymbol(*args.Asc)))
				queryArgs = append(queryArgs, me, *args.After)
			}
		}
		createdOnSecondarySort := ""
		if args.Sort != list.SortCreatedOn {
			createdOnSecondarySort = ", createdOn"
		}
		query.WriteString(Sprintf(` ORDER BY %s%s %s, id LIMIT %d`, args.Sort, createdOnSecondarySort, sql.Asc(*args.Asc), limit))
	}
	srv.Data().Query(func(rows isql.Rows) {
		for rows.Next() {
			if len(args.IDs) == 0 && len(res.Set)+1 == limit {
				res.More = true
				break
			}
			l := &list.List{}
			PanicOn(rows.Scan(&l.ID, &l.CreatedOn, &l.Name, &l.TodoItemCount, &l.CompletedItemCount))
			res.Set = append(res.Set, l)
		}
	}, query.String(), queryArgs...)
	return res
}
