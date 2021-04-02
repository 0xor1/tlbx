import vue from 'vue'
import vueRouter from 'vue-router'
import activate from '../views/activate.vue'
import confirmChangeEmail from '../views/confirmChangeEmail.vue'
import register from '../views/register.vue'
import login from '../views/login.vue'
import projects from '../views/projects.vue'
import projectUser from '../views/projectUser.vue'
import publicProjects from '../views/publicProjects.vue'
import task from '../views/task.vue'
import apiDocs from '../views/apiDocs.vue'
import api from '@/api'

vue.use(vueRouter)

// if session is authed and going to login or register
// redirect to my projects
const authedCheck = (to, from, next) => {
  let me = null
  api.user.me().then((res) => {
    me = res
  }).finally(() => {
    if (me != null) {
      next(`/host/${me.id}/projects`)
      return
    }
    next()
  })
}

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
    path: '/public/projects',
    name: 'publicProjects',
    component: publicProjects
  },
  {
    path: '/host/:host/projects',
    name: 'projects',
    component: projects
  },
  {
    path: '/host/:host/project/:project/user/:user',
    name: 'projectUser',
    component: projectUser
  },
  {
    path: '/host/:host/project/:project/task/:task',
    name: 'task',
    component: task
  },
  {
    path: '/api/docs',
    name: 'apiDocs',
    component: apiDocs
  },
  {
    path: '*',
    redirect: '/login'
  }
]

const router = new vueRouter({
  routes
})

export default router
