<template>
  <div class="root">
    <p>Directing...</p>
  </div>
</template>

<script>
  export default {
    name: 'directing',
    data: function() {
      let me = null
      this.$api.user.me().then((res)=>{
        me = res
      }).finally(()=>{
        if (me != null) {
          this.$router.push(`/host/${me.id}/projects`)
        } else {
          this.$router.push('/login')
        }
      })
      return {}
    }
  }
</script>

<style scoped lang="scss">
</style>