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
          dates: false,
          times: true,
          expenses: true,
          files: false,
          tasks: false
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
      for(const [key, value] of Object.entries(this.initState())) {
        this[key] = value
      }
      this.$u._main_init_utils(this)
      if (this._ctx.currentProjectId !== this.$u.rtr.project()) {
        this._ctx.currentProjectId = this.$u.rtr.project()
        this._ctx.loading = true
        this._ctx.pMe = null
        this._ctx.project = null
        let mapi = this.$api.newMDoApi()
        mapi.project.one(this.$u.rtr.host(), this.$u.rtr.project()).then((p)=>{
          this._ctx.project = p
        })
        mapi.project.getMe(this.$u.rtr.host(), this.$u.rtr.project()).then((pMe)=>{
          this._ctx.pMe = pMe
        })
        mapi.sendMDo().finally(()=>{
          this._ctx.loading = false
        })
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
