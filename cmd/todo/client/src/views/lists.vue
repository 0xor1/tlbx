<template>
  <div class="root">
    <div class="header">
      <a @click.stop.prevent="logout" href="">logout</a>
      <h1>lists</h1>
      <input @keydown.enter="create" v-model="createName" placeholder="new list">
      <button @click="create">create</button>
    </div>
    <p v-if="!loading && lists.length === 0">create your first list</p>
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
      <tr class="list" @click="goto(list)" v-for="(list, index) in lists" :key="list.id">
        <td v-if="!list.showEditTools">
          {{ list.name }}
        </td>
        <td v-else>
          <input :ref='"edit_" + list.id' @keydown.esc="toggleEditTools(list)" @keydown.enter="update(list)" @click.stop v-model="list.newName" placeholder="new name">
          <button @click.stop="update(list)">update</button>
        </td>
        <td class="count">
          {{ list.todoItemCount }}
        </td>
        <td class="count">
          {{ list.completedItemCount }}
        </td>
        <td class="action" @click.stop="toggleEditTools(list)">
          <img src="@/assets/edit.svg">
        </td>
        <td class="action" @click.stop="trash(list, index)">
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
    <p v-if="loading">loading lists</p>
  </div>
</template>

<script>
  import api from '@/api'
  import router from '@/router'
  export default {
    name: 'lists',
    data: function() {
      this.load(true)
      return {
        loading: true,
        createName: "",
        lists: [],
        currentEditItem: null,
        err: null,
        more: false
      }
    },
    methods: {
      create: function(){
        api.list.create(this.createName).then(this.goto)
      },
      goto: function(list){
          router.push('/list/'+list.id)
      },
      trash: function(list, index){
        api.list.delete([list.id]).then(()=>{
          this.lists.splice(index, 1)
        })
      },
      toggleEditTools: function(list){
        list.showEditTools = !list.showEditTools
        list.newName = list.name
        if (this.currentEditItem !== null && this.currentEditItem !== list) {
          this.currentEditItem.showEditTools = false
        }
        this.$forceUpdate()
        if (list.showEditTools) {
          this.currentEditItem = list
          this.$nextTick(()=>{
            this.$refs["edit_"+list.id][0].focus()
          })
        }
      },
      update: function(list){
        list.showEditTools = false
        let oldName = list.name
        let newName = list.newName
        this.currentEditItem = null
        this.$forceUpdate()
        if (oldName === newName) {
          return
        }
        api.list.update(list.id, newName).then((res)=>{
          list.name = res.name
        }).catch(()=>{
          list.name = oldName
        })
      },
      load: function(reset){
        if (!this.loading) {
          this.loading = true
          if (reset) {
            this.lists = []
          }
          let args = {}
          if (this.lists !== undefined && this.lists.length > 0 ) {
            args.after = this.lists[this.lists.length - 1].id
          }
          api.list.get(args).then((res) => {
            for (let i = 0; i < res.set.length; i++) {
              let list = res.set[i]
              list.newName = list.name
              list.showEditTools = false
              this.lists.push(res.set[i]) 
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
  tr.list {
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