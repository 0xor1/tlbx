<template>
  <div class="root">
    <div v-if="notFound">
      <notfound :type="'task'"></notfound>
    </div>
    <div v-else-if="loading" class="loading">loading...</div>
    <div v-else class="content">
      <div class="tree"></div>
    </div>
    <a
      href="https://github.com/0xor1/tlbx/blob/develop/cmd/trees/client/src/views/tree.vue#L3"
      >TODO</a
    >
  </div>
</template>

<script>
import notfound from "../components/notfound";
export default {
  name: "tree",
  components: { notfound },
  computed: {},
  data: function () {
    return this.initState();
  },
  methods: {
    initState() {
      return {
        notFound: false,
        loading: true,
        me: null,
        pMe: null,
        project: null,
        tasks: {},
      };
    },
    init() {
      this.$u.copyProps(this.initState(), this);
      this.$api.user
        .me()
        .then((me) => {
          this.me = me;
        })
        .finally(() => {
          this.$root
            .ctx()
            .then((ctx) => {
              if (this.me != null) {
                this.$api.fcm.onMessage(this.fcmHandler);
              }
              this.pMe = ctx.pMe;
              this.project = ctx.project;
              this.tasks[this.project.id] = this.project;
            })
            .finally(() => {
              this.loading = false;
            });
        });
    },
  },
  mounted() {
    this.init();
  },
  destroyed() {
    // remove fcm listener when leaving view
    this.$api.fcm.onMessage(null);
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