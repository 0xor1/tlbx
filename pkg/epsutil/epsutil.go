package epsutil

import (
	"net/http"
	"time"

	"github.com/0xor1/tlbx/cmd/trees/pkg/cnsts"
	. "github.com/0xor1/tlbx/pkg/core"
	"github.com/0xor1/tlbx/pkg/isql"
	"github.com/0xor1/tlbx/pkg/json"
	"github.com/0xor1/tlbx/pkg/web/app"
	"github.com/0xor1/tlbx/pkg/web/app/service"
	"github.com/0xor1/tlbx/pkg/web/app/service/sql"
	"github.com/0xor1/tlbx/pkg/web/app/session/me"
	sqlh "github.com/0xor1/tlbx/pkg/web/app/sql"
)

func SetAncestralChainAggregateValuesFromTask(tx sql.Tx, host, project, task ID) IDs {
	return setAncestralChainAggregateValuesFrom(tx, host, project, task, false)
}

func SetAncestralChainAggregateValuesFromParentOfTask(tx sql.Tx, host, project, task ID) IDs {
	return setAncestralChainAggregateValuesFrom(tx, host, project, task, true)
}

func setAncestralChainAggregateValuesFrom(tx sql.Tx, host, project, task ID, parentOfTask bool) IDs {
	var qry string
	qryArgs := make([]interface{}, 0, 5)
	qryArgs = append(qryArgs, host, project)
	if !parentOfTask {
		qry = `?`
		qryArgs = append(qryArgs, task)
	} else {
		qry = `(SELECT parent FROM tasks WHERE host=? AND project=? AND id=?)`
		qryArgs = append(qryArgs, host, project, task)
	}
	ancestorChain := make(IDs, 0, 20)
	PanicOn(tx.Query(func(rows isql.Rows) {
		for rows.Next() {
			i := ID{}
			PanicOn(rows.Scan(&i))
			ancestorChain = append(ancestorChain, i)
		}
	}, Strf(`CALL setAncestralChainAggregateValuesFromTask(?, ?, %s)`, qry), qryArgs...))
	return ancestorChain
}

func MustGetRole(tlbx app.Tlbx, tx sql.Tx, host, project ID, user ID) cnsts.Role {
	if host.Equal(user) {
		return cnsts.RoleAdmin
	}
	var role cnsts.Role
	row := tx.QueryRow(`SELECT role FROM users WHERE host=? AND project=? AND id=? AND isActive=1`, host, project, user)
	err := row.Scan(&role)
	app.ReturnIf(sqlh.IsNoRows(err), http.StatusForbidden, "")
	PanicOn(err)
	return role
}

func MustHaveAccess(tlbx app.Tlbx, tx sql.Tx, host, project ID, user *ID, role cnsts.Role) {
	userExist := user != nil
	if userExist && user.Equal(host) {
		TaskMustExist(tx, host, project, project)
		return
	}

	if !userExist || role == cnsts.RoleReader {
		row := tx.QueryRow(`SELECT isPublic FROM projects WHERE host=? AND id=?`, host, project)
		isPublic := false
		sqlh.PanicIfIsntNoRows(row.Scan(&isPublic))
		if isPublic {
			return
		}
		// at this point project isnt public so if no active session return 403
		app.ReturnIf(!userExist, http.StatusForbidden, "")
	}

	row := tx.QueryRow(`SELECT 1 FROM users WHERE host=? AND project=? AND id=? AND role<=? AND isActive=1`, host, project, *user, role)
	hasAccess := false
	sqlh.PanicIfIsntNoRows(row.Scan(&hasAccess))
	app.ReturnIf(!hasAccess, http.StatusForbidden, "")
}

func IMustHaveAccess(tlbx app.Tlbx, tx sql.Tx, host, project ID, role cnsts.Role) {
	iExist := me.Exists(tlbx)
	var mePtr *ID
	if iExist {
		meID := me.Get(tlbx)
		mePtr = &meID
	}
	MustHaveAccess(tlbx, tx, host, project, mePtr, role)
}

func MustLockProject(tx sql.Tx, host, id ID) {
	projectExists := false
	row := tx.QueryRow(`SELECT COUNT(*)=1 FROM projectLocks WHERE host=? AND id=? FOR UPDATE`, host, id)
	sqlh.PanicIfIsntNoRows(row.Scan(&projectExists))
	app.ReturnIf(!projectExists, http.StatusNotFound, "no such project")
}

func TaskMustExist(tx sql.Tx, host, project, id ID) {
	taskExists := false
	row := tx.QueryRow(`SELECT COUNT(*)=1 FROM tasks WHERE host=? AND project=? AND id=?`, host, project, id)
	sqlh.PanicIfIsntNoRows(row.Scan(&taskExists))
	app.ReturnIf(!taskExists, http.StatusNotFound, "task not found")
}

func StorePrefix(host ID, projectAndOrTask ...ID) string {
	prefix := host.String()
	if len(projectAndOrTask) > 0 {
		prefix += "/" + projectAndOrTask[0].String()
		if len(projectAndOrTask) > 1 {
			prefix += "/" + projectAndOrTask[1].String()
			PanicIf(len(projectAndOrTask) > 2, "invalid file address")
		}
	}
	return prefix
}

func LogActivity(tlbx app.Tlbx, tx sql.Tx, host, project, task, item ID, itemType cnsts.Type, action cnsts.Action, itemName *string, extraInfo interface{}, fcmExtraInfo interface{}) {
	PanicIf(itemType == cnsts.TypeTask && !task.Equal(item), "item type is task but item and task ids are different")
	me := me.Get(tlbx)
	var ei *string
	var eiStr string
	if extraInfo != nil {
		eiStr = string(json.MustMarshal(extraInfo))
		PanicIf(StrLen(eiStr) > 10000, "extraInfo string is too long")
		ei = &eiStr

	}
	taskDeleted := itemType == cnsts.TypeTask && action == cnsts.ActionDeleted
	itemDeleted := action == cnsts.ActionDeleted
	var nameQry string
	qryArgs := make([]interface{}, 0, 14)
	occurredOn := tlbx.Start()
	qryArgs = append(qryArgs, host, project, task, occurredOn, me, item, itemType, taskDeleted, itemDeleted, action)
	if itemType == cnsts.TypeTask {
		nameQry = `?`
		qryArgs = append(qryArgs, itemName, nil, ei)
	} else {
		nameQry = `(SELECT name FROM tasks WHERE host=? AND project=? AND id=?)`
		qryArgs = append(qryArgs, host, project, task, itemName, ei)
	}
	_, err := tx.Exec(Strf(`INSERT INTO activities(host, project, task, occurredOn, user, item, itemType, taskDeleted, itemDeleted, action, taskName, itemName, extraInfo) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, %s, ?, ?)`, nameQry), qryArgs...)
	PanicOn(err)
	if itemDeleted {
		// if this is deleting an item we need to update all prev activities on this item to have itemDeleted
		if itemType == cnsts.TypeTask {
			_, err = tx.Exec(`UPDATE activities SET taskDeleted=1, itemDeleted=1 WHERE host=? AND project=? AND task=?`, host, project, item)
		} else {
			_, err = tx.Exec(`UPDATE activities SET itemDeleted=1 WHERE host=? AND project=? AND item=?`, host, project, item)
		}
		PanicOn(err)
	}
	// ***************************************
	// start sendind fcm notifications section
	// ***************************************
	if fcmExtraInfo != nil {
		// if fcmExtraInfo is passed use that instead of activity log extraInfo
		eiStr = string(json.MustMarshal(fcmExtraInfo))
	}
	row := tx.QueryRow(`SELECT handle FROM users WHERE host=? AND project=? AND id=?`, host, project, me)
	handle := ""
	PanicOn(row.Scan(&handle))
	d := map[string]string{
		"host":       host.String(),
		"project":    project.String(),
		"task":       task.String(),
		"item":       item.String(),
		"user":       me.String(),
		"userHandle": handle,
		"type":       string(itemType),
		"occurredOn": occurredOn.Format(time.RFC3339),
		"action":     string(action),
		"extraInfo":  eiStr,
	}
	if itemName != nil {
		d["itemName"] = *itemName
	}
	srv := service.Get(tlbx)
	srv.FCM().AsyncSend(IDs{host, project}, map[string]string{}, 0)
}

func ActivityItemRename(tx sql.Tx, host, project, item ID, newItemName string, isTask bool) {
	var qry string
	// keep all projectActivity entries itemName values up to date
	if isTask {
		qry = `UPDATE activities SET taskName=? WHERE host=? AND project=? AND task=?`
	} else {
		qry = `UPDATE activities SET itemName=? WHERE host=? AND project=? AND item=?`
	}
	_, err := tx.Exec(qry, newItemName, host, project, item)
	PanicOn(err)
}
