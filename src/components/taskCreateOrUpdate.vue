<template>
  <div>
    <div v-if="loading">
      loading...
    </div>
    <div v-else>
      <h1>task {{isCreate? 'create': 'update'}}</h1>
      <input v-model="name" placeholder="name" @blur="validate" @keydown.enter="ok">
      <span v-if="nameErr.length > 0" class="err">{{nameErr}}</span>
      <input v-model="description" placeholder="description" @blur="validate" @keydown.enter="ok">
      <span v-if="descriptionErr.length > 0" class="err">{{descriptionErr}}</span>
      <input v-model="user" placeholder="user id" @blur="validate" @keydown.enter="ok">
      <span>
        <label for="checkbox">parallel </label>
        <input type="checkbox" v-model="isParallel" placeholder="isParallel" @keydown.enter="ok">
      </span>
      <input v-model="timeEstDisplay" type="text" placeholder="0h 0m" @blur="validate" @keydown.enter="ok">
      <input v-model="costEstDisplay" type="text" placeholder="0.00" @blur="validate" @keydown.enter="ok">
      <span v-if="!isCreate && task.id != project.id && task.id != $u.rtr.task()">
        <button @click.stop.prevent="showMove=!showMove">move</button>
      </span>
      <div v-if="showMove">
        <input v-model="parentId" placeholder="parent Id" @blur="validate" @keydown.enter="ok">
        <input v-model="previousSiblingId" placeholder="previous sibling Id" @blur="validate" @keydown.enter="ok">
      </div>
      <button @click="ok">{{isCreate? 'create': 'update'}}</button>
      <button @click="close()">close</button>
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
      task: Object,
      children: Array,
      index: Number,
      parentId: String,
      previousSiblingId: String
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
            if (!this.isCreate) {
              let t = this.task
              if (this.index > -1) {
                t = this.children[this.index]
              }
              this.name = t.name
              this.description = t.description
              this.user = t.user
              this.isParallel = t.isParallel
              this.timeEst = t.timeEst
              this.costEst = t.costEst
            } 
            this.timeEstDisplay = this.$u.fmt.duration(this.timeEst)
            this.costEstDisplay = this.$u.fmt.cost(this.costEst, this.project.currencyCode) 
            this.loading = false
        })
      },
      validate(){
        let isOk = true
        if (this.name.length > 250) {
          isOk = false
          this.nameErr = "name must not be over 250 characters"
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
        return isOk
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
              id: this.task.id,
            }
            this.$api.task.update(args).then((res)=>{
              // todo update correct objs
              console.log(res)
              for(const [key, value] of Object.entries(res.parent)) {
                this.task[key] = value
              }
              this.close()
            })
          }
        }
      },
      close(){
        this.$emit('close')
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