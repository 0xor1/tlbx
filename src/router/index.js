import vue from 'vue'
import vueRouter from 'vue-router'
import activate from '../views/activate.vue'
import confirmChangeEmail from '../views/confirmChangeEmail.vue'
import register from '../views/register.vue'
import login from '../views/login.vue'
import projects from '../views/projects.vue'
import projectCreate from '../views/projectCreate.vue'
import task from '../views/task.vue'
import api from '@/api'

vue.use(vueRouter)

const routes = [
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
    path: '/project/create',
    name: 'projectCreate',
    component: projectCreate
  },
  {
    path: '/projects',
    name: 'projects',
    component: projects
  },
  {
    path: '/host/:hostId/project/:projectId/task/:taskId',
    name: 'task',
    component: task
  },
  {
    path: '*',
    redirect: '/projects'
  }
]

const router = new vueRouter({
  routes
})
router.beforeEach((to, from, next)=>{
  api.user.me().then(()=>{
    next()
  }).catch(()=>{
    if (to.name != 'login' && 
    to.name != 'register') {
      next('/login')
    } else {
      next()
    }
  })
})

export default router
