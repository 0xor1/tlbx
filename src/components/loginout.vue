<template>
  <div class="root">
    <p v-if="loading">Loading...</p>
    <a v-else @click.stop.prevent="click" href="">{{ authed? 'logout': 'login'}}</a>
  </div>
</template>

<script>
  export default {
    name: 'loginout',
    data: function() {
      return {
          loading: true,
          authed: false
      }
    },
    methods: {
      click() {
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
        this.$api.user.me().then(()=>{
            this.authed = true
        }).catch(()=>{
            this.authed = false
        }).finally(()=>{
            this.loading = false
        })
    }
  }
</script>

<style scoped lang="scss">
</style>