<template>
  <div class="root" v-html="docs">
  </div>
</template>

<script>
  export default {
    name: 'apiDocs',
    data: function() {
      return this.initState()
    },
    methods: {
      initState (){
        return {
            docs: {}
        }
      },
      init() {
        this.$u.copyProps(this.initState(), this)
        this.$api.docs().then((docs)=>{
            this.docs = docs
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

<style lang="scss" scoped>
</style>