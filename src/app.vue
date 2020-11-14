<template>
    <router-view/>
</template>
<script>
  import api from '@/api'
  import router from '@/router'
  export default {
    name: 'root',
    data () {
      api.user.me().catch(() => {
        router.push('/login')
      })
      return {}
    }
  }
</script>
<style lang="scss">
@import url(https://fonts.googleapis.com/css2?family=Roboto+Mono:wght@400;700&display=swap);
* {
    font-family: 'Roboto Mono', monospace;
    background-color: #000;
    color: #ddd;
    button {
        cursor: pointer;
        &:hover {
            background-color: #555;
        }
        &:active {
            background-color: #222;
        }
    }
    input, button {
        background-color: #222;
        outline: none;
        border: 1px solid #fff;
        border-radius: 2px;
        padding: 5px;
        &::placeholder{
            color: #aaa;
        }
    }
}
</style>