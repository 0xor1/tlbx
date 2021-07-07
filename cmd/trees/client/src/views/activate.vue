<template>
  <div class="root">
    <p>Activating...</p>
  </div>
</template>

<script>
export default {
  name: "activate",
  data: function () {
    let query = this.$router.currentRoute.query;
    this.$api.user.activate(query.me, query.code).then(() => {
      this.$u.rtr.goHome();
    });
    return {};
  },
};
</script>

<style scoped lang="scss">
</style>