<template>
  <div class="list">
    <h1>list</h1>
  </div>
</template>

<script>
  import api from '@/api'
  export default {
    name: 'list',
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