<template>
  <div class="root">
    <span v-if="loading">loading...</span>
    <a v-else-if="user != null" @click.stop.prevent="goto" href="">{{ user.handle }}</a>
    <span v-else>--</span>
  </div>
</template>

<script>
  export default {
    name: 'user',
    props: {
        userId: String,
        goToHome: Boolean
    },
    data() {
      return this.initState()
    },
    methods: {
      initState (){
        return {
          loading: true,
          user: null
        }
      },
      init() {
        this.$u.copyProps(this.initState(), this)
        if (this.userId == null || this.userId === "") {
          this.loading = false
          return
        }
        this.$api.user.one(this.userId).then((user)=>{
            this.loading = false
            this.user = user
        })
      },
      goto() {
        if (this.goToHome !== true && this.$u.rtr.project() != null && this.$u.rtr.name() != "projectUser") {
          this.$u.rtr.goto(`/host/${this.$u.rtr.host()}/project/${this.$u.rtr.project()}/user/${this.userId}`)
          return
        }
        this.$u.rtr.goto(`/host/${this.userId}/projects`)
      }
    },
    mounted(){
      this.init()
    }
  }
</script>

<style scoped lang="scss">
div.root {
  display: inline-block;
  background-color: transparent;
  span, a {
    background-color: transparent;
  }
}
</style>