<template>
  <p>checking for an active game</p>
</template>

<script>
  import api from '@/api'
  import router from '@/router'
  export default {
    name: 'init',
    data () {
      api.game.active().then((res) => {
        if (res != null) {
          router.push('/'+res.type+'/'+res.id)
        } else {
          router.push('/games')
        }
      }).catch(() => {
        router.push('/games')
      })
      return {}
    }
  }
</script>