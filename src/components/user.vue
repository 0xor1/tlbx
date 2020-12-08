<template>
  <div class="root">
    <p v-if="loading">Loading...</p>
    <a v-else @click.stop.prevent="goto" href="">{{ alias }}</a>
  </div>
</template>

<script>
  export default {
    name: 'user',
    props: {
        user: String
    },
    data() {
      return this.initState()
    },
    methods: {
      initState (){
        return {
          loading: true,
          alias: ""
        }
      },
      init() {
        for(const [key, value] of Object.entries(this.initState())) {
          this[key] = value
        }
        this.$api.user.one(this.user).then((user)=>{
            this.loading = false
            this.alias = user.alias
        })
      },
      goto() {
        if (this.$router.currentRoute.path != `/host/${this.user}/projects`) {
            this.$router.push(`/host/${this.user}/projects`)
        }
      }
    },
    mounted(){
      this.init()
    }
  }
</script>

<style scoped lang="scss">
</style>