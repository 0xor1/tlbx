<template>
  <div class="root">
    <p>confirming change email...</p>
  </div>
</template>

<script>
  import api from '@/api'
  import router from '@/router'
  export default {
    name: 'confirmChangeEmail',
    data: function() {
      let query = router.currentRoute.query
      api.user.confirmChangeEmail(query.me, query.code).then(()=>{
        router.push('/')
      })
      return {}
    }
  }
</script>

<style scoped lang="scss">
</style>