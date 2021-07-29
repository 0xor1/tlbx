package testutil

import (
	"github.com/0xor1/tlbx/cmd/trees/pkg/task"
	. "github.com/0xor1/tlbx/pkg/core"
)

// only suitable for small test trees for visual validation
// whilst writing/debugging unit tests

func PrintFullTree(root ID, tree task.GetTreeRes) {
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
			print(tree[*t.FirstChild], append(as, t))
		}
		if t.NextSib != nil {
			print(tree[*t.NextSib], as)
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
	print(tree[root], nil)
}
