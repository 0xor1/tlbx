import Vue from 'vue'
import app from './app.vue'
import router from './router'
import dayjs from 'vue-dayjs'
import toasted from 'vue-toasted'
import api from '@/api'
import fmt from '@/fmt'


Vue.use({install: v => v.prototype.$api = api})
Vue.use({install: v => v.prototype.$goToMyProjects = ()=>{
  api.user.me().then((me)=>{
    router.push(`/host/${me.id}/projects`)
  })
}})
Vue.use(fmt)
Vue.use(dayjs, {lang: 'en'})
Vue.use(toasted, {
  position: 'bottom-right',
  duration: 3000,
})

Vue.config.productionTip = false

let v = new Vue({
  data() {
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
        ctx: this.ctx || {
          host: null,
          project: null,
          task: null,
          pUser: null,
          pUserPromise: null
        }
      }
      if (this.$router.currentRoute.params.host !== res.ctx.host) {
        res.ctx.host = this.nullOr(this.$router.currentRoute.params.host)
      }
      if (this.$router.currentRoute.params.project !== res.ctx.project) {
        res.ctx.pUser = null
        res.ctx.project = this.nullOr(this.$router.currentRoute.params.project)
        if (res.ctx.project != null) {
          res.ctx.pUserPromise = this.$api.project.getMe(res.ctx.host, res.ctx.project).then((me)=>{
            this.ctx.pUser = me
          }).finally(()=>{
            this.ctx.pUserPromise = null
          })
        }
      }
      if (this.$router.currentRoute.params.task !== res.ctx.task) {
        res.ctx.task = this.nullOr(this.$router.currentRoute.params.task)
      }
      return res
    },
    init(){
      for(const [key, value] of Object.entries(this.initState())) {
        this[key] = value
      }
    },
    getCtx(){
      if (this.ctx.pUserPromise) {
        return this.ctx.pUserPromise.finally(()=>{
          return this.ctx
        })
      }
      return new Promise((resolve)=>{
        resolve(this.ctx)
      })
    },
    // set undefined to null
    nullOr(val){
      if (val == null) {
        return null
      }
      return val
    },
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
