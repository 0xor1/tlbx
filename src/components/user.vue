<template>
  <div class="root">
    <p v-if="loading">Loading...</p>
    <a v-else @click.stop.prevent="goto" href="">{{ user.handle }}</a>
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
        for(const [key, value] of Object.entries(this.initState())) {
          this[key] = value
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
  p, a {
    background-color: transparent;
  }
}
</style>