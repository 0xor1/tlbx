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
        }
      }
      return res
    },
    init(){
      for(const [key, value] of Object.entries(this.initState())) {
        this[key] = value
      }
      this.$u._main_init_utils(this)
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
