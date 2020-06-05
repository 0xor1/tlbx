<template>
  <div class="root">
    <div v-if="loading">
      loading...
    </div>
    <div v-if="errors.length > 0">
      <p v-for="(error, index) in errors" :key="index">{{error}}</p>
    </div>
    <div v-if="game != null && game.board != null" class="game">
      <table class=board>
        <tr v-for="(_, y) in boardDims" :key="y">
          <td v-for="(_, x) in boardDims" :key="x" class="cell" :class="['p'+boardCell(x, y), startCellStyleInfo(x, y).is?'p'+startCellStyleInfo(x, y).pieceSet+'start':'']">{{xyToI(x, y, boardDims)}}</td>
        </tr>
      </table>
      <div class="info-and-controls">
        <p v-if="gameIsActive() && game.myId == null">
          Obeserving game, enjoy the show :)
        </p>
        <div v-if="game.state === 0">
          <p>
            players: {{game.players.length}}
          </p>
          <p v-if="game.id !== game.myId">
            Waiting for more players or creator to start the game.
          </p>
          <p v-if="game.id === game.myId">
            Send this link to a friend <br> {{link()}}
          </p>
          <button v-if="game.id === game.myId && game.players.length >= 2" @click.stop.prevent="start">Start</button>
        </div>
        <div v-if="game.state === 1">
          <p>
            turn: {{game.turn + 1}}
          </p>
          <p :class="'piece-set-'+turnPieceSetIdx()">
            P{{turnPlayerIdx()+1}}s turn<span v-if="game.myId === game.players[turnPlayerIdx()]">. That's you!</span>
          </p>
        </div>
        <p v-if="game.state === 2">
          Game is finished
        </p>
        <p v-if="game.state === 3">
          This game was abandoned! shame! shame! shame!
        </p>
        <p v-if="game.state === 1 || game.state === 2">
          // todo print player scores
        </p>
      </div>
      <table class="piece-sets">
         <tr v-for="(_, piece) in pieces.length" :key="piece">
          <td v-for="(_, pieceSet) in pieceSetsCount" :key="pieceSet">
            <table class="piece" :class="'p'+pieceSet" v-if="game.pieceSets[(pieces.length*pieceSet)+piece] === '1'">
              <tr v-for="(_, y) in pieces[piece].bb[1]" :key="y">
                <td v-for="(_, x) in pieces[piece].bb[0]" :key="x" 
                  class="cell" :class="[pieceCell(piece, x, y)===1? 'p'+pieceSet :'' ]">{{piece}}</td>
              </tr>
            </table>
          </td>
        </tr>
      </table>
      <div class="abandon">
          <button v-if="gameIsActive() && game.myId != null" @click.stop.prevent="abandon">Abandon</button>
      </div>
    </div>
  </div>
</template>

<script>
  import api from '@/api'
  import router from '@/router'
  export default {
    name: 'blockers',
    data: function() {
      window.x_api = api // todo delete this line
      return {
        gameType: "blockers",
        boardDims: 20,
        pieceSetsCount: 4,
        minPlayers: 2,
        loading: true,
        myActiveGameRequested: false,
        errors: [],
        myActiveGame: {},
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
      new: function(){
        this.loading = true
        this.myActiveGameRequested = false
        this.errors = []
        this.myActiveGame = {}
        this.game = {}
        router.push('/'+this.gameType+'/new')
        this.get()
      },
      link: function(){
        return window.location.href
      },
      get: function(){
        clearTimeout(this.getTimeoutId)
        let id = router.currentRoute.params.id
        let mapi = api
        let isMDoReq = false
        let promise = null
        if (!this.myActiveGameRequested) {
          isMDoReq = true
          mapi = api.newMDoApi()
          this.myActiveGameRequested = true
          mapi.game.active().then((info)=>{
            if (info != null) {
              this.myActiveGame = info
            }
          })
        }
        if (id === 'new') {
          promise = mapi.blockers.new().then((game)=>{
            this.game = game
            router.push('/'+this.gameType+'/'+this.game.id)
          })
        } else {
          let updatedAfter = null
          if (this.game != null && this.game.updatedOn != null) {
            updatedAfter = this.game.updatedOn
          }
          promise = mapi.blockers.get(id, updatedAfter).then((game)=>{
            if (game == null) {
              // if we asked for data only after updatedAfter
              // null will be returned if there's no updated state.
              return
            }
            this.game = game
          })
        }
        if (isMDoReq) {
          promise = mapi.sendMDo()
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
          let poll = ()=>{
            this.loading = false
            if (this.game.id == null || this.gameIsActive()) {
              this.getTimeoutId = setTimeout(this.get, 5000)
            }
          }
          if (this.game.state === 0 && 
            this.game.myId == null && 
            this.game.players.length < this.pieceSetsCount &&
            this.myActiveGame.id == null) {
              api.blockers.join(this.game.id).then((game)=>{
                this.game = game
                poll()
            })
          } else {
            poll()
          }
        })
      },
      turnPlayerIdx: function(){
        let playerIdx = 0
        if (this.game.players.length === 3) {
          playerIdx = this.game.turn % this.pieceSetsCount
          if (playerIdx == 3) {
            playerIdx = (Math.floor(((this.game.turn) + 1) / this.pieceSetsCount) - 1) % 3
          }
        } else {
          playerIdx = this.game.turn % this.game.players.length
        }
        return playerIdx
      },
      turnPieceSetIdx: function(){
        return this.game.turn % this.pieceSetsCount
      },
      goToMyActiveGame: function(){
        if (this.myActiveGame.id != null) {
          router.push('/'+this.myActiveGame.type+'/'+this.myActiveGame.id)
        }
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
      gameIsActive: function(){
        return this.game.state === 0 || this.game.state === 1
      },
      start: function(){
        api.blockers.start(false).then((game)=>{
          this.game = game
        })
      },
      abandon: function(){
        if (window.confirm("Do you really want to abandon this game?!?!")) {
          clearTimeout(this.getTimeoutId)
          api.blockers.abandon().then(()=>{
            this.game.state = 3
          })
        }
      }
    },
    mounted: function(){
      this.get()
    },
    destroyed: function(){
      clearTimeout(this.getTimeoutId)
    }
  }
</script>

<style scoped lang="scss">
.game > *{
  margin: 1pc;
}
.board, .piece{
  border-collapse: collapse;
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
  margin: 0.5pc;
  cursor: pointer;
}
.piece-sets {
  border-collapse: collapse;
  border: 1px solid grey;
  td, th {
    border-right: 1px solid grey;
  }
}
.info-and-controls{
  .piece-set-0{
    color: #f00;
    span{
      color: inherit;
    }
  }
  .piece-set-1{
    color: #0f0;
    span{
      color: inherit;
    }
  }
  .piece-set-2{
    color: #00f;
    span{
      color: inherit;
    }
  }
  .piece-set-3{
    color: #ff0;
    span{
      color: inherit;
    }
  }
}
.abandon{
  button{
    background: #500;
    &:hover {
        background-color: #700;
    }
    &:active {
        background-color: #500;
    }
  }
}
</style>