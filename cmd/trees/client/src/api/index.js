import axios from 'axios'
import firebase from "firebase/app";
import "firebase/messaging";

firebase.initializeApp({
    apiKey: "AIzaSyAg43CfgwC2HLC9x582IMq2UwM6NQ3FRCc",
    projectId: "trees-82a30",
    messagingSenderId: "69294578877",
    appId: "1:69294578877:web:1edb203c55b78f43956bd4",
});
const fcmVapidKey = "BIrxz8PBCCRX2XekUa2zAKdYnKLhj9uHKhuSW5gc0WXWSCeh4Kx3c3GjHselJg0ARUgNJvcZLkd6roGfErpodRM"
let fcm = firebase.messaging()
let fcmClientId = null
let fcmToken = null
let fcmCurrentTopic = null
let fcmEnabled = false
let fcmOnLogout = null
let fcmOnEnabled = null
let fcmOnDisabled = null

let notAuthed = false
let memCache = {}
let meInFlight = false
let userGetInFlight = {}
let globalErrorHandler = null
let fcmUnregisterFn = ()=>{
  if (fcmClientId != null && navigator.sendBeacon != null) {
    navigator.sendBeacon(`/api/user/unregisterFromFCM?args={"client":"${fcmClientId}"}`)
  }
}
window.addEventListener("unload", fcmUnregisterFn);
document.addEventListener("visibilitychange", ()=>{
  if (document.visibilityState === 'visible') {
    if (fcmEnabled == true && fcmCurrentTopic != null) {
      window.api.user.registerForFCM({topic: fcmCurrentTopic})
    }
  } else {
    fcmUnregisterFn()
  }
});

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
  let doReq = (path, args, headers, progFn) => {
    path = `/api${path}`
    if (!isMDoApi || (isMDoApi && mDoSending && !mDoSent)) {
      headers = headers || {}
      headers["X-Client"] = "tlbx-web-client"
      if (fcmClientId != null) {
        headers["X-Fcm-Client"] = fcmClientId
      }
      return axios({
        method: 'put',
        url: path,
        headers: headers,
        data: args,
        onUploadProgress: progFn
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
    fcm: {
      isEnabled(){
        return fcmEnabled
      },
      init(askForPerm) {
        if (askForPerm === true || Notification.permission === "granted") {
            return Notification.requestPermission().then((permission) => {
                if (permission === 'granted') {
                    return fcm.getToken({vapidKey: fcmVapidKey}).then((token)=>{
                        if (token) {
                            fcmToken = token
                        } else {
                            throw "fcm token error"
                        }
                    })
                } else {
                    throw "fcm notifications permission not given"
                }
            })
        } else {
            return new Promise((res, rej)=>{
                rej("fcm notifications permission not given")
            })
        }
      },
      onLogout(fn){
        fcmOnLogout = fn
      },
      onEnabled(fn){
        fcmOnEnabled = fn
      },
      onDisabled(fn){
        fcmOnDisabled = fn
      },
      onMessage(fn){
        fcm.onMessage((msg)=>{
          if (msg != null && msg.data != null) {
            let d = msg.data
            if (d.extraInfo != null) {
              d.extraInfo = JSON.parse(d.extraInfo)
            }
            console.log(d)
            if (fcmClientId === d['X-Fcm-Client']) {
              console.log("fcm came from action on this client")
              return 
            }
            switch (d['X-Fcm-Type']) {
              case 'data':
                fn(d)
                break
              case 'logout':
                if (fcmOnLogout != null) {
                  fcmOnLogout()
                }
                break
              case 'enabled':
                if (fcmOnEnabled != null) {
                  fcmOnEnabled()
                }
                break
              case 'disabled':
                if (fcmOnDisabled != null) {
                  fcmOnDisabled()
                }
                break
              default:
                throw 'unexpected X-Fcm-Type: ' + d['X-Fcm-Type']
            }
          } else {
            console.log("unexpected fcm msg format received", msg)
          }
        })
      }
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
          fcmEnabled = res.fcmEnabled
          return res
        })
      },
      logout() {
        memCache = {}
        return doReq('/user/logout').then(()=>{
          notAuthed = true
          fcmEnabled = false
          fcmCurrentTopic = null
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
            fcmEnabled = res.fcmEnabled
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
      },
      setFCMEnabled(val){
        // true/false
        return doReq('/user/setFCMEnabled', {val}).then(()=>{
          fcmEnabled = val
        })
      },
      registerForFCM(args){
        // topic
        if (fcmEnabled == false) {
          // if fcm isn't enabled just return
          // empty success promise
          return new Promise((res)=>{
            res()
          })
        }
        fcmCurrentTopic = args.topic
        args.token = fcmToken
        args.client = fcmClientId
        return doReq('/user/registerForFCM', args).then((clientId)=>{
          fcmClientId = clientId
          return null
        })
      },
      unregisterFromFCM(){
        if (fcmEnabled && fcmClientId != null) {
          return doReq('/user/unregisterFromFCM', {client: fcmClientId}).then(()=>{
            fcmCurrentTopic = null
          })
        }
      }
    },
    project: {
      create(args) {
        // name, isPublic, currencyCode, hoursPerDay, daysPerWeek, startOn, endOn
        return doReq('/project/create', args)
      },
      one(args) {
        // host, id
        args.ids = [args.id]
        delete args.id
        return this.get(args).then((res)=>{
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
        // host, others, ids, namePrefix, isArchived, isPublic, createdOnMin, createdOnMax, startOnMin, startOnMax, endOnMin, endOnMax, after, sort, asc, limit
        return doReq('/project/get', args)
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
      addUsers(args) {
        // host, project, users
        return doReq('/project/addUsers', args)
      },
      getMe(args) {
        // host, project
        return doReq('/project/getMe', args)
      },
      getUsers(args) {
        // host, project, ids, role, handlePrefix, after, limit
        return doReq('/project/getUsers', args)
      },
      setUserRoles(args) {
        // host, project, users [{id role}]
        return doReq('/project/setUserRoles', args)
      },
      removeUsers(args) {
        // host, project, users [ids]
        return doReq('/project/removeUsers', args)
      },
      getActivities(args) {
        // host, project, task, item, user, occuredAfter, occuredBefore, limit
        return doReq('/project/getActivities', args)
      }
    },
    task: {
      create(args) {
        // host, project, parent, prevSib, name, description, isParallel, user, timeEst, costEst
        return doReq('/task/create', args)
      },
      update(args) {
        // host, project, id, parent, prevSib, name, description, isParallel, user, timeEst, costEst
        return doReq('/task/update', args)
      },
      delete(args) {
        // host, project, id
        return doReq('/task/delete', args)
      },
      get(args) {
        // host, project, id
        return doReq('/task/get', args)
      },
      getAncestors(args) {
        // host, project, id, limit
        return doReq('/task/getAncestors', args)
      },
      getChildren(args) {
        // host, project, id, after, limit
        return doReq('/task/getChildren', args)
      }
    },
    vitem: {
      create(args) {
        // host, project, task, type, est, inc, note
        return doReq('/vitem/create', args)
      },
      update(args) {
        // host, project, task, type, id, value, note
        return doReq('/vitem/update', args)
      },
      get(args) {
        // host, project, task, type, ids, createOnMin, createdOnMax, createdBy, after, asc, limit
        return doReq('/vitem/get', args)
      },
      delete(args) {
        // host, project, task, type, id
        return doReq('/vitem/delete', args)
      }
    },
    file: {
      create(args, progFn) {
        // host, project, task, name, type, size, content
        return doReq('/file/create', args.content, {
          "Content-Name": args.name,
          //"Content-Length": args.size,
          "Content-Type": args.type,
          "Content-Args": JSON.stringify({
            host: args.host,
            project: args.project,
            task: args.task
          })
        }, progFn)
      },
      getContentUrl(args) {
        // host, project, task, id, isDownload
        return `/api/file/getContent?args=${JSON.stringify(args)}`
      },
      getContent(args) {
        // host, project, task, id, isDownload
        return doReq('/file/getContent', args)
      },
      get(args) {
        // host, project, task, ids, createOnMin, createdOnMax, createdBy, after, asc, limit
        return doReq('/file/get', args)
      },
      delete(args) {
        // host, project, task, id
        return doReq('/file/delete', args)
      }
    },
    comment: {
      create(args) {
        // host, project, task, body
        return doReq('/comment/create', args)
      },
      update(args) {
        // host, project, task, id, body
        return doReq('/comment/update', args)
      },
      get(args) {
        // host, project, task, after, limit
        return doReq('/comment/get', args)
      },
      delete(args) {
        // host, project, task, id
        return doReq('/comment/delete', args)
      }
    }
  }
}

// make it available for console hacking
window.api = newApi(false)
export default window.api