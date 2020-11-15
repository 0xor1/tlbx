<template>
  <div class="root">
    <p>confirming change email...</p>
  </div>
</template>

<script>
  export default {
    name: 'confirmChangeEmail',
    data: function() {
      let query = this.$router.currentRoute.query
      this.$api.user.confirmChangeEmail(query.me, query.code).then(()=>{
        this.$router.push('/')
      })
      return {}
    }
  }
</script>

<style scoped lang="scss">
</style>