<template>
  <div class="root">
    <p>Directing...</p>
  </div>
</template>

<script>
  export default {
    name: 'directing',
    data: function() {
      this.$api.user.me().then((me)=>{
        this.$router.push(`/host/${me.id}/projects`)
      }).catch(()=>{
        this.$router.push('/login')
      })
      return {}
    }
  }
</script>

<style scoped lang="scss">
</style>