<template>
  <div>
    <div v-if="loading">
      loading...
    </div>
    <div v-else>
      <h1>task {{isCrt? 'create': 'update'}}</h1>
      <span>
        <input ref="name" v-model="name" placeholder="name" @keyup="validate" @keydown.enter="ok">
        <label> name</label>
      </span>
      <span v-if="nameErr.length > 0" class="err">{{nameErr}}</span>
      <span>
        <textarea :class="{err: description.length > 1250}" v-model="description" rows="4" placeholder="description" @blur="validate" @keyup="validate"></textarea>
      </span>
      <span v-if="descriptionErr.length > 0" class="err">{{descriptionErr}}</span>
      <span v-if="isCrt || $u.rtr.project() != updateTask.id">
        <input v-model="user" placeholder="user id" @blur="validate" @keydown.enter="ok"> 
        <label> user id</label>
      </span>
      <span>
        <input id="parallel" type="checkbox" v-model="isParallel" @keydown.enter="ok">
        <label for="parallel"> parallel</label>
      </span>
      <span v-if="$root.show.time">
        <input :class="{err: timeEstErr}" v-model="timeEstDisplay" type="text" placeholder="0h 0m" @blur="validate(true)" @keyup="validate" @keydown.enter="ok">
        <label> time estimate</label>
      </span>
      <span v-if="$root.show.cost">
        <input :class="{err: costEstErr}" v-model="costEstDisplay" type="text" placeholder="0.00" @blur="validate(true)" @keyup="validate" @keydown.enter="ok">
        <label> {{$u.fmt.currencySymbol(this.project.currencyCode)}} cost estimate</label>
      </span>
      <span v-if="!isCrt && updateTask.id != project.id">
        <button @click.stop.prevent="showMove=!showMove">move</button>
      </span>
      <div v-if="showMove">
        <span>
          <input v-model="parentId" placeholder="parent id" @blur="validate" @keydown.enter="ok">
          <label> parent id</label>
        </span>
        <span>
          <input v-model="prevSibId" placeholder="previous sibling id" @blur="validate" @keydown.enter="ok">
          <label> previous sibling id</label>
        </span>
      </div>
      <button @click="ok">{{isCrt? 'create': 'update'}}</button>
      <button @click.stop.prevent="close()">close</button>
      <span v-if="err.length > 0" class="err">{{err}}</span>
    </div>
  </div>
</template>

<script>
  export default {
    name: 'taskCreateOrUpdate',
    props: {
      hostId: String,
      projectId: String,
      set: Array,
      crtIdx: Number,
      updIdx: Number
    },
    computed: {
      isCrt(){
        return this.crtIdx > -1
      },
      updateTask(){
        return this.set[this.updIdx]
      },
      currentPrevSibId(){
        if (this.crtIdx <= 1) {
          return null
        } else {
          return this.set[this.crtIdx - 1].id
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
          timeEstErr: false,
          costEst: 0,
          costEstErr: false,
          timeEstDisplay: "",
          costEstDisplay: "",
          parentId: null,
          prevSibId: null,
          project: null,
          err: "",
        }
      },
      init(){
        this.$u.copyProps(this.initState(), this)
        this.$root.ctx().then((ctx)=>{
            this.project = ctx.project
            this.user = this.set[0].user
            if (!this.isCrt) {
              let t = this.updateTask
              this.prevSibId = null
              this.name = t.name
              this.description = t.description
              this.user = t.user
              this.isParallel = t.isParallel
              this.parentId = t.parent
              this.timeEst = t.timeEst
              this.costEst = t.costEst
            } else {
              this.prevSibId = this.currentPrevSibId
            }
            this.timeEstDisplay = this.$u.fmt.time(this.timeEst)
            this.costEstDisplay = this.$u.fmt.cost(this.costEst) 
            this.loading = false
            this.$nextTick(()=>{
              this.$refs.name.focus()
            })
        })
      },
      validate(isBlur){
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
          let parsed = this.$u.parse.time(this.timeEstDisplay)
          if (parsed != null ) {
            this.timeEst = parsed
            this.timeEstErr = false
            if (isBlur === true) {
              this.timeEstDisplay = this.$u.fmt.time(this.timeEst)
            }
          } else {
            this.timeEstErr = true
            isOk = false
          }
        } else {
            this.timeEst = 0
            this.timeEstErr = false
        }
        if (this.costEstDisplay != "") {
          let parsed = this.$u.parse.cost(this.costEstDisplay)
          if (parsed != null ) {
            this.costEst = parsed
            this.costEstErr = false
            if (isBlur === true) {
              this.costEstDisplay = this.$u.fmt.cost(this.costEst)
            }
          } else {
            this.costEstErr = true
            isOk = false
          }
        } else {
            this.costEst = 0
            this.costEstErr = false
        }
        if (this.prevSibId == "") {
          this.prevSibId = null
        }
        if (isOk) {
          return isOk
        }
      },
      ok(){
        if (this.validate()) {
          if (this.isCrt) {
            this.$api.task.create({
              host: this.hostId,
              project: this.projectId,
              parent: this.set[0].id,
              prevSib: this.prevSibId,
              name: this.name,
              description: this.description,
              isParallel: this.isParallel,
              user: this.user,
              timeEst: this.timeEst,
              costEst: this.costEst
            }).then((res)=>{
              this.set.splice(this.crtIdx, 0, res.task)
              this.$u.copyProps(res.parent, this.set[0])
              this.close()
              this.$emit("refreshProjectActivity", true)
            })
          } else {
            let args = {
              host: this.hostId,
              project: this.projectId,
              id: this.updateTask.id,
            }
            let isUpdate = false
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
            if (this.showMove) {
              isUpdate = true
              args.parent = {v: this.parentId}
              this.updateTask.parent = this.parentId
              args.prevSib = {v: this.prevSibId}
            }
            if (isUpdate) {
              this.$api.task.update(args).then((res)=>{
                let oldNextSib = this.set[this.updIdx].nextSib
                this.$u.copyProps(res.task, this.set[this.updIdx])
                if (res.oldParent != null && this.updIdx > 0) {
                  this.$u.copyProps(res.oldParent, this.set[0])
                }
                this.close(res.newParent != null || res.task.nextSib != oldNextSib)
                this.$emit("refreshProjectActivity", true)
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
  input:not([type="checkbox"]), textarea {
    width: 14pc;
  }
}
</style>