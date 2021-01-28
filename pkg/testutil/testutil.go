package testutil

import (
	"github.com/0xor1/tlbx/cmd/trees/pkg/task"
	. "github.com/0xor1/tlbx/pkg/core"
	"github.com/0xor1/tlbx/pkg/web/app/test"
	"github.com/0xor1/trees/pkg/task/taskeps"
)

// only suitable for small test trees for visual validation
// whilst writing/debugging unit tests
func GrabFullTree(r test.Rig, host, project ID) map[string]*task.Task {
	tree := map[string]*task.Task{}
	rows, err := r.Data().Primary().Query(Strf(`SELECT %s FROM tasks t WHERE t.host=? AND t.Project=?`, taskeps.Sql_task_columns_prefixed), host, project)
	if rows != nil {
		defer rows.Close()
	}
	PanicOn(err)
	for rows.Next() {
		t, err := taskeps.Scan(rows)
		PanicOn(err)
		tree[t.ID.String()] = t
	}
	return tree
}

func PrintFullTree(project ID, tree map[string]*task.Task) {
	var print func(t *task.Task, as []*task.Task)
	print = func(t *task.Task, as []*task.Task) {
		p := 0
		if t.IsParallel {
			p = 1
		}
		v := Strf(`[n: %s, p: %d, te: %d, ti: %d, tsm: %d, tse: %d, tsi: %d, ce: %d, ci: %d, cse: %d, csi: %d, fn: %d, fsn: %d, fs: %d, fss: %d]`, t.Name, p, t.TimeEst, t.TimeInc, t.TimeSubMin, t.TimeSubEst, t.TimeSubInc, t.CostEst, t.CostInc, t.CostSubEst, t.CostSubInc, t.FileN, t.FileSubN, t.FileSize, t.FileSubSize)
		if len(as) > 0 {
			pre := ``
			for _, a := range as[1:] {
				if a.NextSib != nil {
					pre += `|    `
				} else {
					pre += `     `
				}
			}
			Println(Strf(`%s|`, pre))
			Println(Strf(`%s|`, pre))
			Println(Strf(`%s|____%s`, pre, v))
		} else {
			Println(v)
		}
		if t.FirstChild != nil {
			print(tree[t.FirstChild.String()], append(as, t))
		}
		if t.NextSib != nil {
			print(tree[t.NextSib.String()], as)
		}
	}
	println("n: name")
	println("p: isParallel")
	println("te: timeEst")
	println("ti: timeInc")
	println("tsm: timeSubMin")
	println("tse: timeSubEst")
	println("tsi: timeSubInc")
	println("ce: costEst")
	println("ci: costInc")
	println("cse: costSubEst")
	println("csi: costSubInc")
	println("fn: fileN")
	println("fsn: fileSubN")
	println("fs: fileSize")
	println("fss: fileSubSize")
	println()
	print(tree[project.String()], nil)
}
