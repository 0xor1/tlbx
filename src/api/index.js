import axios from 'axios'

let notAuthed = false
let memCache = {}
let meInFlight = false
let userGetInFlight = {}
let globalErrorHandler = null

function NewError(status, body) {
  return {
    status,
    body
  }
}

function newApi(isMDoApi) {
  let mDoSending = false
  let mDoSent = false
  let awaitingMDoList = []
  let doReq = (path, args, headers) => {
    path = `/api${path}`
    if (!isMDoApi || (isMDoApi && mDoSending && !mDoSent)) {
      headers = headers || {"X-Client": "tlbx-web-client"}
      return axios({
        method: 'put',
        url: path,
        headers: headers,
        data: args
      }).then((res) => {
        return res.data
      }).catch((err) => {
        let errObj = NewError(err.response.status, err.response.data)
        if (globalErrorHandler != null) {
            // dont show error just for checking if logged in
          globalErrorHandler(errObj.body)
        }
        throw errObj
      })
    } else if (isMDoApi && !mDoSending && !mDoSent) {
      let awaitingMDoObj = {
        path: path,
        args: args,
        resolve: null,
        reject: null
      }
      awaitingMDoList.push(awaitingMDoObj)
      return new Promise((resolve, reject) => {
        awaitingMDoObj.resolve = resolve
        awaitingMDoObj.reject = reject
      })
    } else {
      throw new Error('invalid get call, use the default api object or a new mdo instance from api.newMDoApi()')
    }
  }

  return {
    setGlobalErrorHandler: (fn)=>{
      globalErrorHandler = fn
    },
    newMDoApi: () => {
      return newApi(true)
    },
    sendMDo: () => {
      if (!isMDoApi) {
        throw new Error('MDoes must be made from the api instance returned from api.newMDoApi()')
      } else if (mDoSending || mDoSent) {
        throw new Error('each MDo must be started with a fresh api.newMDoApi(), once used that same instance cannot be reused')
      }
      mDoSending = true
      let asyncIndividualPromisesReady
      asyncIndividualPromisesReady = (resolve) => {
        let ready = true
        for (let i = 0, l = awaitingMDoList.length; i < l; i++) {
          if (awaitingMDoList[i].resolve === null) {
            ready = false
            setTimeout(asyncIndividualPromisesReady, 0, resolve)
            break
          }
        }
        if (ready) {
          resolve()
        }
      }
      let mdoErrors = []
      mdoErrors.isMDoErrors = true
      let mDoCompleterFunc
      mDoCompleterFunc = (resolve, reject) => {
        if (mDoSent) {
          if (mdoErrors.length === 0) {
            resolve()
          } else {
            reject(mdoErrors)
          }
        } else {
          setTimeout(mDoCompleterFunc, 0, resolve, reject)
        }
      }
      new Promise(asyncIndividualPromisesReady).then(() => {
        let mDoObj = {}
        for (let i = 0, l = awaitingMDoList.length; i < l; i++) {
          let key = `${i}`
          mDoObj[key] = {
            path: awaitingMDoList[i].path,
            args: awaitingMDoList[i].args
          }
        }
        doReq('/mdo', mDoObj).then((res) => {
          for (let i = 0, l = awaitingMDoList.length; i < l; i++) {
            let key = `${i}`
            if (res[key].status === 200) {
              awaitingMDoList[i].resolve(res[key].body)
            } else {
              mdoErrors.push(res[key])
              awaitingMDoList[i].reject(res[key])
            }
          }
        }).catch((error) => {
          mdoErrors.push(error)
          for (let i = 0, il = awaitingMDoList.length; i < il; i++) {
            awaitingMDoList[i].reject(error)
          }
        }).finally(()=>{
          mDoSending = false
          mDoSent = true
        })
      })
      return new Promise(mDoCompleterFunc)
    },
    user: {
      register(alias, handle, email, pwd, confirmPwd) {
        return doReq('/user/register', {alias, handle, email, pwd, confirmPwd})
      },
      resendActivateLink(email) {
        return doReq('/user/resendActivateLink', {email})
      },
      activate(email, code) {
        return doReq('/user/activate', {email, code})
      },
      changeEmail(newEmail) {
        return doReq('/user/changeEmail', {newEmail})
      },
      resendChangeEmailLink: () => {
        return doReq('/user/resendChangeEmailLink')
      },
      confirmChangeEmail(me, code) {
        return doReq('/user/confirmChangeEmail', {me, code})
      },
      resetPwd(email) {
        return doReq('/user/resetPwd', {email})
      },
      setHandle(handle) {
        return doReq('/user/setHandle', {handle: handle}).then(()=>{
          memCache.me.handle = handle
        })
      },
      setAlias(alias) {
        return doReq('/user/setAlias', {alias}).then(()=>{
          memCache.me.alias = alias
        })
      },
      setAvatar(avatar) {
        return doReq('/user/setAvatar', avatar).then(()=>{
          memCache.me.hasAvatar = avatar === null
        })
      },
      setPwd(currentPwd, newPwd, confirmNewPwd) {
        return doReq('/user/setPwd', {currentPwd, newPwd, confirmNewPwd})
      },
      delete(pwd) {
        return doReq('/user/delete', {pwd})
      },
      login(email, pwd) {
        return doReq('/user/login', {email, pwd}).then((res)=>{
          notAuthed = false
          memCache.me = res
          memCache[res.id] = res
          return res
        })
      },
      logout() {
        memCache = {}
        return doReq('/user/logout').then(()=>{
          notAuthed = true
        })
      },
      me() {
        if (notAuthed) {
          // here we have already called /user/me and got back a nil
          // so no need to call it again, just return nil immediately
          return new Promise((resolve) => {
            resolve(null)
          })
        }
        if (memCache.me) {
          // here user is already authed
          return new Promise((resolve) => {
            resolve(memCache.me)
          })
        }
        if (meInFlight) {
          let completer = null
          completer = (resolve, reject) => {
            if (meInFlight) {
              // if Im still in flight loop again
              setTimeout(completer, 100, resolve, reject)
              return
            }
            if (memCache.me) {
              resolve(memCache.me)
            } else {
              resolve(null)
            }
          }
          return new Promise(completer)
        }
        meInFlight = true
        return doReq('/user/me').then((res) => {
          if (res != null ) {
            memCache.me = res
            memCache[res.id] = res
          } else {
            notAuthed = true
          }
          return res
        }).finally(()=>{
          meInFlight = false
        })
      },
      one(id){
        return this.get([id]).then((res)=>{
          if (res != null && res.length > 0) {
            return res[0]
          }
          throw NewError(404, "no such user")
        })
      },
      get(ids){
        if (typeof ids === "string") {
          ids = [ids]
        }
        let toGet = []
        let someInFlight = false
        ids.forEach((id)=>{
          if (memCache[id] == null && userGetInFlight[id] == null) {
            toGet.push(id)
          } else if (!someInFlight && userGetInFlight[id]) {
            someInFlight = true
          }
        })
        // if there are none to get and none in flight
        // return a promis that resolves immediately
        if (toGet.length === 0 && !someInFlight) {
          return new Promise((resolve) => {
            let res = []
            // all users are already cached resolve now.
            ids.forEach((id)=>{
              res.push(memCache[id])
            })
            resolve(res)
          })
        }
        // if there are some to get add them to the in flight list
        // and get them, and return a promis to be resolved later.
        if (toGet.length > 0) {
          someInFlight = true
          toGet.forEach((id)=>{
            userGetInFlight[id] = true
          })
          doReq('/user/get', {users: toGet}).then((res) => {
            if (res != null) {
              res.forEach((user)=>{
                memCache[user.id] = user
              })
            }
          }).finally(()=>{
            // must always clear up in flight list no matter what.
            toGet.forEach((id)=>{
              delete userGetInFlight[id]
            })
            someInFlight = false
          })
        }
        let completer = null
        completer = (resolve, reject) => {
          let someStillInFlight = false
          if ((someInFlight && toGet.length > 0)) {
            someStillInFlight = true
          } else if (someInFlight && Object.keys(userGetInFlight).length > 0) {
            // flight may have ended we need to check if they are here yet
            for (let i = 0; i < ids.length; i++) {
              if (userGetInFlight[ids[i]]) {
                someStillInFlight = true
                break
              }
            }
          }
          if (someStillInFlight) {
            // if some are still in flight loop again
            setTimeout(completer, 100, resolve, reject)
            return
          }
          let res = []
          // req is finished, return what users we have
          ids.forEach((id)=>{
            if (memCache[id] != null) {
              res.push(memCache[id])
            }
          })
          resolve(res)
        
        }
        return new Promise(completer)
      }
    },
    project: {
      create(name, isPublic, currencyCode, hoursPerDay, daysPerWeek, startOn, endOn) {
        return doReq('/project/create', {name, isPublic, currencyCode, hoursPerDay, daysPerWeek, startOn, endOn})
      },
      one(host, id) {
        return this.get({host, ids: [id]}).then((res)=>{
          if (res.set.length > 0) {
            return res.set[0]
          }
          throw {
            status: 404,
            body: "no such project"
          }
        })
      },
      get(args) {
        // host, ids, namePrefix, isArchived, isPublic, createdOnMin, createdOnMax, startOnMin, startOnMax, endOnMin, endOnMax, after, sort, asc, limit
        return doReq('/project/get', args)
      },
      getOthers(args) {
        // ids, namePrefix, createdOnMin, createdOnMax, startOnMin, startOnMax, endOnMin, endOnMax, after, sort, asc, limit
        return doReq('/project/getOthers', args)
      },
      update(ps) {
        // [id, name, currencyCode, hoursPerDay, daysPerWeek, startOn, endOn, isArchived, isPublic]       
        return doReq('/project/update', ps)
      },
      updateOne(args) {
        // id, name, currencyCode, hoursPerDay, daysPerWeek, startOn, endOn, isArchived, isPublic       
        return doReq('/project/update', [args]).then((ps)=>{
          return ps[0]
        })
      },
      delete(ids) {
        return doReq('/project/delete', ids)
      },
      addUsers(host, project, users) {
        return doReq('/project/addUsers', {host, project, users})
      },
      getMe(host, project) {
        return doReq('/project/getMe', {host, project})
      },
      getUsers(host, project, ids, role, handlePrefix, after, limit) {
        return doReq('/project/getUsers', {host, project, ids, role, handlePrefix, after, limit})
      },
      setUserRoles(host, project, users) {
        return doReq('/project/setUserRoles', {host, project, users})
      },
      removeUsers(host, project, users) {
        return doReq('/project/removeUsers', {host, project, users})
      },
      getActivities(args) {
        // host, project, task, item, user, occuredAfter, occuredBefore, limit
        return doReq('/project/getActivities', args)
      }
    },
    task: {
      create(host, project, parent, prevSib, name, description, isParallel, user, timeEst, costEst) {
        return doReq('/task/create', {host, project, parent, prevSib, name, description, isParallel, user, timeEst, costEst})
      },
      update(args) {
        // {host, project, id, parent, prevSib, name, description, isParallel, user, timeEst, costEst}
        return doReq('/task/update', args)
      },
      delete(host, project, id) {
        return doReq('/task/delete', {host, project, id})
      },
      get(host, project, id) {
        return doReq('/task/get', {host, project, id})
      },
      getAncestors(host, project, id, limit) {
        return doReq('/task/getAncestors', {host, project, id, limit})
      },
      getChildren(host, project, id, after, limit) {
        return doReq('/task/getChildren', {host, project, id, after, limit})
      }
    },
    time: {
      create(host, project, task, duration, note) {
        return doReq('/time/create', {host, project, task, duration, note})
      },
      update(host, project, task, id, duration, note) {
        return doReq('/time/update', {host, project, task, id, duration, note})
      },
      get(host, project, task, ids, createOnMin, createdOnMax, createdBy, after, asc, limit) {
        return doReq('/time/get', {host, project, task, ids, createOnMin, createdOnMax, createdBy, after, asc, limit})
      },
      delete(host, project, task, id) {
        return doReq('/time/delete', {host, project, task, id})
      }
    },
    cost: {
      create(host, project, task, value, note) {
        return doReq('/cost/create', {host, project, task, value, note})
      },
      update(host, project, task, id, value, note) {
        return doReq('/cost/update', {host, project, task, id, value, note})
      },
      get(host, project, task, ids, createOnMin, createdOnMax, createdBy, after, asc, limit) {
        return doReq('/cost/get', {host, project, task, ids, createOnMin, createdOnMax, createdBy, after, asc, limit})
      },
      delete(host, project, task, id) {
        return doReq('/cost/delete', {host, project, task, id})
      }
    },
    file: {
      create(host, project, task, name, mimeType, size, content) {
        return doReq('/file/getPresignedPutUrl', {host, project, task, name, mimeType, size}).then((res)=>{
          let id = res.id
          return doReq(res.url, content, {
            "Host": (new URL(res.url)).hostname,
            "X-Amz-Acl": "private",
            "Content-Length": size, 
            "Content-Type": mimeType,
            "Content-Disposition": `attachment; filename=${name}`,
          }).then(()=>{
            return doReq("/file/finalize", {host, project, task, id})
          })
        })
      },
      getPresignedGetUrl(host, project, task, id, isDownload) {
        return doReq('/file/getPresignedGetUrl', {host, project, task, id, isDownload})
      },
      get(host, project, task, ids, createOnMin, createdOnMax, createdBy, after, asc, limit) {
        return doReq('/file/get', {host, project, task, ids, createOnMin, createdOnMax, createdBy, after, asc, limit})
      },
      delete(host, project, task, id) {
        return doReq('/file/delete', {host, project, task, id})
      }
    },
    comment: {
      create(host, project, task, body) {
        return doReq('/comment/create', {host, project, task, body})
      },
      update(host, project, task, id, body) {
        return doReq('/comment/update', {host, project, task, id, body})
      },
      get(host, project, task, after, limit) {
        return doReq('/comment/get', {host, project, task, after, limit})
      },
      delete(host, project, task, id) {
        return doReq('/comment/delete', {host, project, task, id})
      }
    }
  }
}

// make it available for console hacking
window.api = newApi(false)
export default window.api