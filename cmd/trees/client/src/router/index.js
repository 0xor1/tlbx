import vue from 'vue'
import vueRouter from 'vue-router'
import init from '../views/init.vue'
import activate from '../views/activate.vue'
import confirmChangeEmail from '../views/confirmChangeEmail.vue'
import register from '../views/register.vue'
import login from '../views/login.vue'
import projects from '../views/projects.vue'
import task from '../views/task.vue'

vue.use(vueRouter)

const routes = [
  {
    path: '/',
    name: 'init',
    component: init
  },
  {
    path: '/activate',
    name: 'activate',
    component: activate
  },
  {
    path: '/confirmChangeEmail',
    name: 'confirmChangeEmail',
    component: confirmChangeEmail
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
    path: '/projects',
    name: 'projects',
    component: projects
  },
  {
    path: '/task/:id',
    name: 'task',
    component: task
  }
]

const router = new vueRouter({
  routes
})

export default router
