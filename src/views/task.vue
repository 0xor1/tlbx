<template>
  <div class="root">
    <div class="loading" v-if="loading">
      loading...
    </div>
    <div v-else-if="showCreate || showUpdate">
      <task-create-or-update 
        :isCreate="showCreate" 
        :taskId="updateId"
        :hostId="$u.rtr.host()" 
        :projectId="$u.rtr.project()"
        @close="showCreate = showUpdate = false">
      </task-create-or-update>
    </div>
    <div v-else class="content" >
      <div class="breadcrumb">
        <span v-if="ancestors.length > 0 && ancestors[0].parent != null">
          <a title="load more ancestors" :click="getMoreAncestors">..</a>
          /
        </span>
        <span v-for="a in ancestors" :key="a.id">
          <a :title="a.name" :href="'/#/host/'+$u.rtr.host()+'/project/'+$u.rtr.project()+'/task/'+a.id">{{$u.fmt.ellipsis(a.name, 20)}}</a>      
        </span>
      </div>
      <div class="summary">
        <table>
          <tr class="header">
            <th :colspan="s.cols.length" :rowspan="s.cols.length == 1? 2: 1" :class="s.name" v-for="(s, index) in sections" :key="index">
              {{s.name}}
            </th>
          </tr>
          <tr class="header">
            <th :class="c.name" v-for="(c, index) in colHeaders" :key="index">
              {{c.name}}
            </th>
          </tr>
          <tr class="row this-task">
            <td :title="task.description" v-bind:class="c.class" v-for="(c, index) in cols" :key="index">
              {{c.name == "user"? "" : c.get(task)}}
              <user v-if="c.name=='user'" :userId="c.get(task)"></user>
            </td>
            <td class="action">
              
            </td>
            <td v-if="$u.perm.canWrite(pMe)" class="action" @click.stop="updateId = task.id; showCreate = true" title="insert first child">
              <img src="@/assets/insert-below.svg">
            </td>
            <td v-if="canEdit(task)" class="action" @click.stop="updateId = task.id; showUpdate = true" title="update">
              <img src="@/assets/edit.svg">
            </td>
            <td v-if="canDelete(task)" class="action" @click.stop="trash(task, -1)" title="delete">
              <img src="@/assets/trash.svg">
            </td>
          </tr>
          <tr class="row" v-for="(t, index) in children" :key="index" @click.stop.prevent="$u.rtr.goto(`/host/${$u.rtr.host()}/project/${$u.rtr.project()}/task/${t.id}`)">
            <td :title="t.description" v-bind:class="c.class" v-for="(c, index) in cols" :key="index">
              {{c.name == "user"? "" : c.get(t)}}
              <user v-if="c.name=='user'" :userId="c.get(t)"></user>
            </td>
            <td v-if="$u.perm.canWrite(pMe)" class="action" @click.stop="updateId = task.id; showCreate = true" title="insert above">
              <img src="@/assets/insert-above.svg">
            </td>
            <td v-if="$u.perm.canWrite(pMe)" class="action" @click.stop="updateId = task.id; showCreate = true" title="insert below">
              <img src="@/assets/insert-below.svg">
            </td>
            <td v-if="canEdit(t)" class="action" @click.stop="updateId = t.id; showUpdate = true" title="update">
              <img src="@/assets/edit.svg">
            </td>
            <td v-if="canDelete(t)" class="action" @click.stop="trash(t, index)" title="delete">
              <img src="@/assets/trash.svg">
            </td>
          </tr>
        </table>
        <button v-if="moreChildren" @click="loadMoreChildren()">load more</button>
      </div>
      <div class="subs">
      </div>
    </div>
  </div>
</template>

<script>
  import user from '../components/user'
  import taskCreateOrUpdate from '../components/taskCreateOrUpdate'
  export default {
    name: 'tasks',
    components: {user, taskCreateOrUpdate},
    data: function() {
      return this.initState()
    },
    computed: {
      sections(){
        return this.commonSections.filter(i => i.show())
      },
      colHeaders(){
        let res = []
        this.sections.forEach((section)=>{
          if (section.cols.length > 1) {
            res = res.concat(section.cols)
          }
        })
        return res
      },
      cols(){
        let res = []
        this.sections.forEach((section)=>{
          res = res.concat(section.cols)
        })
        return res
      }
    },
    methods: {
      initState(){
        return {
          showCreate: false,
          showUpdate: false,
          updating: false,
          loading: true,
          me: null,
          project: null,
          pMe: null,
          ancestors: [],
          moreAncestors: false,
          loadingMoreAncestors: false,
          task: null,
          children: [],
          moreChildren: false,
          loadingMoreChildren: false,
          times: [],
          moreTimes: false,
          expenses: [],
          moreExpenses: false,
          files: [],
          moreFiles: false,
          comments: [],
          moreComments: false,
          commonSections: [
            {
              name: "name",
              show: () => true,
              cols: [
                {
                  name: "name",
                  get: (t)=> t.name
                }
              ]
            },
            {
              name: "created",
              show: () => this.$root.show.date,
              cols: [
                {
                  name: "created",
                  get: (t)=> this.$u.fmt.date(t.createdOn)
                }
              ]
            },
            {
              name: "user",
              show: () => this.$root.show.user,
              cols: [
                {
                  name: "user",
                  get: (t)=> t.user
                }
              ]
            },
            {
              name: "time",
              show: () => this.$root.show.time,
              cols: [
                { 
                  name: "min",
                  get: (t)=> this.$u.fmt.duration(t.timeEst + t.timeSubMin, this.project.hoursPerDay, this.project.daysPerWeek)                  
                },
                {
                  name: "est",
                  get: (t)=> this.$u.fmt.duration(t.timeEst + t.timeSubEst, this.project.hoursPerDay, this.project.daysPerWeek)
                },
                {
                  name: "inc",
                  get: (t)=> this.$u.fmt.duration(t.timeInc + t.timeSubInc, this.project.hoursPerDay, this.project.daysPerWeek)
                }
              ]
            },
            {
              name: "cost",
              show: () => this.$root.show.cost,
              cols: [
                {
                  name: "est",
                  get: (t)=> this.$u.fmt.cost(this.project.currencyCode, t.costEst + t.costSubEst)
                },
                {
                  name: "inc",
                  get: (t)=> this.$u.fmt.cost(this.project.currencyCode, t.costInc + t.costSubInc)
                }
              ]
            },
            {
              name: "file",
              show: () => this.$root.show.file,
              cols: [
                {
                  name: "n",
                  get: (t)=> t.fileN + t.fileSubN
                },
                {
                  name: "size",
                  get: (t)=> this.$u.fmt.bytes(t.fileSize + t.fileSubSize)
                }
              ]
            },
            {
              name: "task",
              show: () => this.$root.show.task,
              cols: [
                {
                  name: "childn",
                  get: (p)=> p.childN
                },
                {
                  name: "descn",
                  get: (p)=> p.descN
                }
              ]
            }
          ]
        }
      },
      init(){
        for(const [key, value] of Object.entries(this.initState())) {
          this[key] = value
        }
        this.$api.user.me().then((me)=>{
          this.me = me
        }).finally(()=>{
          this.$root.ctx().then((ctx)=>{
            this.pMe = ctx.pMe
            this.project = ctx.project
            let mapi = this.$api.newMDoApi()
            mapi.task.getAncestors(this.$u.rtr.host(), this.$u.rtr.project(), this.$u.rtr.task(), 10).then((res)=>{
              this.ancestors = res.set.reverse()
              this.moreAncestors = res.more
            })
            mapi.task.get(this.$u.rtr.host(), this.$u.rtr.project(), this.$u.rtr.task()).then((t)=>{
              this.task = t
              this.estTime = this.$u.fmt.duration(t.estimatedTime, this.project.hoursPerDay, this.project.daysPerWeek)
              this.estExp = this.$u.fmt.cost(this.project.currencyCode, t.estimatedExpense)
            })
            mapi.task.getChildren(this.$u.rtr.host(), this.$u.rtr.project(), this.$u.rtr.task()).then((res)=>{
              this.children = res.set
              this.moreChildren = res.more
            })
            mapi.time.get(this.$u.rtr.host(), this.$u.rtr.project(), this.$u.rtr.task()).then((res)=>{
              this.times = res.set
              this.moreTimes = res.more
            })
            mapi.expense.get(this.$u.rtr.host(), this.$u.rtr.project(), this.$u.rtr.task()).then((res)=>{
              this.expenses = res.set
              this.moreExpenses = res.more
            })
            mapi.file.get(this.$u.rtr.host(), this.$u.rtr.project(), this.$u.rtr.task()).then((res)=>{
              this.file = res.set
              this.moreFiles = res.more
            })
            mapi.comment.get(this.$u.rtr.host(), this.$u.rtr.project(), this.$u.rtr.task()).then((res)=>{
              this.comments = res.set
              this.moreComments = res.more
            })
            mapi.sendMDo().finally(()=>{
              this.loading = false
            })
          })
        })
      },
      getMoreAncestors(){
        if (!this.loadingMoreAncestors) {
          this.loadingMoreAncestors = true;
          let taskId = this.$u.rtr.task()
          if (this.ancestors.length > 0 && this.ancestors[0].id != null) {
            taskId = this.ancestors[0].id
          }
          this.$api.task.getAncestors(this.$u.rtr.host(), this.$u.rtr.project(), taskId, 10).then((res)=>{
            this.ancestors = res.set.reverse().concat(this.ancestors)
            this.moreAncestors = res.more
          }).finally(()=>{
            this.loadingMoreAncestors = false
          })
        }
      },
      getMoreChildren(){
        if (!this.loadingMoreChildren) {
          this.loadingMoreChildren = true;
          this.$api.task.getChildren(this.$u.rtr.host(), this.$u.rtr.project(), this.$u.rtr.task(), this.children[this.children.length - 1].id, 10).then((res)=>{
            this.children = this.children.concat(res.set)
            this.moreAncestors = res.more
          }).finally(()=>{
            this.loadingMoreChildren = false
          })
        }
      },
      canEdit(t){
        if (this.$u.perm.canWrite(this.pMe)) {
          if ((t.parent == null && this.$u.rtr.host() == this.pMe.id) || 
            (t.parent != null && this.$u.perm.canAdmin(this.pMe))) {
            return true
          }
        }
        return false
      },
      canDelete(t){
        if (this.$u.perm.canWrite(this.pMe)) {
          if ((t.parent == null && this.$u.rtr.host() == this.pMe.id) || 
            (t.parent != null && this.$u.perm.canAdmin(this.pMe))) {
            return true
          }
        }
        return false
      },
      trash(t, index){
        this.$api.task.delete(this.$u.rtr.host(), this.$u.rtr.project(), t.id).then(()=>{
          if (index > -1) {
            this.children.splice(index, 1)
          }
          if (t.parent == null) {
            this.$u.rtr.goHome()
          } else {
            this.$u.rtr.goto(`/host/${this.$u.rtr.host()}/project/${this.$u.rtr.project()}/task/${t.parent}`)
          }
        })
      }
    },
    mounted(){
      this.init()
    },
    watch: {
      $route () {
        this.init()
      }
    }
  }
</script>

<style lang="scss">
div.root {
  padding: 2.6pc 0 0 1.3pc;
  > .content{
    > .breadcrumb {
      white-space: nowrap;
      overflow-y: auto;
    }
    table {
      margin: 1pc 0 1pc 0;
      border-collapse: collapse;
      th, td {
        &:not(.action) {
          text-align: center;
          min-width: 100px;
          &.name{
            min-width: 18pc;
          }
        }
      }
      tr.this-task {
        cursor: default;
        .action{
          cursor: pointer;
        }
        font-size: 1.5pc;
        font-weight: bold;
      }
    }
  }
}
</style>