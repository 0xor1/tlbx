<template>
  <div class="lists">
    <h1>lists</h1>
  </div>
</template>

<script>
  import api from '@/api'
  export default {
    name: 'lists',
    data () {
      api.list.get().then((res) => {
        console.log(res)
      }).catch((err) => {
        console.log(err)
      })
      return {}
    }
  }
</script>