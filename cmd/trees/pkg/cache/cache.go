package cache

import (
	"sort"

	. "github.com/0xor1/tlbx/pkg/core"
	"github.com/0xor1/tlbx/pkg/web/app"
)

// leave this code here for now, to be used in the future, it's slowing down
// development of the app and isn't necessary, just nice to have.

type tlbxKey struct{}

func SetupMware() func(app.Tlbx) {
	return func(tlbx app.Tlbx) {

		//tlbx.Set(tlbxKey{}, s)
	}
}

func CleanupMware() func(app.Tlbx) {
	return func(tlbx app.Tlbx) {
		tlbx.Get(tlbxKey{})
	}
}

type CKey struct {
	isGet   bool
	Key     string
	Args    []interface{}
	DlmKeys map[string]bool
}

func NewGet(key string, args ...interface{}) *CKey {
	return &CKey{
		isGet:   true,
		Key:     key,
		Args:    args,
		DlmKeys: map[string]bool{},
	}
}

func NewSetDlms() *CKey {
	return &CKey{
		isGet:   false,
		DlmKeys: map[string]bool{},
	}
}

func (k *CKey) SortedDlmKeys() []string {
	sorted := make([]string, 0, len(k.DlmKeys))
	for key := range k.DlmKeys {
		idx := sort.SearchStrings(sorted, key)
		if idx == len(sorted) {
			sorted = append(sorted, key)
		} else if sorted[idx] != key {
			sorted = append(sorted, "")
			copy(sorted[idx+1:], sorted[idx:])
			sorted[idx] = key
		}
	}
	return sorted
}

func (k *CKey) HostMaster(host ID) *CKey {
	return k.setKey("h_m", host)
}

func (k *CKey) Host(host ID) *CKey {
	if k.isGet {
		k.HostMaster(host)
	}
	return k.setKey("h", host)
}

func (k *CKey) HostProjectsSet(host ID) *CKey {
	if k.isGet {
		k.HostMaster(host)
	}
	return k.setKey("hps", host)
}

func (k *CKey) ProjectMaster(host, project ID) *CKey {
	if k.isGet {
		k.HostMaster(host)
	}
	return k.setKey("p_m", project)
}

func (k *CKey) Project(host, project ID) *CKey {
	if k.isGet {
		k.ProjectMaster(host, project)
	} else {
		k.HostProjectsSet(host)
	}
	return k.setKey("p", project).setKey("t", project) //projects are also tasks
}

func (k *CKey) ProjectActivities(host, project ID) *CKey {
	if k.isGet {
		k.ProjectMaster(host, project)
	}
	return k.setKey("pa", project)
}

func (k *CKey) ProjectUsersSet(host, project ID) *CKey {
	if k.isGet {
		k.ProjectMaster(host, project)
	}
	return k.setKey("pus", project)
}

func (k *CKey) ProjectUser(host, project, user ID) *CKey {
	if k.isGet {
		k.ProjectMaster(host, project)
	} else {
		k.ProjectUsersSet(host, project)
	}
	return k.setKey("pu", user)
}

func (k *CKey) ProjectUsers(host, project ID, users []ID) *CKey {
	for _, member := range users {
		k.ProjectUser(host, project, member)
	}
	return k
}

func (k *CKey) TaskChildrenSet(host, project, parent ID) *CKey {
	if k.isGet {
		k.ProjectMaster(host, project)
	}
	return k.setKey("tcs", parent)
}

func (k *CKey) Task(host, project, task ID) *CKey {
	if project.Equal(task) {
		k.Project(host, project) //let project handle project nodes
	} else {
		if k.isGet {
			k.ProjectMaster(host, project)
		}
		k.setKey("t", task)
	}
	return k
}

func (k *CKey) CombinedTaskAndTaskChildrenSets(host, project ID, tasks []ID) *CKey {
	for _, task := range tasks {
		k.Task(host, project, task)
		k.setKey("tcs", task)
	}
	return k
}

func (k *CKey) ProjectTimeLogSet(host, project ID) *CKey {
	if k.isGet {
		k.ProjectMaster(host, project)
	}
	k.setKey("ptls", project)
	return k
}

func (k *CKey) ProjectUserTimeLogSet(host, project, user ID) *CKey {
	if k.isGet {
		k.ProjectMaster(host, project)
	} else {
		k.ProjectTimeLogSet(host, project)
	}
	k.setKey("putls", user)
	return k
}

func (k *CKey) TaskTimeLogSet(host, project, task ID, user *ID) *CKey {
	if k.isGet {
		k.ProjectMaster(host, project)
		if user != nil {
			k.ProjectUserTimeLogSet(host, project, *user)
		}
	} else {
		PanicIf(user == nil, "missing member in taskTimeLogSet dlm")
		k.ProjectUserTimeLogSet(host, project, *user)
	}
	k.setKey("ttls", task)
	return k
}

func (k *CKey) TimeLog(host, project, timeLog ID, task, user *ID) *CKey {
	if k.isGet {
		k.ProjectMaster(host, project)
		if task != nil {
			k.TaskTimeLogSet(host, project, *task, user)
		}
	} else {
		PanicIf(task == nil, "missing task in taskTimeLog dlm")
		k.TaskTimeLogSet(host, project, *task, user)
	}
	k.setKey("tl", timeLog)
	return k
}

func (k *CKey) setKey(typeKey string, id ...ID) *CKey {
	var key string
	switch {
	case len(id) == 0:
		key = typeKey
	case len(id) == 1:
		key = typeKey + ":" + id[0].String()
	default:
		PanicOn("invalid number of ids passed")
	}
	k.DlmKeys[key] = true
	return k
}
