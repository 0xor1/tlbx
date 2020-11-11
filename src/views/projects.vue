<template>
  <div class="root">
    <div class="header">
      <a @click.stop.prevent="logout" href="">logout</a>
      <h1>projects</h1>
      <input @keydown.enter="create" v-model="createName" placeholder="new project">
      <button @click="create">create</button>
    </div>
    <p v-if="!loading && projects.length === 0">create your first project</p>
    <p v-else>
    <table >
      <tr>
        <th class="name">
          Name
        </th>
        <th class="todo">
          Todo
        </th>
        <th>
          Completed
        </th>
      </tr>
      <tr class="project" @click="goto(project)" v-for="(project, index) in projects" :key="project.id">
        <td v-if="!project.showEditTools">
          {{ project.name }}
        </td>
        <td v-else>
          <input :ref='"edit_" + project.id' @keydown.esc="toggleEditTools(project)" @keydown.enter="update(project)" @click.stop v-model="project.newName" placeholder="new name">
          <button @click.stop="update(project)">update</button>
        </td>
        <td class="count">
          {{ project.todoItemCount }}
        </td>
        <td class="count">
          {{ project.completedItemCount }}
        </td>
        <td class="action" @click.stop="toggleEditTools(project)">
          <img src="@/assets/edit.svg">
        </td>
        <td class="action" @click.stop="trash(project, index)">
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
  import router from '@/router'
  export default {
    name: 'projects',
    data: function() {
      this.load(true)
      return {
        hostId: router.currentRoute.params.hostId,
        loading: true,
        createName: "",
        projects: [],
        currentEditItem: null,
        err: null,
        more: false
      }
    },
    methods: {
      create: function(){
        api.project.create(this.createName).then(this.goto)
      },
      goto: function(project){
          router.push('/host/'+this.hostId+'/project/'+project.id+'/task/'+project.id)
      },
      trash: function(project, index){
        api.project.delete([project.id]).then(()=>{
          this.projects.splice(index, 1)
        })
      },
      toggleEditTools: function(project){
        project.showEditTools = !project.showEditTools
        project.newName = project.name
        if (this.currentEditItem !== null && this.currentEditItem !== project) {
          this.currentEditItem.showEditTools = false
        }
        this.$forceUpdate()
        if (project.showEditTools) {
          this.currentEditItem = project
          this.$nextTick(()=>{
            this.$refs["edit_"+project.id][0].focus()
          })
        }
      },
      update: function(project){
        project.showEditTools = false
        let oldName = project.name
        let newName = project.newName
        this.currentEditItem = null
        this.$forceUpdate()
        if (oldName === newName) {
          return
        }
        api.project.update(project.id, newName).then((res)=>{
          project.name = res.name
        }).catch(()=>{
          project.name = oldName
        })
      },
      load: function(reset){
        if (!this.loading) {
          this.loading = true
          if (reset) {
            this.projects = []
          }
          let args = {host: this.hostId}
          if (this.projects !== undefined && this.projects.length > 0 ) {
            args.after = this.projects[this.projects.length - 1].id
          }
          api.project.get(this.hostId).then((res) => {
            for (let i = 0; i < res.set.length; i++) {
              let project = res.set[i]
              project.newName = project.name
              project.showEditTools = false
              this.projects.push(res.set[i]) 
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
          router.push('/login')
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