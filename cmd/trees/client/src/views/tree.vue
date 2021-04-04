<template>
  <div class="root"></div>
</template>

<script>
export default {
  name: "projectUser",
  data: function () {
    return this.initState();
  },
  methods: {
    initState() {
      return {};
    },
    init() {
      this.$u.copyProps(this.initState(), this);
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