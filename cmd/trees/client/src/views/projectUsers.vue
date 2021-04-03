<template>
  <div class="root">
    <p v-if="loading">loading..</p>
    <div v-else class="content">
      <h1>Project Users</h1>
      <table>
        <tr class="header">
          <th colspan="1" rowspan="2">user</th>
          <th
            :colspan="s.cols.length"
            :rowspan="s.cols.length == 1 ? 2 : 1"
            :class="s.name + ' ' + (index % 2 === 0 ? 'light' : 'dark')"
            v-for="(s, index) in sections"
            :key="index"
          >
            {{ s.name() }}
          </th>
        </tr>
        <tr class="header">
          <th
            :class="c.sectionClass"
            v-for="(c, index) in colHeaders"
            :key="index"
          >
            {{ c.name }}
          </th>
        </tr>
        <tr
          class="row"
          @click="$u.rtr.goto(`/host/${p.host}/project/${p.id}/task/${p.id}`)"
          v-for="u in users"
          :key="u.id"
        >
          <td class="user">
            <user :userId="u.id"></user>
          </td>
          <td
            :class="c.name + ' ' + c.sectionClass"
            v-for="(c, index) in cols"
            :key="index"
          >
            {{ c.get(u) }}
          </td>
        </tr>
      </table>
      <button class="load-more" v-if="more" @click="loadMore()">
        load more
      </button>
    </div>
  </div>
</template>

<script>
import user from "../components/user";
export default {
  name: "projectUsers",
  components: { user },
  data: function () {
    return this.initState();
  },
  computed: {
    sections() {
      return this.commonSections.filter((i) => i.show());
    },
    colHeaders() {
      let res = [];
      this.sections.forEach((section, idx) => {
        section.cols.forEach((col) => {
          if (idx % 2 === 1) {
            col.sectionClass = "dark";
          } else {
            col.sectionClass = "light";
          }
        });
        if (section.cols.length > 1) {
          res = res.concat(section.cols);
        }
      });
      return res;
    },
    cols() {
      let res = [];
      this.sections.forEach((section, idx) => {
        section.cols.forEach((col) => {
          if (idx % 2 === 1) {
            col.sectionClass = "dark";
          } else {
            col.sectionClass = "light";
          }
        });
        res = res.concat(section.cols);
      });
      return res;
    },
  },
  methods: {
    initState() {
      return {
        me: null,
        pMe: null,
        users: null,
        more: false,
        loading: true,
        loadingMore: false,
        deleteIndex: -2,
        commonSections: [
          {
            name: () => "time",
            show: () => this.$root.show.time,
            cols: [
              {
                name: "est",
                get: (p) =>
                  this.$u.fmt.time(p.timeEst, p.hoursPerDay, p.daysPerWeek),
              },
              {
                name: "inc",
                get: (p) =>
                  this.$u.fmt.time(p.timeInc, p.hoursPerDay, p.daysPerWeek),
              },
            ],
          },
          {
            name: () =>
              "cost " + this.$u.fmt.currencySymbol(this.project.currencyCode),
            show: () => this.$root.show.cost,
            cols: [
              {
                name: "est",
                get: (p) => this.$u.fmt.cost(p.costEst, true),
              },
              {
                name: "inc",
                get: (p) => this.$u.fmt.cost(p.costInc, true),
              },
            ],
          },
          {
            name: () => "file",
            show: () => this.$root.show.file && this.project.fileLimit > 0,
            cols: [
              {
                name: "n",
                get: (p) => p.fileN,
              },
              {
                name: "size",
                get: (p) => this.$u.fmt.bytes(p.fileSize),
              },
            ],
          },
          {
            name: () => "tasks",
            show: () => this.$root.show.task,
            cols: [
              {
                name: "n",
                get: (p) => p.taskN,
              },
            ],
          },
        ],
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
          this.$root.ctx().then((ctx) => {
            if (this.me != null) {
              this.$api.fcm.onMessage(this.fcmHandler);
            }
            this.pMe = ctx.pMe;
            this.project = ctx.project;
            this.$api.project
              .getUsers({
                host: this.$u.rtr.host(),
                project: this.$u.rtr.project(),
              })
              .then((res) => {
                this.users = res.set;
                this.more = res.more;
              })
              .finally(() => {
                this.loading = false;
              });
          });
        });
    },
    loadMore() {
      if (this.loadingMore) {
        return;
      } else {
        this.loadingMore = true;
        this.$api.project
          .getUsers({
            host: this.$u.rtr.host(),
            project: this.$u.rtr.project(),
            after: this.users[this.users.length - 1].id,
          })
          .then((res) => {
            this.users = this.users.concat(res.set);
            this.more = res.more;
          })
          .finally(() => {
            this.loadingMore = false;
          });
      }
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
div.root {
  & > .content {
    table {
      td {
        &.user {
          text-align: left;
        }
      }
    }
  }
}
</style>