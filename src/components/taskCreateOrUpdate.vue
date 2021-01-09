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
      <span v-if="nameErr.length > 0" class="err">{{descriptionErr}}</span>
      <span>
        <input type="checkbox" v-model="isParallel" placeholder="isParallel" @keydown.enter="ok">
        <label for="checkbox"> parallel</label>
      </span>
      <input v-model.number="timeEst" :min="0" :max="1440" type="number" placeholder="minutes estimate" @blur="validate" @keydown.enter="ok">
      <input v-model.number="costEst" :min="0" :max="7" type="number" placeholder="cost estimate" @blur="validate" @keydown.enter="ok">
      <span>
        <button @click.stop.prevent="showMove=!showMove">move</button>
      </span>
      <div v-if="showMove">
        <input v-model="newParentId" placeholder="new parent Id" @blur="validate" @keydown.enter="ok">
        <input v-model="newPreviousSiblingId" placeholder="new previous sibling Id" @blur="validate" @keydown.enter="ok">
      </div>
      <button @click="ok">{{isCreate? 'create': 'update'}}</button>
      <button @click="cancel">cancel</button>
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
      taskId: String
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
          isParallel: false,
          timeEst: null,
          costEst: null,
          newParentId: null,
          newPreviousSiblingId: null,
          err: "",
        }
      },
      init(){
        for(const [key, value] of Object.entries(this.initState())) {
          this[key] = value
        }
        if (!this.isCreate) {
          this.$api.task.get(this.hostId, this.projectId, this.taskId).then((t)=>{
            this.name = t.name
            this.description = t.description
            this.isParallel = t.isParallel
            this.timeEst = t.timeEst
            this.costEst = t.costEst
            this.loading = false
          })
        } else {
          this.loading = false
        }
      },
      validate(){
        if (this.name.length > 250) {
            this.nameErr = "name must be less than 250 characters long"
        } else {
            this.nameErr = ""
        }
        if (this.hoursPerDay != null) {
          if (this.hoursPerDay > 24) {
            this.hoursPerDay = 24
          }
          if (this.hoursPerDay < 1) {
            this.hoursPerDay = null
          }
        }
        if (this.daysPerWeek != null) { 
          if (this.daysPerWeek > 7) {
            this.daysPerWeek = 7
          }
          if (this.daysPerWeek < 1) {
            this.daysPerWeek = null
          }
        }
        if (this.startOn != null) {
            this.startOn.setHours(0, 0, 0, 0)
        }
        if (this.endOn != null) {
            this.endOn.setHours(0, 0, 0, 0)
        }
        if (this.startOn != null && 
          this.endOn != null &&
          this.startOn.getTime() >= this.endOn.getTime()) {
            this.endOn.setDate(this.startOn.getDate()+1)
        }
        return this.nameErr.length === 0
      },
      ok(){
        if (this.validate()) {
          if (this.isCreate) {
            this.$api.project.create(this.name, this.isPublic, this.currencyCode, this.hoursPerDay, this.daysPerWeek, this.startOn, this.endOn).then((p)=>{
              this.$u.rtr.goto(`/host/${p.host}/project/${p.id}/task/${p.id}`)
            })
          } else {
            this.$api.project.updateOne({
              id: this.projectId, 
              name: {v: this.name},
              isPublic: {v: this.isPublic},
              currencyCode: {v: this.currencyCode},
              hoursPerDay: {v: this.hoursPerDay},
              daysPerWeek: {v: this.daysPerWeek},
              startOn: {v: this.startOn},
              endOn: {v: this.endOn}
            }).then((p)=>{
              this.$u.rtr.goto(`/host/${p.host}/project/${p.id}/task/${p.id}`)
            })
          }
        }
      },
      cancel(){
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