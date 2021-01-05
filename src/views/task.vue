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
      <div class="task">
        <h1>{{task.name}}</h1>
        <p v-if="task.description != null">{{task.description}}</p>
        <p>parallel: <input :disabled="!$u.perm.canWrite(pMe)" type="checkbox" v-model="task.isParallel"></p>
        <p>est time: <input :disabled="!$u.perm.canWrite(pMe)" type="text" v-model="estTime" placeholder="[#]h [#]m"></p>
        <table>
          <tr class="header">
            <th>
            </th>
            <th v-bind:class="c.class" v-for="(c, index) in cols" :key="index">
              {{c.name}}
            </th>
          </tr>
          <tr class="row">
          <tr class="row">
            <td>
              sub task summary
            </td>
            <td v-bind:class="c.class" v-for="(c, index) in cols" :key="index">
              {{ c.get(task) }}
            </td>
          </tr>
        </table>
      </div>
      <div class="subs">

      </div>
    </div>
  </div>
</template>

<script>
  export default {
    name: 'tasks',
    data: function() {
      return this.initState()
    },
    computed: {
      cols(){
        return this.commonCols.filter(i => i.show())
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
          task: null,
          estTime: "0m",
          children: [],
          moreChildren: false,
          times: [],
          moreTimes: false,
          expenses: [],
          moreExpenses: false,
          files: [],
          moreFiles: false,
          comments: [],
          moreComments: false,
          loadingMoreAncestors: false,
          commonCols: [
            {
              name: "min time",
              class: "minimumTime",
              get: (t)=> this.$u.fmt.duration(t.minimumSubTime, this.project.hoursPerDay, this.project.daysPerWeek),
              show: () => this.$root.show.times
            },
            {
              name: "est time",
              class: "estimatedTime",
              get: (t)=> this.$u.fmt.duration(t.estimatedSubTime, this.project.hoursPerDay, this.project.daysPerWeek),
              show: () => this.$root.show.times
            },
            {
              name: "log time",
              class: "loggedTime",
              get: (t)=> this.$u.fmt.duration(t.loggedSubTime, this.project.hoursPerDay, this.project.daysPerWeek),
              show: () => this.$root.show.times
            },
            {
              name: "est exp",
              class: "estimatedExpense",
              get: (t)=> this.$u.fmt.cost(this.project.currencyCode, t.estimatedSubExpense),
              show: () => this.$root.show.expenses
            },
            {
              name: "log exp",
              class: "loggedExpense",
              get: (t)=> this.$u.fmt.cost(this.project.currencyCode,t.loggedSubExpense),
              show: () => this.$root.show.expenses
            },
            {
              name: "files",
              class: "fileCount",
              get: (t)=> t.fileSubCount,
              show: () => this.$root.show.files
            },
            {
              name: "file size",
              class: "fileSize",
              get: (t) => this.$u.fmt.bytes(t.fileSubSize),
              show: () => this.$root.show.files
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

<style lang="scss" scoped>
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
      }
    }
  }
}
</style>