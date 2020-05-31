<template>
  <div class="root">
    <div v-if="loading">
      loading...
    </div>
    <div v-else-if="err.length > 0">
      {{err}}
    </div>
    <div v-else>
      <table>
        <tr v-for="y in 20" :key="y-1">
          <td v-for="x in 20" :key="x-1">{{x-1}}</td>
        </tr>
      </table>
    </div>
  </div>
</template>

<script>
  import api from '@/api'
  import router from '@/router'
  export default {
    name: 'blockers',
    data: function() {
      var data = {
        loading: true,
        err: "",
        game: {}
      }
      let id = router.currentRoute.params.id
      let promise = null
      if (id === 'new') {
        promise = api.blockers.new().then((game)=>{
          data.game = game
          router.push('/blockers/'+data.game.id)
        })
      } else {
        promise = api.blockers.get(id).then((game)=>{
          data.game = game
        })
      }
      promise.catch((err)=>{
        data.err = err.response.data
      }).finally(()=>{
        data.loading = false
      })
      return data
    },
    methods: {
    }
  }
</script>

<style lang="scss">

</style>