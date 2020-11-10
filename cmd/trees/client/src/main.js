import vue from 'vue'
import app from './app.vue'
import router from './router'
import dayjs from 'vue-dayjs'

vue.config.productionTip = false

vue.use(dayjs)

new vue({
  router,
  render: h => h(app)
}).mount('#app')
