<template>
  <div class="root">
    <p v-if="loading">Loading...</p>
    <a v-else @click.stop.prevent="loginout" href="">{{ authed? 'logout': 'login'}}</a>
  </div>
</template>

<script>
  export default {
    name: 'loginout',
    data() {
      return this.state()
    },
    methods: {
      state (){
        return {
          loading: true,
          authed: false
        }
      },
      init() {
        for(const [key, value] of Object.entries(this.state())) {
          this[key] = value
        }
        this.$api.user.me().then(()=>{
            this.authed = true
        }).catch(()=>{
            this.authed = false
        }).finally(()=>{
            this.loading = false
        })
      },
      loginout() {
        if (this.authed) {
          this.$api.user.logout().then(()=>{
            this.$router.push('/login')
          })
        } else {
          this.$router.push('/login')
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

<style scoped lang="scss">
</style>