<template>
  <div class="root">
    <table>
      <tr></tr>
    </table>
  </div>
</template>

<script>
export default {
  name: "projectUsers",
  data: function () {
    return this.initState();
  },
  methods: {
    initState() {
      return {};
    },
    init() {
      this.$u.copyProps(this.initState(), this);
      this.$api.project.getUsers({
        host: this.$u.rtr.host(),
        project: this.$u.rtr.project(),
      });
    },
  },
  mounted() {
    this.init();
  },
  watch: {
    $route() {
      this.init();
    },
  },
};
</script>

<style lang="scss" scoped>
</style>