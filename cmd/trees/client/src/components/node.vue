<template>
  <div class="root">
    <div class="this-node">
      <div>Name: {{ task.name }}</div>
      <div>
        childN
        <a href="" @click.stop.prevent="showHideChildren()">{{
          task.childN
        }}</a>
      </div>
      <div>
        descN
        <a
          v-if="task.descN <= 1000"
          href=""
          @click.stop.prevent="showHideFullSubTree()"
          >{{ task.descN }}</a
        >
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
          this.$api.task
            .getChildren({
              host: this.project.host,
              project: this.project.id,
              id: this.id,
            })
            .then((tasks) => {
              tasks.set.forEach((task) => {
                this.tasks[task.id] = task;
              });
            })
            .finally(() => {
              this.showChildren = true;
            });
        } else {
          this.showChildren = true;
          return;
        }
      }
      this.showChildren = false;
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
          this.showChildren = true;
          this.myShowFullSubTree = true;
          return;
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
  watch: {
    $route() {
      this.init();
    },
  },
};
</script>

<style scoped lang="scss">
div.root {
  > .children {
  }
}
</style>