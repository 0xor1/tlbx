package testutil

import (
	"bytes"
	"io/ioutil"
	"net/http"

	. "github.com/0xor1/tlbx/pkg/core"
	"github.com/0xor1/tlbx/pkg/web/app"
	"github.com/0xor1/tlbx/pkg/web/app/test"
	"github.com/0xor1/trees/pkg/file"
	"github.com/0xor1/trees/pkg/task"
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
				if a.NextSibling != nil {
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
		if t.NextSibling != nil {
			print(tree[t.NextSibling.String()], as)
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

func MustUploadFile(client *app.Client, host, project, task ID, name, mimeType string, content []byte) *file.File {
	ppur := (&file.GetPresignedPutUrl{
		Host:     host,
		Project:  project,
		Task:     task,
		Size:     uint64(len(content)),
		Name:     name,
		MimeType: mimeType,
	}).MustDo(client)
	req, err := http.NewRequest(http.MethodPut, ppur.URL, bytes.NewBuffer(content))
	PanicOn(err)
	req.Header.Add("X-Amz-Acl", "private")
	req.Header.Add("Content-Length", Strf(`%d`, len(content)))
	req.Header.Add("Content-Type", mimeType)
	req.Header.Add("Content-Disposition", Strf("attachment; filename=%s", name))
	req.Header.Add("Host", req.Host)
	resp, err := http.DefaultClient.Do(req)
	PanicOn(err)
	PanicOn(resp.Body.Close())
	PanicIf(resp.StatusCode != 200, "resp.StatusCode - %d", resp.StatusCode)
	f := (&file.Finalize{
		Host:    host,
		Project: project,
		Task:    task,
		ID:      ppur.ID,
	}).MustDo(client)
	PanicIf(!f.ID.Equal(ppur.ID), "file id unexpected")
	return f
}

func MustDownloadFile(client *app.Client, host, project, task, id ID) []byte {
	pgur := (&file.GetPresignedGetUrl{
		Host:       host,
		Project:    project,
		Task:       task,
		ID:         id,
		IsDownload: true,
	}).MustDo(client)
	req, err := http.NewRequest(http.MethodGet, pgur, nil)
	PanicOn(err)
	req.Header.Add("Host", req.Host)
	resp, err := http.DefaultClient.Do(req)
	PanicOn(err)
	defer resp.Body.Close()
	PanicIf(resp.StatusCode != 200, "resp.StatusCode - %d", resp.StatusCode)
	bs, err := ioutil.ReadAll(resp.Body)
	PanicOn(err)
	return bs
}
