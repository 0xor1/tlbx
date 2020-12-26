<template>
  <div class="root">
    <div v-if="showCreate || showUpdate">
      <project-create-or-update v-bind:isCreate="showCreate" v-bind:projectId="updateId" @close="showCreate = showUpdate = false"></project-create-or-update>
    </div>
    <div v-else>
      <div class="header">
        <h1 v-if="user != null">{{user.handle+"s projects"}}</h1>
        <button v-if="isMe" @click="showCreate = true">create</button>
      </div>
      <p v-if="loading">
        loading projects
      </p>
      <div v-else>
        <div class="projects">
          <table>
            <tr class="header">
              <th v-bind:class="c.class" v-for="(c, index) in cols" :key="index">
                {{c.name}}
              </th>
            </tr>
            <tr class="project" @click="$u.rtr.goto(`/host/${p.host}/project/${p.id}/task/${p.id}`)" v-for="(p, index) in ps" :key="p.id">
              <td v-bind:class="c.class" v-for="(c, index) in cols" :key="index">
                {{ c.get(p) }}
              </td>
              <td v-if="isMe" class="action" @click.stop="updateId = p.id; showUpdate = true">
                <img src="@/assets/edit.svg">
              </td>
              <td v-if="isMe" class="action" @click.stop="trash(p, index)">
                <img src="@/assets/trash.svg">
              </td>
            </tr>
          </table>
          <button class="load-more" v-if="psMore" @click="loadMore()">load more</button>
        </div>
        <div v-if="others.length > 0" class="others">
          <h1>Others Projects</h1>
          <table>
            <tr class="header">
              <th class="host">
                Host
              </th>
              <th v-bind:class="c.class" v-for="(c, index) in cols" :key="index">
                {{c.name}}
              </th>
            </tr>
            <tr class="project" @click="$u.rtr.goto(`/host/${p.host}/project/${p.id}/task/${p.id}`)" v-for="(p) in others" :key="p.id">
              <td >
                <user v-bind:userId="p.host"></user>
              </td><td v-bind:class="c.class" v-for="(c, index) in cols" :key="index">
                {{ c.get(p) }}
              </td>
            </tr>
          </table>
          <button class="load-more" v-if="othersMore" @click="loadMoreOthers()">load more</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
  import user from '../components/user'
  import projectCreateOrUpdate from '../components/projectCreateOrUpdate.vue'
  export default {
    name: 'projects',
    components: {user, projectCreateOrUpdate},
    data: function() {
      return this.initState()
    },
    computed: {
      isMe(){
        return this.me != null && this.me.id === this.host
      },
      cols(){
        return this.commonCols.filter(i => i.show())
      }
    },
    methods: {
      initState (){
        return {
          host: this.$u.rtr.host(),
          showCreate: false,
          showUpdate: false,
          updateId: null,
          me: null,
          user: null,
          loading: true,
          commonCols: [
            {
              name: "name",
              class: "name",
              get: (p)=> p.name,
              show: () => true
            },
            {
              name: "created on",
              class: "createOn",
              get: (p)=> this.$u.fmt.date(p.createdOn),
              show: () => this.$root.show.dates
            },
            {
              name: "start on",
              class: "startOn",
              get: (p)=> this.$u.fmt.date(p.startOn),
              show: () => this.$root.show.dates
            },
            {
              name: "end on",
              class: "endOn",
              get: (p)=> this.$u.fmt.date(p.endOn),
              show: () => this.$root.show.dates
            },
            {
              name: "hours per day",
              class: "hoursPerDay",
              get: (p)=> p.hoursPerDay,
              show: () => this.$root.show.dates
            },
            {
              name: "days per week",
              class: "daysPerWeek",
              get: (p)=> p.daysPerWeek,
              show: () => this.$root.show.dates
            },
            {
              name: "min time",
              class: "minimumTime",
              get: (p)=> this.$u.fmt.duration(p.minimumTime),
              show: () => this.$root.show.times
            },
            {
              name: "est time",
              class: "estimatedTime",
              get: (p)=> this.$u.fmt.duration(p.estimatedTime),
              show: () => this.$root.show.times
            },
            {
              name: "log time",
              class: "loggedTime",
              get: (p)=> this.$u.fmt.duration(p.loggedTime),
              show: () => this.$root.show.times
            },
            {
              name: "est exp",
              class: "estimatedExpense",
              get: (p)=> this.$u.fmt.cost(p.currencyCode, p.estimatedExpense),
              show: () => this.$root.show.expenses
            },
            {
              name: "log exp",
              class: "loggedExpense",
              get: (p)=> this.$u.fmt.cost(p.currencyCode, p.loggedExpense),
              show: () => this.$root.show.expenses
            },
            {
              name: "files",
              class: "fileCount",
              get: (p)=> p.fileCount,
              show: () => this.$root.show.files
            },
            {
              name: "file size",
              class: "fileSize",
              get: (p) => this.$u.fmt.bytes(p.fileSize + p.fileSubSize),
              show: () => this.$root.show.files
            },
            {
              name: "tasks",
              class: "tasks",
              get: (p)=>{return p.descendantCount + 1},
              show: () => this.$root.show.tasks
            }
          ],
          sort: "createon",
          asc: false,
          ps: [],
          others: [],
          psMore: false,
          othersMore: false
        }
      },
      init() {
        for(const [key, value] of Object.entries(this.initState())) {
          this[key] = value
        }
        this.$api.user.me().then((me)=>{
          this.me = me
        }).finally(()=>{
          let mapi = this.$api.newMDoApi()
          mapi.user.one(this.host).then((user)=>{
            this.user = user
          })
          mapi.project.get({host: this.host}).then((res) => {
            for (let i = 0; i < res.set.length; i++) {
              this.ps.push(res.set[i]) 
            }
            this.psMore = res.more
          })
          if (this.isMe) {
            mapi.project.getOthers().then((res) => {
              for (let i = 0; i < res.set.length; i++) {
                this.others.push(res.set[i]) 
              }
              this.othersMore = res.more
            })
          }
          mapi.sendMDo().finally(()=>{
            this.loading = false
          })
        })
      },
      loadMore(){
        let after = this.ps[this.ps.length-1].id
        this.$api.project.get({host: this.host, after}).then((res) => {
          for (let i = 0; i < res.set.length; i++) {
            this.ps.push(res.set[i]) 
          }
          this.psMore = res.psMore
        })
      },
      loadMoreOthers(){
        let after = this.others[this.others.length-1].id
        this.$api.project.getOthers({host: this.host, after}).then((res) => {
          for (let i = 0; i < res.set.length; i++) {
            this.others.push(res.set[i]) 
          }
          this.othersMore = res.psMore
        })
      },
      trash(p, index){
        this.$api.project.delete([p.id]).then(()=>{
            this.ps.splice(index, 1)
        })
      }
    },
    mounted(){
      this.init()
    },
    watch: {
      $route () {
        this.init()
      }
    }
  }
</script>

<style lang="scss" scoped>
div.root {
  padding: 2.6pc 0 0 1.3pc;
}
.column-filters {
  margin: 0.6pc 0 0.6pc 0;
}
table {
  margin: 1pc 0 1pc 0;
  border-collapse: collapse;
  th, td {
    &:not(.action){
      text-align: center;
      min-width: 100px;
    }
    &.name{
      min-width: 250px;
    }
  }
  tr.project {
    cursor: pointer;
    td.action img {
      margin: 2px 2px 0px 2px;
      width: 18px;
      visibility: hidden;
    }
    &:hover td.action img{
      visibility: visible;
      fill: white;
    }
  }
}
</style>