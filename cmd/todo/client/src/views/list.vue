<template>
  <div class="root">
    <div v-if="!loading" class="header">
      <p><a @click.stop.prevent="logout" href="">logout</a></p>
      <p><a @click.stop.prevent="gotoLists" href="">my lists</a></p>
      <h1>{{list.name}}</h1>
      <p>todo:{{list.todoItemCount}} complete:{{list.completedItemCount}}</p>
      <input @keydown.enter="create" v-model="createName" placeholder="new item">
      <button @click="create">create</button>
    </div>
    <p v-if="!loading && items.length === 0">create your first item</p>
    <p v-else>
    <table class="todo">
      <tr>
        <th class="name">
          Name
        </th>
        <th>
          Created On
        </th>
      </tr>
      <tr class="item" v-for="(item, index) in items" :key="item.id">
        <td v-if="!item.showEditTools">
          {{ item.name }}
        </td>
        <td v-else>
          <input :ref="'edit_' + item.id" @keydown.esc="toggleEditTools(item)" @keydown.enter="update(item)" @click.stop v-model="item.newName" placeholder="new name">
          <button @click.stop="update(item)">update</button>
        </td>
        <td class="count">
          {{ item.createdOn | moment("DD-MM-YYYY") }}
        </td>
        <td class="action" @click.stop="complete(item, index)">
          <img src="@/assets/tick.svg">
        </td>
        <td class="action" @click.stop="toggleEditTools(item)">
          <img src="@/assets/edit.svg">
        </td>
        <td class="action" @click.stop="trash(item, index)">
          <img src="@/assets/trash.svg">
        </td>
      </tr>
      <tr v-if="more">
        <td>
          <button @click="load(false, false)">load more</button>
        </td>
      </tr>
    </table>
    </p>
    <p v-if="loading">loading todo items</p>
  </div>
</template>

<script>
  import api from '@/api'
  import router from '@/router'
  export default {
    name: 'lists',
    data: function() {
      this.load(true, false)
      return {
        loading: true,
        createName: "",
        list: {},
        items: [],
        currentEditItem: null,
        err: null,
        more: false
      }
    },
    methods: {
      create: function(){
        api.item.create(this.list.id, this.createName).then((res)=>{
          this.createName = ""
          this.list.todoItemCount++
          this.items.push(res)
        })
      },
      trash: function(item, index){
        api.item.delete(this.list.id, [item.id]).then(()=>{
          this.list.todoItemCount--
          this.items.splice(index, 1)
        })
      },
      toggleEditTools: function(item){
        item.showEditTools = !item.showEditTools
        item.newName = item.name
        if (this.currentEditItem !== null && this.currentEditItem !== item) {
          this.currentEditItem.showEditTools = false
        }
        this.$forceUpdate()
        if (item.showEditTools) {
          this.currentEditItem = item
          this.$nextTick(()=>{
            this.$refs["edit_"+item.id][0].focus()
          })
        }
      },
      update: function(item){
        item.showEditTools = false
        this.currentEditItem = null
        this.$forceUpdate()
        let oldName = item.name
        let newName = item.newName
        this.currentEditItem = null
        this.$forceUpdate()
        if (oldName === newName) {
          return
        }
        api.item.update(this.list.id, item.id, newName).then((res)=>{
          item.name = res.name
        }).catch(()=>{
          item.name = oldName
        })
      },
      complete: function(item, index){
        api.item.update(this.list.id, item.id, undefined, true).then(()=>{
          this.list.todoItemCount--
          this.list.completedItemCount++
          this.items.splice(index, 1)
        })
      },
      load: function(reset, completed){
        let listId = router.currentRoute.params.id
        let mapi = api
        if (!this.loading) {
          this.loading = true
          if (reset) {
            this.items = []
            mapi = api.newMDoApi()
            mapi.list.one(listId).then((res)=>{
              this.list = res
            })
          }
          let args = {list: listId, completed}
          if (this.items !== undefined && this.items.length > 0 ) {
            args.after = this.items[this.items.length - 1].id
          }
          mapi.item.get(args).then((res) => {
            for (let i = 0; i < res.set.length; i++) {
              let item = res.set[i]
              item.newName = item.name
              item.showEditTools = false
              this.items.push(res.set[i]) 
            }
            this.more = res.more
          }).catch((err) => {
            this.err = err
          })
          if (reset) {
            mapi.sendMDo().finally(()=>{
              this.loading = false
            })
          }
        }
      },
      logout: function(){
        api.me.logout().then(()=>{
          router.push('/login')
        })
      },
      gotoLists: function(){
        router.push('/lists')
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
      min-width: 260px;
    }
    &.todo{
      min-width: 100px;
    }
    &.count{
      text-align: center;
    }
  }
  tr.item {
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