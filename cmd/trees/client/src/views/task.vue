<template>
  <div class="root">
    <div v-if="notFound">
      <notfound :type="'task'"></notfound>
    </div>
    <div v-else-if="loading" class="loading">
      loading...
    </div>
    <div v-else-if="task.updIdx > -1 || task.crtIdx > -1">
      <task-create-or-update 
        :hostId="$u.rtr.host()" 
        :projectId="$u.rtr.project()"
        :set="task.set"
        :updIdx="task.updIdx"
        :crtIdx="task.crtIdx"
        @close="taskOnCrtOrUpdClose"
        @refreshProjectActivity="refreshProjectActivity">
      </task-create-or-update>
    </div>
    <div v-else class="content" >
      <div class="breadcrumb">
        <span>
          <user :goToHome="true" :userId="$u.rtr.host()"></user>
          :
        </span>
        <span v-if="task.ancestor.set.length > 0 && task.ancestor.set[0].parent != null">
          <a title="load more ancestors" href="" @click.stop.prevent="taskAncestorLoadMore">..</a>
          /
        </span>
        <span v-for="(a) in task.ancestor.set" :key="a.id">
          <a :title="a.name" :href="'/#/host/'+$u.rtr.host()+'/project/'+$u.rtr.project()+'/task/'+a.id">{{$u.fmt.ellipsis(a.name, 20)}}</a> 
          /     
        </span>
        <span>
          {{$u.fmt.ellipsis(t0.name, 20)}}      
        </span>
      </div>
      <div class="summary">
        <table>
          <tr class="header">
            <th :colspan="s.cols.length" :rowspan="s.cols.length == 1? 2: 1" :class="s.name" v-for="(s, idx) in taskSections" :key="idx">
              {{s.name()}}
            </th>
          </tr>
          <tr class="header">
            <th :class="c.name" v-for="(c, idx) in taskSectionHeaders" :key="idx">
              {{c.name}}
            </th>
          </tr>
          <tr class="row" v-for="(t, idx) in task.set" :key="idx" @click.stop.prevent="$u.rtr.goto(`/host/${$u.rtr.host()}/project/${$u.rtr.project()}/task/${t.id}`)">
            <td :title="taskTitle(t)" v-bind:class="c.name" v-for="(c, idx) in taskCols" :key="idx">
              <span :title="t.isParallel? 'parallel': 'sequential'" :class="{'parallel-indicator': true, 'parallel': t.isParallel}" v-if="c.name == 'name'">{{t.isParallel? "&#8649;": "&#8699;"}}</span>{{c.name == "user"? "" : c.get(t)}}
              <user v-if="c.name=='user'" :userId="c.get(t)"></user>
            </td>
            <td v-if="$u.perm.canWrite(pMe)" class="action" @click.stop="taskShowCrt(idx+1)" title="insert below">
              <img src="@/assets/insert-below.svg">
            </td>
            <td v-if="idx === 0 && task.set.length > 1 && $u.perm.canWrite(pMe)" class="action">
            </td>
            <td v-else-if="idx > 0 && $u.perm.canWrite(pMe)" class="action" @click.stop="taskShowCrt(idx)" title="insert above">
              <img src="@/assets/insert-above.svg">
            </td>
            <td v-if="taskCanUpd(t)" class="action" @click.stop="taskShowUpd(idx)" title="update">
              <img src="@/assets/edit.svg">
            </td>
            <td v-if="taskCanDlt(t)" class="action" @click.stop="taskTglDltIdx(idx)" title="delete safety">
              <img src="@/assets/trash.svg">
            </td>
            <td v-if="task.dltIdx === idx" class="action confirm-delete" @click.stop="taskDlt(idx)" title="delete">
              <img src="@/assets/trash-red.svg">
            </td>
          </tr>
        </table>
        <button v-if="task.more" @click="taskLoadMore">load more</button>
      </div>
      <div>
        <p v-if="t0.description.length > 0" v-html="$u.fmt.md(t0.description)"></p>
      </div>
      <div v-for="(type, idx) in ['time', 'cost']" :key="idx">
        <div v-if="$root.show[type]" :class="['items', type+'s']">
          <div class="heading">{{type}} <span class="medium" v-if="type == 'cost'">{{$u.fmt.currencySymbol(project.currencyCode)}}</span> <span class="medium">{{$u.fmt[type](t0[type+'Inc'])}}</span><span class="medium" v-if="task.set.length > 0"> | {{$u.fmt[type](t0[type+'SubInc'])}}</span></div>
          <div v-if="$u.perm.canWrite(pMe)" class="create-form">
            <div title="note">
              <span>note <span :class="{err: vitem[type].note.length > 250, 'small': true}">({{250 - vitem[type].note.length}})</span></span><br>
              <input :class="{err: vitem[type].note.length > 250, note: true}" v-model="vitem[type].note" type="text" placeholder="note" @blur="vitemValidate(type)" @keyup="vitemValidate(type)" @keydown.enter="vitemSubmit(type)"/>
            </div>
            <div title="incurred">
              <span>inc <span v-if="type == 'cost'" class="small">{{$u.fmt.currencySymbol(project.currencyCode)}}</span></span><br>
              <input :class="{err: vitem[type].incErr}" v-model="vitem[type].incStr" type="text" :placeholder="vitem[type].placeholder" @blur="vitemValidate(type, true)" @keyup="vitemValidate(type)" @keydown.enter="vitemSubmit(type)"/>
            </div>
            <div title="remaining estimate">
              <span>est <span v-if="type == 'cost'" class="small">{{$u.fmt.currencySymbol(project.currencyCode)}}</span></span><br>
              <input :class="{err: vitem[type].estErr}" v-model="vitem[type].estStr" type="text" :placeholder="vitem[type].placeholder" @blur="vitemValidate(type, true)" @keyup="vitemValidate(type)" @keydown.enter="vitemSubmit(type)"/>
            </div>
            <div>
              <button @click.stop="vitemSubmit(type)">create</button>
            </div>
          </div>
          <table v-if="vitem[type].set.length > 0">
            <tr class="header">
              <th class="note">note</th>
              <th v-if="$root.show.date">created</th>
              <th v-if="$root.show.user">user</th>
              <th>inc <span v-if="type == 'cost'" class="small">{{$u.fmt.currencySymbol(project.currencyCode)}}</span></th>
            </tr>
            <tr class="item" v-for="(i, idx) in vitem[type].set" :key="idx">
              <td v-if="vitem[type].updIdx != idx" class="note" v-html="$u.fmt.mdLinkify(i.note)"></td>
              <td v-else class="note"><input :class="{err: vitem[type].updNote > 250}" v-model="vitem[type].updNote" type="text" placeholder="note" @blur="vitemValidateUpd(type, true)" @keyup="vitemValidateUpd(type)" @keydown.enter="vitemSubmitUpd(type)" @keydown.escape="vitemCancelUpd(type)"/></td>
              <td v-if="$root.show.date">{{$u.fmt.date(i.createdOn)}}</td>
              <td v-if="$root.show.user"><user :userId="i.createdBy"></user></td>
              <td v-if="vitem[type].updIdx != idx">{{$u.fmt[type](i.inc)}}</td>
              <td v-else><input :class="{err: vitem[type].updIncErr}" v-model="vitem[type].updIncStr" type="text" :placeholder="vitem[type].placeholder" @blur="vitemValidateUpd(type, true)" @keyup="vitemValidateUpd(type)" @keydown.enter="vitemSubmitUpd(type)" @keydown.escape="vitemCancelUpd(type)"/></td>
              <td v-if="vitemCanUpd(i) && vitem[type].updIdx != idx" class="action" @click.stop="vitemShowUpd(i, idx)" title="update">
                <img src="@/assets/edit.svg">
              </td>
              <td v-if="vitemCanUpd(i) && vitem[type].updIdx != idx" class="action" @click.stop="vitemTglDltIdx(type, idx)" title="delete safety">
                <img src="@/assets/trash.svg">
              </td>
              <td v-if="vitem[type].dltIdx === idx" class="action confirm-delete" @click.stop="vitemDlt(i, idx)" title="delete">
                <img src="@/assets/trash-red.svg">
              </td>
            </tr>
          </table>
          <div v-if="vitem[type].more"><button @click.stop.prevent="vitemLoadMore(type)">load more</button></div>
        </div>
      </div>
      <div v-if="$root.show.file" class="items files">
        <div class="heading">file <span class="medium">{{$u.fmt.bytes(t0.fileSize)}}</span><span class="medium" v-if="task.set.length > 0"> | {{$u.fmt.bytes(t0.fileSubSize)}}</span></div>
        <div v-if="$u.perm.canWrite(pMe)" class="create-form">
          <div @click.stop="fileButtonClick" class="file-selector" title="choose file">
            <input ref="fileInput" id="file" class="file" type="file" @change="fileSelectorChange"/>
            <label ref="fileLabel" class="small" for="file" @click.stop>avail space ({{$u.fmt.bytes(project.fileLimit - (project.fileSize + project.fileSubSize))}}) <span v-if="fileUploadProgress > -1">| uploading {{fileUploadProgress}}%</span></label><br>
            <span v-if="selectedFile != null" class="input-file">{{$u.fmt.ellipsis(fileSelectedName, 34)}}</span>
            <span v-else class="input-file">select file</span>
          </div>
          <div>
            <button @click.stop="fileSubmit()">upload</button>
          </div>
        </div>
        <table v-if="files.length > 0">
          <tr class="header">
            <th class="name">name</th>
            <th v-if="$root.show.date">created</th>
            <th v-if="$root.show.user">user</th>
            <th>size</th>
          </tr>
          <tr class="item" v-for="(f, idx) in files" :key="idx">
            <td class="note">
              <a v-if="fileIsImageType(f)" :href="fileGetDownloadUrl(f, false)" target="_blank">{{$u.fmt.ellipsis(f.name, 35)}}</a>
              <a v-else :href="fileGetDownloadUrl(f, true)">{{$u.fmt.ellipsis(f.name, 35)}}</a>
            </td>
            <td v-if="$root.show.date">{{$u.fmt.date(f.createdOn)}}</td>
            <td v-if="$root.show.user"><user :userId="f.createdBy"></user></td>
            <td>{{$u.fmt.bytes(f.size)}}</td>
            <td class="action" title="download">
              <a :href="fileGetDownloadUrl(f, true)"><img src="@/assets/download.svg"></a>
            </td>
            <td v-if="vitemCanUpd(f)" class="action" @click.stop="fileTglDltIdx(idx)" title="delete safety">
              <img src="@/assets/trash.svg">
            </td>
            <td v-if="fileDltIdx === idx" class="action confirm-delete" @click.stop="fileDlt(idx)" title="delete">
              <img src="@/assets/trash-red.svg">
            </td>
          </tr>
        </table>
        <div v-if="moreFiles"><button @click.stop.prevent="fileLoadMore()">load more</button></div>
      </div>
    </div>
  </div>
</template>

<script>
  import user from '../components/user'
  import taskCreateOrUpdate from '../components/taskCreateOrUpdate'
  import notfound from '../components/notfound'
  export default {
    name: 'task',
    components: {user, taskCreateOrUpdate, notfound},
    data: function() {
      return this.initState()
    },
    computed: {
      t0(){
        return this.task.set[0]
      },
      taskSections(){
        return this.task.sections.filter(i => i.show())
      },
      taskSectionHeaders(){
        let res = []
        this.taskSections.forEach((section)=>{
          if (section.cols.length > 1) {
            res = res.concat(section.cols)
          }
        })
        return res
      },
      taskCols(){
        let res = []
        this.taskSections.forEach((section)=>{
          res = res.concat(section.cols)
        })
        return res
      },
      fileSelectedName(){
        if (this.selectedFile != null) {
          return this.selectedFile.name
        }
        return ""
      },
    },
    methods: {
      initState(){
        let vitemFn = function(ph){
          return {
              dltIdx: -1,
              set: [],
              more: false,
              loading: false,
              estStr: "",
              estErr: false,
              incStr: "",
              incErr: false,
              note: "",
              placeholder: ph,
              updIdx: -1,
              updIncStr: "",
              updIncErr: false,
              updNote: ""
            }
        }
        return {
          notFound: false,
          loading: true,
          me: null,
          pMe: null,
          project: null,
          task: {
            ancestor: {
              set: [],
              more: false,
              loading: false,
            },
            set: [],
            more: false,
            loading: false,
            crtIdx: -1,
            updIdx: -1,
            dltIdx: -1,
            sections: [
              {
                name: () => "name",
                show: () => true,
                cols: [
                  {
                    name: "name",
                    get: (t)=> this.$u.fmt.ellipsis(t.name, 35)
                  }
                ]
              },
              {
                name: () => "created",
                show: () => this.$root.show.date,
                cols: [
                  {
                    name: "created",
                    get: (t)=> this.$u.fmt.date(t.createdOn)
                  }
                ]
              },
              {
                name: () => "user",
                show: () => this.$root.show.user,
                cols: [
                  {
                    name: "user",
                    get: (t)=> t.user
                  }
                ]
              },
              {
                name: () => "time",
                show: () => this.$root.show.time,
                cols: [
                  { 
                    name: "min",
                    get: (t)=> this.$u.fmt.time(t.timeEst + t.timeSubMin, this.project.hoursPerDay, this.project.daysPerWeek)                  
                  },
                  {
                    name: "est",
                    get: (t)=> this.$u.fmt.time(t.timeEst + t.timeSubEst, this.project.hoursPerDay, this.project.daysPerWeek)
                  },
                  {
                    name: "inc",
                    get: (t)=> this.$u.fmt.time(t.timeInc + t.timeSubInc, this.project.hoursPerDay, this.project.daysPerWeek)
                  }
                ]
              },
              {
                name: () => `cost ${this.$u.fmt.currencySymbol(this.project.currencyCode)}`,
                show: () => this.$root.show.cost,
                cols: [
                  {
                    name: "est",
                    get: (t)=> this.$u.fmt.cost(t.costEst + t.costSubEst, true)
                  },
                  {
                    name: "inc",
                    get: (t)=> this.$u.fmt.cost(t.costInc + t.costSubInc, true)
                  }
                ]
              },
              {
                name: () => "file",
                show: () => this.$root.show.file,
                cols: [
                  {
                    name: "n",
                    get: (t)=> t.fileN + t.fileSubN
                  },
                  {
                    name: "size",
                    get: (t)=> this.$u.fmt.bytes(t.fileSize + t.fileSubSize)
                  }
                ]
              },
              {
                name: () => "task",
                show: () => this.$root.show.task,
                cols: [
                  {
                    name: "childn",
                    get: (t)=> t.childN
                  },
                  {
                    name: "descn",
                    get: (t)=> t.descN
                  }
                ]
              }
            ]
          },
          vitem: {
            time: vitemFn("0h 0m"),
            cost: vitemFn("0.00")
          },
          files: [],
          moreFiles: false,
          comments: [],
          moreComments: false,
          selectedFile: null,
          fileDltIdx: -1,
          loadingFiles: false,
          fileUploadProgress: -1,
          commentDisplay: "",
          commentErr: "",
        }
      },
      init(){
        this.$u.copyProps(this.initState(), this)
        this.$api.user.me().then((me)=>{
          this.me = me
        }).finally(()=>{
          this.$root.ctx().then((ctx)=>{
            this.pMe = ctx.pMe
            this.project = ctx.project
            let mapi = this.$api.newMDoApi()
            mapi.task.getAncestors({
              host: this.$u.rtr.host(), 
              project: this.$u.rtr.project(), 
              id: this.$u.rtr.task(), 
              limit: 10
            }).then((res)=>{
              this.task.ancestor.set = res.set.reverse()
              this.task.ancestor.more = res.more
            })
            mapi.task.get({
              host: this.$u.rtr.host(),
              project: this.$u.rtr.project(),
              id: this.$u.rtr.task()
            }).then((t)=>{
              this.task.set = [t].concat(this.task.set)
              this.vitem.time.estStr = this.$u.fmt.time(this.t0.timeEst)
              this.vitem.cost.estStr = this.$u.fmt.cost(this.t0.costEst)
            }).catch((err)=>{
              if (err.status == 404) {
                this.notFound = true
              }
            })
            mapi.task.getChildren({
              host: this.$u.rtr.host(),
              project: this.$u.rtr.project(), 
              id: this.$u.rtr.task()
            }).then((res)=>{
              this.task.set = this.task.set.concat(res.set)
              this.task.more = res.more
            })
            mapi.vitem.get({
              host: this.$u.rtr.host(),
              project: this.$u.rtr.project(), 
              task: this.$u.rtr.task(), 
              type: this.$u.cnsts.time
            }).then((res)=>{
              this.vitem.time.set = res.set
              this.vitem.time.more = res.more
            })
            mapi.vitem.get({
              host: this.$u.rtr.host(),
              project: this.$u.rtr.project(), 
              task: this.$u.rtr.task(), 
              type: this.$u.cnsts.cost
            }).then((res)=>{
              this.vitem.cost.set = res.set
              this.vitem.cost.more = res.more
            })
            mapi.file.get({
              host: this.$u.rtr.host(),
              project: this.$u.rtr.project(), 
              task: this.$u.rtr.task()
            }).then((res)=>{
              this.files = res.set
              this.moreFiles = res.more
            })
            mapi.comment.get({
              host: this.$u.rtr.host(), 
              project: this.$u.rtr.project(), 
              task: this.$u.rtr.task()
            }).then((res)=>{
              this.comments = res.set
              this.moreComments = res.more
            })
            mapi.sendMDo().finally(()=>{
              this.loading = false
            })
          })
        })
      },
      taskAncestorLoadMore(){
        let obj = this.task.ancestor
        if (obj.loading) {
          return
        }
        obj.loading = true
        this.$api.task.getAncestors({
          host: this.$u.rtr.host(),
          project: this.$u.rtr.project(),
          id: obj.set[0].id, 
          limit: 10
        }).then((res)=>{
          obj.set = res.set.reverse().concat(obj.set)
          obj.more = res.more
        }).finally(()=>{
          obj.loading = false
        })
      },
      taskLoadMore(){
        let obj = this.task
        if (obj.loading) {
          return
        }
        obj.loading = true;
        this.$api.task.getChildren({
          host: this.$u.rtr.host(), 
          project: this.$u.rtr.project(), 
          id: this.$u.rtr.task(), 
          after: obj.set[obj.set.length - 1].id
        }).then((res)=>{
          obj.set = obj.set.concat(res.set)
          obj.more = res.more
        }).finally(()=>{
          obj.loading = false
        })
      },
      taskCanUpd(t){
        if (this.pMe == null) {
          return false
        }
        if (this.$u.rtr.host() == this.pMe.id || 
          (t.parent != null && this.$u.perm.canWrite(this.pMe))) {
          // if I'm the host I can edit anything,
          // or if I'm an admin or writer I can edit any none root node
          return true
        }
        return false
      },
      taskCanDlt(t){
        if (t.descN > 100) {
          // can't delete a task that would 
          // result in deleting more than 100 
          // sub tasks in one go.
          return false
        }
        if (this.pMe == null) {
          return false
        }
        if (this.$u.rtr.host() == this.pMe.id || 
          (t.parent != null && this.$u.perm.canAdmin(this.pMe))) {
          // if I'm the host I can delete anything,
          // or if I'm an admin I can delete any none root node
          return true
        }
        if (this.$u.perm.canWrite(this.pMe) && 
          t.createdBy == this.pMe.id &&
          (Date.now() - (new Date(t.createdOn))) < 3600000 &&
          t.descN == 0) {
          // writers may only delete their own tasks within an hour of creating them
          // and if the have no children tasks.
          return true
        }
        return false
      },
      taskTglDltIdx(idx){
        if (this.task.dltIdx === idx) {
          this.task.dltIdx = -1
        } else {
          this.task.dltIdx = idx
        }
      },
      taskDlt(idx){
        let dltT = this.task.set[idx] 
        if (dltT.id == this.$u.rtr.project()) {
          this.$api.project.delete([dltT.id]).then(()=>{
            this.$u.rtr.goHome()
            this.refreshProjectActivity(true)
          })
        } else {
          this.$api.task.delete({
            host: this.$u.rtr.host(),
            project: this.$u.rtr.project(),
            id: dltT.id
          }).then((t)=>{
            if (idx > 0) {
              this.task.set.splice(idx, 1)
              this.$u.copyProps(t, this.task[0])
              this.task.dltIdx = -1
              this.project.fileSubSize - (dltT.fileSize + dltT.fileSubSize)
            } else {
                this.$u.rtr.goto(`/host/${this.$u.rtr.host()}/project/${this.$u.rtr.project()}/task/${t.id}`)
            }
            this.refreshProjectActivity(true)
          })
        }
      },
      taskShowCrt(idx){
        this.task.crtIdx = idx
      },
      taskShowUpd(idx){
        this.task.updIdx = idx
      },
      taskOnCrtOrUpdClose(fullRefresh){
        this.task.crtIdx = -1
        this.task.updIdx = -1
        if (fullRefresh) {
          this.init()
        }
      },
      taskTitle(t){
        let res = t.name
        if (t.description != "") {
          res += " - " + t.description
        }
        return res
      },
      vitemTglDltIdx(type, idx){
        if (this.vitem[type].dltIdx === idx) {
          this.vitem[type].dltIdx = -1
        } else {
          this.vitem[type].dltIdx = idx
        }
      },
      vitemValidate(type, isBlur){
        let isOK = true
        let obj = this.vitem[type]
        if (obj.estStr != null && obj.estStr.length > 0) {
          let parsed = this.$u.parse[type](obj.estStr)
          if (parsed == null) {
            obj.estErr = true
            isOK = false
          } else {
            if (isBlur === true) {
              obj.estStr = this.$u.fmt[type](parsed)
            }
            obj.estErr = false
          }
        } else {
          obj.estErr = false
        }
        if (obj.incStr != null && obj.incStr.length > 0) {
          let parsed = this.$u.parse[type](obj.incStr)
          if (parsed == null) {
            obj.incErr = true
            isOK = false
          } else {
            if (isBlur === true) {
              obj.incStr = this.$u.fmt[type](parsed)
            }
            obj.incErr = false
          }
        }else {
          obj.incErr = false
        }
        obj.note = obj.note.substring(0, 250)
        return isOK
      },
      vitemSubmit(type){
        if (this.vitemValidate(type)) {
          let obj = this.vitem[type]
          if (obj.loading) {
            return
          }
          let est = this.$u.parse[type](obj.estStr)
          let inc = this.$u.parse[type](obj.incStr)
          if ((inc == null || inc == 0) &&
            (est != null && est != this.t0[type+'Est'])) {
            // only changing est value
            let args = {
              host: this.$u.rtr.host(),
              project: this.$u.rtr.project(),
              id: this.$u.rtr.task(),
            }
            args[type+'Est'] = {v:est}
            obj.loading = true
            this.$api.task.update(args).then((res)=>{
              this.$u.copyProps(res.task, this.task.set[0])
              this.refreshProjectActivity(true)
            }).finally(()=>{
              obj.loading = false
            })
          } else if (inc != null && inc != 0) {
            let args = {
              host: this.$u.rtr.host(),
              project: this.$u.rtr.project(),
              task: this.$u.rtr.task(),
              type: type,
              inc: inc,
              note: obj.note
            }
            if (est != null && est != this.t0[type+'Est']) {
              args.est = est
            }
            obj.loading = true
            this.$api.vitem.create(args).then((res)=>{
              this.$u.copyProps(res.task, this.task.set[0])
              obj.inc = 0
              obj.incStr = ""
              obj.note = ""
              obj.set.splice(0, 0, res.item)
              this.refreshProjectActivity(true)
            }).finally(()=>{
              obj.loading = false
            })
          }
        }
      },
      vitemValidateUpd(type, isBlur){
        let isOK = true
        let obj = this.vitem[type]
        if (obj.updIncStr != null && obj.updIncStr.length > 0) {
          let parsed = this.$u.parse[type](obj.updIncStr)
          if (parsed == null) {
            obj.updIncErr = true
            isOK = false
          } else {
            if (isBlur === true) {
              obj.updIncStr = this.$u.fmt[type](parsed)
            }
            obj.updIncErr = false
          }
        }else {
          obj.updIncErr = false
        }
        obj.updNote = obj.updNote.substring(0, 250)
        return isOK
      },
      vitemSubmitUpd(type){
        if (this.vitemValidateUpd(type)) {
          let obj = this.vitem[type]
          let curItem = obj.set[obj.updIdx]
          if (obj.loading) {
            return
          }
          let inc = this.$u.parse[type](obj.updIncStr)
          if (inc != null && inc != 0 && 
            (obj.updNote != curItem.note ||
            inc != curItem.inc)) {
            let args = {
              host: this.$u.rtr.host(),
              project: this.$u.rtr.project(),
              task: this.$u.rtr.task(),
              type: type,
              id: curItem.id,
              inc: {v:inc},
              note: {v:obj.updNote}
            }
            obj.loading = true
            this.$api.vitem.update(args).then((res)=>{
              this.$u.copyProps(res.task, this.task.set[0])
              obj.set[obj.updIdx] = res.item
              this.vitemCancelUpd(type)
              this.refreshProjectActivity(true)
            }).finally(()=>{
              obj.loading = false
            })
          } else {
            this.vitemCancelUpd(type)
          }
        }
      },
      vitemCancelUpd(type){
        let obj = this.vitem[type]
        obj.updIdx = -1
        obj.updIncStr = ""
        obj.updIncErr = false
        obj.updNote = ""
      },
      vitemCanUpd(i){
        if (this.pMe == null) {
          return false
        }
        return this.$u.perm.canAdmin(this.pMe) ||
          (this.$u.perm.canWrite(this.pMe) && 
          i.createdBy == this.pMe.id &&
          (Date.now() - (new Date(i.createdOn))) < 3600000 )
      },
      vitemShowUpd(i, idx) {
        this.vitem[i.type].updIdx = idx
        this.vitem[i.type].updIncStr = this.$u.fmt[i.type](i.inc)
        this.vitem[i.type].updIncErr = false
        this.vitem[i.type].updNote = i.note
      },
      vitemDlt(i, idx) {
        let obj = this.vitem[i.type]
        if (obj.loading) {
          return
        }
        obj.loading = true
        this.$api.vitem.delete({
          host: this.$u.rtr.host(),
          project: this.$u.rtr.project(),
          task: this.$u.rtr.task(),
          type: i.type,
          id: i.id
        }).then((t)=>{
          this.$u.copyProps(t, this.task.set[0])
          obj.set.splice(idx, 1)
          obj.dltIdx = -1
          this.refreshProjectActivity(true)
        }).finally(()=>{
          obj.loading = false
        })
      },
      vitemLoadMore(type) {
        let obj = this.vitem[type]
        if (obj.loading) {
          return
        }
        obj.loading = true
        this.$api.vitem.get({
          host: this.$u.rtr.host(),
          project: this.$u.rtr.project(),
          task: this.$u.rtr.task(),
          type: type,
          after: obj.set[obj.set.length - 1].id
        }).then((res)=>{
          obj.set = obj.set.concat(res.set)
          obj.more = res.more
        }).finally(()=>{
          obj.loading = false
        })
      },
      fileLoadMore() {
        if (this.loadingFiles) {
          return
        }
        this.loadingFiles = true
        this.$api.file.get({
          host: this.$u.rtr.host(),
          project: this.$u.rtr.project(),
          task: this.$u.rtr.task(),
          after: this.files[this.files.length - 1].id
        }).then((res)=>{
          this.files = this.files.concat(res.set)
          this.moreFiles = res.more
        }).finally(()=>{
          this.loadingFiles = false
        })
      },
      fileButtonClick(){
        this.$refs.fileLabel.click()
      },
      fileSelectorChange(event){
        if (event == null) {
          this.selectedFile = null
        } else {
          this.selectedFile = this.$refs.fileInput.files[0]
        }
      },
      fileGetDownloadUrl(f, isDownload){
        return this.$api.file.getContentUrl({
          host: this.$u.rtr.host(),
          project: this.$u.rtr.project(),
          task: this.$u.rtr.task(),
          id: f.id,
          isDownload
        })
      },
      fileIsImageType(f){
        return f.type.startsWith("image/")
      },
      fileTglDltIdx(idx) {
        if (this.fileDltIdx === idx) {
          this.fileDltIdx = -1
        } else {
          this.fileDltIdx = idx
        }
      },
      fileDlt(idx) {
        if (this.loadingFiles) {
          return
        }
        let f = this.files[idx]
        this.loadingFiles = true
        this.$api.file.delete({
          host: this.$u.rtr.host(),
          project: this.$u.rtr.project(),
          task: this.$u.rtr.task(),
          id: f.id
        }).then((t)=>{
          this.$u.copyProps(t, this.task.set[0])
          this.files.splice(idx, 1)
          this.fileDltIdx = -1
          this.project.fileSubSize -= f.size
          this.refreshProjectActivity(true)
        }).finally(()=>{
          this.loadingFiles = false
        })
      },
      fileSubmit(){
        if (this.selectedFile != null && !this.loadingFiles) {
          this.loadingFiles = true
          this.$api.file.create({
            host: this.$u.rtr.host(), 
            project: this.$u.rtr.project(),
            task: this.$u.rtr.task(),
            name: this.selectedFile.name, 
            type: this.selectedFile.type,
            size: this.selectedFile.size,
            content: this.selectedFile
          }, (e)=>{
            this.fileUploadProgress = Math.round( (e.loaded * 100) / e.total )
          }).then((res)=>{
            this.task = res.task
            this.files.splice(0, 0, res.file)
            this.selectedFile = null
            this.$refs.fileInput.value = null
            this.project.fileSize += res.file.size
            this.refreshProjectActivity(true)
          }).finally(()=>{
            this.fileUploadProgress = -1
            this.loadingFiles = false
          })
        }
      },
      refreshProjectActivity(force){
        if (this.t0 != null) {
          this.vitem.time.estStr = this.$u.fmt.time(this.t0.timeEst)
          this.vitem.cost.estStr = this.$u.fmt.cost(this.t0.costEst)
        }
        this.$emit("refreshProjectActivity", force)
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
@import "../style.scss";
div.root {
  > .content{
    > .breadcrumb {
      white-space: nowrap;
      overflow-y: auto;
    }
    table {
      margin: 1pc 0 1pc 0;
      border-collapse: collapse;
      tr{
        td.action:not(.confirm-delete) img {
          visibility: hidden;
        }
        &:hover td.action img{
          visibility: visible;
        }
        &:hover td img{
          visibility: initial;
        }
        th {
          text-align: center;
          min-width: 5pc;
        }
        td {
          height: 1pc;
          .parallel-indicator {
            font-size: 1.5pc;
            padding: 0.2pc;
            background: transparent;
            color: orange;
            &.parallel {
              color: green;
            }
          }
          &.action {
            cursor: pointer;
          }
          &:not(.action) {
            text-align: right;
            &.name{
              text-align: left;
              min-width: 20pc;
            }
          }
          &.confirm-delete {
            img {
              visibility: initial;
            }
          }
          img{
            background-color: transparent;
          }
        }
        &.row:nth-child(3){
          cursor: default;
          .action{
            cursor: pointer;
          }
          > * {
            background-color: #333;
          }
          font-weight: bold;
        }
      }
    }
    .items {
      &.files{
        margin-top: 0.3pc;
      }
      > .heading {
        margin-top: 1.5pc;
        font-size: 1.5pc;
        font-weight: bold;
        border-bottom: 1px solid #777;
      }
      .small{
        font-size: 0.8pc;
      }
      .medium{
        font-size: 1pc;
      }
      th.note, th.name {
        min-width: 21pc;
      }
      td.note{
        text-align: left;
        input {
          width: calc(100% - 0.8pc);
        }
        > * {
          // for markdown <p> elements
          margin: 0;
        }
      }
      > .create-form {
        > div {
          display: inline-block;
          margin: 1pc 1pc 0 0;
          &.file-selector {
            .input-file {
              cursor: pointer;
              //height: 1.8pc;
              @include border();
              border-radius: 0.15pc;
              width: 20.62pc;
              background: $inputColor;
              display: inline-block;
              padding: 0.22pc;
            }
          }
          > input {
            width: 5pc;
            &.note, &.file {
              width: 20pc;
            }
            &.file {
              padding: 0;
              margin: 0;
              width: 0;
              height: 0;
              opacity: 0;
              overflow: hidden;
              position: absolute;
              z-idx: -1;
            }
          }
        }
      }
      .btn {
        cursor: pointer;
      }
      img {
        height: 1pc;
        width: 1pc;
      }
    }
  }
}
</style>