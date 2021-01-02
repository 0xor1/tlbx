<template>
  <div class="root">
    <div class="loading" v-if="loading">
      loading...
    </div>
    <div class="content" v-else>
      <div class="breadcrumb">
        <span v-if="ancestors.length > 0 && ancestors[ancestors.length - 1].parent != null">
          <a :click="getMoreAncestors">..</a>
          /
        </span>
        <span v-for="a in ancestors" :key="a.id">
          <a :href="'/#/host/'+$u.rtr.host()+'/project/'+$u.rtr.project()+'/task/'+a.id">{{a.name}}</a>
          /           
        </span>
        <span>
          <a :href="'/#/host/'+$u.rtr.host()+'/project/'+$u.rtr.project()+'/task/'+task.id">{{task.name}}</a>   
        </span>
      </div>
      <div class="task">
        <h2>{{task.name}}</h2>
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
    methods: {
      initState(){
        return {
          loading: true,
          me: null,
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
          loadingMoreAncestors: false
        }
      },
      init(){
        for(const [key, value] of Object.entries(this.initState())) {
          this[key] = value
        }
        this.$api.user.me().then((me)=>{
          this.me = me
        }).finally(()=>{
          let mapi = this.$api.newMDoApi()
          if (this.me != null) {
            mapi.project.getMe(this.$u.rtr.host(), this.$u.rtr.project()).then((pMe)=>{
              if (pMe != null && pMe.isActive) {
                this.pMe = pMe
              }
            })
          }
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
  }
}
</style>