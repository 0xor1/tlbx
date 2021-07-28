package taskeps

import (
	"net/http"
	"strings"
	"time"

	"github.com/0xor1/tlbx/cmd/trees/pkg/cnsts"
	"github.com/0xor1/tlbx/cmd/trees/pkg/task"
	. "github.com/0xor1/tlbx/pkg/core"
	"github.com/0xor1/tlbx/pkg/field"
	"github.com/0xor1/tlbx/pkg/isql"
	"github.com/0xor1/tlbx/pkg/ptr"
	"github.com/0xor1/tlbx/pkg/web/app"
	"github.com/0xor1/tlbx/pkg/web/app/service"
	"github.com/0xor1/tlbx/pkg/web/app/service/sql"
	"github.com/0xor1/tlbx/pkg/web/app/session/me"
	sqlh "github.com/0xor1/tlbx/pkg/web/app/sql"
	"github.com/0xor1/tlbx/pkg/web/app/validate"
	"github.com/0xor1/trees/pkg/epsutil"
)

var (
	Eps = []*app.Endpoint{
		{
			Description:  "Create a new task",
			Path:         (&task.Create{}).Path(),
			Timeout:      500,
			MaxBodyBytes: app.KB,
			IsPrivate:    false,
			GetDefaultArgs: func() interface{} {
				return &task.Create{}
			},
			GetExampleArgs: func() interface{} {
				return &task.Create{
					Host:        app.ExampleID(),
					Project:     app.ExampleID(),
					Parent:      app.ExampleID(),
					PrevSib:     ptr.ID(app.ExampleID()),
					Name:        "do it",
					Description: "do the thing you're supposed to do",
					IsParallel:  true,
					User:        ptr.ID(app.ExampleID()),
					TimeEst:     40,
				}
			},
			GetExampleResponse: func() interface{} {
				return &task.CreateRes{
					Parent: ExampleTask,
					Task:   ExampleTask,
				}
			},
			Handler: func(tlbx app.Tlbx, a interface{}) interface{} {
				args := a.(*task.Create)
				me := me.AuthedGet(tlbx)
				args.Name = StrTrimWS(args.Name)
				validate.Str(tlbx, "name", args.Name, nameMinLen, nameMaxLen)
				args.Description = StrTrimWS(args.Description)
				validate.Str(tlbx, "description", args.Description, 0, descriptionMaxLen)
				srv := service.Get(tlbx)
				tx := srv.Data().BeginWrite()
				defer tx.Rollback()
				epsutil.IMustHaveAccess(tlbx, tx, args.Host, args.Project, cnsts.RoleWriter)
				t := &task.Task{
					ID:          tlbx.NewID(),
					Parent:      &args.Parent,
					FirstChild:  nil,
					NextSib:     nil,
					User:        args.User,
					Name:        args.Name,
					Description: args.Description,
					CreatedBy:   me,
					CreatedOn:   tlbx.Start(),
					TimeEst:     args.TimeEst,
					TimeInc:     0,
					TimeSubMin:  0,
					TimeSubEst:  0,
					TimeSubInc:  0,
					CostEst:     args.CostEst,
					CostInc:     0,
					CostSubEst:  0,
					CostSubInc:  0,
					FileN:       0,
					FileSize:    0,
					FileSubN:    0,
					FileSubSize: 0,
					ChildN:      0,
					DescN:       0,
					IsParallel:  args.IsParallel,
				}
				if args.User != nil && !args.User.Equal(me) {
					// if Im assigning to someone that isnt me,
					// validate that user has write access to this
					// project
					epsutil.MustHaveAccess(tlbx, tx, args.Host, args.Project, args.User, cnsts.RoleWriter)
				}
				// lock project, required for any action that will change aggregate values nad/or tree structure
				epsutil.MustLockProject(tx, args.Host, args.Project)
				// get correct next sib value from either prevSib if
				// specified or parent.FirstChild otherwise. Then update prevSibs nextSib value
				// or parents firstChild value depending on the scenario.
				var prevSib *task.Task
				if args.PrevSib != nil {
					prevSib = GetOne(tx, args.Host, args.Project, *args.PrevSib)
					app.ReturnIf(prevSib == nil, http.StatusNotFound, "prevSib not found")
					app.BadReqIf(prevSib.Parent == nil || !prevSib.Parent.Equal(args.Parent), "prevSib and parent args mismatch")
					t.NextSib = prevSib.NextSib
					prevSib.NextSib = &t.ID
					// point prev sib at new task
					_, err := tx.Exec(`UPDATE tasks SET nextSib=? WHERE host=? AND project=? AND id=?`, t.ID, args.Host, args.Project, prevSib.ID)
					PanicOn(err)
				} else {
					// else newTask is being inserted as firstChild, so set any current firstChild
					// as newTask's NextSib
					// get parent for updating child/descendant counts and firstChild if required
					parent := GetOne(tx, args.Host, args.Project, args.Parent)
					app.ReturnIf(parent == nil, http.StatusNotFound, "parent not found")
					t.NextSib = parent.FirstChild
					// increment parents child and descendant counters and firstChild pointer incase that was changed
					_, err := tx.Exec(`UPDATE tasks SET firstChild=? WHERE host=? AND project=? AND id=?`, t.ID, args.Host, args.Project, t.Parent)
					PanicOn(err)
				}
				// insert new task
				_, err := tx.Exec(Strf(`INSERT INTO tasks (host, project, %s) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, sql_task_columns), args.Host, args.Project, t.ID, t.Parent, t.FirstChild, t.NextSib, t.User, t.Name, t.Description, t.CreatedBy, t.CreatedOn, t.TimeEst, t.TimeInc, t.TimeSubMin, t.TimeSubEst, t.TimeSubInc, t.CostEst, t.CostInc, t.CostSubEst, t.CostSubInc, t.FileN, t.FileSize, t.FileSubN, t.FileSubSize, t.ChildN, t.DescN, t.IsParallel)
				PanicOn(err)
				// at this point the tree structure has been updated so all tasks are pointing to the correct new positions
				// all that remains to do is update aggregate values
				ancestors := epsutil.SetAncestralChainAggregateValuesFromTask(tx, args.Host, args.Project, args.Parent)
				epsutil.LogActivity(tlbx, tx, args.Host, args.Project, t.ID, t.ID, cnsts.TypeTask, cnsts.ActionCreated, &t.Name, nil, nil, ancestors)
				p := GetOne(tx, args.Host, args.Project, *t.Parent)
				tx.Commit()
				return &task.CreateRes{
					Parent: p,
					Task:   t,
				}
			},
		},
		{
			Description:  "Update a task",
			Path:         (&task.Update{}).Path(),
			Timeout:      500,
			MaxBodyBytes: app.KB,
			IsPrivate:    false,
			GetDefaultArgs: func() interface{} {
				return &task.Update{}
			},
			GetExampleArgs: func() interface{} {
				return &task.Update{
					Host:        app.ExampleID(),
					Project:     app.ExampleID(),
					ID:          app.ExampleID(),
					Parent:      &field.ID{V: app.ExampleID()},
					PrevSib:     &field.IDPtr{V: ptr.ID(app.ExampleID())},
					Name:        &field.String{V: "new name"},
					Description: &field.String{V: "new description"},
					IsParallel:  &field.Bool{V: true},
					User:        &field.IDPtr{V: ptr.ID(app.ExampleID())},
					TimeEst:     &field.UInt64{V: 123},
					CostEst:     &field.UInt64{V: 123},
				}
			},
			GetExampleResponse: func() interface{} {
				return &task.UpdateRes{
					OldParent: ExampleTask,
					NewParent: ExampleTask,
					Task:      ExampleTask,
				}
			},
			Handler: func(tlbx app.Tlbx, a interface{}) interface{} {
				args := a.(*task.Update)
				me := me.AuthedGet(tlbx)
				if args.Parent == nil &&
					args.PrevSib == nil &&
					args.Name == nil &&
					args.Description == nil &&
					args.IsParallel == nil &&
					args.User == nil &&
					args.TimeEst == nil &&
					args.CostEst == nil {
					// nothing to update
					return nil
				}
				tx := service.Get(tlbx).Data().BeginWrite()
				defer tx.Rollback()
				if args.ID.Equal(args.Project) {
					app.ReturnIf(!me.Equal(args.Host), http.StatusForbidden, "only the host may edit the project root node")
					app.ReturnIf(args.User != nil, http.StatusForbidden, "user value is not settable on the project root node")
					app.ReturnIf(args.Parent != nil, http.StatusForbidden, "parent value is not settable on the project root node")
					app.ReturnIf(args.PrevSib != nil, http.StatusForbidden, "prevSib value is not settable on the project root node")
				} else {
					epsutil.IMustHaveAccess(tlbx, tx, args.Host, args.Project, cnsts.RoleWriter)
				}
				treeUpdateRequired := false
				simpleUpdateRequired := false
				if args.Parent != nil ||
					args.PrevSib != nil ||
					args.CostEst != nil ||
					args.TimeEst != nil ||
					args.IsParallel != nil {
					// if moving the task or setting an aggregate value effecting property
					// we must lock
					epsutil.MustLockProject(tx, args.Host, args.Project)
				}
				t := GetOne(tx, args.Host, args.Project, args.ID)
				app.ReturnIf(t == nil, http.StatusNotFound, "task not found")
				var newParent, newPrevSib, oldParent, oldPrevSib *task.Task
				if args.Parent != nil {
					if !args.Parent.V.Equal(*t.Parent) {
						var newNextSib *ID
						// validate new parent exists
						newParent = GetOne(tx, args.Host, args.Project, args.Parent.V)
						app.ReturnIf(newParent == nil, http.StatusNotFound, "parent not found")
						// we must ensure we dont allow recursive loops
						row := tx.QueryRow(Strf("%s SELECT COUNT(*)=1 FROM ancestors a WHERE id=?", sql_ancestors_cte), args.Host, args.Project, args.Parent.V, args.Host, args.Project, args.ID)
						ancestorLoopDetected := false
						PanicOn(row.Scan(&ancestorLoopDetected))
						app.BadReqIf(ancestorLoopDetected || t.ID.Equal(args.Parent.V), "ancestor loop detected, invalid parent value")
						if args.PrevSib != nil && args.PrevSib.V != nil {
							app.BadReqIf(args.PrevSib.V.Equal(args.ID), "sib loop detected, invalid prevSib value")
							newPrevSib = GetOne(tx, args.Host, args.Project, *args.PrevSib.V)
							app.ReturnIf(newPrevSib == nil, http.StatusNotFound, "prevSib not found")
							app.BadReqIf(!newPrevSib.Parent.Equal(args.Parent.V), "prevSibs parent does not match the specified parent arg")
							newNextSib = newPrevSib.NextSib
							newPrevSib.NextSib = &t.ID
						} else {
							newNextSib = newParent.FirstChild
							newParent.FirstChild = &t.ID
						}
						// need to reconnect oldPrevSib with oldNextSib
						if newParent.NextSib != nil && newParent.NextSib.Equal(args.ID) {
							// !!!SPECIAL CASE!!! oldPrevSib may be newParent
							oldPrevSib = newParent
						} else {
							oldPrevSib = getPrevSib(tx, args.Host, args.Project, args.ID)
						}
						// need to get old parent for ancestor value updates
						if newPrevSib != nil && newPrevSib.ID.Equal(*t.Parent) {
							// !!!SPECIAL CASE!!! oldParent may be the newPrevSib
							oldParent = newPrevSib
						} else {
							oldParent = GetOne(tx, args.Host, args.Project, *t.Parent)
						}
						app.ReturnIf(oldParent == nil, http.StatusNotFound, "oldParent not found")
						if oldPrevSib != nil {
							oldPrevSib.NextSib = t.NextSib
						} else {
							// need to update oldParent firstChild as t is it
							oldParent.FirstChild = t.NextSib
						}
						t.Parent = &newParent.ID
						t.NextSib = newNextSib
						treeUpdateRequired = true
					} else {
						// if args.Parent is set and it's equal to the current value
						// i.e. no change is being made then let's just nil it out
						args.Parent = nil
					}
				}
				if args.Parent == nil && args.PrevSib != nil {
					// we now know we are doing a purely horizontal move, i.e. not changing parent node
					// get oldPrevSib
					oldPrevSib = getPrevSib(tx, args.Host, args.Project, args.ID)
					if !((oldPrevSib == nil && args.PrevSib.V == nil) ||
						(oldPrevSib != nil && args.PrevSib.V != nil &&
							oldPrevSib.ID.Equal(*args.PrevSib.V))) {
						var newNextSib *ID
						// here we know that an actual change is being attempted
						oldParent = GetOne(tx, args.Host, args.Project, *t.Parent)
						PanicIf(oldParent == nil, "oldParent not found")
						if args.PrevSib.V != nil {
							// moving to a non first child position
							if oldParent.FirstChild.Equal(t.ID) {
								//moving the old first child away so need to repoint oldParent.firstChild
								oldParent.FirstChild = t.NextSib
							} else {
								// not moving first child therefor nil out oldParent to
								// save an update query
								oldParent = nil
							}
							app.BadReqIf(args.PrevSib.V.Equal(args.ID), "sib loop detected, invalid prevSib value")
							newPrevSib = GetOne(tx, args.Host, args.Project, *args.PrevSib.V)
							app.ReturnIf(newPrevSib == nil, http.StatusNotFound, "prevSib not found")
							app.BadReqIf(!newPrevSib.Parent.Equal(*t.Parent), "prevSibs parent does not match the current tasks parent")
							newNextSib = newPrevSib.NextSib
							newPrevSib.NextSib = &t.ID
						} else {
							// moving to the first child position
							newNextSib = oldParent.FirstChild
							oldParent.FirstChild = &t.ID
						}
						// need to reconnect oldPrevSib with oldNextSib
						if oldPrevSib != nil {
							oldPrevSib.NextSib = t.NextSib
						}
						t.NextSib = newNextSib
						treeUpdateRequired = true
					} else {
						// here we know no change is being made so nil out oldPrevSib to save a sql update call
						oldPrevSib = nil
						// and nil this out to remove redundant data from activity entry
						args.PrevSib = nil
					}
				}
				nameUpdated := false
				// at this point all the moving has been done
				if args.Name != nil && t.Name != args.Name.V {
					args.Name.V = StrTrimWS(args.Name.V)
					validate.Str(tlbx, "name", args.Name.V, nameMinLen, nameMaxLen)
					t.Name = args.Name.V
					simpleUpdateRequired = true
					nameUpdated = true
				}
				if args.Description != nil && args.Description.V != t.Description {
					args.Description.V = StrTrimWS(args.Description.V)
					validate.Str(tlbx, "description", args.Description.V, 0, descriptionMaxLen)
					t.Description = args.Description.V
					simpleUpdateRequired = true
				}
				isParallelChanged := false
				if args.IsParallel != nil && t.IsParallel != args.IsParallel.V {
					t.IsParallel = args.IsParallel.V
					treeUpdateRequired = true
					isParallelChanged = true
				}
				if args.User != nil &&
					((args.User.V == nil && t.User != nil) ||
						(args.User.V != nil && t.User == nil) ||
						(args.User.V != nil && t.User != nil && !t.User.Equal(*args.User.V))) {
					if args.User.V != nil && !args.User.V.Equal(me) {
						epsutil.MustHaveAccess(tlbx, tx, args.Host, args.Project, args.User.V, cnsts.RoleWriter)
					}
					t.User = args.User.V
					simpleUpdateRequired = true
				}
				if args.TimeEst != nil && t.TimeEst != args.TimeEst.V {
					t.TimeEst = args.TimeEst.V
					treeUpdateRequired = true
				}
				if args.CostEst != nil && t.CostEst != args.CostEst.V {
					t.CostEst = args.CostEst.V
					treeUpdateRequired = true
				}
				update := func(ts ...*task.Task) {
					updated := map[string]bool{}
					for _, t := range ts {
						if t != nil {
							idStr := t.ID.String()
							// this check is for the two special cases that newPrevSib may also be the oldParent
							// or the newParent may also be the oldPrevSib
							// thus saving a duplicated update query
							if !updated[idStr] {
								updated[idStr] = true
								_, err := tx.Exec(`UPDATE tasks SET parent=?, firstChild=?, nextSib=?, name=?, description=?, isParallel=?, user=?, timeEst=?, costEst=? WHERE host=? AND project=? AND id=?`, t.Parent, t.FirstChild, t.NextSib, t.Name, t.Description, t.IsParallel, t.User, t.TimeEst, t.CostEst, args.Host, args.Project, t.ID)
								PanicOn(err)
							}
						}
					}
				}
				if simpleUpdateRequired || treeUpdateRequired {
					var ancestors IDs
					update(t, oldParent, oldPrevSib, newParent, newPrevSib)
					if treeUpdateRequired {
						if isParallelChanged {
							// if parallel has been changed need to recalculate agg values on this task as it effects minimum time
							ancestors = epsutil.SetAncestralChainAggregateValuesFromTask(tx, args.Host, args.Project, t.ID)
							if len(ancestors) > 0 {
								// here we know the task was updated as it's id was returned in the ancestors set
								// so we need to get it again as the timeMin value has changed
								t = GetOne(tx, args.Host, args.Project, t.ID)
							}
						}
						if newParent != nil {
							// if we moved parent we must recalculate aggregate values on the new and old parents ancestral chains
							if len(ancestors) < 2 {
								// if ancestors has less than 2 entries the new parent hasn't been updated yet
								ancestors = epsutil.SetAncestralChainAggregateValuesFromTask(tx, args.Host, args.Project, newParent.ID)
							}
							moreAncestors := epsutil.SetAncestralChainAggregateValuesFromTask(tx, args.Host, args.Project, oldParent.ID)
							ancestors = IDsMerge(ancestors, moreAncestors)
						} else if t.Parent != nil {
							// need to do the t.Parent nil check here incase it's the root project node having est value updated
							// here a tree update is required but the task has not been moved parents
							// so it must have had an est value changed on it, so update from its parent
							if len(ancestors) < 2 {
								// if ancestors has less than 2 entries the parent hasn't been updated yet by the call in the isParallelChanged section above.
								ancestors = epsutil.SetAncestralChainAggregateValuesFromTask(tx, args.Host, args.Project, *t.Parent)
							}
						}

					}
					if args.Name != nil {
						args.Name.V = StrEllipsis(args.Name.V, 50)
					}
					if args.Description != nil {
						args.Description.V = StrEllipsis(args.Description.V, 50)
					}
					epsutil.LogActivity(tlbx, tx, args.Host, args.Project, args.ID, args.ID, cnsts.TypeTask, cnsts.ActionUpdated, &t.Name, struct {
						Parent      *field.ID     `json:"parent,omitempty"`
						PrevSib     *field.IDPtr  `json:"prevSib,omitempty"`
						Name        *field.String `json:"name,omitempty"`
						Description *field.String `json:"description,omitempty"`
						IsParallel  *field.Bool   `json:"isParallel,omitempty"`
						User        *field.IDPtr  `json:"user,omitempty"`
						TimeEst     *field.UInt64 `json:"timeEst,omitempty"`
						CostEst     *field.UInt64 `json:"costEst,omitempty"`
					}{
						Parent:      args.Parent,
						PrevSib:     args.PrevSib,
						Name:        args.Name,
						Description: args.Description,
						IsParallel:  args.IsParallel,
						User:        args.User,
						TimeEst:     args.TimeEst,
						CostEst:     args.CostEst,
					}, nil, ancestors)
				}

				if nameUpdated {
					epsutil.ActivityItemRename(tx, args.Host, args.Project, args.ID, t.Name, true)
				}
				res := &task.UpdateRes{
					Task: t,
				}
				if args.Parent != nil {
					// if the task moved parent get both old and new parents for response
					res.OldParent = GetOne(tx, args.Host, args.Project, oldParent.ID)
					res.NewParent = GetOne(tx, args.Host, args.Project, args.Parent.V)
				} else if treeUpdateRequired && t.Parent != nil {
					// task didnt move parent, and it's not the root project node
					// but an est value was updated so return old parent
					res.OldParent = GetOne(tx, args.Host, args.Project, *t.Parent)
				}
				tx.Commit()
				return res
			},
		},
		{
			Description:  "Delete a task (returns the parent of the deleted task)",
			Path:         (&task.Delete{}).Path(),
			Timeout:      0,
			MaxBodyBytes: app.KB,
			IsPrivate:    false,
			GetDefaultArgs: func() interface{} {
				return &task.Delete{}
			},
			GetExampleArgs: func() interface{} {
				return &task.Delete{
					Host:    app.ExampleID(),
					Project: app.ExampleID(),
					ID:      app.ExampleID(),
				}
			},
			GetExampleResponse: func() interface{} {
				return ExampleTask
			},
			Handler: func(tlbx app.Tlbx, a interface{}) interface{} {
				args := a.(*task.Delete)
				app.BadReqIf(args.ID.Equal(args.Project), "use project delete endpoint to delete a project node")
				me := me.AuthedGet(tlbx)
				srv := service.Get(tlbx)
				tx := srv.Data().BeginWrite()
				defer tx.Rollback()
				role := epsutil.MustGetRole(tlbx, tx, args.Host, args.Project, me)
				app.ReturnIf(role == cnsts.RoleReader, http.StatusForbidden, "you don't have permission to delete a task")
				epsutil.MustLockProject(tx, args.Host, args.Project)
				// at this point we need to get the task
				t := GetOne(tx, args.Host, args.Project, args.ID)
				app.ReturnIf(t == nil, http.StatusNotFound, "task not found")
				app.BadReqIf(t.DescN > 100, "may not delete more than 100 task per delete action")
				app.ReturnIf(role == cnsts.RoleWriter && (!t.CreatedBy.Equal(me) || t.DescN > 0 || t.CreatedOn.Before(Now().Add(-1*time.Hour))), http.StatusForbidden, "you may only delete your own tasks within an hour of creating them and they must have no children")
				prevNode := getPrevSib(tx, args.Host, args.Project, args.ID)
				if prevNode == nil {
					prevNode = GetOne(tx, args.Host, args.Project, *t.Parent)
					PanicIf(!prevNode.FirstChild.Equal(t.ID), "invalid data detected, deleting task %s", t.ID)
					prevNode.FirstChild = t.NextSib
				} else {
					prevNode.NextSib = t.NextSib
				}
				tasksWithFiles := make(IDs, 0, t.DescN+1)
				tasksToDelete := make(IDs, 0, t.DescN+1)
				tx.Query(func(rows isql.Rows) {
					for rows.Next() {
						i := ID{}
						hasFiles := false
						PanicOn(rows.Scan(&i, &hasFiles))
						tasksToDelete = append(tasksToDelete, i)
						if hasFiles {
							tasksWithFiles = append(tasksWithFiles, i)
						}
					}
				}, `WITH RECURSIVE descendants (id, hasFiles) AS (SELECT id, fileN > 0 AS hasFiles  FROM tasks WHERE host=? AND project=? AND id=? UNION SELECT t.id, t.fileN > 0 AS hasFiles FROM tasks t, descendants d WHERE t.host=? AND t.project=? AND t.parent=d.id) CYCLE id RESTRICT SELECT id, hasFiles FROM descendants`, args.Host, args.Project, args.ID, args.Host, args.Project)
				if len(tasksToDelete) > 0 {
					queryArgs := make([]interface{}, 0, len(tasksToDelete)+2)
					queryArgs = append(queryArgs, args.Host, args.Project)
					queryArgs = append(queryArgs, tasksToDelete.ToIs()...)
					_, err := tx.Exec(Strf(`DELETE FROM tasks WHERE host=? AND project=? %s`, sqlh.InCondition(true, `id`, len(tasksToDelete))), queryArgs...)
					PanicOn(err)
					_, err = tx.Exec(`UPDATE tasks SET firstChild=?, nextSib=? WHERE host=? AND project=? AND id=?`, prevNode.FirstChild, prevNode.NextSib, args.Host, args.Project, prevNode.ID)
					PanicOn(err)
					ancestors := epsutil.SetAncestralChainAggregateValuesFromTask(tx, args.Host, args.Project, *t.Parent)
					t.Name = StrEllipsis(t.Name, 50)
					t.Description = StrEllipsis(t.Description, 50)
					epsutil.LogActivity(tlbx, tx, args.Host, args.Project, args.ID, args.ID, cnsts.TypeTask, cnsts.ActionDeleted, &t.Name, t, nil, ancestors)

					sql_in_tasks := sqlh.InCondition(true, `task`, len(tasksToDelete))
					// first get all time/cost/file/comment ids being deleted then
					_, err = tx.Exec(Strf(`DELETE FROM vitems WHERE host=? AND project=? %s`, sql_in_tasks), queryArgs...)
					PanicOn(err)
					_, err = tx.Exec(Strf(`DELETE FROM files WHERE host=? AND project=? %s`, sql_in_tasks), queryArgs...)
					PanicOn(err)
					_, err = tx.Exec(Strf(`DELETE FROM comments WHERE host=? AND project=? %s`, sql_in_tasks), queryArgs...)
					PanicOn(err)

					// set all relevant activity logs as taskDeleted =1, itemDeleted=1
					_, err = tx.Exec(Strf(`UPDATE activities SET taskDeleted=1, itemDeleted=1 WHERE host=? AND project=? %s`, sql_in_tasks), queryArgs...)
					PanicOn(err)

					for _, t := range tasksWithFiles {
						srv.Store().MustDeletePrefix(cnsts.FileBucket, epsutil.StorePrefix(args.Host, args.Project, t))
					}
				}
				parent := GetOne(tx, args.Host, args.Project, *t.Parent)
				tx.Commit()
				// return parent to show aggregate value changes
				return parent
			},
		},
		{
			Description:  "get a task",
			Path:         (&task.Get{}).Path(),
			Timeout:      500,
			MaxBodyBytes: app.KB,
			IsPrivate:    false,
			GetDefaultArgs: func() interface{} {
				return &task.Get{}
			},
			GetExampleArgs: func() interface{} {
				return &task.Get{
					Host:    app.ExampleID(),
					Project: app.ExampleID(),
					ID:      app.ExampleID(),
				}
			},
			GetExampleResponse: func() interface{} {
				return ExampleTask
			},
			Handler: func(tlbx app.Tlbx, a interface{}) interface{} {
				args := a.(*task.Get)
				tx := service.Get(tlbx).Data().BeginRead()
				defer tx.Rollback()
				epsutil.IMustHaveAccess(tlbx, tx, args.Host, args.Project, cnsts.RoleReader)
				t := GetOne(tx, args.Host, args.Project, args.ID)
				app.ReturnIf(t == nil, http.StatusNotFound, "task not found")
				tx.Commit()
				return t
			},
		},
		{
			Description:  "get task ancestors",
			Path:         (&task.GetAncestors{}).Path(),
			Timeout:      500,
			MaxBodyBytes: app.KB,
			IsPrivate:    false,
			GetDefaultArgs: func() interface{} {
				return &task.GetAncestors{
					Limit: 10,
				}
			},
			GetExampleArgs: func() interface{} {
				return &task.GetAncestors{
					Host:    app.ExampleID(),
					Project: app.ExampleID(),
					ID:      app.ExampleID(),
					Limit:   20,
				}
			},
			GetExampleResponse: func() interface{} {
				return &task.GetSetRes{
					Set:  []*task.Task{ExampleTask},
					More: true,
				}
			},
			Handler: func(tlbx app.Tlbx, a interface{}) interface{} {
				args := a.(*task.GetAncestors)
				tx := service.Get(tlbx).Data().BeginRead()
				defer tx.Rollback()
				epsutil.IMustHaveAccess(tlbx, tx, args.Host, args.Project, cnsts.RoleReader)
				args.Limit = sqlh.Limit100(args.Limit)
				res := &task.GetSetRes{
					Set:  make([]*task.Task, 0, args.Limit),
					More: false,
				}
				PanicOn(tx.Query(func(rows isql.Rows) {
					iLimit := int(args.Limit)
					for rows.Next() {
						if len(res.Set)+1 == iLimit {
							res.More = true
							break
						}
						t, err := Scan(rows)
						PanicOn(err)
						res.Set = append(res.Set, t)
					}
				}, Strf(`%s SELECT %s FROM tasks t JOIN ancestors a ON t.id = a.id WHERE t.host=? AND t.project=? ORDER BY a.n ASC LIMIT ?`, sql_ancestors_cte, Sql_task_columns_prefixed), args.Host, args.Project, args.ID, args.Host, args.Project, args.Host, args.Project, args.Limit))
				tx.Commit()
				return res
			},
		},
		{
			Description:  "get task children",
			Path:         (&task.GetChildren{}).Path(),
			Timeout:      500,
			MaxBodyBytes: app.KB,
			IsPrivate:    false,
			GetDefaultArgs: func() interface{} {
				return &task.GetChildren{
					Limit: 10,
				}
			},
			GetExampleArgs: func() interface{} {
				return &task.GetChildren{
					Host:    app.ExampleID(),
					Project: app.ExampleID(),
					ID:      app.ExampleID(),
					After:   ptr.ID(app.ExampleID()),
					Limit:   20,
				}
			},
			GetExampleResponse: func() interface{} {
				return &task.GetSetRes{
					Set:  []*task.Task{ExampleTask},
					More: true,
				}
			},
			Handler: func(tlbx app.Tlbx, a interface{}) interface{} {
				args := a.(*task.GetChildren)
				tx := service.Get(tlbx).Data().BeginRead()
				defer tx.Rollback()
				epsutil.IMustHaveAccess(tlbx, tx, args.Host, args.Project, cnsts.RoleReader)
				args.Limit = sqlh.Limit100(args.Limit)
				res := &task.GetSetRes{
					Set:  make([]*task.Task, 0, args.Limit),
					More: false,
				}
				var sql_filter string
				queryArgs := make([]interface{}, 0, 10)
				queryArgs = append(queryArgs, args.Host, args.Project)
				if args.After == nil {
					sql_filter = `firstChild`
					queryArgs = append(queryArgs, args.ID)
				} else {
					sql_filter = `nextSib`
					queryArgs = append(queryArgs, *args.After)
				}
				queryArgs = append(queryArgs, args.Host, args.Project, args.Host, args.Project, args.Limit)
				PanicOn(tx.Query(func(rows isql.Rows) {
					iLimit := int(args.Limit)
					for rows.Next() {
						if len(res.Set)+1 == iLimit {
							res.More = true
							break
						}
						t, err := Scan(rows)
						PanicOn(err)
						res.Set = append(res.Set, t)
					}
				}, Strf(`WITH RECURSIVE sibs (n, id) AS (SELECT 0 AS n, %s AS id FROM tasks WHERE host=? AND project=? AND id=? UNION SELECT s.n + 1 AS n, t.nextSib AS id FROM tasks t, sibs s WHERE t.host=? AND t.project=? AND t.id = s.id) CYCLE id RESTRICT SELECT %s FROM tasks t JOIN sibs s ON t.id = s.id WHERE t.host=? AND t.project=? ORDER BY s.n ASC LIMIT ?`, sql_filter, Sql_task_columns_prefixed), queryArgs...))
				tx.Commit()
				return res
			},
		},
		{
			Description:  "get task tree",
			Path:         (&task.GetTree{}).Path(),
			Timeout:      500,
			MaxBodyBytes: app.KB,
			IsPrivate:    false,
			GetDefaultArgs: func() interface{} {
				return &task.GetTree{}
			},
			GetExampleArgs: func() interface{} {
				return &task.GetTree{
					Host:    app.ExampleID(),
					Project: app.ExampleID(),
					ID:      app.ExampleID(),
				}
			},
			GetExampleResponse: func() interface{} {
				return &task.GetTreeRes{
					app.ExampleID(): ExampleTask,
				}
			},
			Handler: func(tlbx app.Tlbx, a interface{}) interface{} {
				args := a.(*task.GetTree)
				tx := service.Get(tlbx).Data().BeginRead()
				defer tx.Rollback()
				epsutil.IMustHaveAccess(tlbx, tx, args.Host, args.Project, cnsts.RoleReader)
				t := GetOne(tx, args.Host, args.Project, args.ID)
				app.ReturnIf(t == nil, http.StatusNotFound, "task not found")
				app.BadReqIf(t.DescN > 1000, "get tree may only be called on a task with descN <= 1000")
				res := task.GetTreeRes{
					t.ID: t,
				}
				if t.DescN > 0 {
					queryArgs := make([]interface{}, 0, 10)
					queryArgs = append(queryArgs, args.Host, args.Project, args.ID, args.Host, args.Project, args.Host, args.Project)
					PanicOn(tx.Query(func(rows isql.Rows) {
						for rows.Next() {
							t, err := Scan(rows)
							PanicOn(err)
							res[t.ID] = t
						}
					}, Strf(`WITH RECURSIVE nodes (id) AS (SELECT id FROM tasks WHERE host=? AND project=? AND parent=? UNION SELECT t.id FROM tasks t JOIN nodes n ON t.parent = n.id WHERE t.host=? AND t.project=?) CYCLE id RESTRICT SELECT %s FROM tasks t JOIN nodes n ON t.id = n.id WHERE t.host=? AND t.project=?`, Sql_task_columns_prefixed), queryArgs...))
				}
				tx.Commit()
				return res
			},
		},
	}

	nameMinLen        = 1
	nameMaxLen        = 250
	descriptionMaxLen = 1250
	ExampleTask       = &task.Task{
		ID:          app.ExampleID(),
		Parent:      ptr.ID(app.ExampleID()),
		FirstChild:  ptr.ID(app.ExampleID()),
		NextSib:     ptr.ID(app.ExampleID()),
		User:        ptr.ID(app.ExampleID()),
		Name:        "do it",
		Description: "do that thing you're supposed to do",
		CreatedBy:   app.ExampleID(),
		CreatedOn:   app.ExampleTime(),
		TimeSubMin:  100,
		TimeEst:     100,
		TimeInc:     100,
		TimeSubEst:  100,
		TimeSubInc:  100,
		CostEst:     100,
		CostInc:     100,
		CostSubEst:  100,
		CostSubInc:  100,
		FileN:       100,
		FileSize:    100,
		FileSubN:    100,
		FileSubSize: 100,
		ChildN:      100,
		DescN:       100,
		IsParallel:  true,
	}
)

func GetOne(tx sql.Tx, host, project, id ID) *task.Task {
	row := tx.QueryRow(Strf(`SELECT %s FROM tasks t WHERE t.host=? AND t.project=? AND id=?`, Sql_task_columns_prefixed), host, project, id)
	t, err := Scan(row)
	sqlh.PanicIfIsntNoRows(err)
	return t
}

func getPrevSib(tx sql.Tx, host, project, nextSib ID) *task.Task {
	row := tx.QueryRow(Strf(`SELECT %s FROM tasks t WHERE t.host=? AND t.project=? AND t.nextSib=?`, Sql_task_columns_prefixed), host, project, nextSib)
	t, err := Scan(row)
	sqlh.PanicIfIsntNoRows(err)
	return t
}

func Scan(r isql.Row) (*task.Task, error) {
	t := &task.Task{}
	err := r.Scan(
		&t.ID,
		&t.Parent,
		&t.FirstChild,
		&t.NextSib,
		&t.User,
		&t.Name,
		&t.Description,
		&t.CreatedBy,
		&t.CreatedOn,
		&t.TimeEst,
		&t.TimeInc,
		&t.TimeSubMin,
		&t.TimeSubEst,
		&t.TimeSubInc,
		&t.CostEst,
		&t.CostInc,
		&t.CostSubEst,
		&t.CostSubInc,
		&t.FileN,
		&t.FileSize,
		&t.FileSubN,
		&t.FileSubSize,
		&t.ChildN,
		&t.DescN,
		&t.IsParallel)
	if t.ID.IsZero() {
		t = nil
	}
	return t, err
}

var (
	Sql_task_columns_prefixed = `t.id, t.parent, t.firstChild, t.nextSib, t.user, t.name, t.description, t.createdBy, t.createdOn, t.timeEst, t.timeInc, t.timeSubMin, t.timeSubEst, t.timeSubInc, t.costEst, t.costInc, t.costSubEst, t.costSubInc, t.fileN, t.fileSize, t.fileSubN, t.fileSubSize, t.childN, t.descN, t.isParallel`
	sql_task_columns          = strings.ReplaceAll(Sql_task_columns_prefixed, `t.`, ``)
	sql_ancestors_cte         = `WITH RECURSIVE ancestors (n, id) AS (SELECT 0, parent FROM tasks WHERE host=? AND project=? AND id=? UNION SELECT a.n + 1, t.parent FROM tasks t, ancestors a WHERE t.host=? AND t.project=? AND t.id = a.id) CYCLE id RESTRICT`
)
