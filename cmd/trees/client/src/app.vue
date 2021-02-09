<template>
    <div class="app-root">
        <div v-bind:class="{'show-menu':showMenu, 'show-project-activity': showProjectActivity && showProjectActivityToggle}" class="body">
            <router-view @refreshProjectActivity="refreshProjectActivity"/>
        </div>
        <!-- main menu -->
        <div v-if="showMenu" v-bind:class="{'show':showMenu}" class="menu slide-out">
          <div class="opts">
            <div class="btn" v-if="me" @click.stop.prevent="goHome()">
              home
            </div>
            <div class="btn" @click.stop.prevent="showFields=!showFields">
              show
            </div>
            <div class="fields" v-if="showFields">
              <div v-for="(f, index) in availFields" :key="index" @click.stop.prevent="$root.show[f] = !$root.show[f]">
                <span>{{f}}<input @click.stop type="checkbox" v-model="$root.show[f]"></span>
              </div>
            </div>
            <div class="btn" @click.stop.prevent="loginout">
              {{ me? 'logout': 'login'}}
            </div>
          </div>

          <div title="report a bug" class="bug">
            <a href="https://github.com/0xor1/tlbx/issues/new?assignees=0xor1&labels=bug%2C+trees&template=trees-bug-report.md&title=%5Btrees%5D+%5Bbug%5D" target="_blank"><img src="@/assets/bug.svg"></a>
          </div>
        </div>
        <div title="menu" @click.stop.prevent="showMenu=!showMenu" class="slide-out-toggle menu-toggle">
          <div></div>
          <div></div>
          <div></div>
        </div>
        <!-- project activity stream -->
        <div v-if="showProjectActivityToggle && showProjectActivity" v-bind:class="{'show':showProjectActivity}" class="project-activity slide-out">
          <div class="exclude-deleted" @click.stop="toggleProjectActivityExcludeDeleted"><input id="hide-deleted" :checked="projectActivityExcludeDeleted" type="checkbox"><label for=""> hide deleted</label></div>
          <div v-if="me != null" class="enable-realtime" @click.stop.prevent="toggleRealtime"><input id="enable-realtime" :checked="realtimeEnabled" type="checkbox"><label for=""> realtime</label></div>        
          <div class="entries">
            <div :class="{entry: true, 'task-deleted': a.taskDeleted, deleted: a.itemDeleted}" v-for="(a, index) in projectActivity" :key="index" @click.stop.prevent="gotoActivityTask(a)">
              <user :userId="a.user"></user> 
              <span v-if="a.itemType == `task`">
                {{a.action}} {{a.itemType}} {{a.taskName}}
              </span>
              <span v-else>
                {{a.action}} {{a.itemType}}<span v-if="a.itemName != null"> {{a.itemName}}</span>, on task {{a.taskName}}
              </span>
              <br>
              <span class="datetime"> {{$u.fmt.datetime(a.occurredOn)}}</span>
            </div>
          </div>
          <button v-if="!loadingProjectActivity && moreProjectActivity" @click.stop.prevent="loadMoreProjectActivity">load more</button>
          <div v-if="loadingProjectActivity && projectActivity.length > 0">loading...</div>
        </div>
        <div title="project activity" v-if="showProjectActivityToggle" @click.stop.prevent="projectActivityToggle()" class="slide-out-toggle project-activity-toggle">
          <img src="@/assets/activity.svg">
        </div>
    </div>
</template>

<script>
  import user from './components/user'
  export default {
    name: 'app',
    components: {user},
    computed: {
      availFields(){
        return this.fields.filter((f)=>{
          return f != "file" || this.project == null || this.project.fileLimit > 0
        })
      }
    },
    data() {
      return this.initState()
    },
    methods: {
      initState(){
        return {
          loading: true,
          currentProjectId: this.currentProjectId || null,
          project: null,
          showMenu: this.showMenu || false,
          showProjectActivity: this.showProjectActivity || false,
          showProjectActivityToggle: this.$u.rtr.project() != null,
          showFields: this.showFields || false,
          projectActivity: this.projectActivity || [],
          moreProjectActivity: this.moreProjectActivity || false,
          projectActivityExcludeDeleted: this.projectActivityExcludeDeleted === false? false: true,
          realtimeEnabled: Notification.permission === "granted",
          loadingProjectActivity: false,
          projectActivityLastGotOn: null,
          projectActivityCurrentPollDelayMs: 60000,
          projectActivityMinPollDelayMs: 60000,
          projectActivityMaxPollDelayMs: 3600000,
          projectActivitySetTimeoutId: null,
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
        this.$u.copyProps(this.initState(), this)
        this.$api.fcm.onLogout(()=>{
          this.loginout()
        })
        this.$api.fcm.onEnabled(()=>{
          if (!this.realtimeEnabled) {
            this.toggleRealtime()
          }
        })
        this.$api.fcm.onDisabled(()=>{
          if (this.realtimeEnabled) {
            this.toggleRealtime()
          }
        })
        if (window.innerWidth <= 480) {
          this.showMenu = false
          this.showProjectActivity = false
        }
        this.$api.user.me().then((me)=>{
          this.me = me
          this.$root.ctx().then((ctx)=>{
            this.project = ctx.project
          }).finally(()=>{
            this.loading = false
          })
        })
        if (this.currentProjectId !== this.$u.rtr.project()) {
          this.currentProjectId = this.$u.rtr.project()
          this.projectActivity = []
          this.moreProjectActivity = false
          this.projectActivityLastGotOn = null
          this.refreshProjectActivity()
        }
      },
      toggleRealtime(){
        if (this.togglingRealtimeEnabled) {
          return
        }
        this.togglingRealtimeEnabled = true
        if (this.realtimeEnabled) {
          this.$api.user.setFCMEnabled(false).then(()=>{
            this.realtimeEnabled = false
          }).finally(()=>{
            this.togglingRealtimeEnabled = false
          })
        } else {
          this.$api.fcm.init(true).then(()=>{
            this.$api.user.setFCMEnabled(true).then(()=>{
              this.realtimeEnabled = true
              return this.$api.user.registerForFCM({
                topic: [this.$u.rtr.host(), this.$u.rtr.project()]
              }).then(()=>{
                this.realtimeEnabled = true
              }).finally(()=>{
              this.togglingRealtimeEnabled = false
            })
            }).finally(()=>{
              this.togglingRealtimeEnabled = false
            })
          })
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
      gotoActivityTask(a){
        if (!a.taskDeleted) {
          this.$u.rtr.goto(`/host/${this.$u.rtr.host()}/project/${this.$u.rtr.project()}/task/${a.task}`)
          if (window.innerWidth <= 480) {
            this.showProjectActivity = false
          }
        }
      },
      goto(path){
        this.$u.rtr.goto(path)
        if (window.innerWidth <= 480) {
          this.showMenu = false
        }
      },
      toggleProjectActivityExcludeDeleted(){
        if (!this.loadingProjectActivity) {
          this.projectActivityExcludeDeleted = !this.projectActivityExcludeDeleted
          this.refreshProjectActivity(true)
        }
      },
      loadMoreProjectActivity(){
        if (this.$u.rtr.project() != null && this.showProjectActivity && this.moreProjectActivity) {
          this.loadingProjectActivity = true
          this.$api.project.getActivities({
            host: this.$u.rtr.host(),
            project: this.$u.rtr.project(),
            excludeDeletedItems: this.projectActivityExcludeDeleted,
            occurredBefore: this.projectActivity[this.projectActivity.length - 1].occurredOn
          }).then((res)=>{
            this.loadingProjectActivity = false
            this.projectActivity = this.projectActivity.concat(res.set)
            this.moreProjectActivity = res.more
          })
        }
      },
      projectActivityToggle() {
        this.showProjectActivity = !this.showProjectActivity
        if (this.showProjectActivity) {
          // if opening project activity refresh data
          this.refreshProjectActivity();
        } else {
          // else if closing then clear polling timeout
          clearTimeout(this.projectActivitySetTimeoutId)
          this.projectActivitySetTimeoutId = null
        }
      },
      refreshProjectActivity(force){
        if (force) {
          // if we're forcing an update then
          // set lastGotOn to null to refresh now.
          this.projectActivityLastGotOn = null
          clearTimeout(this.projectActivitySetTimeoutId)
          this.projectActivitySetTimeoutId = null
        }
        if (this.$u.rtr.project() != null && 
          this.showProjectActivity && 
          this.projectActivitySetTimeoutId == null) {
          if (this.projectActivityLastGotOn + this.projectActivityCurrentPollDelayMs >= Date.now()) {
            // set up correct poll again!
            let delay = this.projectActivityLastGotOn + this.projectActivityCurrentPollDelayMs + 1000 - Date.now()
            this.projectActivitySetTimeoutId = setTimeout(()=>{
              this.projectActivitySetTimeoutId = null
              this.refreshProjectActivity()
            }, delay)
          } else {
            this.projectActivityLastGotOn = Date.now()
            this.loadingProjectActivity = true
            this.$api.project.getActivities({
              host: this.$u.rtr.host(),
              project: this.$u.rtr.project(),
              excludeDeletedItems: this.projectActivityExcludeDeleted,
            }).then((res)=>{
              this.loadingProjectActivity = false
              if ((this.projectActivity.length == 0 && res.set.length == 0) ||
                (this.projectActivity.length > 0 &&
                res.set.length > 0 &&
                this.projectActivity[0].occurredOn == res.set[0].occurredOn)) {
                // there was no updated entries so half polling rate
                this.projectActivityCurrentPollDelayMs *= 2
                if (this.projectActivityCurrentPollDelayMs > this.projectActivityMaxPollDelayMs) {
                  this.projectActivityCurrentPollDelayMs = this.projectActivityMaxPollDelayMs
                }
              } else {
                // there was some new activity so reset polling rate to minimum rate
                this.projectActivityCurrentPollDelayMs = this.projectActivityMinPollDelayMs
              }
              this.projectActivity = res.set
              this.moreProjectActivity = res.more
              // poll again
              this.projectActivitySetTimeoutId = setTimeout(()=>{
                this.projectActivitySetTimeoutId = null
                this.refreshProjectActivity()
              }, this.projectActivityCurrentPollDelayMs+1000)
            })
          }
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
@import "./style.scss";

.app-root{
  > .body{
    position: absolute;
    width: 100%;
    height: 100%;
    overflow: auto;
    > :first-child {
      margin: 2.6pc 1.3pc 0 1.3pc;
    }
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
    > div {
      &.exclude-deleted, &.enable-realtime {
        margin-top: 0;
      }
    }
  }
  > .menu{
    left: 0;
    &.show{
      @include border($dir: right);
    }
    > .opts{
      margin-top: 2.6pc;
      display: flex;
      flex-direction: column;
      justify-content: flex-start;
      align-items: stretch;
      overflow-y: auto;
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
    > .bug {
      margin: 0;
      background: transparent;
      cursor: pointer;
      position: absolute;
      bottom: 0;
      left: 0.5pc;
      a img{
        width: 1.5pc;
        height: 1.5pc;
      }
    }
  }
  > .project-activity{
    right: 0;
    &.show{
      @include border($dir: left);
    }
    > .entries {
      > .entry{
        overflow-wrap: anywhere;
        padding: 0.6pc;
        &:not(.task-deleted) {
          cursor: pointer;
          &:hover {
            background-color: $inputActiveColor;
          }
        }
        &.deleted{
          background-color: #400;
          &:hover:not(.task-deleted) {
            background-color: #600;
          }
        }
        * {
          background-color: transparent;
        }
        .datetime{
          font-size: 0.7pc;
        }
        @include border($dir: 'bottom');
      }
    }
  }
  .err{
    color: #c33;
  }
  div.markdown {
    img{
      max-height: 20pc;
      max-width: 20pc;
    }
    pre {
        @include border();
        padding: 0.5pc;
        border-radius: 0.5pc;
        background: $inputActiveColor;
        code{
          background: transparent;
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
textarea, input, button {
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
    &.err {
    @include border($color: $errColor);
    }
}
</style>
