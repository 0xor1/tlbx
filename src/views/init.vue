<template>
  <p>checking session status</p>
</template>

<script>
  import api from '@/api'
  import router from '@/router'
  export default {
    name: 'init',
    data () {
      api.user.me().then(() => {
        router.push('/lists')
      }).catch(() => {
        router.push('/login')
      })
      return {}
    }
  }
</script>