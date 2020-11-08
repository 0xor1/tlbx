import axios from 'axios'

let memCache = {}

let newApi = (isMDoApi) => {
  let mDoSending = false
  let mDoSent = false
  let awaitingMDoList = []
  let doReq = (path, args) => {
    path = '/api'+path
    if (!isMDoApi || (isMDoApi && mDoSending && !mDoSent)) {
      return axios({
        method: 'put',
        url: path,
        headers: {"X-Client": "tlbx-web-client"},
        data: args
      }).then((res) => {
        return res.data
      }).catch((err) => {
        throw {
          status: err.response.status,
          body: err.response.data
        }
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
          }
        }
        if (ready) {
          resolve()
        }
      }
      let mdoErrors = []
      mdoErrors.isMDoErrors = true
      let mDoComplete = false
      let mDoCompleterFunc
      mDoCompleterFunc = (resolve, reject) => {
        if (mDoComplete) {
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
          let key = '' + i
          mDoObj[key] = {
            path: awaitingMDoList[i].path,
            args: awaitingMDoList[i].args
          }
        }
        doReq('/mdo', mDoObj).then((res) => {
          for (let i = 0, l = awaitingMDoList.length; i < l; i++) {
            let key = '' + i
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
          mDoComplete = true
          mDoSending = false
          mDoSent = true
        })
      })
      return new Promise(mDoCompleterFunc)
    },
    user: {
      register: (alias, email, pwd, confirmPwd) => {
        return doReq('/user/register', {alias, email, pwd, confirmPwd})
      },
      resendActivateLink: (email) => {
        return doReq('/user/resendActivateLink', {email})
      },
      activate: (email, code) => {
        return doReq('/user/activate', {email, code})
      },
      changeEmail: (newEmail) => {
        return doReq('/user/changeEmail', {newEmail})
      },
      resendChangeEmailLink: () => {
        return doReq('/user/resendChangeEmailLink')
      },
      confirmChangeEmail: (me, code) => {
        return doReq('/user/confirmChangeEmail', {me, code})
      },
      resetPwd: (email) => {
        return doReq('/user/resetPwd', {email})
      },
      setAlias: (alias) => {
        return doReq('/user/setAlias', {alias}).then(()=>{
          memCache.me.alias = alias
        })
      },
      setPwd: (currentPwd, newPwd, confirmNewPwd) => {
        return doReq('/user/setPwd', {currentPwd, newPwd, confirmNewPwd})
      },
      delete: (pwd) => {
        return doReq('/user/delete', {pwd})
      },
      login: (email, pwd) => {
        return doReq('/user/login', {email, pwd}).then((res)=>{
          memCache.me = res
          memCache[res.id] = res
          return res
        })
      },
      logout: () => {
        memCache = {}
        return doReq('/user/logout')
      },
      me: () => {
        if (memCache.me) {
          return new Promise((resolve) => {
            resolve(memCache.me)
          })
        }
        return doReq('/user/me').then((res) => {
          memCache.me = res
          memCache[res.id] = res
          return res
        })
      },
      get: (ids) => {
        let toGet = []
        let found = []
        ids.forEach((id)=>{
          if (memCache[id]) {
            found.push(memCache[id])
          } else {
            found.push(memCache[id])
            toGet.push(id)
          }
        })
        if (toGet.length === 0) {
          return new Promise((resolve) => {
            resolve(found)
          })
        }
        return doReq('/user/get', {users: toGet}).then((res) => {
          if (res != null) {
            res.forEach((user)=>{
              memCache[user.id] = user
            })
          }
          return res
        })
      }
    },
    project: {
      
    }
  }
}

// make it available for console hacking
window.api = newApi(false)
export default window.api
