<template>
  <div class="list">
    <h1>list {{loading}}</h1>
  </div>
</template>

<script>
  import api from '@/api'
  import router from '@/router'
  export default {
    name: 'list',
    data () {
      let list = router.currentRoute.params.id
      let mapi = api.newMDoApi()
      let data = {
        loading: true
      }
      mapi.list.one(list).then((res)=>{
        console.log(res)
      })
      mapi.item.get(list).then((res)=>{
        console.log(res)        
      })
      mapi.sendMDo().then(()=>{
        console.log("mdo finished")
      }).catch((err)=>{
        console.log(err)
      }).finally(()=>{
        data.loading = false
      })
      return data
    }
  }
</script>