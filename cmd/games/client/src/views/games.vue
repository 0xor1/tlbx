<template>
  <div class="root">
    <button class="game" @click.stop.prevent="router.push('/blockers/new')">
      blockers
    </button>
  </div>
</template>

<script>
  import router from '@/router'
  export default {
    name: 'games',
    data: function() {
      return {router}
    }
  }
</script>

<style lang="scss">

</style>