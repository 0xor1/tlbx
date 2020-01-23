package common

import (
	. "github.com/0xor1/wtf/pkg/core"
	"github.com/0xor1/wtf/pkg/iredis"
	"github.com/0xor1/wtf/pkg/isql"
	"github.com/0xor1/wtf/pkg/web/app"
)

type tlbxKey struct{}

type ReplicaSet struct {
	Primary string
	Slaves  []string
}

type PersistLayer interface {
	Cache() iredis.Pool
	User() isql.ReplicaSet
	Pwd() isql.ReplicaSet
	Data() isql.ReplicaSet
}

func Mware(user, pwd, data ReplicaSet, redisCache string) func(app.Toolbox) {
	cache := iredis.CreatePool(redisCache)
	userDB, err := isql.NewReplicaSet(user.Primary, user.Slaves...)
	PanicOn(err)
	pwdDB, err := isql.NewReplicaSet(pwd.Primary, pwd.Slaves...)
	PanicOn(err)
	dataDB, err := isql.NewReplicaSet(data.Primary, data.Slaves...)
	PanicOn(err)
	return func(tlbx app.Toolbox) {
		tlbx.Set(tlbxKey{}, &persist{
			cache: cache,
			user:  userDB,
			pwd:   pwdDB,
			data:  dataDB,
		})
	}
}

func Persist(tlbx app.Toolbox) PersistLayer {
	return tlbx.Get(tlbxKey{}).(PersistLayer)
}

type persist struct {
	cache iredis.Pool
	user  isql.ReplicaSet
	pwd   isql.ReplicaSet
	data  isql.ReplicaSet
}

func (d *persist) Cache() iredis.Pool {
	return d.cache
}

func (d *persist) User() isql.ReplicaSet {
	return d.user
}

func (d *persist) Pwd() isql.ReplicaSet {
	return d.pwd
}

func (d *persist) Data() isql.ReplicaSet {
	return d.data
}
