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
          OBSERVING GAME, ENJOY :)
        </p>
        <div v-if="game.state === 0">
          <p>
            PLAYERS: {{game.players.length}}
          </p>
          <p v-if="game.id !== game.myId">
            WAITING FOR MORE PLAYERS OR CREATOR TO START GAME
          </p>
          <p v-if="game.id === game.myId">
            SEND LINK TO FRIENDS:<br>{{link()}}
          </p>
          <button v-if="game.id === game.myId && game.players.length >= 2" @click.stop.prevent="start">START GAME</button>
        </div>
        <p v-if="game.state >= 1">
          TURN: {{game.turn + 1}}
        </p>
        <p v-if="game.state === 2">
          GAME OVER
        </p>
        <p v-if="game.state === 2" class="winner" :class="'ps'+getWinningPlayerIdx()">
          P{{getWinningPlayerIdx()+1}} WON!
        </p>
        <p v-if="game.state === 3">
          GAME ABANDONED! SHAME! SHAME! SHAME!
        </p>
        <button v-if="game.state === 2 || game.state === 3" @click.stop.prevent="goToGames()">
          BACK TO GAMES
        </button>
      </div>
      <div class="piece-sets">
         <div v-for="(_, pieceSet) in pieceSetsCount" :key="pieceSet" :class="['piece-set',game.state === 1 && turnPieceSetIdx() === pieceSet?'active':'','ps'+pieceSet]">
          <div class="piece-set-header">
            <p v-if="game.state === 0" class="player-tag">
              WAITING TO START
            </p>
            <p v-if="game.state >= 1" class="player-tag">
              {{pieceSetPlayerLabel(pieceSet)}} <span v-if="game.players[turnPieceSetIdx()] === game.myId">THAT'S YOU!</span>
            </p>
            <button :disabled="turnPieceSetIdx() !== pieceSet || game.state !== 1" class="red" v-if="game.pieceSetsEnded[pieceSet] === '0'" @click.stop.prevent="end(pieceSet)">
              END SET
            </button>
            <p class="set-state" v-else>
              SET ENDED
            </p>
            <p class="set-score">
              SET SCORE: {{pieceSetScore(pieceSet)}}
            </p>
          </div>
          <div v-for="(_, piece) in pieces.length" :key="piece" class="piece">
            <table class="piece" :class="'ps'+pieceSet" v-if="game.pieceSets[(pieces.length*pieceSet)+piece] === '1'">
              <tr v-for="(_, y) in pieces[piece].bb[1]" :key="y">
                <td v-for="(_, x) in pieces[piece].bb[0]" :key="x" 
                  class="cell" :class="[pieceCell(piece, x, y)===1? 'ps'+pieceSet :'dead' ]">{{piece}}</td>
              </tr>
            </table>
          </div>
        </div>
      </div>
        <button class="red" v-if="gameIsActive() && game.myId != null" @click.stop.prevent="abandon">Abandon</button>
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
              this.getTimeoutId = setTimeout(this.get, 3000)
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
      pieceSetPlayerLabel: function(pieceSet){
        switch (pieceSet) {
          case 0:
          case 1:
            return 'P'+(pieceSet+1)
          case 2:
            switch (this.game.players.length) {
              case 2:
                return 'P1'
              default:
                return 'P3'
            }
          case 3:
            switch (this.game.players.length) {
              case 2:
                return 'P2'
              case 4:
                return 'P4'
              case 3:
                if (this.turnPieceSetIdx() !== 3) {
                  return 'ROTA'
                }
                return 'P'+ (this.turnPlayerIdx()+1);
            }
        }
      },
      pieceSetScore: function(pieceSet){
        let score = 0
        for (let i = 0, l = this.pieces.length; i < l; i++) {
          if (this.game.pieceSets[(this.pieces.length*pieceSet)+i] === '1') {
            score += this.pieces[i].score
          }
        }
        return score
      },
      getWinningPlayerIdx: function() {
        let scores = []
        switch (this.game.players.length) {
          case 2:
            scores = [this.pieceSetScore(0)+this.pieceSetScore(2), this.pieceSetScore(1)+this.pieceSetScore(3)]
            break
          case 3:
            scores = [this.pieceSetScore(0), this.pieceSetScore(1), this.pieceSetScore(2)]
            break
          case 4:
            scores = [this.pieceSetScore(0), this.pieceSetScore(1), this.pieceSetScore(2), this.pieceSetScore(3)]
        }
        let winningPlayer = 0
        for (let i = 1, l = scores.length; i < l; i++) {
          if (scores[winningPlayer] >= scores[i]) {
            winningPlayer = i
          }
        }
        return winningPlayer
      },
      goToMyActiveGame: function(){
        if (this.myActiveGame.id != null) {
          router.push('/'+this.myActiveGame.type+'/'+this.myActiveGame.id)
        }
      },
      goToGames: function(){
        router.push('/games')
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
            return "ps0 shade"
          } else if (x === this.boardDims - 1 && y === 0) {
            return "ps1 shade"
          } else if (x === this.boardDims - 1 && y === this.boardDims - 1) {
            return "ps2 shade"
          } else if (x === 0 && y === this.boardDims - 1) {
            return "ps3 shade"
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
      end: function(pieceSet){
        if (this.turnPieceSetIdx() === pieceSet && this.game.players[this.turnPlayerIdx()] === this.game.myId) {
          api.blockers.takeTurn(0, 0, 0, 0, 1).then((game)=>{
            this.game = game
          })
        }
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

@mixin shade($base) {
  background: repeating-linear-gradient(
      45deg,
      color.adjust($base, $lightness: -26),
      color.adjust($base, $lightness: -26) 5px,
      color.adjust($base, $lightness: -15) 5px,
      color.adjust($base, $lightness: -15) 10px
    );
}
@mixin cell($root, $base) {
  &.#{$root}{
    color: white; //todo set this back to $base
    background: $base;
    &.dead{
      background: transparent;
    }
    &.shade{
      @include shade($base)
    }
  }
}
@mixin pieceSet($root, $base) {
  &.#{$root}{
    @include shade($base);
    &:not(.active){
      @include shade(color.adjust($base, $lightness: -27));
      .cell {
        background: color.adjust($base, $lightness: -36);
        &.dead{
          background: transparent;
        }
      }
    }
    &.active{
      .piece-set-header{
      @include shade(#fff);
        p {
          color: $base;
        }
      }
    }
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
  cursor: pointer;
  width: $cellSize;
  height: $cellSize;
  &:not(.dead){
    border: $border
  }
  @include cell("ps0", $ps0);
  @include cell("ps1", $ps1);
  @include cell("ps2", $ps2);
  @include cell("ps3", $ps3);
  @include cell("ps4", $empty);
}

.piece-sets {
    white-space: nowrap;
  > .piece-set {
    vertical-align: top;
    display: inline-block;
    border: 1px solid #555;
    @include pieceSet('ps0', $ps0);
    @include pieceSet('ps1', $ps1);
    @include pieceSet('ps2', $ps2);
    @include pieceSet('ps3', $ps3);

    > .piece-set-header{
      font-weight: 1000;
      align-content: center;
      @include shade(#333);
      p {
        margin: 10px 0;
        color: #555;
        background: transparent;
        text-align: center;
        > span {
          font-weight: inherit;
          color: inherit;
          background: transparent;
        }
      }
      button.red{
        position: relative;
        width: 5pc;
        left: calc(50% - 2.5pc);
        &:disabled, &[disabled]{
          background: #100;
          color: #555;
        }
      }
    }
    > .piece {
      background: transparent;
      float: left;
      clear: both;
      > table {
        margin: 0.5pc;
      }
    }
    &.active > .piece:hover td:not(.dead){
      border: 1px solid white;      
    }
  }
}

.info-and-controls{
  .winner{
    &.ps0{
      color: $ps0;
    }
    &.ps1{
      color: $ps1;
    }
    &.ps2{
      color: $ps2;
    }
    &.ps3{
      color: $ps3;
    }
  }
}

button.red{
  background: #500;
  border: 1px solid #300;
  &:hover {
      background-color: #700;
  }
  &:active {
      background-color: #500;
  }
}
</style>