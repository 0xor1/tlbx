<template>
  <div class="root">
    <div class="loading" v-if="loading">
      loading...
    </div>
    <div class="content" v-else>
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
            <td v-bind:class="c.class" v-for="(c, index) in cols" :key="index">
              {{c.name == "user"? "" : c.get(task)}}
              <user v-if="c.name=='user'" :userId="c.get(task)"></user>
            </td>
          </tr>
          <tr class="row" v-for="(t, index) in children" :key="index">
            <td v-bind:class="c.class" v-for="(c, index) in cols" :key="index">
              {{c.name == "user"? "" : c.get(t)}}
              <user v-if="c.name=='user'" :userId="c.get(t)"></user>
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
  export default {
    name: 'tasks',
    components: {user},
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
          updating: false,
          loading: true,
          me: null,
          project: null,
          pMe: null,
          ancestors: [],
          moreAncestors: false,
          loadingMoreAncestors: false,
          task: null,
          estTime: "0m",
          estExp: "0.00",
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
        text-align: center;
        min-width: 100px;
        &.name{
          min-width: 18pc;
        }
      }
      tr.this-task {
        font-weight: 1000;
      }
    }
  }
}
</style>