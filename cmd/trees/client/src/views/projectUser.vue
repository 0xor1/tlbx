<template>
  <div class="root">
    <a
      href="https://github.com/0xor1/tlbx/blob/develop/cmd/trees/client/src/views/projectUser.vue#L3"
      >TODO</a
    >
  </div>
</template>

<script>
export default {
  name: "tree",
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