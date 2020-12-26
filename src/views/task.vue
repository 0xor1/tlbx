<template>
  <div class="root">
    <div v-if="loading">
    </div>
    <div v-else>
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
            this.ancestors = res.set
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

</style>