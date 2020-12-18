<template>
    <div class="app-root">
        <div v-bind:class="{'show-menu':showMenu}" class="body">
            <router-view/>
        </div>
        <div v-if="showMenu" v-bind:class="{'show-menu':showMenu}" class="menu">
          <div class="opts">
            <div class="btn" v-if="me" @click.stop.prevent="goHome()">
              home
            </div>
            <div class="btn" v-if="me" @click.stop.prevent="showFields=!showFields">
              fields
            </div>
            <div class="fields" v-if="showFields">
              <div @click.stop.prevent="$root.show.dates = !$root.show.dates">
                <label @click.stop for="dates">dates</label>
                <input @click.stop type="checkbox" v-model="$root.show.dates" id="dates">
              </div>
              <div @click.stop.prevent="$root.show.times = !$root.show.times">
                <label @click.stop for="times">times</label>
                <input @click.stop type="checkbox" v-model="$root.show.times" id="times">
              </div>
              <div @click.stop.prevent="$root.show.expenses = !$root.show.expenses">
                <label @click.stop for="expenses">expenses</label>
                <input @click.stop type="checkbox" v-model="$root.show.expenses" id="expenses">
              </div>
              <div @click.stop.prevent="$root.show.files = !$root.show.files">
                <label @click.stop for="files">files</label>
                <input @click.stop type="checkbox" v-model="$root.show.files" id="files">
              </div>
              <div @click.stop.prevent="$root.show.tasks = !$root.show.tasks">
                <label @click.stop for="tasks">tasks</label>
                <input @click.stop type="checkbox" v-model="$root.show.tasks" id="tasks">
              </div>
            </div>
            <div class="btn" @click.stop.prevent="loginout">
              {{ me? 'logout': 'login'}}
            </div>
          </div>
        </div>
        <div @click.stop.prevent="showMenu=!showMenu" class="menu-toggle">
          <div></div>
          <div></div>
          <div></div>
        </div>
    </div>
</template>

<script>
  export default {
    name: 'app',
    data() {
      return this.initState()
    },
    methods: {
      initState(){
        return {
          loading: true,
          showMenu: this.showMenu || false,
          showFields: this.showFields || false,
          me: null
        }
      },
      init(){
        for(const [key, value] of Object.entries(this.initState())) {
          this[key] = value
        }
        this.$api.user.me().then((me)=>{
          this.me = me
        }).finally(()=>{
          this.loading = false
        })
      },
      loginout() {
        if (this.me != null) {
          this.$api.user.logout().then(()=>{
            this.goto('/login')
          })
        } else {
          this.goto('/login')
        }
      },
      goHome(){
        this.goto(`/host/${this.me.id}/projects`)
      },
      goto(path){
        if (this.$router.currentRoute.path != path) {
          this.$router.push(path)
        }
      }
    },mounted(){
      this.init()
    },
    watch: {
      $route () {
        this.init()
      }
    }
  }
</script>

<style lang="scss">
$color: #ddd;
$borderColor: #777;
$bgColor: #000;
$inputColor: #222;
$inputHoverColor: #555;
$inputActiveColor: #222;
$inputPlaceholderColor: #aaa;

@mixin border(
  $dir: false,
  $ticc: 1px, 
  $style: solid, 
  $color: $color) {
  @if $dir {
    border-#{$dir}: $ticc $style $color;
  } @else {
    border: $ticc $style $color;
  }
}

@mixin basic(
  $display: block,
  $width: 100%,
  $height: 100%,
  $margin: 0,
  $padding: 0,
  $overflow: hidden) {
    display: $display;
    width: $width;
    height: $height;
    margin: $margin;
    padding: $padding;
    overflow: $overflow;
}

.app-root{
  > .body{
    position: absolute;
    width: 100%;
    height: 100%;
    overflow: auto;
    &.show-menu{
      @media only screen and (min-width: 480px) {
        left: 15.1pc;
        width: calc(100% - 15.1pc);
      }
    }
  }
  > .menu-toggle{
    position: absolute;
    top: 0;
    left: 0;
    height: 2.5pc;
    width: 2.5pc;
    cursor: pointer;
    display: flex;
    flex-direction: column;
    justify-content: center;
    align-items: center;
    background-color: $inputColor;
    &:hover{
      background-color: $inputHoverColor
    }
    &:active{
      background-color: $inputActiveColor
    }
    @include border();
    > div{
      margin: 0.1pc;
      width: 1.8pc;
      height: 0.35pc;
      background-color: $color;
    }
  }
  > .menu{
    position: absolute;
    top: 0;
    left: 0;
    padding-top: 2.6pc;
    overflow: hidden;
    width: 0;
    height: 100%;
    &.show-menu{
      @include border($dir: right);
      width: 100%;
      overflow-y: auto;
      @media only screen and (min-width: 480px) {
        width: 15pc;
      }
    }
    > .opts{
      display: flex;
      flex-direction: column;
      justify-content: flex-start;
      align-items: stretch;
      overflow-y: auto;
      height: calc(100% - 2.5pc);
      > .btn{
        cursor: pointer;
        text-align: center;
        height: 2pc;
        line-height: 2pc;
        &:hover{
          background-color: $inputHoverColor
        }
        &:active{
          background-color: $inputActiveColor
        }
      }
      > .fields{
        text-align: center;
        > div{
          width: 100%;
          cursor: pointer;
          &:hover{
            background-color: $inputHoverColor
          }
          &:active{
            background-color: $inputActiveColor
          }
          > label, >input {
            cursor: pointer;
            background: transparent;
          }
        }
        &:hover{
          background-color: $inputHoverColor
        }
        &:active{
          background-color: $inputActiveColor
        }
      }
    }
  }
}
// what follows is essentially a prelude
// for the entire app
html, body, .app-root {
    @include basic;
}
table{
  th, td{
    @include border($ticc: 0.1pc);
  }
}
@import url(https://fonts.googleapis.com/css2?family=Roboto+Mono:wght@400;700&display=swap);
* {
    font-family: 'Roboto Mono', monospace;
    background-color: $bgColor;
    color: $color;
}
button {
    cursor: pointer;
    &:hover {
        background-color: $inputHoverColor;
    }
    &:active {
        background-color: $inputActiveColor;
    }
}
input, button {
    background-color: $inputColor;
    outline: none;
    border: 1px solid $color;
    border-radius: 2px;
    padding: 5px;
    &::placeholder{
        color: $inputPlaceholderColor;
    }
}
</style>