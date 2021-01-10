<template>
  <div class="root">
    <div v-if="showCreate || showUpdate">
      <project-create-or-update 
        :isCreate="showCreate"
        :project="update"
        @close="showCreate = showUpdate = false">
      </project-create-or-update>
    </div>
    <div v-else>
      <div class="header">
        <h1 v-if="user != null">{{user.handle+"s projects"}}</h1>
        <button v-if="isMe" @click="showCreate = true">create</button>
      </div>
      <p v-if="loading">
        loading projects
      </p>
      <div v-else>
        <div class="projects">
          <table>
            <tr class="header">
              <th :colspan="s.cols.length" :rowspan="s.cols.length == 1? 2: 1" :class="s.name" v-for="(s, index) in sections" :key="index">
                {{s.name}}
              </th>
            </tr>
            <tr class="header">
              <th :class="c.name" v-for="(c, index) in colHeaders" :key="index">
                {{c.name}}
              </th>
            </tr>
            <tr class="row" @click="$u.rtr.goto(`/host/${p.host}/project/${p.id}/task/${p.id}`)" v-for="(p, index) in ps" :key="p.id">
              <td :class="c.class" v-for="(c, index) in cols" :key="index">
                {{ c.get(p) }}
              </td>
              <td v-if="isMe" class="action" @click.stop="update = p; showUpdate = true" title="update">
                <img src="@/assets/edit.svg">
              </td>
              <td v-if="isMe" class="action" @click.stop="trash(p, index)" title="delete">
                <img src="@/assets/trash.svg">
              </td>
            </tr>
          </table>
          <button class="load-more" v-if="psMore" @click="loadMore()">load more</button>
        </div>
        <div v-if="others.length > 0" class="others">
          <h1>Others Projects</h1>
          <table>
            <tr class="header">
              <th colspan="1" rowspan="2">
                host
              </th>
              <th :colspan="s.cols.length" :rowspan="s.cols.length == 1? 2: 1" :class="s.name" v-for="(s, index) in sections" :key="index">
                {{s.name}}
              </th>
            </tr>
            <tr class="header">
              <th :class="c.name" v-for="(c, index) in colHeaders" :key="index">
                {{c.name}}
              </th>
            </tr>
            <tr class="project" @click="$u.rtr.goto(`/host/${p.host}/project/${p.id}/task/${p.id}`)" v-for="(p) in others" :key="p.id">
              <td >
                <user :userId="p.host"></user>
              </td><td :class="c.class" v-for="(c, index) in cols" :key="index">
                {{ c.get(p) }}
              </td>
            </tr>
          </table>
          <button class="load-more" v-if="othersMore" @click="loadMoreOthers()">load more</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
  import user from '../components/user'
  import projectCreateOrUpdate from '../components/projectCreateOrUpdate.vue'
  export default {
    name: 'projects',
    components: {user, projectCreateOrUpdate},
    data: function() {
      return this.initState()
    },
    computed: {
      isMe(){
        return this.me != null && this.me.id === this.host
      },
      sections(){
        return this.commonSections.filter(i => i.show())
      },
      colHeaders(){
        let res = []
        this.sections.forEach((section)=>{
          if (section.cols.length > 1) {
            res = res.concat(section.cols)
          }
        })
        return res
      },
      cols(){
        let res = []
        this.sections.forEach((section)=>{
          res = res.concat(section.cols)
        })
        return res
      }
    },
    methods: {
      initState (){
        return {
          host: this.$u.rtr.host(),
          showCreate: false,
          showUpdate: false,
          updateId: null,
          me: null,
          user: null,
          loading: true,
          commonSections: [
            {
              name: "name",
              show: () => true,
              cols: [
                {
                  name: "name",
                  get: (p)=> p.name
                }
              ]
            },
            {
              name: "date",
              show: () => this.$root.show.date,
              cols: [
                {
                  name: "created",
                  get: (p)=> this.$u.fmt.date(p.createdOn)
                },
                {
                  name: "start",
                  get: (p)=> this.$u.fmt.date(p.startOn)
                },
                {
                  name: "end",
                  get: (p)=> this.$u.fmt.date(p.endOn)
                },
                {
                  name: "hrs/day",
                  get: (p)=> p.hoursPerDay
                },
                {
                  name: "days/wk",
                  get: (p)=> p.daysPerWeek
                }
              ]
            },
            {
              name: "time",
              show: () => this.$root.show.time,
              cols: [
                { 
                  name: "min",
                  get: (p)=> this.$u.fmt.duration(p.timeEst + p.timeSubMin, p.hoursPerDay, p.daysPerWeek)                  
                },
                {
                  name: "est",
                  get: (p)=> this.$u.fmt.duration(p.timeEst + p.timeSubEst, p.hoursPerDay, p.daysPerWeek)
                },
                {
                  name: "inc",
                  get: (p)=> this.$u.fmt.duration(p.timeInc + p.timeSubInc, p.hoursPerDay, p.daysPerWeek)
                }
              ]
            },
            {
              name: "cost",
              show: () => this.$root.show.cost,
              cols: [
                {
                  name: "est",
                  get: (p)=> this.$u.fmt.cost(p.costEst + p.costSubEst, p.currencyCode)
                },
                {
                  name: "inc",
                  get: (p)=> this.$u.fmt.cost(p.costInc + p.costSubInc, p.currencyCode)
                }
              ]
            },
            {
              name: "file",
              show: () => this.$root.show.file,
              cols: [
                {
                  name: "n",
                  get: (p)=> p.fileN + p.fileSubN
                },
                {
                  name: "size",
                  get: (p)=> this.$u.fmt.bytes(p.fileSize + p.fileSubSize)
                }
              ]
            },
            {
              name: "task",
              show: () => this.$root.show.task,
              cols: [
                {
                  name: "childn",
                  get: (p)=> p.childN
                },
                {
                  name: "descn",
                  get: (p)=> p.descN
                }
              ]
            }
          ],
          sort: "createon",
          asc: false,
          ps: [],
          others: [],
          psMore: false,
          othersMore: false
        }
      },
      init() {
        for(const [key, value] of Object.entries(this.initState())) {
          this[key] = value
        }
        this.$api.user.me().then((me)=>{
          this.me = me
        }).finally(()=>{
          let mapi = this.$api.newMDoApi()
          mapi.user.one(this.host).then((user)=>{
            this.user = user
          })
          mapi.project.get({host: this.host}).then((res) => {
            for (let i = 0; i < res.set.length; i++) {
              this.ps.push(res.set[i]) 
            }
            this.psMore = res.more
          })
          if (this.isMe) {
            mapi.project.getOthers().then((res) => {
              for (let i = 0; i < res.set.length; i++) {
                this.others.push(res.set[i]) 
              }
              this.othersMore = res.more
            })
          }
          mapi.sendMDo().finally(()=>{
            this.loading = false
          })
        })
      },
      loadMore(){
        let after = this.ps[this.ps.length-1].id
        this.$api.project.get({host: this.host, after}).then((res) => {
          for (let i = 0; i < res.set.length; i++) {
            this.ps.push(res.set[i]) 
          }
          this.psMore = res.psMore
        })
      },
      loadMoreOthers(){
        let after = this.others[this.others.length-1].id
        this.$api.project.getOthers({host: this.host, after}).then((res) => {
          for (let i = 0; i < res.set.length; i++) {
            this.others.push(res.set[i]) 
          }
          this.othersMore = res.psMore
        })
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
div.root {
  padding: 2.6pc 0 0 1.3pc;
}
.column-filters {
  margin: 0.6pc 0 0.6pc 0;
}
table {
  margin: 1pc 0 1pc 0;
  border-collapse: collapse;
  th, td {
    &:not(.action){
      text-align: center;
      min-width: 100px;
    }
    &.name{
      min-width: 18pc;
    }
  }
  tr.row {
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