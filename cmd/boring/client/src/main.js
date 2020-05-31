import Vue from 'vue'
import app from './app.vue'
import router from './router'
import moment from 'vue-moment'

Vue.config.productionTip = false

Vue.use(moment)

new Vue({
  router,
  render: h => h(app)
}).$mount('#app')
