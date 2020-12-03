<template>
  <div class="root">
    <div class="header">
      <loginout></loginout>
      <h1>projects</h1>
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
  import loginout from '../components/loginout'
  export default {
    name: 'projects',
    components: { loginout },
    data: function() {
      this.load(true)
      this.$api.user.me().then((me)=>{
        this.isMe = this.$router.currentRoute.params.hostId == me.id
      })
      return {
        isMe: false,
        loading: true,
        showDates: false,
        showTimes: true,
        showExpenses: true,
        showFiles: false,
        showTasks: false,
        ps: [],
        err: null,
        more: false,
      }
    },
    methods: {
      trash: function(p, index){
        this.$api.project.delete([p.id]).then(()=>{
            this.ps.splice(index, 1)
        })
      },
      load: function(reset){
        if (!this.loading) {
          this.loading = true
          if (reset) {
            this.ps = []
          }
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
          }).catch((err) => {
            this.err = err
          }).finally(()=>{
            this.loading = false
          })
        }
      },
      logout: function(){
        this.$api.user.logout().then(()=>{
          this.$router.push('/login')
        })
      }
    },
    watch: {
      $route () {
        this.load(true)
        this.$api.user.me().then((me)=>{
          this.isMe = this.$router.currentRoute.params.hostId == me.id
        })
      }
    }
  }
</script>

<style lang="scss">
.column-filters {
  margin: 0.5pc 0 0.5pc 0;
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