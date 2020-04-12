<template>
  <div class="root">
    <div class="header">
      <h1>lists</h1>
      <input v-on:keyup.enter="create" v-model="createName" placeholder="new list">
      <button v-on:click="create">create</button>
    </div>
    <p v-if="!loading && lists.length === 0">create your first list</p>
    <p v-else>
    <table >
      <tr>
        <th>
          Name
        </th>
        <th>
          Todo
        </th>
        <th>
          Completed
        </th>
        <th>
          
        </th>
      </tr>
      <tr class="list" v-on:click="goto(list)" v-for="(list, index) in lists" :key="list.id">
        <th>
          {{ list.name }}
        </th>
        <th>
          {{ list.todoItemCount }}
        </th>
        <th>
          {{ list.completedItemCount }}
        </th>
        <th class="delete" v-on:click="del($event, list, index)">
          <img src="@/assets/delete.svg">
        </th>
      </tr>
      <tr v-if="more">
        <th>
          <button v-on:click="load(false)">load more</button>
        </th>
        <th>
        </th>
        <th>
        </th>
        <th>
        </th>
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
      this.load()
      return {
        loading: true,
        createName: "",
        lists: [],
        err: null,
        more: false
      }
    },
    methods: {
      create: function(){
        api.list.create(this.createName).then((list)=>{
          router.push('/list/'+list.id)
        })
      },
      goto: function(list){
          router.push('/list/'+list.id)
      },
      del: function(event, list, index){
        event.stopPropagation()
        api.list.delete([list.id]).then(()=>{
          this.lists.splice(index, 1)
          this.load(true)
        })
      },
      load: function(onDelete){
        if (!this.loading) {
          let args = {}
          if (onDelete) {
            args.limit = 1
          }
          if (this.lists !== undefined && this.lists.length > 0 ) {
            args.after = this.lists[this.lists.length - 1].id
          }
          api.list.get(args).then((res) => {
            if (onDelete) {
              this.more = res.set.length === 1
            } else {
              for (let i = 0; i < res.set.length; i++) {
                this.lists.push(res.set[i]) 
              }
              this.more = res.more
            }
          }).catch((err) => {
            this.err = err
          }).finally(()=>{
            this.loading = false
          })
        }
      }
    }
  }
</script>

<style lang="scss">
table {
  border-collapse: collapse;
  th, td {
    border: 1px solid black;
  }
  tr.list {
    cursor: pointer;
    th.delete img{
      width: 16px;
      visibility: hidden;
    }
    &:hover th.delete img{
      visibility: visible;
    }
  }
}
</style>