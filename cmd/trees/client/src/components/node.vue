<template>
  <div class="root fx-c fx-start">
    <div class="this-node fx-c" v-if="tasks[id] != null">
      <span class="name">
        <span
          v-if="task.descN > 0"
          :title="task.isParallel ? 'parallel' : 'sequential'"
          :class="{ 'parallel-indicator': true, parallel: task.isParallel }"
          class="mr large"
          >{{ task.isParallel ? "&#8649;" : "&#8699;" }}</span
        ><a
          :href="`/#/host/${project.host}/project/${project.id}/task/${task.id}`"
          >{{ task.name }}</a
        >
      </span>
      <div v-if="$root.show.time" class="time">
        <img
          title="time"
          class="icon small mr mt"
          src="@/assets/sand-clock.svg"
        />
        <span title="minimum time" class="time-min" v-if="task.childN > 0">{{
          $u.fmt.time(
            task.timeEst + task.timeSubMin,
            project.hoursPerDay,
            project.daysPerWeek
          )
        }}</span>
        <span v-if="task.childN > 0"> / </span>
        <span title="estimated time" class="time-est">{{
          $u.fmt.time(
            task.timeEst + task.timeSubEst,
            project.hoursPerDay,
            project.daysPerWeek
          )
        }}</span>
        /
        <span title="incurred time" class="time-inc">{{
          $u.fmt.time(
            task.timeInc + task.timeSubInc,
            project.hoursPerDay,
            project.daysPerWeek
          )
        }}</span>
      </div>
      <div v-if="$root.show.cost" class="cost">
        <img
          title="cost"
          class="icon small mr mt"
          src="@/assets/calculator.svg"
        />
        <span title="estimated cost" class="cost-est">{{
          $u.fmt.currencySymbol(project.currencyCode) +
          $u.fmt.cost(task.costEst + task.costSubEst)
        }}</span>
        /
        <span title="incurred cost" class="cost-inc">{{
          $u.fmt.currencySymbol(project.currencyCode) +
          $u.fmt.cost(task.costInc + task.costSubInc)
        }}</span>
      </div>
      <div v-if="$root.show.file && project.fileLimit > 0" class="file">
        <img title="file" class="icon small mr mt" src="@/assets/file.svg" />
        <span title="used space" class="file-used">{{
          $u.fmt.bytes(task.fileSize + task.fileSubSize)
        }}</span>
        /
        <span title="file count" class="file-n">{{
          task.fileN + task.fileSubN
        }}</span>
      </div>
      <div v-if="task.descN > 0">
        <img
          title="sub tasks"
          class="icon small mr mt"
          src="@/assets/hierarchy.svg"
        />
        <span
          title="children"
          class="blue clk"
          href=""
          @click.stop.prevent="showHideChildren()"
          ><span class="dark-blue small">({{ showChildren ? "-" : "+" }})</span>
          {{ task.childN }}</span
        >
        /
        <span
          title="descendants"
          class="blue clk"
          v-if="task.descN <= 1000"
          @click.stop.prevent="showHideFullSubTree()"
        >
          <span class="dark-blue small"
            >({{ myShowFullSubTree ? "-" : "+" }})</span
          >
          {{ task.descN }}
        </span>
        <span v-else title="descendants" class="blue">{{ task.descN }}</span>
      </div>
      <div ref="scrollhandle"></div>
    </div>
    <div class="this-node" v-else>
      <button @click.stop.prevent="loadMeAndMore(id)">
        load next siblings
      </button>
    </div>
    <div v-if="showChildren && children.length > 0" class="fx-c">
      <div class="bb-stem"></div>
      <div
        class="children"
        :class="{
          //'fx-c': task.isParallel,
          //parallel: task.isParallel,
          'fx-r': !task.isParallel,
        }"
      >
        <div
          class="child-bb-container"
          v-for="child in children"
          :key="child.id"
        >
          <div class="bb-container">
            <div class="bb-bridge" v-if="task.firstChild !== child.id"></div>
            <div
              class="bb-prev"
              :class="{ show: task.firstChild !== child.id }"
            ></div>
            <div class="bb-next" :class="{ show: child.nextSib != null }"></div>
          </div>
          <div class="child-container">
            <div class="bb-padding" v-if="task.firstChild !== child.id"></div>
            <node
              :project="project"
              :id="child.id"
              :tasks="tasks"
              :showFullSubTree="myShowFullSubTree"
              :initExpandPath="initExpandPath"
            ></node>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
export default {
  name: "node",
  props: {
    project: Object,
    id: String,
    tasks: Object,
    showFullSubTree: Boolean,
    initExpandPath: Object,
  },
  data: function () {
    return this.initState();
  },
  computed: {
    children() {
      // we simply reference this.showChildren here
      // to force this computed 'children' to be re-evalutated
      // when this.showChildren is changed.
      this.showChildren;
      let children = [];
      if (this.task.firstChild == null) {
        return children;
      }
      let id = this.task.firstChild;
      while (id != null) {
        let c = this.tasks[id];
        if (c == null) {
          break;
        }
        children.push(c);
        id = c.nextSib;
      }
      return children;
    },
  },
  methods: {
    initState() {
      return {
        task: this.tasks[this.id],
        showChildren: this.showFullSubTree,
        myShowFullSubTree: this.showFullSubTree,
      };
    },
    init() {
      this.$u.copyProps(this.initState(), this);
      if (this.id === this.$u.rtr.task()) {
        this.$nextTick(() => {
          this.$refs.scrollhandle.scrollIntoView({
            behavior: "smooth",
            block: "end",
          });
        });
      }
      if (this.initExpandPath[this.id] != null && !this.showChildren) {
        this.showHideChildren();
      }
      if (this.task.id == this.$u.rtr.task()) {
        // if expanded down to target node, delete all keys in initExpandPath
        // so that collapsing and expanding ancestor nodes doesnt keep auto exapnding
        // all the way down to this task again.
        setTimeout(() => {
          Object.keys(this.initExpandPath).forEach((key) => {
            delete this.initExpandPath[key];
          });
        }, 20);
      }
    },
    showHideChildren() {
      if (this.task.childN == 0) {
        return;
      }
      if (!this.showChildren) {
        // check tasks to see if I already have them
        let children = this.children;
        if (!(children.length == this.task.childN || children.length >= 100)) {
          let fut = null;
          if (this.task.childN <= 1000) {
            fut = this.$api.task
              .getTree({
                host: this.project.host,
                project: this.project.id,
                id: this.id,
              })
              .then((tasksMap) => {
                for (const [key, value] of Object.entries(tasksMap)) {
                  this.tasks[key] = value;
                }
              });
          } else {
            fut = this.$api.task
              .getChildren({
                host: this.project.host,
                project: this.project.id,
                id: this.id,
              })
              .then((tasks) => {
                tasks.set.forEach((task) => {
                  this.tasks[task.id] = task;
                });
              });
          }
          fut.finally(() => {
            this.showChildren = true;
            if (this.task.childN == this.task.descN) {
              this.myShowFullSubTree = true;
            }
          });
        } else {
          this.showChildren = true;
          if (this.task.childN == this.task.descN) {
            this.myShowFullSubTree = true;
          }
          return;
        }
      }
      this.showChildren = false;
      this.myShowFullSubTree = false;
    },
    showHideFullSubTree() {
      if (this.task.descN == 0 || this.task.descN > 1000) {
        return;
      }
      if (!this.myShowFullSubTree) {
        // check tasks to see if I already have them all
        let descendants = this.inMemoryDescendantsOf(this.id);
        if (descendants.length != this.task.descN) {
          this.$api.task
            .getTree({
              host: this.project.host,
              project: this.project.id,
              id: this.id,
            })
            .then((tasksMap) => {
              for (const [key, value] of Object.entries(tasksMap)) {
                this.tasks[key] = value;
              }
            })
            .finally(() => {
              this.showChildren = true;
              this.myShowFullSubTree = true;
            });
        } else {
          // this settimeout is used to to force vue to re render the full sub tree
          // incase the children were already shown, they must be hidden then re shown
          // to make sure the are initialised with :showFullSubTree="true"
          setTimeout(() => {
            this.showChildren = true;
            this.myShowFullSubTree = true;
          }, 0);
        }
      }
      this.showChildren = false;
      this.myShowFullSubTree = false;
    },
    inMemoryDescendantsOf(id, includeThisOne) {
      let descendants = [];
      let t = this.tasks[id];
      if (t == null) {
        return descendants;
      }
      if (t.descN > 0) {
        let fc = this.tasks[t.firstChild];
        if (fc != null) {
          descendants = descendants.concat(
            this.inMemoryDescendantsOf(fc.id, true)
          );
        }
      }
      if (includeThisOne) {
        descendants.push(t);
        t = this.tasks[t.nextSib];
        if (t != null) {
          descendants = descendants.concat(
            this.inMemoryDescendantsOf(t.id, true)
          );
        }
      }
      return descendants;
    },
  },
  mounted() {
    this.init();
  },
  // watch: {
  //   $route() {
  //     this.init();
  //   },
  // },
};
</script>

<style scoped lang="scss">
@import "../style.scss";
$spacer: 1pc;
* {
  white-space: nowrap;
}
.clk {
  cursor: pointer;
}
.icon.small {
  display: inline-block;
  height: 1pc;
  width: 1pc;
}
.blue {
  color: #22a0dd;
}
.dark-blue {
  color: #1160aa;
}
.small {
  font-size: 0.8pc;
}
.mr {
  margin-right: 0.5pc;
}
.mt {
  margin-top: 0.5pc;
}
.parallel-indicator {
  color: orange;
  &.parallel {
    color: green;
  }
}
.time-min {
  color: #31ff38;
}
.time-est {
  color: #ff9100;
}
.time-inc {
  color: #ffe138;
}
.fx-c {
  display: inline-flex;
  flex-direction: column;
}
.fx-r {
  display: inline-flex;
  flex-direction: row;
}
.fx-start {
  justify-content: flex-start;
  align-items: flex-start;
}
.bb-stem {
  height: $spacer;
  width: $spacer;
  border-right: 1px solid white;
}
.children {
  //padding-left: $spacer;
  > .child-bb-container {
    display: inline-flex;
    flex-direction: column;
    > .bb-container {
      display: inline-flex;
      flex-direction: row;
      height: $spacer;
      > .bb-bridge {
        width: $spacer;
        border-top: 1px solid $color;
      }
      > .bb-prev {
        width: $spacer;
        &.show {
          border-top: 1px solid $color;
        }
        border-right: 1px solid $color;
      }
      > .bb-next {
        flex-grow: 1;
        &.show {
          border-top: 1px solid $color;
        }
      }
    }
    > .child-container {
      display: inline-flex;
      flex-direction: row;
      > .bb-padding {
        width: $spacer;
      }
    }
  }
  // &.parallel {
  //   > .child-bb-container {
  //     flex-direction: row;
  //     > .bb-container {
  //       flex-direction: column;
  //       > .bb-bridge {
  //         border-top: none;
  //         border-left: 1px solid $color;
  //       }
  //       > .bb-prev {
  //         height: $spacer;
  //         border-top: none;
  //         border-right: none;
  //         border-left: 1px solid $color;
  //         border-bottom: 1px solid $color;
  //       }
  //       > .bb-next {
  //         flex-grow: 1;
  //         border-top: none;
  //         &.show {
  //           border-left: 1px solid $color;
  //         }
  //       }
  //     }
  //     > .child-container {
  //       flex-direction: column;
  //       > .bb-padding {
  //         height: $spacer;
  //       }
  //     }
  //   }
  // }
}

div.root {
  .this-node {
    @include border();
    border-radius: 0.2pc;
    background: transparent;
    * {
      background: transparent;
    }
    padding: 0.5pc;
    .name {
      a {
        text-decoration: none;
      }
      span {
        font-size: 2pc;
        line-height: 50%;
        display: inline-block;
      }
    }
  }
}
</style>