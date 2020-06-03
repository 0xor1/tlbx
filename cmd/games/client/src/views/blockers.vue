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
          <td v-for="(_, x) in boardDims" :key="x" class="cell" :class="['p'+boardCell(x, y), startCellStyleInfo(x, y).is?'p'+startCellStyleInfo(x, y).pieceSet+'start':'']"></td>
        </tr>
      </table>
      <table class="piece-sets">
         <tr v-for="(_, piece) in pieces.length" :key="piece">
          <td v-for="(_, pieceSet) in pieceSetsCount" :key="pieceSet">
            <table class="piece" :class="'p'+pieceSet">
              <tr v-for="(_, y) in pieces[piece].bb[1]" :key="y">
                <td v-for="(_, x) in pieces[piece].bb[0]" :key="x" 
                  class="cell" :class="[pieceCell(piece, x, y)===1? 'p'+pieceSet :'' ]"></td>
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
          let updatedAfter = null
          if (this.game != null && this.game.updatedOn != null) {
            updatedAfter = this.game.updatedOn
          }
          promise = api.blockers.get(id, updatedAfter).then((game)=>{
            if (game == null) {
              // if we asked for data only after updatedAfter
              // null will be returned if there's no updated state.
              return
            }
            this.game = game
            if (game.state === 0 && game.myId == null) {
              api.blockers.join(game.id)
            }
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
          if (this.gameIsActive()) {
            setTimeout(this.get, 5000)
          }
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
        return this.game.board[this.xyToI(x, y, this.boardDims)]
      },
      startCellStyleInfo: function(x, y) {
        console.log(this.boardCell(x, y))
        if (this.boardCell(x, y) === '4') {
          // only apply start cell styles to start cells
          // with no pieces over the top of them.
          if (x === 0 && y === 0) {
            return {is: true, pieceSet: 0}
          } else if (x === this.boardDims - 1 && y === 0) {
            return {is: true, pieceSet: 1}
          } else if (x === this.boardDims - 1 && y === this.boardDims - 1) {
            return {is: true, pieceSet: 2}
          } else if (x === 0 && y === this.boardDims - 1) {
            return {is: true, pieceSet: 3}
          }
        }
        return {is: false, pieceSet: null}
      },
      pieceCell: function(piece, x, y){
        let p = this.pieces[piece]
        return p.shape[this.xyToI(x, y, p.bb[0])]
      },
      gameState: function(){
        return this.game.state
      },
      gameIsActive: function(){
        return this.gameState() === 0 || this.gameState() === 1
      }
    }
  }
</script>

<style lang="scss">
.board, .piece{
  border-collapse: collapse;
  border: 1px solid black;
  .cell{
    border: 1px solid black;
    width: 2pc;
    height: 2pc;
    &.p0 {
      background: #f00;
    }
    &.p1 {
      background: #0f0;
    }
    &.p2 {
      background: #00f;
    }
    &.p3 {
      background: #ff0;
    }
    &.p0start {
      background: repeating-linear-gradient(
        45deg,
        #300,
        #300 5px,
        #500 5px,
        #500 10px
      );
    }
    &.p1start {
      background: repeating-linear-gradient(
        45deg,
        #030,
        #030 5px,
        #050 5px,
        #050 10px
      );
    }
    &.p2start {
      background: repeating-linear-gradient(
        45deg,
        #003,
        #003 5px,
        #005 5px,
        #005 10px
      );
    }
    &.p3start {
      background: repeating-linear-gradient(
        45deg,
        #330,
        #330 5px,
        #550 5px,
        #550 10px
      );
    }
  }
}
.board .cell {
  background: #222;
}
.piece{
  cursor: pointer;
  margin: 1pc;
}

</style>