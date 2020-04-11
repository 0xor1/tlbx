<template>
  <div></div>
</template>

<script>
  import api from '@/api'
  import router from '@/router'
  export default {
    name: 'init',
    data () {
      api.me.get().then(() => {
        console.log("active session")
        router.push('/lists')
      }).catch(() => {
        console.log("missing session")
        router.push('/login')
      })
      return {}
    }
  }
</script>