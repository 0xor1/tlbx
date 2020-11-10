<template>
  <div class="root">
    <p>Activating...</p>
  </div>
</template>

<script>
  import api from '@/api'
  import router from '@/router'
  export default {
    name: 'activate',
    data: function() {
      let query = router.currentRoute.query
      api.user.activate(query.email, query.code).then(()=>{
        router.push('/')
      })
      return {}
    }
  }
</script>

<style scoped lang="scss">
</style>