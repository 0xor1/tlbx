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
          <a :title="a.name" :href="'/#/host/'+$u.rtr.host()+'/project/'+$u.rtr.project()+'/task/'+a.id">{{$u.fmt.ellipsis(a.name, 10)}}</a>
          /           
        </span>
        <span>
          <a :title="task.name" :href="'/#/host/'+$u.rtr.host()+'/project/'+$u.rtr.project()+'/task/'+task.id">{{$u.fmt.ellipsis(task.name, 10)}}</a>   
        </span>
      </div>
      <div class="task">
        <h1 >{{task.name}}</h1>
        <p v-if="task.description != null">{{task.description}}</p>
        <p>is parallel: {{task.isParallel}}</p>
        <p>min sub time: {{$u.fmt.duration(task.estimatedTime + task.minimumSubTime, project.hoursPerDay, project.daysPerWeek)}}</p>
        <table>
          <tr class="header">
            <th>
            </th>
            <th v-bind:class="c.class" v-for="(c, index) in cols" :key="index">
              {{c.name}}
            </th>
          </tr>
          <tr class="row">
            <td>
              this task
            </td>
            <td v-bind:class="c.class" v-for="(c, index) in cols" :key="index">
              {{ c.get(task) }}
            </td>
          </tr>
          <tr class="row">
            <td>
              sub tasks
            </td>
            <td v-bind:class="c.class" v-for="(c, index) in cols" :key="index">
              {{ c.sub(task) }}
            </td>
          </tr>
          <tr class="row">
            <td>
              summary
            </td>
            <td v-bind:class="c.class" v-for="(c, index) in cols" :key="index">
              {{ c.summary(task) }}
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
          loading: true,
          me: null,
          project: null,
          pMe: null,
          ancestors: [],
          moreAncestors: false,
          task: null,
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
              name: "est time",
              class: "estimatedTime",
              summary: (t)=> this.$u.fmt.duration(t.estimatedTime + t.estimatedSubTime, this.project.hoursPerDay, this.project.daysPerWeek),
              get: (t)=> this.$u.fmt.duration(t.estimatedTime, this.project.hoursPerDay, this.project.daysPerWeek),
              sub: (t)=> this.$u.fmt.duration(t.estimatedSubTime, this.project.hoursPerDay, this.project.daysPerWeek),
              show: () => this.$root.show.times
            },
            {
              name: "log time",
              class: "loggedTime",
              summary: (t)=> this.$u.fmt.duration(t.loggedTime + t.loggedSubTime, this.project.hoursPerDay, this.project.daysPerWeek),
              get: (t)=> this.$u.fmt.duration(t.loggedTime, this.project.hoursPerDay, this.project.daysPerWeek),
              sub: (t)=> this.$u.fmt.duration(t.loggedSubTime, this.project.hoursPerDay, this.project.daysPerWeek),
              show: () => this.$root.show.times
            },
            {
              name: "est exp",
              class: "estimatedExpense",
              summary: (t)=> this.$u.fmt.cost(this.project.currencyCode, t.estimatedExpense + t.estimatedSubExpense),
              get: (t)=> this.$u.fmt.cost(this.project.currencyCode, t.estimatedExpense),
              sub: (t)=> this.$u.fmt.cost(this.project.currencyCode, t.estimatedSubExpense),
              show: () => this.$root.show.expenses
            },
            {
              name: "log exp",
              class: "loggedExpense",
              summary: (t)=> this.$u.fmt.cost(this.project.currencyCode, t.loggedExpense + t.loggedSubExpense),
              get: (t)=> this.$u.fmt.cost(this.project.currencyCode, t.loggedExpense),
              sub: (t)=> this.$u.fmt.cost(this.project.currencyCode,t.loggedSubExpense),
              show: () => this.$root.show.expenses
            },
            {
              name: "files",
              class: "fileCount",
              summary: (t)=> t.fileCount + t.fileSubCount,
              get: (t)=> t.fileCount,
              sub: (t)=> t.fileSubCount,
              show: () => this.$root.show.files
            },
            {
              name: "file size",
              class: "fileSize",
              summary: (t) => this.$u.fmt.bytes(t.fileSize + t.fileSubSize),
              get: (t) => this.$u.fmt.bytes(t.fileSize),
              sub: (t) => this.$u.fmt.bytes(t.fileSubSize),
              show: () => this.$root.show.files
            },
            {
              name: "tasks",
              class: "tasks",
              summary: (t)=> t.descendantCount + 1,
              get: ()=> 1,
              sub: (t)=> t.descendantCount,
              show: () => this.$root.show.tasks
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