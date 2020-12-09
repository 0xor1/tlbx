<template>
  <div class="root">
    <div class="header">
      <h1 v-if="user != null">{{user.handle+" projects"}}</h1>
      <button v-if="isMe" @click="$router.push('/project/create')">create</button>
    </div>
    <p v-if="loading">
      loading projects
    </p>
    <p v-else-if="isMe && ps.length === 0">
      create your first project
    </p>
    <p v-else-if="ps.length === 0">
      user has no projects
    </p>
    <div v-else>
      <div class="column-filters">
        show: 
        <input type="checkbox" v-model="showDates"><label>dates </label>
        <input type="checkbox" v-model="showTimes"><label>times </label>
        <input type="checkbox" v-model="showExpenses"><label>expenses </label>
        <input type="checkbox" v-model="showFiles"><label>files </label>
        <input type="checkbox" v-model="showTasks"><label>tasks</label>
      </div>
      <table >
        <tr class="header">
          <th class="name">
            Name
          </th>
          <th v-if="showDates" class="createdon">
            Created On
          </th>
          <th v-if="showDates" class="starton">
            Start On
          </th>
          <th v-if="showDates" class="endon">
            End On
          </th>
          <th v-if="showDates" class="hoursperday">
            Hours Per Day
          </th>
          <th v-if="showDates" class="daysperweek">
            Days Per Week
          </th>
          <th v-if="showTimes" class="mintime">
            Minimum Time
          </th>
          <th v-if="showTimes" class="esttime">
            Estimated Time
          </th>
          <th v-if="showTimes" class="logtime">
            Logged Time
          </th>
          <th v-if="showExpenses" class="estexp">
            Estimated Expense
          </th>
          <th v-if="showExpenses" class="logexp">
            Logged Expense
          </th>
          <th v-if="showFiles" class="filecount">
            File Count
          </th>
          <th v-if="showFiles" class="filesize">
            File Size
          </th>
          <th v-if="showTasks" class="tasks">
            Tasks
          </th>
        </tr>
        <tr class="project" @click="$router.push(`/host/${p.host}/project/${p.id}/task/${p.id}`)" v-for="(p, index) in ps" :key="p.id">
          <td>
            {{ p.name }}
          </td>
          <td v-if="showDates" class="createdon">
            {{ $fmt.date(p.createdOn) }}
          </td>
          <td v-if="showDates" class="starton">
            {{ $fmt.date(p.startOn) }}
          </td>
          <td v-if="showDates" class="endon">
            {{ $fmt.date(p.endOn) }}
          </td>
          <td v-if="showDates" class="hoursperday">
            {{ p.hoursPerDay }}
          </td>
          <td v-if="showDates" class="daysperweek">
            {{ p.daysPerWeek }}
          </td>
          <td v-if="showTimes" class="mintime">
            {{ $fmt.duration(p.minimumTime) }}
          </td>
          <td v-if="showTimes" class="esttime">
            {{ $fmt.duration(p.estimatedTime) }}
          </td>
          <td v-if="showTimes" class="logtime">
            {{ $fmt.duration(p.loggedTime) }}
          </td>
          <td v-if="showExpenses" class="estexp">
            {{ $fmt.cost(p.currencyCode, p.estimatedExpense)}}
          </td>
          <td v-if="showExpenses" class="logexp">
            {{ $fmt.cost(p.currencyCode, p.loggedExpense) }}
          </td>
          <td v-if="showFiles" class="filecount">
            {{ p.fileCount + p.fileSubCount }}
          </td>
          <td v-if="showFiles" class="filesize">
            {{ $fmt.bytes(p.fileSize + p.fileSubSize) }}
          </td>
          <td v-if="showTasks" class="tasks">
            {{ p.descendantCount + 1 }}
          </td>
          <td class="action" @click.stop="$router.push(`/project/${p.id}/update`)">
            <img src="@/assets/edit.svg">
          </td>
          <td class="action" @click.stop="trash(p, index)">
            <img src="@/assets/trash.svg">
          </td>
        </tr>
        <tr v-if="more">
          <td>
            <button @click="load(false)">load more</button>
          </td>
        </tr>
      </table>
    </div>
  </div>
</template>

<script>
  export default {
    name: 'projects',
    data: function() {
      return this.initState()
    },
    computed: {
      isMe(){
        return this.me != null && this.me.id === this.hostId
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
              prop: "name",
              show: () => true,
            },
            {
              name: "Created On",
              prop: "createOn",
              show: () => this.showDates,
            }
          ],
          sort: "createon",
          asc: false,
          ps: [],
          others: [],
          more: false,
        }
      },
      init() {
        for(const [key, value] of Object.entries(this.initState())) {
          this[key] = value
        }
        let mapi = this.$api.newMDoApi()
        mapi.user.me().then((me)=>{
          this.me = me
        })
        mapi.user.one(this.hostId).then((user)=>{
          this.user = user
        })
        mapi.project.get(this.hostId).then((res) => {
          for (let i = 0; i < res.set.length; i++) {
            let project = res.set[i]
            project.newName = project.name
            project.showEditTools = false
            this.ps.push(res.set[i]) 
          }
          this.more = res.more
        })
        mapi.project.getOthers(this.hostId).then((res) => {
          for (let i = 0; i < res.set.length; i++) {
            let project = res.set[i]
            project.newName = project.name
            project.showEditTools = false
            this.ps.push(res.set[i]) 
          }
          this.more = res.more
        })
        mapi.sendMDo().finally(()=>{
          this.loading = false
        })

      },
      loadMore(){
        if (!this.loading) {
          this.loading = true
          let args = {host: this.$router.currentRoute.params.hostId}
          if (this.ps !== undefined && this.ps.length > 0 ) {
            args.after = this.ps[this.ps.length - 1].id
          }
          this.$api.project.get(this.$router.currentRoute.params.hostId).then((res) => {
            for (let i = 0; i < res.set.length; i++) {
              let project = res.set[i]
              project.newName = project.name
              project.showEditTools = false
              this.ps.push(res.set[i]) 
            }
            this.more = res.more
          }).finally(()=>{
            this.loading = false
          })
        }
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
</style>