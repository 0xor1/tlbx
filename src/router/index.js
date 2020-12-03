import vue from 'vue'
import vueRouter from 'vue-router'
import direct from '../views/direct.vue'
import activate from '../views/activate.vue'
import confirmChangeEmail from '../views/confirmChangeEmail.vue'
import register from '../views/register.vue'
import login from '../views/login.vue'
import projects from '../views/projects.vue'
import projectCreate from '../views/projectCreate.vue'
import task from '../views/task.vue'
import api from '@/api'

vue.use(vueRouter)

// if session not authed redirect to login
const notAuthedCheck = (to, from, next)=>{
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
}

// if session is authed and going to login or register
// redirect to my projects
const authedCheck = (to, from, next)=>{
  api.user.me().then((me)=>{
    next(`/host/${me.id}/projects`)
  }).catch(()=>{
    next()
  })
}

const routes = [
  {
    path: '/',
    name: 'direct',
    component: direct
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
    component: register,
    beforeEnter: authedCheck
  },
  {
    path: '/login',
    name: 'login',
    component: login,
    beforeEnter: authedCheck
  },
  {
    path: '/project/create',
    name: 'projectCreate',
    component: projectCreate,
    beforeEnter: notAuthedCheck
  },
  {
    path: '/host/:hostId/projects',
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
    redirect: '/'
  }
]

const router = new vueRouter({
  routes
})

export default router
