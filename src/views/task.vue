<template>
  <div class="root">
    <div v-if="notFound">
      <notfound :type="'task'"></notfound>
    </div>
    <div class="loading" v-else-if="loading">
      loading...
    </div>
    <div v-else-if="showCreateOrUpdate">
      <task-create-or-update 
        :isCreate="isCreate"
        :hostId="$u.rtr.host()" 
        :projectId="$u.rtr.project()"
        :task="task"
        :children="children"
        :parentUserId="task.user"
        :index="index"
        @close="onCreateOrUpdateClose"
        @refreshProjectActivity="refreshProjectActivity">
      </task-create-or-update>
    </div>
    <div v-else class="content" >
      <div class="breadcrumb">
        <span>
          <user :goToHome="true" :userId="$u.rtr.host()"></user>
          :
        </span>
        <span v-if="ancestors.length > 0 && ancestors[0].parent != null">
          <a title="load more ancestors" @click.stop.prevent="getMoreAncestors">..</a>
          /
        </span>
        <span v-for="(a) in ancestors" :key="a.id">
          <a :title="a.name" :href="'/#/host/'+$u.rtr.host()+'/project/'+$u.rtr.project()+'/task/'+a.id">{{$u.fmt.ellipsis(a.name, 20)}}</a> 
          /     
        </span>
        <span>
          {{$u.fmt.ellipsis(task.name, 20)}}      
        </span>
      </div>
      <div class="summary">
        <table>
          <tr class="header">
            <th :colspan="s.cols.length" :rowspan="s.cols.length == 1? 2: 1" :class="s.name" v-for="(s, index) in sections" :key="index">
              {{s.name()}}
            </th>
          </tr>
          <tr class="header">
            <th :class="c.name" v-for="(c, index) in colHeaders" :key="index">
              {{c.name}}
            </th>
          </tr>
          <tr class="row this-task">
            <td :title="descriptionTitle(task)" :class="c.name" v-for="(c, index) in cols" :key="index">
              {{c.name == "user"? "" : c.get(task)}}
              <user v-if="c.name=='user'" :userId="c.get(task)"></user>
            </td>
            <td v-if="$u.perm.canWrite(pMe)" class="action" @click.stop="showCreate(0)" title="insert first child">
              <img src="@/assets/insert-below.svg">
            </td>
            <td v-if="this.children.length > 0" class="action">
              
            </td>
            <td v-if="canUpdate(task)" class="action" @click.stop="showUpdate(-1)" title="update">
              <img src="@/assets/edit.svg">
            </td>
            <td v-if="canDelete(task)" class="action" @click.stop="toggleDeleteIndex(-1)" title="delete safety">
              <img src="@/assets/trash.svg">
            </td>
            <td v-if="deleteIndex === -1" class="action confirm-delete" @click.stop="trash(task, -1)" title="delete">
              <img src="@/assets/trash-red.svg">
            </td>
          </tr>
          <tr class="row" v-for="(t, index) in children" :key="index" @click.stop.prevent="$u.rtr.goto(`/host/${$u.rtr.host()}/project/${$u.rtr.project()}/task/${t.id}`)">
            <td :title="descriptionTitle(t)" v-bind:class="c.name" v-for="(c, index) in cols" :key="index">
              {{c.name == "user"? "" : c.get(t)}}
              <user v-if="c.name=='user'" :userId="c.get(t)"></user>
            </td>
            <td v-if="$u.perm.canWrite(pMe)" class="action" @click.stop="showCreate(index+1)" title="insert below">
              <img src="@/assets/insert-below.svg">
            </td>
            <td v-if="$u.perm.canWrite(pMe)" class="action" @click.stop="showCreate(index)" title="insert above">
              <img src="@/assets/insert-above.svg">
            </td>
            <td v-if="canUpdate(t)" class="action" @click.stop="showUpdate(index)" title="update">
              <img src="@/assets/edit.svg">
            </td>
            <td v-if="canDelete(t)" class="action" @click.stop="toggleDeleteIndex(index)" title="delete safety">
              <img src="@/assets/trash.svg">
            </td>
            <td v-if="deleteIndex === index" class="action confirm-delete" @click.stop="trash(t, index)" title="delete">
              <img src="@/assets/trash-red.svg">
            </td>
          </tr>
        </table>
        <button v-if="moreChildren" @click="getMoreChildren()">load more</button>
      </div>
      <div>
        <p v-if="children.length > 0 && task.isParallel">parallel</p>
        <p v-if="children.length > 0 && !task.isParallel">not parallel</p>
        <p v-if="task.description.length > 0" v-html="$u.fmt.mdLinkify(task.description)"></p>
      </div>
      <div v-for="(type, index) in ['time', 'cost']" :key="index">
        <div v-if="$root.show[type]" :class="['items', type+'s']">
          <div class="heading">{{type}} <span class="medium" v-if="type == 'cost'">{{$u.fmt.currencySymbol(project.currencyCode)}}</span> <span class="medium">{{$u.fmt[type](task[type+'Inc'])}} | {{$u.fmt[type](task[type+'SubInc'])}}</span></div>
          <div v-if="$u.perm.canWrite(pMe)" class="create-form">
            <div title="note">
              <span>note <span :class="{err: vitems[type].note.length > 250, 'small': true}">({{250 - vitems[type].note.length}})</span></span><br>
              <input :class="{err: vitems[type].note.length > 250, note: true}" v-model="vitems[type].note" type="text" placeholder="note" @blur="validate(type)" @keyup="validate(type)" @keydown.enter="submit(type)"/>
            </div>
            <div title="incurred">
              <span>inc <span v-if="type == 'cost'" class="small">{{$u.fmt.currencySymbol(project.currencyCode)}}</span></span><br>
              <input :class="{err: vitems[type].incErr}" v-model="vitems[type].incDisplay" type="text" :placeholder="vitems[type].placeholder" @blur="validate(type, true)" @keyup="validate(type)" @keydown.enter="submit(type)"/>
            </div>
            <div title="remaining estimate">
              <span>est <span v-if="type == 'cost'" class="small">{{$u.fmt.currencySymbol(project.currencyCode)}}</span></span><br>
              <input :class="{err: vitems[type].estErr}" v-model="vitems[type].estDisplay" type="text" :placeholder="vitems[type].placeholder" @blur="validate(type, true)" @keyup="validate(type)" @keydown.enter="submit(type)"/>
            </div>
            <div>
              <button @click.stop="submit(type)">create</button>
            </div>
          </div>
          <table v-if="vitems[type].set.length > 0">
            <tr class="header">
              <th class="note">note</th>
              <th v-if="$root.show.date">created</th>
              <th v-if="$root.show.user">user</th>
              <th>inc <span v-if="type == 'cost'" class="small">{{$u.fmt.currencySymbol(project.currencyCode)}}</span></th>
            </tr>
            <tr class="item" v-for="(i, index) in vitems[type].set" :key="index">
              <td v-if="vitems[type].updateIndex != index" class="note" v-html="$u.fmt.mdLinkify(i.note)"></td>
              <td v-else class="note"><input :class="{err: vitems[type].updateNote > 250}" v-model="vitems[type].updateNote" type="text" placeholder="note" @blur="validateUpdate(type, true)" @keyup="validateUpdate(type)" @keydown.enter="submitUpdate(type)" @keydown.escape="cancelUpdate(type)"/></td>
              <td v-if="$root.show.date">{{$u.fmt.date(i.createdOn)}}</td>
              <td v-if="$root.show.user"><user :userId="i.createdBy"></user></td>
              <td v-if="vitems[type].updateIndex != index">{{$u.fmt[type](i.inc)}}</td>
              <td v-else><input :class="{err: vitems[type].updateIncErr}" v-model="vitems[type].updateIncDisplay" type="text" :placeholder="vitems[type].placeholder" @blur="validateUpdate(type, true)" @keyup="validateUpdate(type)" @keydown.enter="submitUpdate(type)" @keydown.escape="cancelUpdate(type)"/></td>
              <td v-if="canUpdateVitem(i) && vitems[type].updateIndex != index" class="action" @click.stop="showVitemUpdate(i, index)" title="update">
                <img src="@/assets/edit.svg">
              </td>
              <td v-if="canUpdateVitem(i) && vitems[type].updateIndex != index" class="action" @click.stop="toggleVitemDeleteIndex(type, index)" title="delete safety">
                <img src="@/assets/trash.svg">
              </td>
              <td v-if="vitems[type].deleteIndex === index" class="action confirm-delete" @click.stop="trashVitem(i, index)" title="delete">
                <img src="@/assets/trash-red.svg">
              </td>
            </tr>
          </table>
          <div v-if="vitems[type].more"><button @click.stop.prevent="loadMoreVitems(type)">load more</button></div>
        </div>
      </div>
      <div v-if="$root.show.file" class="items files">
        <div class="heading">file <span class="medium">{{$u.fmt.bytes(task.fileSize)}} | {{$u.fmt.bytes(task.fileSubSize)}}</span></div>
        <div v-if="$u.perm.canWrite(pMe)" class="create-form">
          <div @click.stop="fileButtonClick" class="file-selector" title="choose file">
            <input ref="fileInput" id="file" class="file" type="file" @change="fileSelectorChange"/>
            <button><label ref="fileLabel" class="btn" for="file" @click.stop>Choose a file</label></button>
            <span class="input-file">{{$u.fmt.ellipsis(selectedFileName, 21)}}</span>
          </div>
          <div>
            <button @click.stop="submitFile()">upload</button>
          </div>
        </div>
        <table v-if="files.length > 0">
          <tr class="header">
            <th class="name">name</th>
            <th v-if="$root.show.date">created</th>
            <th v-if="$root.show.user">user</th>
            <th>size</th>
          </tr>
          <tr class="item" v-for="(f, index) in files" :key="index">
            <td class="note">
              <a v-if="isImageType(f)" :href="getFileDownloadUrl(f, false)" target="_blank">{{$u.fmt.ellipsis(f.name, 35)}}</a>
              <a v-else :href="getFileDownloadUrl(f, true)">{{$u.fmt.ellipsis(f.name, 35)}}</a>
            </td>
            <td v-if="$root.show.date">{{$u.fmt.date(f.createdOn)}}</td>
            <td v-if="$root.show.user"><user :userId="f.createdBy"></user></td>
            <td>{{$u.fmt.bytes(f.size)}}</td>
            <td class="action" title="download">
              <a :href="getFileDownloadUrl(f, true)"><img src="@/assets/download.svg"></a>
            </td>
            <td v-if="canUpdateVitem(f)" class="action" @click.stop="toggleFileDeleteIndex(index)" title="delete safety">
              <img src="@/assets/trash.svg">
            </td>
            <td v-if="fileDeleteIndex === index" class="action confirm-delete" @click.stop="trashFile(index)" title="delete">
              <img src="@/assets/trash-red.svg">
            </td>
          </tr>
        </table>
        <div v-if="moreFiles"><button @click.stop.prevent="loadMoreFiles()">load more</button></div>
      </div>
    </div>
  </div>
</template>

<script>
  import user from '../components/user'
  import taskCreateOrUpdate from '../components/taskCreateOrUpdate'
  import notfound from '../components/notfound'
  export default {
    name: 'tasks',
    components: {user, taskCreateOrUpdate, notfound},
    data: function() {
      return this.initState()
    },
    computed: {
      selectedFileName(){
        if (this.selectedFile != null) {
          return this.selectedFile.name
        }
        return ""
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
      initState(){
        return {
          notFound: false,
          showCreateOrUpdate: false,
          index: null,
          loading: true,
          me: null,
          project: null,
          pMe: null,
          ancestors: [],
          moreAncestors: false,
          loadingMoreAncestors: false,
          task: null,
          children: [],
          moreChildren: false,
          loadingMoreChildren: false,
          files: [],
          moreFiles: false,
          comments: [],
          moreComments: false,
          deleteIndex: -2,
          selectedFile: null,
          fileDeleteIndex: -2,
          loadingFiles: false,
          vitems: {
            time: {
              deleteIndex: -2,
              set: [],
              more: false,
              estDisplay: "",
              estErr: false,
              incDisplay: "",
              incErr: false,
              note: "",
              placeholder: "0h 0m",
              loading: false,
              updateIndex: -1,
              updateIncDisplay: "",
              updateIncErr: false,
              updateNote: ""
            },
            cost: {
              deleteIndex: -2,
              set: [],
              more: false,
              estDisplay: "",
              estErr: false,
              incDisplay: "",
              incErr: false,
              note: "",
              placeholder: "0.00",
              loading: false,
              updateIndex: -1,
              updateIncDisplay: "",
              updateIncErr: false,
              updateNote: ""
            }
          },
          commentDisplay: "",
          commentErr: "",
          commonSections: [
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
                  get: (p)=> p.childN
                },
                {
                  name: "descn",
                  get: (p)=> p.descN
                }
              ]
            }
          ]
        }
      },
      init(){
        for(const [key, value] of Object.entries(this.initState())) {
          this[key] = value
        }
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
              this.ancestors = res.set.reverse()
              this.moreAncestors = res.more
            })
            mapi.task.get({
              host: this.$u.rtr.host(),
              project: this.$u.rtr.project(),
              id: this.$u.rtr.task()
            }).then((t)=>{
              this.task = t
              this.vitems.time.estDisplay = this.$u.fmt.time(this.task.timeEst)
              this.vitems.cost.estDisplay = this.$u.fmt.cost(this.task.costEst)
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
              this.children = res.set
              this.moreChildren = res.more
            })
            mapi.vitem.get({
              host: this.$u.rtr.host(),
              project: this.$u.rtr.project(), 
              task: this.$u.rtr.task(), 
              type: this.$u.cnsts.time
            }).then((res)=>{
              this.vitems.time.set = res.set
              this.vitems.time.more = res.more
            })
            mapi.vitem.get({
              host: this.$u.rtr.host(),
              project: this.$u.rtr.project(), 
              task: this.$u.rtr.task(), 
              type: this.$u.cnsts.cost
            }).then((res)=>{
              this.vitems.cost.set = res.set
              this.vitems.cost.more = res.more
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
      getMoreAncestors(){
        if (!this.loadingMoreAncestors) {
          this.loadingMoreAncestors = true;
          let taskId = this.$u.rtr.task()
          if (this.ancestors.length > 0 && this.ancestors[0].id != null) {
            taskId = this.ancestors[0].id
          }
          this.$api.task.getAncestors({
            host: this.$u.rtr.host(),
            project: this.$u.rtr.project(),
            id: taskId, 
            limit: 10
          }).then((res)=>{
            this.ancestors = res.set.reverse().concat(this.ancestors)
            this.moreAncestors = res.more
          }).finally(()=>{
            this.loadingMoreAncestors = false
          })
        }
      },
      getMoreChildren(){
        if (!this.loadingMoreChildren) {
          this.loadingMoreChildren = true;
          this.$api.task.getChildren({
            host: this.$u.rtr.host(), 
            project: this.$u.rtr.project(), 
            id: this.$u.rtr.task(), 
            after: this.children[this.children.length - 1].id, 
            limit: 10
          }).then((res)=>{
            this.children = this.children.concat(res.set)
            this.moreChildren = res.more
          }).finally(()=>{
            this.loadingMoreChildren = false
          })
        }
      },
      canUpdate(t){
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
      canDelete(t){
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
          if ((t.parent == null && this.$u.rtr.host() == this.pMe.id) || 
            (t.parent != null && this.$u.perm.canAdmin(this.pMe))) {
            return true
          }
        }
        return false
      },
      showCreate(index) {
        this.showCreateOrUpdate = true
        this.isCreate = true
        this.index = index
      },
      showUpdate(index) {
        this.showCreateOrUpdate = true
        this.isCreate = false
        if (index == -1) {
          this.update = this.task
        } else {
          this.update = this.children[index]
        }
        this.parentId = this.update.parent
        this.index = index
      },
      onCreateOrUpdateClose(fullRefresh) {
        this.showCreateOrUpdate = false
        this.index = null
        if (fullRefresh) {
          this.init()
        }
      },
      toggleDeleteIndex(index){
        if (this.deleteIndex === index) {
          this.deleteIndex = -2
        } else {
          this.deleteIndex = index
        }
      },
      toggleVitemDeleteIndex(type, index){
        if (this.vitems[type].deleteIndex === index) {
          this.vitems[type].deleteIndex = -2
        } else {
          this.vitems[type].deleteIndex = index
        }
      },
      trash(t, index){
        if (t.id == this.$u.rtr.project()) {
          this.$api.project.delete([this.$u.rtr.project()]).then(()=>{
            this.$u.rtr.goHome()
            this.refreshProjectActivity(true)
          })
        } else {
          this.$api.task.delete({
            host: this.$u.rtr.host(),
            project: this.$u.rtr.project(),
            id: t.id
          }).then((t)=>{
            if (index > -1) {
              this.children.splice(index, 1)
              this.task = t
              this.deleteIndex = -2
            } else {
                this.$u.rtr.goto(`/host/${this.$u.rtr.host()}/project/${this.$u.rtr.project()}/task/${t.id}`)
            }
            this.refreshProjectActivity(true)
          })
        }
      },
      validate(type, isBlur){
        let isOK = true
        let obj = this.vitems[type]
        if (obj.estDisplay != null && obj.estDisplay.length > 0) {
          let parsed = this.$u.parse[type](obj.estDisplay)
          if (parsed == null) {
            obj.estErr = true
            isOK = false
          } else {
            if (isBlur === true) {
              obj.estDisplay = this.$u.fmt[type](parsed)
            }
            obj.estErr = false
          }
        } else {
          obj.estErr = false
        }
        if (obj.incDisplay != null && obj.incDisplay.length > 0) {
          let parsed = this.$u.parse[type](obj.incDisplay)
          if (parsed == null) {
            obj.incErr = true
            isOK = false
          } else {
            if (isBlur === true) {
              obj.incDisplay = this.$u.fmt[type](parsed)
            }
            obj.incErr = false
          }
        }else {
          obj.incErr = false
        }
        obj.note = obj.note.substring(0, 250)
        return isOK
      },
      submit(type){
        if (this.validate(type)) {
          let obj = this.vitems[type]
          if (obj.loading) {
            return
          }
          let est = this.$u.parse[type](obj.estDisplay)
          let inc = this.$u.parse[type](obj.incDisplay)
          if ((inc == null || inc == 0) &&
            (est != null && est != this.task[type+'Est'])) {
            // only changing est value
            let args = {
              host: this.$u.rtr.host(),
              project: this.$u.rtr.project(),
              id: this.$u.rtr.task(),
            }
            args[type+'Est'] = {v:est}
            obj.loading = true
            this.$api.task.update(args).then((res)=>{
              this.task = res.task
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
            if (est != null && est != this.task[type+'Est']) {
              args.est = est
            }
            obj.loading = true
            this.$api.vitem.create(args).then((res)=>{
              this.task = res.task
              obj.inc = 0
              obj.incDisplay = ""
              obj.note = ""
              obj.set.splice(0, 0, res.item)
              this.refreshProjectActivity(true)
            }).finally(()=>{
              obj.loading = false
            })
          }
        }
      },
      validateUpdate(type, isBlur){
        let isOK = true
        let obj = this.vitems[type]
        if (obj.updateIncDisplay != null && obj.updateIncDisplay.length > 0) {
          let parsed = this.$u.parse[type](obj.updateIncDisplay)
          if (parsed == null) {
            obj.updateIncErr = true
            isOK = false
          } else {
            if (isBlur === true) {
              obj.updateIncDisplay = this.$u.fmt[type](parsed)
            }
            obj.updateIncErr = false
          }
        }else {
          obj.updateIncErr = false
        }
        obj.updateNote = obj.updateNote.substring(0, 250)
        return isOK
      },
      submitUpdate(type){
        if (this.validateUpdate(type)) {
          let obj = this.vitems[type]
          let curItem = obj.set[obj.updateIndex]
          if (obj.loading) {
            return
          }
          let inc = this.$u.parse[type](obj.updateIncDisplay)
          if (inc != null && inc != 0 && 
            (obj.updateNote != curItem.note ||
            inc != curItem.inc)) {
            let args = {
              host: this.$u.rtr.host(),
              project: this.$u.rtr.project(),
              task: this.$u.rtr.task(),
              type: type,
              id: curItem.id,
              inc: {v:inc},
              note: {v:obj.updateNote}
            }
            obj.loading = true
            this.$api.vitem.update(args).then((res)=>{
              this.task = res.task
              obj.set[obj.updateIndex] = res.item
              this.cancelUpdate(type)
              this.refreshProjectActivity(true)
            }).finally(()=>{
              obj.loading = false
            })
          } else {
            this.cancelUpdate(type)
          }
        }
      },
      cancelUpdate(type){
        let obj = this.vitems[type]
        obj.updateIndex = -1
        obj.updateIncDisplay = ""
        obj.updateIncErr = false
        obj.updateNote = ""
      },
      canUpdateVitem(i){
        if (this.pMe == null) {
          return false
        }
        return this.$u.perm.canAdmin(this.pMe) ||
          (this.$u.perm.canWrite(this.pMe) && 
          i.createdBy == this.pMe.id &&
          (Date.now() - (new Date(i.createdOn))) < 3600000 )
      },
      showVitemUpdate(i, index) {
        this.vitems[i.type].updateIndex = index
        this.vitems[i.type].updateIncDisplay = this.$u.fmt[i.type](i.inc)
        this.vitems[i.type].updateIncErr = false
        this.vitems[i.type].updateNote = i.note
      },
      trashVitem(i, index) {
        let obj = this.vitems[i.type]
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
          this.task = t
          obj.set.splice(index, 1)
          obj.deleteIndex = -2
          this.refreshProjectActivity(true)
        }).finally(()=>{
          obj.loading = false
        })
      },
      loadMoreVitems(type) {
        let obj = this.vitems[type]
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
      loadMoreFiles() {
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
      descriptionTitle(t) {
        let res = t.name
        if (t.description != "") {
          res += " - " + t.description
        }
        return res
      },
      refreshProjectActivity(force){
        if (this.task != null) {
          this.vitems.time.estDisplay = this.$u.fmt.time(this.task.timeEst)
          this.vitems.cost.estDisplay = this.$u.fmt.cost(this.task.costEst)    
        }
        this.$emit("refreshProjectActivity", force)
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
      getFileDownloadUrl(f, isDownload){
        return this.$api.file.getContentUrl({
          host: this.$u.rtr.host(),
          project: this.$u.rtr.project(),
          task: this.$u.rtr.task(),
          id: f.id,
          isDownload
        })
      },
      isImageType(f){
        return f.type.startsWith("image/")
      },
      toggleFileDeleteIndex(index) {
        if (this.fileDeleteIndex === index) {
          this.fileDeleteIndex = -2
        } else {
          this.fileDeleteIndex = index
        }
      },
      trashFile(index) {
        if (this.loadingFiles) {
          return
        }
        let f = this.files[index]
        this.loadingFiles = true
        this.$api.file.delete({
          host: this.$u.rtr.host(),
          project: this.$u.rtr.project(),
          task: this.$u.rtr.task(),
          id: f.id
        }).then((t)=>{
          this.task = t
          this.files.splice(index, 1)
          this.fileDeleteIndex = -2
          this.refreshProjectActivity(true)
        }).finally(()=>{
          this.loadingFiles = false
        })
      },
      submitFile(){
        if (this.selectedFile != null) {
          this.$api.file.create({
            host: this.$u.rtr.host(), 
            project: this.$u.rtr.project(),
            task: this.$u.rtr.task(),
            name: this.selectedFile.name, 
            type: this.selectedFile.type,
            size: this.selectedFile.size,
            content: this.selectedFile
          }).then((res)=>{
            this.task = res.task
            this.files.splice(0, 0, res.file)
            this.selectedFile = null
            this.$refs.fileInput.value = null
            this.refreshProjectActivity(true)
          })
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
          min-width: 8pc;
        }
        td {
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
      }
      tr.this-task {
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
    .items {
      &.files{
        margin-top: 0.3pc;
      }
      > .heading {
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
            cursor: pointer;
            width: 20.6pc;
            @include border();
            border-radius: 0.1pc;
            background: $inputColor;
            label.btn {
              width: 100%;
              height: 100%;
              background: transparent;
            }
            span {
              margin-left: 0.3pc;
              background: transparent;
            }
          }
          > input {
            width: 5pc;
            &.note, &.file {
              width: 20pc;
            }
            &.file {
              width: 0.1px;
              height: 0.1px;
              opacity: 0;
              overflow: hidden;
              position: absolute;
              z-index: -1;
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