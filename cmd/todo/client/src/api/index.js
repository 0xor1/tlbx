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
    me: {
      register: (email, pwd, confirmPwd) => {
        return doReq('/me/register', {email, pwd, confirmPwd})
      },
      resendActivateLink: (email) => {
        return doReq('/me/resendActivateLink', {email})
      },
      activate: (email, code) => {
        return doReq('/me/activate', {email, code})
      },
      changeEmail: (newEmail) => {
        return doReq('/me/changeEmail', {newEmail})
      },
      resendChangeEmailLink: () => {
        return doReq('/me/resendChangeEmailLink')
      },
      confirmChangeEmail: (me, code) => {
        return doReq('/me/confirmChangeEmail', {me, code})
      },
      resetPwd: (email) => {
        return doReq('/me/resetPwd', {email})
      },
      setPwd: (currentPwd, newPwd, confirmNewPwd) => {
        return doReq('/me/setPwd', {currentPwd, newPwd, confirmNewPwd})
      },
      delete: (pwd) => {
        return doReq('/me/delete', {pwd})
      },
      login: (email, pwd) => {
        return doReq('/me/login', {email, pwd})
      },
      logout: () => {
        return doReq('/me/logout')
      },
      get: () => {
        if (memCache.me) {
          return new Promise((resolve) => {
            resolve(memCache.me)
          })
        }
        return doReq('/me/get').then((res) => {
          memCache.me = res
          return res
        })
      }
    },
    list: {
      create: (name) => {
        return doReq('/list/create', {name})
      },
      one: (id) => {
        return doReq('/list/get', {ids: [id]}).then((res)=>{
          if (res.set.length === 1) {
            return res.set[0]
          }
          return null
        })
      },
      get: (args) => {
        return doReq('/list/get', args)
      },
      update: (id, name) => {
        return doReq('/list/update', {id, name: {v: name}})
      },
      delete: (ids) => {
        return doReq('/list/delete', {ids})
      }
    },
    item: {
      create: (list, name) => {
        return doReq('/item/create', {list, name})
      },
      one: (list, id) => {
        return doReq('/item/get', {list, ids: [id]}).then((res)=>{
          if (res.set.length === 1) {
            return res.set[0]
          }
          return null
        })
      },
      get: (args) => {
        return doReq('/item/get', args)
      },
      update: (list, id, name, complete) => {
        let args = {
          list,
          id
        }
        if (name !== undefined) {
          args.name = {v: name}
        }
        if (complete !== undefined) {
          args.complete = {v: complete}
        }
        return doReq('/item/update', args)
      },
      delete: (list, ids) => {
        return doReq('/item/delete', {list, ids})
      }
    }
  }
}

// make it available for console hacking
window.api = newApi(false)
export default window.api
