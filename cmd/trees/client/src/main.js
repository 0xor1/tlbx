import Vue from 'vue'
import app from './app.vue'
import router from './router'
import dayjs from 'vue-dayjs'
import toasted from 'vue-toasted'
import api from '@/api'
import util from '@/util'

Vue.use({install: v => v.prototype.$api = api})
Vue.use(util)
Vue.use(dayjs, {lang: 'en'})
Vue.use(toasted, {
  position: 'bottom-right',
  duration: 3000,
})

Vue.config.productionTip = false

let v = new Vue({
  data() {
    this.$u._main_init_utils(this)
    return this.initState()
  },
  methods: {
    initState(){
      let res = {
        show: this.show || {
          date: false,
          user: false,
          time: true,
          cost: true,
          file: true,
          task: true
        },
        _ctx: this._ctx || {
          currentProjectId: null,
          loading: false,
          pMe: null,
          project: null
        }
      }
      return res
    },
    init(){
      this.$u.copyProps(this.initState(), this)
      this.$u._main_init_utils(this)
      if (this._ctx.currentProjectId !== this.$u.rtr.project()) {
        this._ctx.currentProjectId = this.$u.rtr.project()
        this._ctx.pMe = null
        this._ctx.project = null
        if (this._ctx.currentProjectId != null) {
          this._ctx.loading = true
          let mapi = this.$api.newMDoApi()
          mapi.project.one({
            host: this.$u.rtr.host(), 
            id: this.$u.rtr.project()
          }).then((p)=>{
            this._ctx.project = p
          })
          mapi.project.getMe({
            host: this.$u.rtr.host(), 
            project: this.$u.rtr.project()
          }).then((pMe)=>{
            this._ctx.pMe = pMe
          })
          mapi.sendMDo().finally(()=>{
            this._ctx.loading = false
            if (this._ctx.pMe != null && this._ctx.project != null) {
              this.$api.project.registerForFCM({
                host: this.$u.rtr.host(), 
                id: this.$u.rtr.project(),
              }).then((fcm)=>{
                fcm.onMessage((msg)=>{
                  console.log(msg)
                })
                
              }).catch((err)=>{
                console.log(err)
              })
            }
          })
        }
      }
    },
    ctx(){
      let completer = null
      completer = (resolve)=>{
        if (this._ctx.loading) {
          setTimeout(completer, 10, resolve)
        } else {
          resolve({
            pMe: this._ctx.pMe,
            project: this._ctx.project
          })
        }
      }
      return new Promise(completer)
    }
  },
  mounted(){
    this.init()
  },
  watch: {
    $route () {
      this.init()
    }
  },
  router,
  render: h => h(app)
})
v.$mount('#app')
api.setGlobalErrorHandler(v.$toasted.error)
