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

func CleanupMware(authKey64s, encrKey32s [][]byte, isLocal bool) func(app.Tlbx) {
	return func(tlbx app.Tlbx) {
		tlbx.Get(tlbxKey{})
	}
}

type Key struct {
	isGet   bool
	Key     string
	Args    []interface{}
	DlmKeys map[string]bool
}

func NewGet(key string, args ...interface{}) *Key {
	return &Key{
		isGet:   true,
		Key:     key,
		Args:    args,
		DlmKeys: map[string]bool{},
	}
}

func NewSetDlms() *Key {
	return &Key{
		isGet:   false,
		DlmKeys: map[string]bool{},
	}
}

func (k *Key) SortedDlmKeys() []string {
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

func (k *Key) HostMaster(host ID) *Key {
	return k.setKey("h_m", host)
}

func (k *Key) Host(host ID) *Key {
	if k.isGet {
		k.HostMaster(host)
	}
	return k.setKey("h", host)
}

func (k *Key) HostProjectsSet(host ID) *Key {
	if k.isGet {
		k.HostMaster(host)
	}
	return k.setKey("hps", host)
}

func (k *Key) ProjectMaster(host, project ID) *Key {
	if k.isGet {
		k.HostMaster(host)
	}
	return k.setKey("p_m", project)
}

func (k *Key) Project(host, project ID) *Key {
	if k.isGet {
		k.ProjectMaster(host, project)
	} else {
		k.HostProjectsSet(host)
	}
	return k.setKey("p", project).setKey("t", project) //projects are also tasks
}

func (k *Key) ProjectActivities(host, project ID) *Key {
	if k.isGet {
		k.ProjectMaster(host, project)
	}
	return k.setKey("pa", project)
}

func (k *Key) ProjectUsersSet(host, project ID) *Key {
	if k.isGet {
		k.ProjectMaster(host, project)
	}
	return k.setKey("pus", project)
}

func (k *Key) ProjectUser(host, project, user ID) *Key {
	if k.isGet {
		k.ProjectMaster(host, project)
	} else {
		k.ProjectUsersSet(host, project)
	}
	return k.setKey("pu", user)
}

func (k *Key) ProjectUsers(host, project ID, users []ID) *Key {
	for _, member := range users {
		k.ProjectUser(host, project, member)
	}
	return k
}

func (k *Key) TaskChildrenSet(host, project, parent ID) *Key {
	if k.isGet {
		k.ProjectMaster(host, project)
	}
	return k.setKey("tcs", parent)
}

func (k *Key) Task(host, project, task ID) *Key {
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

func (k *Key) CombinedTaskAndTaskChildrenSets(host, project ID, tasks []ID) *Key {
	for _, task := range tasks {
		k.Task(host, project, task)
		k.setKey("tcs", task)
	}
	return k
}

func (k *Key) ProjectTimeLogSet(host, project ID) *Key {
	if k.isGet {
		k.ProjectMaster(host, project)
	}
	k.setKey("ptls", project)
	return k
}

func (k *Key) ProjectUserTimeLogSet(host, project, user ID) *Key {
	if k.isGet {
		k.ProjectMaster(host, project)
	} else {
		k.ProjectTimeLogSet(host, project)
	}
	k.setKey("putls", user)
	return k
}

func (k *Key) TaskTimeLogSet(host, project, task ID, user *ID) *Key {
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

func (k *Key) TimeLog(host, project, timeLog ID, task, user *ID) *Key {
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

func (k *Key) setKey(typeKey string, id ...ID) *Key {
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
