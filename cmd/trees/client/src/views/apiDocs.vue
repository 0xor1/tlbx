<template>
  <div class="root">
    <h1>{{docs.name}}</h1>
    <p>{{docs.description}}</p>
    <p>
      all endpoints can be called with <strong>GET</strong>, <strong>PUT</strong> or <strong>POST</strong>
      http methods, <strong>args</strong> can be passed as <strong>JSON</strong> in the request body or
      as stringified json in the query parameter args e.g. <strong>?args={"name":"val"}</strong>
    </p>
    <div class="ep" v-for="(ep, idx) in docs.endpoints" :key="idx">
      <h3>{{ep.path}}</h3>
      <p>{{ep.description}}</p>
      <p>max body size: {{ep.maxBodyBytes === 1000? '1KB': $u.fmt.bytes(ep.maxBodyBytes)}}</p>
      <p>timeout: {{ep.timeoutMilli}}ms</p>
      <div>
        <h4>default args</h4>
        <div v-html="$u.fmt.md('```\n'+JSON.stringify(ep.defaultArgs, null, 4)+'\n```')"></div>
      </div>
      <div>
        <h4>example args</h4>
        <div v-html="$u.fmt.md('```\n'+JSON.stringify(ep.exampleArgs, null, 4)+'\n```')"></div>
      </div>
      <div>
        <h4>example response</h4>
        <div v-html="$u.fmt.md('```\n'+JSON.stringify(ep.exampleResponse, null, 4)+'\n```')"></div>
      </div>
    </div>
  </div>
</template>

<script>
  export default {
    name: 'apiDocs',
    data: function() {
      return this.initState()
    },
    methods: {
      initState (){
        return {
            docs: {}
        }
      },
      init() {
        this.$u.copyProps(this.initState(), this)
        this.$api.docs().then((docs)=>{
            this.docs = docs
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
@import "../style.scss";

.ep{
  border-top: 1px solid $color;
}
</style>