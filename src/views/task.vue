<template>
  <div class="root">
   
  </div>
</template>

<script>
  export default {
    name: 'tasks',
    data: function() {
      return this.initState()
    },
    methods: {
      initState(){
        return {
          loading: true,
          me: null
        }
      },
      init(){
        for(const [key, value] of Object.entries(this.initState())) {
          this[key] = value
        }
        this.$api.user.me().then((me)=>{
          this.me = me
        }).finally(()=>{
          
        })
      }
    }
  }
</script>

<style lang="scss">

</style>