<template>
  <div class="root">
    <p>Logging In...</p>
  </div>
</template>

<script>
export default {
  name: "loginLinkLogin",
  data: function () {
    let query = this.$router.currentRoute.query;
    this.$api.user.loginLinkLogin(query.me, query.code).then(() => {
      this.$u.rtr.goHome();
    });
    return {};
  },
};
</script>

<style scoped lang="scss">
</style>