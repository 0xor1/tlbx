import vue from 'vue'
import vueRouter from 'vue-router'
import init from '../views/init.vue'
import register from '../views/register.vue'
import login from '../views/login.vue'
import lists from '../views/lists.vue'

vue.use(vueRouter)

const routes = [
  {
    path: '/',
    name: 'init',
    component: init
  },
  {
    path: '/register',
    name: 'register',
    component: register
  },
  {
    path: '/login',
    name: 'login',
    component: login
  },
  {
    path: '/lists',
    name: 'lists',
    component: lists
  }
]

const router = new vueRouter({
  routes
})

export default router
