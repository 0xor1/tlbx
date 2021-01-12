<template>
    <div class="app-root">
        <div v-bind:class="{'show-menu':showMenu, 'show-project-activity': showProjectActivity && showProjectActivityToggle}" class="body">
            <router-view/>
        </div>
        <!-- main menu -->
        <div v-if="showMenu" v-bind:class="{'show':showMenu}" class="menu slide-out">
          <div class="opts">
            <div class="btn" v-if="me" @click.stop.prevent="goHome()">
              home
            </div>
            <div class="btn" @click.stop.prevent="showFields=!showFields">
              fields
            </div>
            <div class="fields" v-if="showFields">
              <div v-for="(f, index) in fields" :key="index" @click.stop.prevent="$root.show[f] = !$root.show[f]">
                <span>{{f}}<input @click.stop type="checkbox" v-model="$root.show[f]"></span>
              </div>
            </div>
            <div class="btn" @click.stop.prevent="loginout">
              {{ me? 'logout': 'login'}}
            </div>
          </div>
        </div>
        <div title="menu" @click.stop.prevent="showMenu=!showMenu" class="slide-out-toggle menu-toggle">
          <div></div>
          <div></div>
          <div></div>
        </div>
        <!-- project activity stream -->
        <div v-if="showProjectActivityToggle && showProjectActivity" v-bind:class="{'show':showProjectActivity}" class="project-activity slide-out">
          <div class="entry" v-for="(a, index) in projectActivity" :key="index">
            <p>
              user: <user :userId="a.user"></user>
            </p>
            <p>
              action: {{a.action}}
            </p>
            <p>
              type: {{a.itemType}}
            </p>
          </div>
        </div>
        <div title="project activity" v-if="showProjectActivityToggle" @click.stop.prevent="showProjectActivity=!showProjectActivity" class="slide-out-toggle project-activity-toggle">
          <img src="@/assets/activity.svg">
        </div>
    </div>
</template>

<script>
  import user from './components/user'
  export default {
    name: 'app',
    components: {user},
    data() {
      return this.initState()
    },
    methods: {
      initState(){
        return {
          loading: true,
          currentProjectId: this.currentProjectId || null,
          showMenu: this.showMenu || false,
          showProjectActivity: this.showProjectActivity || false,
          showProjectActivityToggle: this.$u.rtr.project() != null,
          showFields: this.showFields || false,
          projectActivity: this.projectActivity || [],
          moreProjectActivity: this.moreProjectActivity || false,
          fields: [
            "date",
            "user",
            "time",
            "cost",
            "file",
            "task"
          ],
          me: null
        }
      },
      init(){
        for(const [key, value] of Object.entries(this.initState())) {
          this[key] = value
        }
        if (window.innerWidth <= 480) {
          this.showMenu = false
          this.showProjectActivity = false
        }
        this.$api.user.me().then((me)=>{
          this.me = me
        }).finally(()=>{
          this.loading = false
        })
        if (this.currentProjectId !== this.$u.rtr.project()) {
          this.currentProjectId = this.$u.rtr.project()
          this.projectActivity = []
          this.moreProjectActivity = false
          if (this.$u.rtr.project() != null) {
            this.$api.project.getActivities(this.$u.rtr.host(), this.$u.rtr.project()).then((res)=>{
              this.projectActivity = res.set
              this.moreProjectActivity = res.more
            })
          }
        }
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
        this.$u.rtr.goto(path)
        if (window.innerWidth <= 480) {
          this.showMenu = false
        }
      }
    },
    mounted(){
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
    &.show-project-activity{
      @media only screen and (min-width: 480px) {
        left: 0c;
        width: calc(100% - 15.1pc);
      }
    }
    &.show-menu.show-project-activity{
      @media only screen and (min-width: 480px) {
        left: 15.1pc;
        width: calc(100% - 30.2pc);
      }
    }
  }
  > .slide-out-toggle{
    position: absolute;
    top: 0;
    height: 2.5pc;
    width: 2.5pc;
    cursor: pointer;
    background-color: $inputColor;
    &:hover{
      background-color: $inputHoverColor
    }
    &:active{
      background-color: $inputActiveColor
    }
    img {
      margin: 0.25pc;
      background-color: transparent;
      width: 2pc;
      height: 2pc;
    }
    @include border();
  }
  > .menu-toggle{
    left: 0;
    display: flex;
    flex-direction: column;
    justify-content: center;
    align-items: center;
    > div{
      margin: 0.1pc;
      width: 1.8pc;
      height: 0.35pc;
      background-color: $color;
    }
  }
  > .project-activity-toggle {
    right: 0;
  }
  > .slide-out {
    position: absolute;
    top: 0;
    padding-top: 2.6pc;
    overflow: hidden;
    width: 0;
    height: 100%;
    &.show{
      width: 100%;
      overflow-y: auto;
      @media only screen and (min-width: 480px) {
        width: 15pc;
      }
    }
  }
  > .menu{
    left: 0;
    &.show{
      @include border($dir: right);
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
          background: $inputColor;
          width: 100%;
          cursor: pointer;
          &:hover{
            background-color: $inputHoverColor
          }
          &:active{
            background-color: $inputActiveColor
          }
          >span, >input {
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
  > .project-activity{
    right: 0;
    &.show{
      @include border($dir: left);
    }
    > .entry{
      width: 100%;
      overflow-wrap: anywhere;
      margin-bottom: 1pc;
      > p {
        margin: 0.2pc 0;
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
    &:disabled {
      background-color: $bgColor;
    }
}
</style>