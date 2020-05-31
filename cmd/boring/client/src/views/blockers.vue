<template>
  <div class="root">
    <div v-if="loading">
      loading...
    </div>
    <div v-if="errors.length > 0">
      <p v-for="(error, index) in errors" :key="index">{{error}}</p>
    </div>
    <div v-if="game != null && game.board != null">
      <table class=board>
        <tr v-for="(_, y) in boardDims" :key="y">
          <td v-for="(_, x) in boardDims" :key="x" class="cell" :class="'p'+game.board[y*20+x]"></td>
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
      this.get()
      window.x_api = api
      return {
        gameType: "blockers",
        boardDims: 20,
        pieceSetsCount: 4,
        minPlayers: 2,
        loading: true,
        errors: [],
        game: {}
      }
    },
    methods: {
      get: function(){
        let id = router.currentRoute.params.id
        let promise = null
        let mapi = api
        if (id === 'new') {
          mapi = api.newMDoApi()
          mapi.game.active().then((info)=>{
            if (info != null) {
              router.push('/'+info.type+'/'+info.id)
              this.get()
            }
          }).catch((error)=>{
            this.errors.push(error.body)
          })
          mapi.blockers.new().then((game)=>{
            this.game = game
            router.push('/'+this.gameType+'/'+this.game.id)
          }).catch((error)=>{
            this.errors.push(error.body)
          })
          promise = mapi.sendMDo()
        } else {
          promise = api.blockers.get(id).then((game)=>{
            this.game = game
          }).catch((error)=>{
            this.errors.push(error.body)
          })
        }
        promise.catch((err)=>{
          console.log(err)
          this.errors.push(err)
        }).finally(()=>{
          this.loading = false
        })
      },
      xyToi: function(x, y, xDim){
        return xDim*y + x
      }
    }
  }
</script>

<style lang="scss">
.board{
  border-collapse: collapse;
  border: 1px solid black;
  .cell{
    border: 1px solid black;
    background: #222;
    width: 2pc;
    height: 2pc;
    &.p0 {
      background: red;
    }
    &.p1 {
      background: green;
    }
    &.p2 {
      background: blue;
    }
    &.p3 {
      background: yellow;
    }
  }
}
</style>