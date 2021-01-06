<template>
  <div class="root">
  </div>
</template>

<script>
  export default {
    name: 'projectUser',
    data: function() {
      return this.initState()
    },
    methods: {
      initState (){
        return {}
      },
      init() {
        for(const [key, value] of Object.entries(this.initState())) {
          this[key] = value
        }
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

<style lang="scss" scoped>
div.root {
  padding: 2.6pc 0 0 1.3pc;
}
</style>