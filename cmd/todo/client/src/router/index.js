import vue from 'vue'
import vueRouter from 'vue-router'
import init from '../views/init.vue'
import register from '../views/register.vue'
import login from '../views/login.vue'
import lists from '../views/lists.vue'
import list from '../views/list.vue'

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
  },
  {
    path: '/list/:id',
    name: 'list',
    component: list
  }
]

const router = new vueRouter({
  routes
})

export default router
