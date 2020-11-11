<template>
  <p>checking session status</p>
</template>

<script>
  import api from '@/api'
  import router from '@/router'
  export default {
    name: 'init',
    data () {
      api.user.me().then((me) => {
        router.push('/host/'+me.id+'/projects')
      }).catch(() => {
        router.push('/login')
      })
      return {}
    }
  }
</script>