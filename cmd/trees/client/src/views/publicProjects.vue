<template>
  <div class="root">
    <div class="header">
      <h1>Latest Public Projects</h1>
    </div>
    <p v-if="loading">loading projects</p>
    <div v-else>
      <div class="projects">
        <table>
          <tr class="header">
            <th colspan="1" rowspan="2">host</th>
            <th
              :colspan="s.cols.length"
              :rowspan="s.cols.length == 1 ? 2 : 1"
              :class="s.name + ' ' + (index % 2 !== 0 ? 'light' : 'dark')"
              v-for="(s, index) in sections"
              :key="index"
            >
              {{ s.name }}
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
            v-for="p in ps"
            :key="p.id"
          >
            <td class="host">
              <user :userId="p.host"></user>
            </td>
            <td
              :class="c.name + ' ' + c.sectionClass"
              v-for="(c, index) in cols"
              :key="index"
            >
              {{ c.get(p) }}
            </td>
          </tr>
        </table>
      </div>
    </div>
  </div>
</template>

<script>
import user from "../components/user";
export default {
  name: "publicProjects",
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
          if (idx % 2 === 0) {
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
          if (idx % 2 === 0) {
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
        ps: [],
        loading: true,
        commonSections: [
          {
            name: "name",
            show: () => true,
            cols: [
              {
                name: "name",
                get: (p) => this.$u.fmt.ellipsis(p.name, 30),
              },
            ],
          },
          {
            name: "date",
            show: () => this.$root.show.date,
            cols: [
              {
                name: "created",
                get: (p) => this.$u.fmt.date(p.createdOn),
              },
              {
                name: "start",
                get: (p) => this.$u.fmt.date(p.startOn),
              },
              {
                name: "end",
                get: (p) => this.$u.fmt.date(p.endOn),
              },
              {
                name: "hrs/day",
                get: (p) => p.hoursPerDay,
              },
              {
                name: "days/wk",
                get: (p) => p.daysPerWeek,
              },
            ],
          },
          {
            name: "time",
            show: () => this.$root.show.time,
            cols: [
              {
                name: "min",
                get: (p) =>
                  this.$u.fmt.time(
                    p.timeEst + p.timeSubMin,
                    p.hoursPerDay,
                    p.daysPerWeek
                  ),
              },
              {
                name: "est",
                get: (p) =>
                  this.$u.fmt.time(
                    p.timeEst + p.timeSubEst,
                    p.hoursPerDay,
                    p.daysPerWeek
                  ),
              },
              {
                name: "inc",
                get: (p) =>
                  this.$u.fmt.time(
                    p.timeInc + p.timeSubInc,
                    p.hoursPerDay,
                    p.daysPerWeek
                  ),
              },
            ],
          },
          {
            name: "cost",
            show: () => this.$root.show.cost,
            cols: [
              {
                name: "est",
                get: (p) =>
                  this.$u.fmt.currencySymbol(p.currencyCode) +
                  this.$u.fmt.cost(p.costEst + p.costSubEst, true),
              },
              {
                name: "inc",
                get: (p) =>
                  this.$u.fmt.currencySymbol(p.currencyCode) +
                  this.$u.fmt.cost(p.costInc + p.costSubInc, true),
              },
            ],
          },
          {
            name: "file",
            show: () => this.$root.show.file,
            cols: [
              {
                name: "n",
                get: (p) => p.fileN + p.fileSubN,
              },
              {
                name: "size",
                get: (p) => this.$u.fmt.bytes(p.fileSize + p.fileSubSize),
              },
            ],
          },
          {
            name: "task",
            show: () => this.$root.show.task,
            cols: [
              {
                name: "childn",
                get: (p) => p.childN,
              },
              {
                name: "descn",
                get: (p) => p.descN,
              },
            ],
          },
        ],
      };
    },
    init() {
      this.$u.copyProps(this.initState(), this);
      this.$api.project
        .getLatestPublic()
        .then((res) => {
          this.ps = res.set;
        })
        .finally(() => {
          this.loading = false;
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