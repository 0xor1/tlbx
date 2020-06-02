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
          <td v-for="(_, x) in boardDims" :key="x" class="cell" :class="'p'+boardCell(x, y)"></td>
        </tr>
      </table>
      <table class="piece-sets">
         <tr v-for="(_, piece) in pieces.length" :key="piece">
          <td v-for="(_, pieceSet) in pieceSetsCount" :key="pieceSet" class="piece" :class="'p'+pieceSet">
            <table class=piece>
              <tr v-for="(_, y) in pieces[piece].bb[1]" :key="y">
                <td v-for="(_, x) in pieces[piece].bb[0]" :key="x" 
                  class="cell" :class="['p'+pieceSet, pieceCell(piece, x, y)===1? 'active' :'dead' ]">{{pieceCell(piece, x, y)}}</td>
              </tr>
            </table>
          </td>
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
        game: {},
        pieces: [
          // 0
          // #
          {bb: [1, 1], shape: [1]},

          // 1
          // ##
          {bb: [2, 1], shape: [1, 1]},

          // 2
          // ###
          {bb: [3, 1], shape: [1, 1, 1]},

          // 3
          // #
          // ##
          {bb: [2, 2], shape: [1, 0, 1, 1]},

          // 4
          // ####
          {bb: [4, 1], shape: [1, 1, 1, 1]},

          // 5
          // ##
          // ##
          {bb: [2, 2], shape: [1, 1, 1, 1]},

          // 6
          //  #
          // ###
          {bb: [3, 2], shape: [0, 1, 0, 1, 1, 1]},

          // 7
          //   #
          // ###
          {bb: [3, 2], shape: [0, 0, 1, 1, 1, 1]},

          // 8
          //  ##
          // ##
          {bb: [3, 2], shape: [0, 1, 1, 1, 1, 0]},

          // 9
          // #####
          {bb: [5, 1], shape: [1, 1, 1, 1, 1]},

          // 10
          // ###
          // ##
          {bb: [3, 2], shape: [1, 1, 1, 1, 1, 0]},

          // 11
          //  #
          // ###
          //  #
          {bb: [3, 3], shape: [0, 1, 0, 1, 1, 1, 0, 1, 0]},

          // 12
          // #
          // ###
          //   #
          {bb: [3, 3], shape: [1, 0, 0, 1, 1, 1, 0, 0, 1]},

          // 13
          //    #
          // ####
          {bb: [4, 2], shape: [0, 0, 0, 1, 1, 1, 1, 1]},

          // 14
          //   #
          // ####
          {bb: [4, 2], shape: [0, 0, 1, 0, 1, 1, 1, 1]},

          // 15
          // ###
          //   ##
          {bb: [4, 2], shape: [1, 1, 1, 0, 0, 0, 1, 1]},

          // 16
          // #
          // ###
          //  #
          {bb: [3, 3], shape: [1, 0, 0, 1, 1, 1, 0, 1, 0]},

          // 17
          // ###
          // # #
          {bb: [3, 2], shape: [1, 1, 1, 1, 0, 1]},

          // 18
          // #
          // ###
          // #
          {bb: [3, 3], shape: [1, 0, 0, 1, 1, 1, 1, 0, 0]},

          // 19
          // ##
          //  ##
          //   #
          {bb: [3, 3], shape: [1, 1, 0, 0, 1, 1, 0, 0, 1]},

          // 20
          // #
          // #
          // ###
          {bb: [3, 3], shape: [1, 0, 0, 1, 0, 0, 1, 1, 1]}
        ]
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
          })
          mapi.blockers.new().then((game)=>{
            this.game = game
            router.push('/'+this.gameType+'/'+this.game.id)
          })
          promise = mapi.sendMDo()
        } else {
          promise = api.blockers.get(id).then((game)=>{
            this.game = game
          })
        }
        promise.catch((error)=>{
          if (error.isMDoErrors === true) {
            for (let i = 0, l = error.length; i < l; i++) {
              this.errors.push(error[i])
            }
          } else {
            this.errors.push(error)
          }
        }).finally(()=>{
          this.loading = false
        })
      },
      xyToI: function(x, y, xDim){
        return xDim*y + x
      },
      iToXY: function(i, xDim, yDim){
        return {
          x: i % xDim,
          y: Math.floor(i / yDim)
        }
      },
      boardCell: function(x, y){
        return this.game[this.xyToI(x, y, this.boardDims)]
      },
      pieceCell: function(piece, x, y){
        let p = this.pieces[piece]
        return p.shape[this.xyToI(x, y, p.bb[0])]
      }
    },
    watch: {
      pieces: {
        deep: true
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