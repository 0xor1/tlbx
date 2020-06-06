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
          <td v-for="(_, x) in boardDims" :key="x" class="cell" :class="boardCellClass(x, y)">{{xyToI(x, y, boardDims)}}</td>
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
        </div>
        <p v-if="game.state === 2">
          Game is finished
        </p>
        <p v-if="game.state === 3">
          This game was abandoned! shame! shame! shame!
        </p>
      </div>
      <div class="piece-sets">
         <div v-for="(_, pieceSet) in pieceSetsCount" :key="pieceSet" :class="[game.state === 1 && turnPieceSetIdx() === pieceSet?'active':'','piece-set',''+pieceSet]">
          <!-- <div v-if="game.state >= 1" class="piece-set-header">
            P{{pieceSetPlayerIdx(pieceSet)}} score: {{pieceSetScore(pieceSet)}}
          </div> -->
          <div v-for="(_, piece) in pieces.length" :key="piece">
            <table class="piece" :class="'p'+pieceSet" v-if="game.pieceSets[(pieces.length*pieceSet)+piece] === '1'">
              <tr v-for="(_, y) in pieces[piece].bb[1]" :key="y">
                <td v-for="(_, x) in pieces[piece].bb[0]" :key="x" 
                  class="cell" :class="[pieceCell(piece, x, y)===1? 'ps'+pieceSet :'dead' ]">{{piece}}</td>
              </tr>
            </table>
          </div>
        </div>
      </div>
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
          {bb: [1, 1], shape: [1], score: 1},

          // 1
          // ##
          {bb: [2, 1], shape: [1, 1], score: 2},

          // 2
          // ###
          {bb: [3, 1], shape: [1, 1, 1], score: 3},

          // 3
          // #
          // ##
          {bb: [2, 2], shape: [1, 0, 1, 1], score: 3},

          // 4
          // ####
          {bb: [4, 1], shape: [1, 1, 1, 1], score: 4},

          // 5
          // ##
          // ##
          {bb: [2, 2], shape: [1, 1, 1, 1], score: 4},

          // 6
          //  #
          // ###
          {bb: [3, 2], shape: [0, 1, 0, 1, 1, 1], score: 4},

          // 7
          //   #
          // ###
          {bb: [3, 2], shape: [0, 0, 1, 1, 1, 1], score: 4},

          // 8
          //  ##
          // ##
          {bb: [3, 2], shape: [0, 1, 1, 1, 1, 0], score: 4},

          // 9
          // #####
          {bb: [5, 1], shape: [1, 1, 1, 1, 1], score: 5},

          // 10
          // ###
          // ##
          {bb: [3, 2], shape: [1, 1, 1, 1, 1, 0], score: 5},

          // 11
          //  #
          // ###
          //  #
          {bb: [3, 3], shape: [0, 1, 0, 1, 1, 1, 0, 1, 0], score: 5},

          // 12
          // #
          // ###
          //   #
          {bb: [3, 3], shape: [1, 0, 0, 1, 1, 1, 0, 0, 1], score: 5},

          // 13
          //    #
          // ####
          {bb: [4, 2], shape: [0, 0, 0, 1, 1, 1, 1, 1], score: 5},

          // 14
          //   #
          // ####
          {bb: [4, 2], shape: [0, 0, 1, 0, 1, 1, 1, 1], score: 5},

          // 15
          // ###
          //   ##
          {bb: [4, 2], shape: [1, 1, 1, 0, 0, 0, 1, 1], score: 5},

          // 16
          // #
          // ###
          //  #
          {bb: [3, 3], shape: [1, 0, 0, 1, 1, 1, 0, 1, 0], score: 5},

          // 17
          // ###
          // # #
          {bb: [3, 2], shape: [1, 1, 1, 1, 0, 1], score: 5},

          // 18
          // #
          // ###
          // #
          {bb: [3, 3], shape: [1, 0, 0, 1, 1, 1, 1, 0, 0], score: 5},

          // 19
          // ##
          //  ##
          //   #
          {bb: [3, 3], shape: [1, 1, 0, 0, 1, 1, 0, 0, 1], score: 5},

          // 20
          // #
          // #
          // ###
          {bb: [3, 3], shape: [1, 0, 0, 1, 0, 0, 1, 1, 1], score: 5}
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
              this.getTimeoutId = setTimeout(this.get, 2000)
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
      boardCellClass: function(x, y) {
        if (this.boardCell(x, y) === '4') {
          // only apply start cell styles to start cells
          // with no pieces over the top of them.
          if (x === 0 && y === 0) {
            return "ps0 start"
          } else if (x === this.boardDims - 1 && y === 0) {
            return "ps1 start"
          } else if (x === this.boardDims - 1 && y === this.boardDims - 1) {
            return "ps2 start"
          } else if (x === 0 && y === this.boardDims - 1) {
            return "ps3 start"
          }
        }
        return "ps"+this.boardCell(x, y)
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
@use "sass:color";

$ps0: #d00;
$ps1: #0d0;
$ps2: #00d;
$ps3: #dd0;
$empty: #222;
$border: 1px solid black;
$cellSize: 2pc;

@mixin cell($base) {
  color: white;
  background: $base;
  &.start{
    background: repeating-linear-gradient(
      45deg,
      color.adjust($base, $lightness: -33),
      color.adjust($base, $lightness: -33) 5px,
      color.adjust($base, $lightness: -26) 5px,
      color.adjust($base, $lightness: -26) 10px
    );
  }
}

.game > *{
  margin: 1pc;
}

table, tr, td{
  border-collapse: collapse;
  background: transparent;
}

.cell{
  width: $cellSize;
  height: $cellSize;
  &:not(.dead){
    border: $border
  }
  &.ps0 {
    @include cell($ps0);
  }
  &.ps1 {
    @include cell($ps1);
  }
  &.ps2 {
    @include cell($ps2);
  }
  &.ps3 {
    @include cell($ps3);
  }
  &.ps4 {
    @include cell($empty);
  }
}

.cell{
  cursor: pointer;
}

.piece-sets {
  > div {
    vertical-align: top;
    display: inline-block;
    border: 1px solid #555;
    &.active{
      background: #555;
    }
    > * {
      background: transparent;
      margin: 0.75pc;
    }
  }
}

.info-and-controls{
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