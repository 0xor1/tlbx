<template>
  <div class="root">
    <div class="header">
      <h1 v-if="user != null">{{user.handle+"s projects"}}</h1>
      <button v-if="isMe" @click="$router.push('/project/create')">create</button>
    </div>
    <p v-if="loading">
      loading projects
    </p>
    <div v-else>
      <div class="projects">
        <div class="column-filters">
          show: 
          <input type="checkbox" v-model="showDates"><label>dates </label>
          <input type="checkbox" v-model="showTimes"><label>times </label>
          <input type="checkbox" v-model="showExpenses"><label>expenses </label>
          <input type="checkbox" v-model="showFiles"><label>files </label>
          <input type="checkbox" v-model="showTasks"><label>tasks</label>
        </div>
        <table>
          <tr class="header">
            <th v-bind:class="c.class" v-for="(c, index) in cols" :key="index">
              {{c.name}}
            </th>
          </tr>
          <tr class="project" @click="$router.push(`/host/${p.host}/project/${p.id}/task/${p.id}`)" v-for="(p, index) in ps" :key="p.id">
            <td v-bind:class="c.class" v-for="(c, index) in cols" :key="index">
              {{ c.get(p) }}
            </td>
            <td class="action" @click.stop="$router.push(`/project/${p.id}/update`)">
              <img src="@/assets/edit.svg">
            </td>
            <td class="action" @click.stop="trash(p, index)">
              <img src="@/assets/trash.svg">
            </td>
          </tr>
        </table>
        <button class="load-more" v-if="isMe && psMore" @click="loadMore()">load more</button>
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
          <tr class="project" @click="$router.push(`/host/${p.host}/project/${p.id}/task/${p.id}`)" v-for="(p) in others" :key="p.id">
            <td >
              <user v-bind:userId="p.host"></user>
            </td><td v-bind:class="c.class" v-for="(c, index) in cols" :key="index">
              {{ c.get(p) }}
            </td>
          </tr>
        </table>
        <button class="load-more" v-if="isMe && othersMore" @click="loadMore()">load more</button>
      </div>
    </div>
  </div>
</template>

<script>
  import user from '../components/user'
  export default {
    name: 'projects',
    components: {user},
    data: function() {
      return this.initState()
    },
    computed: {
      isMe(){
        return this.me != null && this.me.id === this.hostId
      },
      cols(){
        return this.commonCols.filter(i => i.show())
      }
    },
    methods: {
      initState (){
        return {
          hostId: this.$router.currentRoute.params.hostId,
          me: null,
          user: null,
          loading: true,
          showDates: false,
          showTimes: true,
          showExpenses: true,
          showFiles: false,
          showTasks: false,
          commonCols: [
            {
              name: "Name",
              class: "name",
              get: (p)=> p.name,
              show: () => true,
            },
            {
              name: "Created On",
              class: "createOn",
              get: (p)=> this.$fmt.date(p.createdOn),
              show: () => this.showDates,
            },
            {
              name: "Start On",
              class: "startOn",
              get: (p)=> this.$fmt.date(p.startOn),
              show: () => this.showDates,
            },
            {
              name: "End On",
              class: "endOn",
              get: (p)=> this.$fmt.date(p.endOn),
              show: () => this.showDates,
            },
            {
              name: "Hours Per Day",
              class: "hoursPerDay",
              get: (p)=> p.hoursPerDay,
              show: () => this.showDates,
            },
            {
              name: "Days Per Week",
              class: "daysPerWeek",
              get: (p)=> p.daysPerWeek,
              show: () => this.showDates,
            },
            {
              name: "Minimum Time",
              class: "minimumTime",
              get: (p)=> this.$fmt.duration(p.minimumTime),
              show: () => this.showTimes,
            },
            {
              name: "Estimated Time",
              class: "estimatedTime",
              get: (p)=> this.$fmt.duration(p.estimatedTime),
              show: () => this.showTimes,
            },
            {
              name: "Logged Time",
              class: "loggedTime",
              get: (p)=> this.$fmt.duration(p.loggedTime),
              show: () => this.showTimes,
            },
            {
              name: "Estimated Expense",
              class: "estimatedExpense",
              get: (p)=> this.$fmt.cost(p.currencyCode, p.estimatedExpense),
              show: () => this.showExpenses,
            },
            {
              name: "Logged Expense",
              class: "loggedExpense",
              get: (p)=> this.$fmt.cost(p.currencyCode, p.loggedExpense),
              show: () => this.showExpenses,
            },
            {
              name: "File Count",
              class: "fileCount",
              get: (p)=> p.fileCount,
              show: () => this.showFiles,
            },
            {
              name: "File Size",
              class: "fileSize",
              get: (p) => this.$fmt.bytes(p.fileSize + p.fileSubSize),
              show: () => this.showFiles,
            },
            {
              name: "Tasks",
              class: "tasks",
              get: (p)=>{return p.descendantCount + 1},
              show: () => this.showTasks,
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
          mapi.user.one(this.hostId).then((user)=>{
            this.user = user
          })
          mapi.project.get({host: this.hostId}).then((res) => {
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
        this.$api.project.get({host: this.hostId, after}).then((res) => {
          for (let i = 0; i < res.set.length; i++) {
            this.ps.push(res.set[i]) 
          }
          this.psMore = res.psMore
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

<style lang="scss">
.column-filters {
  margin: 0.6pc 0 0.6pc 0;
}
table {
  margin: 1pc 0 1pc 0;
  border-collapse: collapse;
  th, td {
    border: 1px solid #ddd;
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
button.load-more{
  margin: 1pc 0 1pc 0;
}
</style>