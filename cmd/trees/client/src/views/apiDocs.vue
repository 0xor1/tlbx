<template>
  <div class="root">
    <h1>{{docs.name}}</h1>
    <p>{{docs.description}}</p>
    <p>
      all endpoints can be called with <strong>GET</strong>, <strong>PUT</strong> or <strong>POST</strong>
      http methods, <strong>args</strong> can be passed as <strong>JSON</strong> in the request body or
      as stringified json in the query parameter args e.g. <strong>?args={"name":"val"}</strong>
    </p>
    <div v-for="(sec, idx) in docs.sections" :key="idx">
      <h2 class="expandable" @click.stop.prevent="toggleSection(idx)">{{sec.name}} [{{sec.collapse?'+':'-'}}]</h2>
      <div v-if="!sec.collapse">
        <div v-for="(ep, epIdx) in sec.endpoints" :key="epIdx">
          <h3 class="expandable" @click.stop.prevent="toggleEp(idx, epIdx)">{{ep.path}} [{{ep.collapse?'+':'-'}}]</h3>
          <div v-if="!ep.collapse">
            <p>{{ep.description}}<br>max body size: {{ep.maxBodyBytes === 1000? '1KB': $u.fmt.bytes(ep.maxBodyBytes)}}<br>timeout: {{ep.timeout}}ms</p>
            <div v-if="ep.argsTypes != null">
              <h4 class="expandable" @click.stop.prevent="ep.collapseArgsTypes = !ep.collapseArgsTypes">args types [{{ep.collapseArgsTypes?'+':'-'}}]</h4>
              <div v-if="!ep.collapseArgsTypes" v-html="$u.fmt.md('```\n'+JSON.stringify(ep.argsTypes, null, 4)+'\n```')"></div>
            </div>
            <div v-if="ep.resTypes != null">
              <h4 class="expandable" @click.stop.prevent="ep.collapseResTypes = !ep.collapseResTypes">res types [{{ep.collapseResTypes?'+':'-'}}]</h4>
              <div v-if="!ep.collapseResTypes" v-html="$u.fmt.md('```\n'+JSON.stringify(ep.resTypes, null, 4)+'\n```')"></div>
            </div>
            <div v-if="ep.defaultArgs != null">
              <h4 class="expandable" @click.stop.prevent="ep.collapseDefaultArgs = !ep.collapseDefaultArgs">default args [{{ep.collapseDefaultArgs?'+':'-'}}]</h4>
              <div v-if="!ep.collapseDefaultArgs" v-html="$u.fmt.md('```\n'+JSON.stringify(ep.defaultArgs, null, 4)+'\n```')"></div>
            </div>
            <div v-if="ep.exampleArgs != null">
              <h4 class="expandable" @click.stop.prevent="ep.collapseExampleArgs = !ep.collapseExampleArgs">example args [{{ep.collapseExampleArgs?'+':'-'}}]</h4>
              <div v-if="!ep.collapseExampleArgs" v-html="$u.fmt.md('```\n'+JSON.stringify(ep.exampleArgs, null, 4)+'\n```')"></div>
            </div>
            <div v-if="ep.exampleRes != null">
              <h4 class="expandable" @click.stop.prevent="ep.collapseExampleRes = !ep.collapseExampleRes">example response [{{ep.collapseExampleRes?'+':'-'}}]</h4>
              <div v-if="!ep.collapseExampleRes" v-html="$u.fmt.md('```\n'+JSON.stringify(ep.exampleRes, null, 4)+'\n```')"></div>
            </div>
          </div>
        </div>
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
          let sectionsArray = []
          let sectionsMap = {}
          docs.endpoints.forEach((ep)=>{
            let rawSegs = ep.path.replace('/api/', '').split('/')
            if (rawSegs.length > 1) {
              ep.collapse = true
              ep.collapseArgsTypes = true
              ep.collapseResTypes = true
              ep.collapseDefaultArgs = true
              ep.collapseExampleArgs = true
              ep.collapseExampleRes = true
              let section = {name: rawSegs[0], collapse: true, endpoints: []}
              if (sectionsMap[rawSegs[0]] == null) {
                sectionsMap[rawSegs[0]] = section
                sectionsArray.push(section)
              } else {
                section = sectionsMap[rawSegs[0]]
              }
              section.endpoints.push(ep)
            }
          })
          docs.sections = sectionsArray
          this.docs = docs
        })
      },
      toggleSection(idx) {
        let s = this.docs.sections[idx]
        if (s.collapse) {
          s.collapse = false
        } else {
          // if collapsing section, collapse all children for next opening
          s.collapse = true
          s.endpoints.forEach((ep)=>{
            ep.collapse = true
            ep.collapseArgsTypes = true
            ep.collapseResTypes = true
            ep.collapseDefaultArgs = true
            ep.collapseExampleArgs = true
            ep.collapseExampleRes = true
          })
        }
      },
      toggleEp(idx, epIdx) {
        let ep = this.docs.sections[idx].endpoints[epIdx]
         if (ep.collapse) {
          ep.collapse = false
        } else {
          // if collapsing ep, collapse all children for next opening
          ep.collapse = true
          ep.collapseArgsTypes = true
          ep.collapseResTypes = true
          ep.collapseDefaultArgs = true
          ep.collapseExampleArgs = true
          ep.collapseExampleRes = true
        }
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

.expandable{
  cursor: pointer;
  text-decoration: underline;
}
</style>