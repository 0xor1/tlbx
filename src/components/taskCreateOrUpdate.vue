<template>
  <div>
    <div v-if="loading">
      loading...
    </div>
    <div v-else>
      <h1>task {{isCreate? 'create': 'update'}}</h1>
      <span>
        <input ref="name" v-model="name" placeholder="name" @keydown.enter="ok">
        <label> name</label>
      </span>
      <span v-if="nameErr.length > 0" class="err">{{nameErr}}</span>
      <span>
        <input v-model="description" placeholder="description" @blur="validate" @keydown.enter="ok">
        <label> description</label>
      </span>
      <span v-if="descriptionErr.length > 0" class="err">{{descriptionErr}}</span>
      <span v-if="$u.rtr.project() != updateTask.id">
        <input v-model="user" placeholder="user id" @blur="validate" @keydown.enter="ok"> 
        <label> user id</label>
      </span>
      <span>
        <input type="checkbox" v-model="isParallel" @keydown.enter="ok">
        <label> parallel</label>
      </span>
      <span>
        <input v-model="timeEstDisplay" type="text" placeholder="0h 0m" @blur="validate" @keydown.enter="ok">
        <label> time estimate</label>
      </span>
      <span>
        <input v-model="costEstDisplay" type="text" placeholder="0.00" @blur="validate" @keydown.enter="ok">
        <label> cost estimate</label>
      </span>
      <span v-if="!isCreate && updateTask.id != project.id && updateTask.id != $u.rtr.task()">
        <button @click.stop.prevent="showMove=!showMove">move</button>
      </span>
      <div v-if="showMove">
        <span>
          <input v-model="parentId" placeholder="parent id" @blur="validate" @keydown.enter="ok">
          <label> parent id</label>
        </span>
        <span>
          <input v-model="previousSiblingId" placeholder="previous sibling d" @blur="validate" @keydown.enter="ok">
          <label> previous sibling id</label>
        </span>
      </div>
      <button @click="ok">{{isCreate? 'create': 'update'}}</button>
      <button @click.stop.prevent="close()">close</button>
      <span v-if="err.length > 0" class="err">{{err}}</span>
    </div>
  </div>
</template>

<script>
  export default {
    name: 'taskCreateOrUpdate',
    props: {
      isCreate: Boolean,
      hostId: String,
      projectId: String,
      parentUserId: String,
      task: Object,
      children: Array,
      index: Number
    },
    computed: {
      updateTask(){
        if (this.index == -1) {
          return this.task
        } else {
          return this.children[this.index]
        }
      },
      currentPreviousSiblingId(){
        if (this.index < 1) {
          return null
        } else {
          return this.children[this.index - 1].id
        }
      }
    },
    data: function() {
      return this.initState()
    },
    methods: {
      initState (){
        return {
          loading: true,
          showMove: false,
          name: "",
          nameErr: true,
          description: "",
          descriptionErr: true,
          user: null,
          isParallel: false,
          timeEst: 0,
          costEst: 0,
          timeEstDisplay: "",
          costEstDisplay: "",
          parentId: null,
          previousSiblingId: null,
          project: null,
          err: "",
        }
      },
      init(){
        for(const [key, value] of Object.entries(this.initState())) {
          this[key] = value
        }
        this.$root.ctx().then((ctx)=>{
            this.project = ctx.project
            this.previousSiblingId = this.currentPreviousSiblingId
            this.user = this.parentUserId
            if (!this.isCreate) {
              let t = this.updateTask
              this.name = t.name
              this.description = t.description
              this.user = t.user
              this.isParallel = t.isParallel
              this.parentId = t.parent
              this.timeEst = t.timeEst
              this.costEst = t.costEst
            } 
            this.timeEstDisplay = this.$u.fmt.duration(this.timeEst)
            this.costEstDisplay = this.$u.fmt.cost(this.costEst, this.project.currencyCode) 
            this.loading = false
            this.$nextTick(()=>{
              this.$refs.name.focus()
            })
        })
      },
      validate(){
        let isOk = true
        if (this.name.length < 1 || this.name.length > 250) {
          isOk = false
          this.nameErr = "name must 1 to 250 characters"
        } else {
          this.nameErr = ""
        }
        if (this.description.length > 1250) {
          isOk = false
          this.descriptionErr = "description must not be over 1250 characters"
        } else {
          this.descriptionErr = ""
        }
        if (this.user == "") {
          this.user = null
        }
        if (this.timeEstDisplay != "") {
          let match = this.timeEstDisplay.match(/\D*((\d+)h)?\D*((\d+)m)?\D*/)
          if (match != null && match[0] != null && match[0].length > 0) {
            let value = 0
            if (match[2] != null) {
              value += parseInt(match[2], 10) * 60
            }
            if (match[4] != null) {
              value += parseInt(match[4], 10)
            }
            if (!isNaN(value) && value != null) {
              this.timeEst = value
            }
          }
        } 
        this.timeEstDisplay = this.$u.fmt.duration(this.timeEst)
        if (this.costEstDisplay != "") {
          let match = this.costEstDisplay.match(/[^\d.,]*(\d*)(\.|,)?(\d*)?\D*/)
          if (match != null && match[0] != null && match[0].length > 0) {
            let value = parseFloat(match[1]+"."+match[3]) * 100
            value = Math.floor(value)
            if (!isNaN(value) && value != null) {
              this.costEst = value
            }
          }
        } 
        this.costEstDisplay = this.$u.fmt.cost(this.costEst, this.project.currencyCode)
        if (this.previousSiblingId == "") {
          this.previousSiblingId = null
        }
        if (isOk) {
          return isOk
        }
      },
      ok(){
        if (this.validate()) {
          if (this.isCreate) {
            this.$api.task.create(this.hostId, this.projectId, this.task.id, this.previousSiblingId, this.name, this.description, this.isParallel, this.user, this.timeEst, this.costEst).then((res)=>{
              this.children.splice(this.index, 0, res.task)
              for(const [key, value] of Object.entries(res.parent)) {
                this.task[key] = value
              }
              this.close()
            })
          } else {
            let args = {
              host: this.hostId,
              project: this.projectId,
              id: this.updateTask.id,
            }
            let isUpdate = false
            let moved = false
            if (this.updateTask.name != this.name) {
              isUpdate = true
              args.name = {v: this.name}
              this.updateTask.name = this.name
            }
            if (this.updateTask.description != this.description) {
              isUpdate = true
              args.description = {v: this.description}
              this.updateTask.description = this.description
            }
            if (this.updateTask.isParallel != this.isParallel) {
              isUpdate = true
              args.isParallel = {v: this.isParallel}
              this.updateTask.isParallel = this.isParallel
            }
            if (this.updateTask.timeEst != this.timeEst) {
              isUpdate = true
              args.timeEst = {v: this.timeEst}
              this.updateTask.timeEst = this.timeEst
            }
            if (this.updateTask.costEst != this.costEst) {
              isUpdate = true
              args.costEst = {v: this.costEst}
              this.updateTask.costEst = this.costEst
            }
            if (this.updateTask.parent != this.parentId) {
              isUpdate = true
              moved = true
              args.parent = {v: this.parentId}
              this.updateTask.parent = this.parentId
            }
            if (this.currentPreviousSiblingId != this.previousSiblingId) {
              isUpdate = true
              moved = true
              args.previousSibling = {v: this.previousSiblingId}
            }
            if (isUpdate) {
              this.$api.task.update(args).then((res)=>{
                if (res.parent != null) {
                  for(const [key, value] of Object.entries(res.parent)) {
                    this.task[key] = value
                  }
                }
                this.close(moved)
              })
            } else {
              this.close()
            }
          }
        }
      },
      close(fullRefresh){
        if (fullRefresh !== true) {
          fullRefresh = false
        }
        this.$emit('close', fullRefresh)
      },
      handleEsc(e){
        if (e.key == "Escape") {
          this.close()
        }
      }
    },
    mounted(){
      this.init()
      window.addEventListener('keydown', this.handleEsc)
    },
    destroyed(){
      window.removeEventListener('keydown', this.handleEsc)
    },
    watch: {
      $route () {
        this.init()
      }
    }
  }
</script>

<style scoped lang="scss">
div > div {
  & > * {
    display: block;
    margin-bottom: 5px;
  }
  button, a{
    display: inline;
    margin-right: 15px;
  }
  input[type="number"] {
    width: 10pc;
  }
}
.err{
  color: #c33;
}
</style>