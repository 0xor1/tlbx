<template>
  <div class="root">
    <h1>register</h1>
    <input ref="alias" v-model="alias" placeholder="alias" @blur="validate" @keydown.enter="register">
    <span v-if="aliasErr.length > 0" class="err">{{aliasErr}}</span>
    <input v-model="handle" placeholder="handle" @blur="validate" @keydown.enter="register">
    <span v-if="handleErr.length > 0" class="err">{{handleErr}}</span>
    <input v-model="email" placeholder="email" @blur="validate" @keydown.enter="register">
    <span v-if="!emailIsValid" class="err">email is not valid</span>
    <input v-model="pwd" placeholder="pwd" type="password" @blur="validate" @keydown.enter="register">
    <span v-if="pwdErr.length > 0" class="err">{{pwdErr}}</span>
    <input v-model="confirmPwd" placeholder="confirm pwd" type="password" @blur="validate" @keydown.enter="register">
    <span v-if="!pwdsMatch" class="err">pwds don't match</span>
    <button @click="register">register</button>
    <a href="/#/login">login</a>
    <span v-if="registered">check your emails for confirmation link</span>
    <span v-if="alreadyLoggedIn" class="err">already logged in <a href="/#/lists">go to your lists</a></span>
    <span v-if="registerErr.length > 0" class="err">{{registeredErr}}</span>
  </div>
</template>

<script>
  export default {
    name: 'login',
    data: function() {
      return {
        aliasErr: true,
        alias: "",
        handleErr: true,
        handle: "",
        emailIsValid: true,
        email: "",
        pwdErr: "",
        pwd: "",
        pwdsMatch: true,
        confirmPwd: "",
        registered: false,
        alreadyLoggedIn: false,
        registerErr: ""
      }
    },
    methods: {
      validate: function(){
        if (this.alias.length > 20) {
            this.aliasErr = "alias must be less than 20 characters long"
        } else {
            this.aliasErr = ""
        }
        this.handle = this.handle.toLowerCase()
        if (this.handle.length > 0 || this.email.length > 0) {
          if (this.handle.length < 1 || this.handle.length > 20) {
              this.handleErr = "handle must be 1 - 20 characters long"
          } else if (!/^[_a-z0-9]{1,20}$/.test(this.handle)) {
            this.handleErr = "handle may only contain underscores and alphanumerical characters"
          } else {
              this.handleErr = ""
          }
        } else {
              this.handleErr = ""
          }
        if (this.email.length > 0) {
          this.emailIsValid = /^.+@.+\..+$/.test(this.email)
        }
        if (this.pwd.length > 0) {
          if (this.pwd.length < 8) {
            this.pwdErr = "pwd must be 8 at least 8 characters long"
          } else if (!/[0-9]/.test(this.pwd)) {
            this.pwdErr = "pwd must contain a number"
          } else if (!/[a-z]/.test(this.pwd)) {
            this.pwdErr = "pwd must contain a lowercase letter"
          } else if (!/[A-Z]/.test(this.pwd)) {
            this.pwdErr = "pwd must contain an uppercase letter"
          } else if (!/[\W]/.test(this.pwd)) {
            this.pwdErr = "pwd must contain an non alphernumeric character"
          } else {
            this.pwdErr = ""
          }
        }
        this.pwdsMatch = this.confirmPwd.length === 0 || this.pwd === this.confirmPwd
        return this.emailIsValid && this.pwdErr.length === 0 && this.pwdsMatch
      },
      register: function(){
        if (this.validate()) {
          this.$api.user.register(this.alias, this.handle, this.email, this.pwd, this.confirmPwd).then(()=>{
            this.registered = true
          }).catch((err)=>{
            this.alreadyLoggedIn = err.response.data === "already logged in"
            if (!this.alreadyLoggedIn) {
              this.registerErr = err.response.data
            } 
          })
        }
      }
    },
    mounted(){
      this.$nextTick(()=>{
        this.$refs.alias.focus()
      })
    }
  }
</script>

<style scoped lang="scss">
div.root {
  padding: 2.6pc 0 0 1.3pc;
  & > * {
    display: block;
    margin-bottom: 5px;
  }
  button, a{
    display: inline;
    margin-right: 15px;
  }
}
.err{
  color: #c33;
}
</style>