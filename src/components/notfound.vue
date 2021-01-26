<template>
  <div class="root">
    <h1>{{type}} not found :(</h1>
    <p v-if="me != null">return <a :href="`/#/host/${me.id}/projects`">home</a></p>
    <p v-else>go to <a href="/#/login">login</a></p>
  </div>
</template>

<script>
  export default {
    name: 'notfound',
    props: {
      type: String
    },
    data: function() {
      return {
        me: null
      }
    },
    methods: {
      init(){
        this.$api.user.me().then((me)=>{
            this.me = me
        })
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