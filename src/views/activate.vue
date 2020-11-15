<template>
  <div class="root">
    <p>Activating...</p>
  </div>
</template>

<script>
  export default {
    name: 'activate',
    data: function() {
      let query = this.$router.currentRoute.query
      this.$api.user.activate(query.email, query.code).then(()=>{
        this.$router.push('/')
      })
      return {}
    }
  }
</script>

<style scoped lang="scss">
</style>