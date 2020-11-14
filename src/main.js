import Vue from 'vue'
import app from './app.vue'
import router from './router'
import dayjs from 'vue-dayjs'
import toasted from 'vue-toasted'
import api from '@/api'
 
Vue.use(dayjs)
Vue.use(toasted, {
  position: 'bottom-right',
  duration: 3000,
})

Vue.config.productionTip = false


let v = new Vue({
  router,
  render: h => h(app)
})
v.$mount('#app')
api.setGlobalErrorHandler(v.$toasted.error)
