<template>
  <div class="root">
    <div v-if="notFound">
      <notfound :type="'task'"></notfound>
    </div>
    <div v-else-if="loading" class="loading">loading...</div>
    <div v-else class="content">
      <node
        :project="project"
        :id="project.id"
        :tasks="tasks"
        :showFullSubTree="false"
      ></node>
    </div>
  </div>
</template>

<script>
import notfound from "../components/notfound";
import node from "../components/node";
export default {
  name: "tree",
  components: { notfound, node },
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
@import "../style.scss";
div.root {
  > .content {
  }
}
</style>