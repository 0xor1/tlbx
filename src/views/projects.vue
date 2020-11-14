<template>
  <div class="root">
    <div class="header">
      <a @click.stop.prevent="logout" href="">logout</a>
      <h1>projects</h1>
      <button @click="$router.push('/project/create')">create</button>
    </div>
    <p v-if="!loading && ps.length === 0">create your first project</p>
    <p v-else>
    <table >
      <tr class="header">
        <th class="name">
          Name
        </th>
        <th class="nodes">
          Nodes
        </th>
        <th class="mintime">
          Minimum Time
        </th>
        <th class="esttime">
          Estimated Time
        </th>
        <th class="logtime">
          Logged Time
        </th>
        <th class="estexp">
          Estimated Expense
        </th>
        <th class="logexp">
          Logged Expense
        </th>
        <th class="createdon">
          Created On
        </th>
        <th class="starton">
          Start On
        </th>
        <th class="endon">
          End On
        </th>
      </tr>
      <tr class="project" @click="$router.push('/host/'+p.host+'/project/'+p.id+'/task/'+p.id)" v-for="(p, index) in ps" :key="p.id">
        <td>
          {{ p.name }}
        </td>
        <td class="nodes">
          {{ p.descendantCount + 1 }}
        </td>
        <th class="mintime">
          {{ p.minimumTime }}
        </th>
        <th class="esttime">
          {{ p.estimatedTime }}
        </th>
        <th class="logtime">
          {{ p.loggedTime }}
        </th>
        <th class="estexp">
          {{ p.currencyCode }} {{ p.estimatedExpense / 100}}
        </th>
        <th class="logexp">
          {{p.currencyCode}} {{ p.loggedExpense / 100}}
        </th>
        <th class="createdon">
          {{ p.createdOn }}
        </th>
        <th class="starton">
          {{ p.startOn }}
        </th>
        <th class="endon">
          {{ p.endOn }}
        </th>
        <td class="action" @click.stop="$router.push('project/'+p.id+'/update')">
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
    </p>
    <p v-if="loading">loading projects</p>
  </div>
</template>

<script>
  import api from '@/api'
  export default {
    name: 'projects',
    data: function() {
      this.load(true)
      return {
        loading: true,
        createName: "",
        ps: [],
        currentEditItem: null,
        err: null,
        more: false
      }
    },
    methods: {
      trash: function(p, index){
        api.project.delete([p.id]).then(()=>{
            this.ps.splice(index, 1)
        })
      },
      load: function(reset){
        if (!this.loading) {
          this.loading = true
          if (reset) {
            this.ps = []
          }
          let args = {host: this.hostId}
          if (this.ps !== undefined && this.ps.length > 0 ) {
            args.after = this.ps[this.ps.length - 1].id
          }
          api.project.get(this.hostId).then((res) => {
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
        api.user.logout().then(()=>{
          this.$router.push('/login')
        })
      }
    }
  }
</script>

<style lang="scss">
table {
  border-collapse: collapse;
  th, td {
    border: 1px solid #ddd;
    &.name{
      min-width: 250px;
    }
    &.todo{
      min-width: 100px;
    }
    &.count{
      text-align: center;
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