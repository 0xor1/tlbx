<template>
  <div class="root">
    <div class="this-node">
      <div class="name">{{ task.name }}</div>
      <div>
        childn
        <a href="" @click.stop.prevent="showHideChildren()"
          ><span v-if="task.childN > 0" class="small"
            >({{ showChildren ? "-" : "+" }})</span
          >
          {{ task.childN }}</a
        >
      </div>
      <div>
        descn
        <a
          v-if="task.descN <= 1000"
          href=""
          @click.stop.prevent="showHideFullSubTree()"
        >
          <span v-if="task.descN > 0" class="small"
            >({{ myShowFullSubTree ? "-" : "+" }})</span
          >
          {{ task.descN }}
        </a>
        <a v-else>{{ task.descN }}</a>
      </div>
    </div>
    <div
      v-if="showChildren"
      :class="{ parallel: task.isParallel }"
      class="children"
    >
      <node
        v-for="child in children"
        :key="child.id"
        :project="project"
        :id="child.id"
        :tasks="tasks"
        :showFullSubTree="myShowFullSubTree"
      ></node>
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
  },
  data: function () {
    return this.initState();
  },
  computed: {
    task() {
      return this.tasks[this.id];
    },
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
        showChildren: this.showFullSubTree,
        myShowFullSubTree: this.showFullSubTree,
      };
    },
    init() {
      this.$u.copyProps(this.initState(), this);
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
* {
  white-space: nowrap;
}
.small {
  color: $borderColor;
  font-size: 0.8pc;
}
div.root {
  padding: 10px;
  display: inline-flex;
  flex-direction: column;
  @include border();
  > .children {
    margin-top: 10px;
    display: inline-flex;
    &.parallel {
      flex-direction: column;
    }
  }
}
</style>