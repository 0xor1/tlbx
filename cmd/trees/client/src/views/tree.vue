<template>
  <div class="root">
    <div v-if="notFound">
      <notfound :type="'task'"></notfound>
    </div>
    <div v-else-if="loading" class="loading">loading...</div>
    <div v-else class="content">
      <div class="tools">
        <img
          title="users"
          class="icon"
          src="@/assets/users.svg"
          @click.stop.prevent="
            $u.rtr.goto(
              `/host/${$u.rtr.host()}/project/${$u.rtr.project()}/users`
            )
          "
        />
      </div>
      <div class="breadcrumb">
        <span>
          <user :goToHome="true" :userId="$u.rtr.host()"></user>
          :
        </span>
        <span
          v-if="ancestors.set.length > 0 && ancestors.set[0].parent != null"
        >
          <a
            title="load more ancestors"
            href=""
            @click.stop.prevent="loadMoreAncestors"
            >..</a
          >
          /
        </span>
        <span v-for="a in ancestors" :key="a.id">
          <a
            :title="a.name"
            :href="
              '/#/host/' +
              $u.rtr.host() +
              '/project/' +
              $u.rtr.project() +
              '/tree/' +
              a.id
            "
            >{{ $u.fmt.ellipsis(a.name, 20) }}</a
          >
          /
        </span>
        <span>
          {{ $u.fmt.ellipsis(t0.name, 20) }}
        </span>
      </div>
      <div class="tree">
        
      </div>
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
  computed: {
    t0() {
      return this.tasks[this.$u.rtr.task()];
    },
    ancestors() {
      return [];
    },
  },
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