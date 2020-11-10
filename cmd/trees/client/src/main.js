import Vue from 'vue'
import app from './app.vue'
import router from './router'
import dayjs from 'vue-dayjs'

Vue.config.productionTip = false

Vue.use(dayjs)

new Vue({
  router,
  render: h => h(app)
}).$mount('#app')
