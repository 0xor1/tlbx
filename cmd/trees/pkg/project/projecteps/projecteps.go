package projecteps

import (
	"bytes"
	"net/http"
	"time"

	"github.com/0xor1/tlbx/cmd/trees/pkg/cnsts"
	"github.com/0xor1/tlbx/cmd/trees/pkg/epsutil"
	"github.com/0xor1/tlbx/cmd/trees/pkg/project"
	"github.com/0xor1/tlbx/cmd/trees/pkg/task"
	. "github.com/0xor1/tlbx/pkg/core"
	"github.com/0xor1/tlbx/pkg/field"
	"github.com/0xor1/tlbx/pkg/isql"
	"github.com/0xor1/tlbx/pkg/ptr"
	"github.com/0xor1/tlbx/pkg/web/app"
	"github.com/0xor1/tlbx/pkg/web/app/service"
	"github.com/0xor1/tlbx/pkg/web/app/session/me"
	"github.com/0xor1/tlbx/pkg/web/app/sql"
	"github.com/0xor1/tlbx/pkg/web/app/user"
	"github.com/0xor1/tlbx/pkg/web/app/validate"
)

var (
	Eps = []*app.Endpoint{
		{
			Description:  "Create a new project",
			Path:         (&project.Create{}).Path(),
			Timeout:      500,
			MaxBodyBytes: app.KB,
			IsPrivate:    false,
			GetDefaultArgs: func() interface{} {
				return &project.Create{
					CurrencyCode: "USD",
					HoursPerDay:  8,
					DaysPerWeek:  5,
					IsPublic:     false,
				}
			},
			GetExampleArgs: func() interface{} {
				return &project.Create{
					CurrencyCode: "USD",
					HoursPerDay:  8,
					DaysPerWeek:  5,
					StartOn:      ptr.Time(app.ExampleTime()),
					DueOn:        ptr.Time(app.ExampleTime().Add(24 * time.Hour)),
					IsPublic:     false,
					Name:         "My New Project",
				}
			},
			GetExampleResponse: func() interface{} {
				return exampleProject
			},
			Handler: func(tlbx app.Tlbx, a interface{}) interface{} {
				args := a.(*project.Create)
				me := me.Get(tlbx)
				validate.Str("name", args.Name, tlbx, nameMinLen, nameMaxLen)
				if args.CurrencyCode == "" {
					args.CurrencyCode = "USD"
				}
				validate.CurrencyCode(tlbx, args.CurrencyCode)
				app.BadReqIf(args.HoursPerDay < 1 || args.HoursPerDay > 24, "invalid hoursPerDay must be > 0 and <= 24")
				app.BadReqIf(args.DaysPerWeek < 1 || args.DaysPerWeek > 7, "invalid daysPerWeek must be > 0 and <= 7")
				app.BadReqIf(args.StartOn != nil && args.DueOn != nil && !args.StartOn.Before(*args.DueOn), "invalid startOn must be before dueOn")
				p := &project.Project{
					Task: task.Task{
						ID:         tlbx.NewID(),
						Name:       args.Name,
						CreatedBy:  me,
						CreatedOn:  NowMilli(),
						IsParallel: true,
					},
					Base: project.Base{
						CurrencyCode: args.CurrencyCode,
						HoursPerDay:  args.HoursPerDay,
						DaysPerWeek:  args.DaysPerWeek,
						StartOn:      args.StartOn,
						DueOn:        args.DueOn,
						IsPublic:     args.IsPublic,
					},
					Host:       me,
					IsArchived: false,
				}
				srv := service.Get(tlbx)

				u := &user.User{}
				row := srv.User().QueryRow(`SELECT id, handle, alias, hasAvatar FROM users WHERE id=?`, me)
				PanicOn(row.Scan(&u.ID, &u.Handle, &u.Alias, &u.HasAvatar))

				tx := srv.Data().Begin()
				defer tx.Rollback()
				_, err := tx.Exec(`INSERT INTO projectLocks (host, id) VALUES (?, ?)`, p.Host, p.ID)
				PanicOn(err)
				_, err = tx.Exec(`INSERT INTO users (host, project, id, handle, alias, hasAvatar, isActive, role) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`, p.Host, p.ID, me, u.Handle, u.Alias, u.HasAvatar, true, cnsts.RoleAdmin)
				PanicOn(err)
				_, err = tx.Exec(`INSERT INTO projects (host, id, isArchived, name, createdOn, currencyCode, hoursPerDay, daysPerWeek, startOn, dueOn, isPublic) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, p.Host, p.ID, p.IsArchived, p.Name, p.CreatedOn, p.CurrencyCode, p.HoursPerDay, p.DaysPerWeek, p.StartOn, p.DueOn, p.IsPublic)
				PanicOn(err)
				_, err = tx.Exec(`INSERT INTO tasks (host, project, id, parent, firstChild, nextSibling, user, name, description, isParallel, createdBy, createdOn, minimumTime, estimatedTime, loggedTime, estimatedSubTime, loggedSubTime, estimatedExpense, loggedExpense, estimatedSubExpense, loggedSubExpense, fileCount, fileSize, fileSubCount, fileSubSize, childCount, descendantCount) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, p.Host, p.ID, p.ID, p.Parent, p.FirstChild, p.NextSibling, p.User, p.Name, p.Description, p.IsParallel, p.CreatedBy, p.CreatedOn, p.MinimumTime, p.EstimatedTime, p.LoggedTime, p.EstimatedSubTime, p.LoggedSubTime, p.EstimatedExpense, p.LoggedExpense, p.EstimatedSubExpense, p.LoggedSubExpense, p.FileCount, p.FileSize, p.FileSubCount, p.FileSubSize, p.ChildCount, p.DescendantCount)
				PanicOn(err)
				epsutil.LogActivity(tlbx, tx, me, p.ID, p.ID, cnsts.TypeProject, cnsts.ActionCreated, ptr.String(p.Name), nil)
				tx.Commit()
				return p
			},
		},
		{
			Description:  "Get a project set",
			Path:         (&project.Get{}).Path(),
			Timeout:      500,
			MaxBodyBytes: app.KB,
			IsPrivate:    false,
			GetDefaultArgs: func() interface{} {
				return &project.Get{
					IsArchived: false,
					Sort:       cnsts.SortCreatedOn,
					Asc:        ptr.Bool(true),
					Limit:      100,
				}
			},
			GetExampleArgs: func() interface{} {
				return &project.Get{
					IsArchived:   false,
					NamePrefix:   ptr.String("My Proj"),
					CreatedOnMin: ptr.Time(app.ExampleTime()),
					CreatedOnMax: ptr.Time(app.ExampleTime()),
					After:        ptr.ID(app.ExampleID()),
					Sort:         cnsts.SortName,
					Asc:          ptr.Bool(true),
					Limit:        50,
				}
			},
			GetExampleResponse: func() interface{} {
				return &project.GetRes{
					Set: []*project.Project{
						exampleProject,
					},
					More: true,
				}
			},
			Handler: func(tlbx app.Tlbx, a interface{}) interface{} {
				return getSet(tlbx, a.(*project.Get))
			},
		},
		{
			Description:  "Update a project",
			Path:         (&project.Updates{}).Path(),
			Timeout:      500,
			MaxBodyBytes: app.KB,
			IsPrivate:    false,
			GetDefaultArgs: func() interface{} {
				return &project.Updates{}
			},
			GetExampleArgs: func() interface{} {
				return &project.Updates{
					{
						ID:           app.ExampleID(),
						Name:         &field.String{V: "Renamed Project"},
						CurrencyCode: &field.String{V: "EUR"},
						HoursPerDay:  &field.UInt8{V: 6},
						DaysPerWeek:  &field.UInt8{V: 4},
						StartOn:      &field.TimePtr{V: ptr.Time(app.ExampleTime())},
						DueOn:        &field.TimePtr{V: ptr.Time(app.ExampleTime().Add(24 * time.Hour))},
						IsArchived:   &field.Bool{V: false},
						IsPublic:     &field.Bool{V: true},
					},
				}
			},
			GetExampleResponse: func() interface{} {
				return []*project.Project{
					exampleProject,
				}
			},
			Handler: func(tlbx app.Tlbx, a interface{}) interface{} {
				args := *(a.(*project.Updates))
				if len(args) == 0 {
					return nil
				}
				app.BadReqIf(len(args) > 100, "can not update more than 100 projects at a time")
				me := me.Get(tlbx)
				ids := make(IDs, len(args))
				namesSet := make([]bool, len(args))
				dupes := map[string]bool{}
				for i := 0; i < len(args); i++ {
					u := args[i]
					idStr := u.ID.String()
					app.BadReqIf(dupes[idStr], "duplicate entry detected")
					dupes[idStr] = true
					// if there are no changes to be made, remove this entry
					if u.Name == nil &&
						u.CurrencyCode == nil &&
						u.HoursPerDay == nil &&
						u.DaysPerWeek == nil &&
						u.StartOn == nil &&
						u.DueOn == nil &&
						u.IsArchived == nil &&
						u.IsPublic == nil {
						copy(args[i:], args[i+1:])
						args[len(args)-1] = nil
						args = args[:len(args)-1]
					}
				}
				for i, u := range args {
					ids[i] = u.ID
				}
				ps := getSet(tlbx, &project.Get{Host: &me, IDs: ids}).Set
				for i, p := range ps {
					a := args[i]
					if a.CurrencyCode != nil {
						validate.CurrencyCode(tlbx, a.CurrencyCode.V)
						p.CurrencyCode = a.CurrencyCode.V
					}
					// validate name
					if a.Name != nil {
						validate.Str("name", a.Name.V, tlbx, nameMinLen, nameMaxLen)
						p.Name = a.Name.V
						namesSet[i] = true
					}
					// validate startOn and dueOn
					switch {
					case a.StartOn != nil && a.DueOn != nil:
						app.BadReqIf(a.StartOn.V != nil && a.DueOn.V != nil && !a.StartOn.V.Before(*a.DueOn.V), "invalid startOn must be before dueOn")
					case a.StartOn != nil && p.DueOn != nil:
						app.BadReqIf(a.StartOn.V != nil && p.DueOn != nil && !a.StartOn.V.Before(*p.DueOn), "invalid startOn must be before dueOn")
					case a.DueOn != nil && p.StartOn != nil:
						app.BadReqIf(p.StartOn != nil && a.DueOn.V != nil && !p.StartOn.Before(*a.DueOn.V), "invalid startOn must be before dueOn")
					}
					if a.StartOn != nil {
						p.StartOn = a.StartOn.V
					}
					if a.DueOn != nil {
						p.DueOn = a.DueOn.V
					}
					if a.HoursPerDay != nil {
						app.BadReqIf(a.HoursPerDay.V < 1 || a.HoursPerDay.V > 24, "invalid hoursPerDay must be > 0 and <= 24")
						p.HoursPerDay = a.HoursPerDay.V
					}
					if a.DaysPerWeek != nil {
						app.BadReqIf(a.DaysPerWeek.V < 1 || a.DaysPerWeek.V > 7, "invalid daysPerWeek must be > 0 and <= 7")
						p.DaysPerWeek = a.DaysPerWeek.V
					}
					if a.IsArchived != nil {
						p.IsArchived = a.IsArchived.V
					}
					if a.IsPublic != nil {
						p.IsPublic = a.IsPublic.V
					}
				}
				srv := service.Get(tlbx)
				tx := srv.Data().Begin()
				defer tx.Rollback()
				for i, p := range ps {
					_, err := tx.Exec(`UPDATE projects SET name=?, currencyCode=?, hoursPerDay=?, daysPerWeek=?, startOn=?, dueOn=?, isArchived=?, isPublic=? WHERE host=? AND id=?`, p.Name, p.CurrencyCode, p.HoursPerDay, p.DaysPerWeek, p.StartOn, p.DueOn, p.IsArchived, p.IsPublic, me, p.ID)
					PanicOn(err)
					if namesSet[i] {
						_, err = tx.Exec(`UPDATE tasks SET name=? WHERE host=? AND project=? AND id=?`, p.Name, me, p.ID, p.ID)
						PanicOn(err)
						epsutil.ActivityItemRename(tx, me, p.ID, p.ID, p.Name)
					}
					epsutil.LogActivity(tlbx, tx, me, p.ID, p.ID, cnsts.TypeProject, cnsts.ActionUpdated, ptr.String(p.Name), args[i])
				}
				tx.Commit()
				return ps
			},
		},
		{
			Description:  "delete projects",
			Path:         (&project.Delete{}).Path(),
			Timeout:      500,
			MaxBodyBytes: app.KB,
			IsPrivate:    false,
			GetDefaultArgs: func() interface{} {
				return &project.Delete{}
			},
			GetExampleArgs: func() interface{} {
				return &project.Delete{
					app.ExampleID(),
				}
			},
			GetExampleResponse: func() interface{} {
				return nil
			},
			Handler: func(tlbx app.Tlbx, a interface{}) interface{} {
				args := *(a.(*project.Delete))
				if len(args) == 0 {
					return nil
				}
				app.BadReqIf(len(args) > 100, "can not delete more than 100 projects at a time")
				me := me.Get(tlbx)
				queryArgs := append([]interface{}{me}, IDs(args).ToIs()...)
				inID := sql.InCondition(true, "id", len(args))
				inProject := sql.InCondition(true, "project", len(args))
				srv := service.Get(tlbx)
				tx := srv.Data().Begin()
				defer tx.Rollback()
				_, err := tx.Exec(Strf(`DELETE FROM projectLocks WHERE host=? %s`, inID), queryArgs...)
				PanicOn(err)
				_, err = tx.Exec(Strf(`DELETE FROM users WHERE host=? %s`, inProject), queryArgs...)
				PanicOn(err)
				_, err = tx.Exec(Strf(`DELETE FROM activities WHERE host=? %s`, inProject), queryArgs...)
				PanicOn(err)
				_, err = tx.Exec(Strf(`DELETE FROM projects WHERE host=? %s`, inID), queryArgs...)
				PanicOn(err)
				_, err = tx.Exec(Strf(`DELETE FROM tasks WHERE host=? %s`, inProject), queryArgs...)
				PanicOn(err)
				_, err = tx.Exec(Strf(`DELETE FROM times WHERE host=? %s`, inProject), queryArgs...)
				PanicOn(err)
				_, err = tx.Exec(Strf(`DELETE FROM expenses WHERE host=? %s`, inProject), queryArgs...)
				PanicOn(err)
				_, err = tx.Exec(Strf(`DELETE FROM files WHERE host=? %s`, inProject), queryArgs...)
				PanicOn(err)
				_, err = tx.Exec(Strf(`DELETE FROM comments WHERE host=? %s`, inProject), queryArgs...)
				PanicOn(err)
				for _, p := range args {
					srv.Store().MustDeletePrefix(cnsts.FileBucket, epsutil.StorePrefix(me, p))
				}
				tx.Commit()
				return nil
			},
		},
		{
			Description:  "add project users",
			Path:         (&project.AddUsers{}).Path(),
			Timeout:      500,
			MaxBodyBytes: app.KB,
			IsPrivate:    false,
			GetDefaultArgs: func() interface{} {
				return &project.AddUsers{}
			},
			GetExampleArgs: func() interface{} {
				return &project.AddUsers{
					Project: app.ExampleID(),
					Users: []*project.SendUser{
						{
							ID:   app.ExampleID(),
							Role: cnsts.RoleAdmin,
						},
					},
				}
			},
			GetExampleResponse: func() interface{} {
				return nil
			},
			Handler: func(tlbx app.Tlbx, a interface{}) interface{} {
				args := a.(*project.AddUsers)
				lenUsers := len(args.Users)
				if lenUsers == 0 {
					return nil
				}
				app.BadReqIf(lenUsers > 100, "can not add more than 100 users to a project at a time")
				epsutil.IMustHaveAccess(tlbx, args.Host, args.Project, cnsts.RoleAdmin)
				srv := service.Get(tlbx)

				// need two sets for id IN (?, ...) and ORDER BY FIELD (id, ?, ...)
				ids := make([]interface{}, 0, 2*lenUsers)
				for i := 0; i < 2; i++ {
					for _, u := range args.Users {
						ids = append(ids, u.ID)
					}
				}
				// get userTx and lock all user rows, to ensure they are not changed whilst inserting into data db
				userTx := srv.User().Begin()
				defer userTx.Rollback()
				users := make([]*project.User, 0, lenUsers)
				PanicOn(userTx.Query(func(rows isql.Rows) {
					for rows.Next() {
						u := &project.User{}
						PanicOn(rows.Scan(&u.ID, &u.Handle, &u.Alias, &u.HasAvatar))
						u.IsActive = true
						users = append(users, u)
					}
				}, Strf(`SELECT id, handle, alias, hasAvatar FROM users WHERE 1=1 %s %s FOR UPDATE`, sql.InCondition(true, `id`, lenUsers), sql.OrderByField(`id`, lenUsers)), ids...))

				app.BadReqIf(len(users) != lenUsers, "users specified: %d, users found: %d", lenUsers, len(users))

				tx := srv.Data().Begin()
				defer tx.Rollback()
				for i, u := range users {
					app.BadReqIf(u.ID.Equal(args.Host), "can not add host to project")
					u.Role = args.Users[i].Role
					_, err := tx.Exec(`INSERT INTO users (host, project, id, handle, alias, hasAvatar, isActive, role) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`, args.Host, args.Project, u.ID, u.Handle, u.Alias, u.HasAvatar, u.IsActive, u.Role)
					PanicOn(err)
					epsutil.LogActivity(tlbx, tx, args.Host, args.Project, u.ID, cnsts.TypeUser, cnsts.ActionCreated, nil, u.Role)
				}
				tx.Commit()
				userTx.Commit()
				return nil
			},
		},
		{
			Description:  "get my project user value",
			Path:         (&project.GetMe{}).Path(),
			Timeout:      500,
			MaxBodyBytes: app.KB,
			IsPrivate:    false,
			GetDefaultArgs: func() interface{} {
				return &project.GetMe{}
			},
			GetExampleArgs: func() interface{} {
				return &project.GetMe{
					Host:    app.ExampleID(),
					Project: app.ExampleID(),
				}
			},
			GetExampleResponse: func() interface{} {
				return exampleUser
			},
			Handler: func(tlbx app.Tlbx, a interface{}) interface{} {
				args := a.(*project.GetMe)
				return getUsers(tlbx, &project.GetUsers{
					Host:    args.Host,
					Project: args.Project,
					IDs:     IDs{me.Get(tlbx)},
				}).Set[0]
			},
		},
		{
			Description:  "get project users",
			Path:         (&project.GetUsers{}).Path(),
			Timeout:      500,
			MaxBodyBytes: app.KB,
			IsPrivate:    false,
			GetDefaultArgs: func() interface{} {
				return &project.GetUsers{
					Limit: 100,
				}
			},
			GetExampleArgs: func() interface{} {
				r := cnsts.RoleReader
				return &project.GetUsers{
					Host:         app.ExampleID(),
					Project:      app.ExampleID(),
					IDs:          IDs{app.ExampleID()},
					Role:         &r,
					HandlePrefix: ptr.String("my_frien"),
					After:        ptr.ID(app.ExampleID()),
					Limit:        100,
				}
			},
			GetExampleResponse: func() interface{} {
				return &project.GetUsersRes{
					Set: []*project.User{
						exampleUser,
					},
					More: true,
				}
			},
			Handler: func(tlbx app.Tlbx, a interface{}) interface{} {
				return getUsers(tlbx, a.(*project.GetUsers))
			},
		},
		{
			Description:  "set project user roles",
			Path:         (&project.SetUserRoles{}).Path(),
			Timeout:      500,
			MaxBodyBytes: app.KB,
			IsPrivate:    false,
			GetDefaultArgs: func() interface{} {
				return &project.SetUserRoles{}
			},
			GetExampleArgs: func() interface{} {
				return &project.SetUserRoles{
					Project: app.ExampleID(),
					Users: []*project.SendUser{
						{
							ID:   app.ExampleID(),
							Role: cnsts.RoleAdmin,
						},
					},
				}
			},
			GetExampleResponse: func() interface{} {
				return nil
			},
			Handler: func(tlbx app.Tlbx, a interface{}) interface{} {
				args := a.(*project.SetUserRoles)
				lenUsers := len(args.Users)
				if lenUsers == 0 {
					return nil
				}
				app.BadReqIf(lenUsers > 100, "can not set more than 100 user roles in a project at a time")
				epsutil.IMustHaveAccess(tlbx, args.Host, args.Project, cnsts.RoleAdmin)
				srv := service.Get(tlbx)
				tx := srv.Data().Begin()
				defer tx.Rollback()
				for _, u := range args.Users {
					app.ReturnIf(u.ID.Equal(args.Host), http.StatusForbidden, "can not set hosts role")
					res, err := tx.Exec(`UPDATE users SET role=? WHERE host=? AND project=? AND id=?`, u.Role, args.Host, args.Project, u.ID)
					PanicOn(err)
					count, err := res.RowsAffected()
					PanicOn(err)
					app.ReturnIf(count != 1, http.StatusNotFound, "user: %s not found", u.ID)
					epsutil.LogActivity(tlbx, tx, args.Host, args.Project, u.ID, cnsts.TypeUser, cnsts.ActionUpdated, nil, u.Role)
				}
				tx.Commit()
				return nil
			},
		},
		{
			Description:  "remove project users",
			Path:         (&project.RemoveUsers{}).Path(),
			Timeout:      500,
			MaxBodyBytes: app.KB,
			IsPrivate:    false,
			GetDefaultArgs: func() interface{} {
				return &project.RemoveUsers{}
			},
			GetExampleArgs: func() interface{} {
				return &project.RemoveUsers{
					Host:    app.ExampleID(),
					Project: app.ExampleID(),
					Users: IDs{
						app.ExampleID(),
					},
				}
			},
			GetExampleResponse: func() interface{} {
				return nil
			},
			Handler: func(tlbx app.Tlbx, a interface{}) interface{} {
				args := a.(*project.RemoveUsers)
				if len(args.Users) == 0 {
					return nil
				}
				epsutil.IMustHaveAccess(tlbx, args.Host, args.Project, cnsts.RoleAdmin)
				queryArgs := make([]interface{}, 0, len(args.Users)+2)
				queryArgs = append(queryArgs, args.Host, args.Project)
				tx := service.Get(tlbx).Data().Begin()
				defer tx.Rollback()
				for _, u := range args.Users {
					app.BadReqIf(u.Equal(args.Host), "can not remove host from project")
					queryArgs = append(queryArgs, u)
				}
				_, err := tx.Exec(Strf(`UPDATE users SET isActive=0 WHERE host=? AND project=? %s`, sql.InCondition(true, `id`, len(args.Users))), queryArgs...)
				PanicOn(err)
				for _, u := range args.Users {
					epsutil.LogActivity(tlbx, tx, args.Host, args.Project, u, cnsts.TypeUser, cnsts.ActionDeleted, nil, nil)
				}
				tx.Commit()
				return nil
			},
		},
		{
			Description:  "get project activities",
			Path:         (&project.GetActivities{}).Path(),
			Timeout:      500,
			MaxBodyBytes: app.KB,
			IsPrivate:    false,
			GetDefaultArgs: func() interface{} {
				return &project.GetActivities{
					Limit: 100,
				}
			},
			GetExampleArgs: func() interface{} {
				return &project.GetActivities{
					Host:          app.ExampleID(),
					Project:       app.ExampleID(),
					Item:          ptr.ID(app.ExampleID()),
					User:          ptr.ID(app.ExampleID()),
					OccuredAfter:  ptr.Time(app.ExampleTime()),
					OccuredBefore: ptr.Time(app.ExampleTime().Add(24 * time.Hour)),
					Limit:         100,
				}
			},
			GetExampleResponse: func() interface{} {
				return &project.GetActivitiesRes{
					Set: []*project.Activity{
						{
							OccurredOn:         app.ExampleTime(),
							User:               app.ExampleID(),
							Item:               app.ExampleID(),
							ItemType:           cnsts.TypeTask,
							ItemHasBeenDeleted: false,
							Action:             cnsts.ActionUpdated,
							ItemName:           ptr.String("my task"),
							ExtraInfo:          ptr.String(`{"isParallel":true}`),
						},
					},
				}
			},
			Handler: func(tlbx app.Tlbx, a interface{}) interface{} {
				args := a.(*project.GetActivities)
				args.Limit = sql.Limit100(args.Limit)
				app.BadReqIf(args.OccuredAfter != nil && args.OccuredBefore != nil, "only one of occurredBefore or occurredAfter may be used")
				epsutil.IMustHaveAccess(tlbx, args.Host, args.Project, cnsts.RoleReader)
				query := bytes.NewBufferString(`SELECT occurredOn, user, item, itemType, itemHasBeenDeleted, action, itemName, extraInfo FROM activities WHERE host=? AND project=?`)
				queryArgs := make([]interface{}, 0, 7)
				queryArgs = append(queryArgs, args.Host, args.Project)
				if args.Item != nil {
					query.WriteString(` AND item=?`)
					queryArgs = append(queryArgs, *args.Item)
				}
				if args.User != nil {
					query.WriteString(` AND user=?`)
					queryArgs = append(queryArgs, *args.User)
				}
				asc := false
				if args.OccuredAfter != nil {
					asc = true
					query.WriteString(` AND occurredOn>?`)
					queryArgs = append(queryArgs, *args.OccuredAfter)
				}
				if args.OccuredBefore != nil {
					query.WriteString(` AND occurredOn<?`)
					queryArgs = append(queryArgs, *args.OccuredBefore)
				}
				query.WriteString(sql.OrderLimit100(`occurredOn`, asc, args.Limit))
				res := &project.GetActivitiesRes{
					Set: make([]*project.Activity, 0, args.Limit),
				}
				PanicOn(service.Get(tlbx).Data().Query(func(rows isql.Rows) {
					iLimit := int(args.Limit)
					for rows.Next() {
						if len(res.Set)+1 == iLimit {
							res.More = true
							break
						}
						pa := &project.Activity{}
						PanicOn(rows.Scan(&pa.OccurredOn, &pa.User, &pa.Item, &pa.ItemType, &pa.ItemHasBeenDeleted, &pa.Action, &pa.ItemName, &pa.ExtraInfo))
						res.Set = append(res.Set, pa)
					}
				}, query.String(), queryArgs...))
				return res
			},
		},
	}

	nameMinLen     = 1
	nameMaxLen     = 250
	exampleProject = &project.Project{
		Task: task.Task{
			ID:        app.ExampleID(),
			Name:      "My Project",
			CreatedOn: app.ExampleTime(),
		},
		Base: project.Base{
			CurrencyCode: "USD",
			HoursPerDay:  8,
			DaysPerWeek:  5,
			StartOn:      ptr.Time(app.ExampleTime()),
			DueOn:        ptr.Time(app.ExampleTime().Add(24 * time.Hour)),
			IsPublic:     false,
		},
		Host:       app.ExampleID(),
		IsArchived: false,
	}
	exampleUser = &project.User{
		User: user.User{
			ID:        app.ExampleID(),
			Handle:    ptr.String("joe_bloggs"),
			Alias:     ptr.String("joe soap"),
			HasAvatar: ptr.Bool(true),
		},
		Role:     cnsts.RoleReader,
		IsActive: true,
	}
)

func OnSetSocials(tlbx app.Tlbx, user *user.User) error {
	srv := service.Get(tlbx)
	tx := srv.Data().Begin()
	defer tx.Rollback()
	_, err := tx.Exec(`UPDATE users SET handle=?, alias=?, hasAvatar=? WHERE id=?`, user.Handle, user.Alias, user.HasAvatar, user.ID)
	PanicOn(err)
	tx.Commit()
	return nil
}

func OnDelete(tlbx app.Tlbx, me ID) {
	srv := service.Get(tlbx)
	tx := srv.Data().Begin()
	defer tx.Rollback()
	_, err := tx.Exec(`DELETE FROM projectLocks WHERE host=?`, me)
	PanicOn(err)
	_, err = tx.Exec(`DELETE FROM users WHERE host=?`, me)
	PanicOn(err)
	_, err = tx.Exec(`DELETE FROM activities WHERE host=?`, me)
	PanicOn(err)
	_, err = tx.Exec(`DELETE FROM projects WHERE host=?`, me)
	PanicOn(err)
	_, err = tx.Exec(`DELETE FROM tasks WHERE host=?`, me)
	PanicOn(err)
	_, err = tx.Exec(`DELETE FROM times WHERE host=?`, me)
	PanicOn(err)
	_, err = tx.Exec(`DELETE FROM expenses WHERE host=?`, me)
	PanicOn(err)
	_, err = tx.Exec(`DELETE FROM files WHERE host=?`, me)
	PanicOn(err)
	_, err = tx.Exec(`DELETE FROM comments WHERE host=?`, me)
	PanicOn(err)
	_, err = tx.Exec(`UPDATE users set isActive=0 WHERE id=?`, me)
	PanicOn(err)
	srv.Store().MustDeletePrefix(cnsts.FileBucket, epsutil.StorePrefix(me))
	tx.Commit()
}

func getSet(tlbx app.Tlbx, args *project.Get) *project.GetRes {
	validate.MaxIDs(tlbx, "ids", args.IDs, 100)
	app.BadReqIf(
		args.CreatedOnMin != nil &&
			args.CreatedOnMax != nil &&
			args.CreatedOnMin.After(*args.CreatedOnMax),
		"createdOnMin must be before createdOnMax")
	app.BadReqIf(
		args.StartOnMin != nil &&
			args.StartOnMax != nil &&
			args.StartOnMin.After(*args.StartOnMax),
		"startOnMin must be before startOnMax")
	app.BadReqIf(
		args.DueOnMin != nil &&
			args.DueOnMax != nil &&
			args.DueOnMin.After(*args.DueOnMax),
		"dueOnMin must be before dueOnMax")
	app.BadReqIf(
		args.StartOnMin != nil &&
			args.DueOnMin != nil &&
			args.StartOnMin.After(*args.DueOnMax),
		"startOnMin must be before dueOnMin")
	app.BadReqIf(
		args.StartOnMax != nil &&
			args.DueOnMax != nil &&
			args.StartOnMax.After(*args.DueOnMax),
		"startOnMax must be before dueOnMax")
	args.Limit = sql.Limit100(args.Limit)
	srv := service.Get(tlbx)
	res := &project.GetRes{
		Set: make([]*project.Project, 0, args.Limit),
	}
	query := bytes.NewBufferString(`SELECT p.host, p.id, p.isArchived, p.name, p.createdOn, p.currencyCode, p.hoursPerDay, p.daysPerWeek, p.startOn, p.dueOn, p.isPublic, t.parent, t.firstChild, t.nextSibling, t.user, t.name, t.description, t.createdBy, t.createdOn, t.minimumTime, t.estimatedTime, t.loggedTime, t.estimatedSubTime, t.loggedSubTime, t.estimatedExpense, t.loggedExpense, t.estimatedSubExpense, t.loggedSubExpense, t.fileCount, t.fileSize, t.fileSubCount, t.fileSubSize, t.childCount, t.descendantCount, t.isParallel FROM projects p JOIN tasks t ON (t.host=p.host AND t.project=p.id AND t.id=p.id) WHERE`)
	queryArgs := make([]interface{}, 0, 14)
	idsLen := len(args.IDs)
	if args.Host != nil {
		query.WriteString(` p.host=?`)
		queryArgs = append(queryArgs, args.Host)
		if me.Exists(tlbx) {
			me := me.Get(tlbx)
			if !me.Equal(*args.Host) {
				query.WriteString(` AND (p.isPublic=1 OR p.id IN (SELECT u.project FROM users u WHERE u.host=? AND u.isActive=1 AND u.id=?))`)
				queryArgs = append(queryArgs, args.Host, me)
			}
		} else {
			query.WriteString(` AND p.isPublic=1`)
		}
	} else {
		PanicIf(!me.Exists(tlbx), "if no host is specified, the request must come from an active user session")
		query.WriteString(` (p.host=? OR p.id IN (SELECT u.project FROM users u WHERE u.isActive=1 AND u.id=? AND u.host<>?))`)
		me := me.Get(tlbx)
		queryArgs = append(queryArgs, me, me, me)
	}
	if idsLen > 0 {
		query.WriteString(sql.InCondition(true, `p.id`, idsLen))
		query.WriteString(sql.OrderByField(`p.id`, idsLen))
		Is := args.IDs.ToIs()
		queryArgs = append(queryArgs, Is...)
		queryArgs = append(queryArgs, Is...)
	} else {
		query.WriteString(` AND p.isArchived=?`)
		queryArgs = append(queryArgs, args.IsArchived)
		if args.IsPublic != nil {
			query.WriteString(` AND p.isPublic=?`)
			queryArgs = append(queryArgs, *args.IsPublic)
		}
		if ptr.StringOr(args.NamePrefix, "") != "" {
			query.WriteString(` AND p.name LIKE ?`)
			queryArgs = append(queryArgs, Strf(`%s%%`, *args.NamePrefix))
		}
		if args.CreatedOnMin != nil {
			query.WriteString(` AND p.createdOn >=?`)
			queryArgs = append(queryArgs, *args.CreatedOnMin)
		}
		if args.CreatedOnMax != nil {
			query.WriteString(` AND p.createdOn <= ?`)
			queryArgs = append(queryArgs, *args.CreatedOnMax)
		}
		if args.StartOnMin != nil {
			query.WriteString(` AND p.startOn >=?`)
			queryArgs = append(queryArgs, *args.StartOnMin)
		}
		if args.StartOnMax != nil {
			query.WriteString(` AND p.startOn <=?`)
			queryArgs = append(queryArgs, *args.StartOnMax)
		}
		if args.DueOnMin != nil {
			query.WriteString(` AND p.dueOn >=?`)
			queryArgs = append(queryArgs, *args.DueOnMin)
		}
		if args.DueOnMax != nil {
			query.WriteString(` AND p.dueOn <=?`)
			queryArgs = append(queryArgs, *args.DueOnMax)
		}
		if args.After != nil {
			query.WriteString(Strf(` AND %s %s= (SELECT p.%s FROM projects p WHERE p.host=? AND p.id=?) AND p.id <> ?`, args.Sort, sql.GtLtSymbol(*args.Asc), args.Sort))
			queryArgs = append(queryArgs, args.Host, *args.After, *args.After)
			if args.Sort != cnsts.SortCreatedOn {
				query.WriteString(Strf(` AND p.createdOn %s (SELECT p.createdOn FROM projects p WHERE p.host=? AND p.id=?)`, sql.GtLtSymbol(*args.Asc)))
				queryArgs = append(queryArgs, args.Host, *args.After)
			}
		}
		createdOnSecondarySort := ", p.id"
		if args.Sort != cnsts.SortCreatedOn {
			createdOnSecondarySort = ", p.createdOn, p.id"
		}
		query.WriteString(sql.OrderLimit100("p."+string(args.Sort)+createdOnSecondarySort, *args.Asc, args.Limit))
	}
	PanicOn(srv.Data().Query(func(rows isql.Rows) {
		iLimit := int(args.Limit)
		for rows.Next() {
			if len(args.IDs) == 0 && len(res.Set)+1 == iLimit {
				res.More = true
				break
			}
			p := &project.Project{}
			PanicOn(rows.Scan(&p.Host, &p.ID, &p.IsArchived, &p.Name, &p.CreatedOn, &p.CurrencyCode, &p.HoursPerDay, &p.DaysPerWeek, &p.StartOn, &p.DueOn, &p.IsPublic, &p.Parent, &p.FirstChild, &p.NextSibling, &p.User, &p.Name, &p.Description, &p.CreatedBy, &p.CreatedOn, &p.MinimumTime, &p.EstimatedTime, &p.LoggedTime, &p.EstimatedSubTime, &p.LoggedSubTime, &p.EstimatedExpense, &p.LoggedExpense, &p.EstimatedSubExpense, &p.LoggedSubExpense, &p.FileCount, &p.FileSize, &p.FileSubCount, &p.FileSubSize, &p.ChildCount, &p.DescendantCount, &p.IsParallel))
			res.Set = append(res.Set, p)
		}
	}, query.String(), queryArgs...))
	return res
}

func getUsers(tlbx app.Tlbx, args *project.GetUsers) *project.GetUsersRes {
	validate.MaxIDs(tlbx, "ids", args.IDs, 100)
	app.BadReqIf(args.HandlePrefix != nil && StrLen(*args.HandlePrefix) >= 15, "handlePrefix must be < 15 chars long")
	epsutil.IMustHaveAccess(tlbx, args.Host, args.Project, cnsts.RoleReader)
	limit := sql.Limit100(args.Limit)
	srv := service.Get(tlbx)
	res := &project.GetUsersRes{
		Set: make([]*project.User, 0, limit),
	}
	query := bytes.NewBufferString(`SELECT id, handle, alias, hasAvatar, isActive, role FROM users WHERE host=? AND project=?`)
	queryArgs := make([]interface{}, 0, 14)
	queryArgs = append(queryArgs, args.Host, args.Project)
	idsLen := len(args.IDs)
	if idsLen > 0 {
		query.WriteString(sql.InCondition(true, `id`, idsLen))
		query.WriteString(sql.OrderByField(`id`, idsLen))
		Is := args.IDs.ToIs()
		queryArgs = append(queryArgs, Is...)
		queryArgs = append(queryArgs, Is...)
	} else {
		query.WriteString(` AND isActive=1`)
		if ptr.StringOr(args.HandlePrefix, "") != "" {
			query.WriteString(` AND handle LIKE ?`)
			queryArgs = append(queryArgs, Strf(`%s%%`, *args.HandlePrefix))
		}
		if args.Role != nil {
			query.WriteString(` AND role=?`)
			queryArgs = append(queryArgs, *args.Role)
		}
		if args.After != nil {
			if ptr.StringOr(args.HandlePrefix, "") == "" {
				query.WriteString(` AND role >= (SELECT role FROM users WHERE host=? AND project=? AND id=?)`)
				queryArgs = append(queryArgs, args.Host, args.Project, *args.After)
			}
			query.WriteString(` AND handle > (SELECT handle FROM users WHERE host=? AND project=? AND id=?)`)
			queryArgs = append(queryArgs, args.Host, args.Project, *args.After)
		}
		query.WriteString(` ORDER BY`)
		if ptr.StringOr(args.HandlePrefix, "") == "" {
			query.WriteString(` role ASC,`)
		}
		query.WriteString(Strf(` handle ASC LIMIT %d`, limit))
	}
	PanicOn(srv.Data().Query(func(rows isql.Rows) {
		for rows.Next() {
			if len(args.IDs) == 0 && len(res.Set)+1 == int(limit) {
				res.More = true
				break
			}
			u := &project.User{}
			PanicOn(rows.Scan(&u.ID, &u.Handle, &u.Alias, &u.HasAvatar, &u.IsActive, &u.Role))
			res.Set = append(res.Set, u)
		}
	}, query.String(), queryArgs...))
	return res
}
