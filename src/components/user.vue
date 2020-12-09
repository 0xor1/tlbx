<template>
  <div class="root">
    <p v-if="loading">Loading...</p>
    <a v-else @click.stop.prevent="goto" href="">{{ user.handle }}</a>
  </div>
</template>

<script>
  export default {
    name: 'user',
    props: {
        userId: String
    },
    data() {
      return this.initState()
    },
    methods: {
      initState (){
        return {
          loading: true,
          user: null
        }
      },
      init() {
        for(const [key, value] of Object.entries(this.initState())) {
          this[key] = value
        }
        this.$api.user.one(this.userId).then((user)=>{
            this.loading = false
            this.user = user
        })
      },
      goto() {
        if (this.$router.currentRoute.path != `/host/${this.userId}/projects`) {
            this.$router.push(`/host/${this.userId}/projects`)
        }
      }
    },
    mounted(){
      this.init()
    }
  }
</script>

<style scoped lang="scss">
div {
  display: inline-block;
}
</style>