<template>
  <div class="root">
    <p>Activating...</p>
  </div>
</template>

<script>
  import api from '@/api'
  export default {
    name: 'activate',
    data: function() {
      let query = this.$router.currentRoute.query
      api.user.activate(query.email, query.code).then(()=>{
        this.$router.push('/')
      })
      return {}
    }
  }
</script>

<style scoped lang="scss">
</style>