<template>
  <div class="root">
    <p v-if="loading">Loading...</p>
    <p v-else>
      <user v-if="authed" v-bind:userId="me.id"></user>&nbsp;
      <a v-if="!/\/(login|register)/.test($router.currentRoute.path)" @click.stop.prevent="loginout" href="">{{ authed? 'logout': 'login'}}</a>
    </p>
  </div>
</template>

<script>
  import user from './user'
  export default {
    name: 'loginout',
    components: { user },
    data() {
      return this.initState()
    },
    methods: {
      initState (){
        return {
          loading: true,
          authed: false,
          me: null
        }
      },
      init() {
        for(const [key, value] of Object.entries(this.initState())) {
          this[key] = value
        }
        this.$api.user.me().then((me)=>{
          this.me = me
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
user{
  display: inline-block;
}
</style>