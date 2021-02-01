<template>
  <div class="root">
    <div v-if="notFound">
      <notfound :type="'host'"></notfound>
    </div>
    <div v-else-if="showCreateOrUpdate">
      <project-create-or-update 
        :project="update"
        @close="onCreateOrUpdateClose()">
      </project-create-or-update>
    </div>
    <div v-else>
      <div class="header">
        <h1 v-if="user != null">{{user.handle+"s projects"}}</h1>
        <button v-if="isMe" @click="showCreate()">create</button>
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
              <td :class="c.name" v-for="(c, index) in cols" :key="index">
                {{ c.get(p) }}
              </td>
              <td v-if="isMe" class="action" @click.stop="showUpdate(p)" title="update">
                <img src="@/assets/edit.svg">
              </td>
              <td v-if="isMe" class="action" @click.stop="toggleDeleteIndex(index)" title="delete safety">
                <img src="@/assets/trash.svg">
              </td>
              <td v-if="isMe && deleteIndex === index" class="action confirm-delete" @click.stop="trash(p, index)" title="delete">
                <img src="@/assets/trash-red.svg">
              </td>
            </tr>
          </table>
          <button class="load-more" v-if="psMore" @click="loadMore()">load more</button>
        </div>
        <div v-if="others.length > 0" class="others">
          <h1>others projects</h1>
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
            <tr class="row" @click="$u.rtr.goto(`/host/${p.host}/project/${p.id}/task/${p.id}`)" v-for="(p) in others" :key="p.id">
              <td class="host">
                <user :userId="p.host"></user>
              </td><td :class="c.name" v-for="(c, index) in cols" :key="index">
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
  import notfound from '../components/notfound.vue'
  export default {
    name: 'projects',
    components: {user, projectCreateOrUpdate, notfound},
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
          notFound: false,
          host: this.$u.rtr.host(),
          showCreateOrUpdate: false,
          update: null,
          me: null,
          user: null,
          loading: true,
          deleteIndex: -2,
          commonSections: [
            {
              name: "name",
              show: () => true,
              cols: [
                {
                  name: "name",
                  get: (p)=> this.$u.fmt.ellipsis(p.name, 30)
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
                  get: (p)=> this.$u.fmt.time(p.timeEst + p.timeSubMin, p.hoursPerDay, p.daysPerWeek)                  
                },
                {
                  name: "est",
                  get: (p)=> this.$u.fmt.time(p.timeEst + p.timeSubEst, p.hoursPerDay, p.daysPerWeek)
                },
                {
                  name: "inc",
                  get: (p)=> this.$u.fmt.time(p.timeInc + p.timeSubInc, p.hoursPerDay, p.daysPerWeek)
                }
              ]
            },
            {
              name: "cost",
              show: () => this.$root.show.cost,
              cols: [
                {
                  name: "est",
                  get: (p)=> this.$u.fmt.currencySymbol(p.currencyCode) + this.$u.fmt.cost(p.costEst + p.costSubEst, true)
                },
                {
                  name: "inc",
                  get: (p)=> this.$u.fmt.currencySymbol(p.currencyCode) + this.$u.fmt.cost(p.costInc + p.costSubInc, true)
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
        this.$u.copyProps(this.initState(), this)
        this.$api.user.me().then((me)=>{
          this.me = me
        }).finally(()=>{
          let mapi = this.$api.newMDoApi()
          mapi.user.one(this.host).then((user)=>{
            this.user = user
          }).catch((err)=>{
            if (err.status == 404) {
              this.notFound = true
            }
          })
          mapi.project.get({host: this.host}).then((res) => {
            for (let i = 0; i < res.set.length; i++) {
              this.ps.push(res.set[i]) 
            }
            this.psMore = res.more
          })
          mapi.project.get({host: this.host, others: true}).then((res) => {
            for (let i = 0; i < res.set.length; i++) {
              this.others.push(res.set[i]) 
            }
            this.othersMore = res.more
          })
          
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
        this.$api.project.get({host: this.host, others: true, after}).then((res) => {
          for (let i = 0; i < res.set.length; i++) {
            this.others.push(res.set[i]) 
          }
          this.othersMore = res.psMore
        })
      },
      showCreate(){
        this.showCreateOrUpdate = true
        this.update = null
      },
      showUpdate(p){
        this.showCreateOrUpdate = true
        this.update = p
      },
      onCreateOrUpdateClose(){
        this.showCreateOrUpdate = false
        this.update = null
      },
      toggleDeleteIndex(index){
        if (this.deleteIndex === index) {
          this.deleteIndex = -2
        } else {
          this.deleteIndex = index
        }
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
.column-filters {
  margin: 0.6pc 0 0.6pc 0;
}
table {
  margin: 1pc 0 1pc 0;
  border-collapse: collapse;
  th {
    text-align: center;
  }
  td {
    &:not(.action){
      text-align: right;
      min-width: 5pc;
    }
    &.host {
      text-align: left;
    }
    &.name{
      text-align: left;
      min-width: 20pc;
    }
    img{
      background-color: transparent;
    }
  }
  tr.row {
    cursor: pointer;
    td.action img {
      margin: 2px 2px 0px 2px;
      width: 18px;
    }
    td.action:not(.confirm-delete) img {
      visibility: hidden;
    }
    &:hover td.action img{
      visibility: visible;
    }
  }
}
</style>