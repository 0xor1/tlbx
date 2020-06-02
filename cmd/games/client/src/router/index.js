import vue from 'vue'
import vueRouter from 'vue-router'
import init from '../views/init.vue'
import games from '../views/games.vue'
import blockers from '../views/blockers.vue'

vue.use(vueRouter)

const routes = [
  {
    path: '/',
    name: 'init',
    component: init
  },
  {
    path: '/games',
    name: 'games',
    component: games
  },
  {
    path: '/blockers/:id',
    name: 'blockers',
    component: blockers
  }
]

const router = new vueRouter({
  routes
})

export default router
