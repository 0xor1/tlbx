<template>
  <div class="root">
    <h1>Register</h1>
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
  import api from '@/api'
  export default {
    name: 'login',
    data: function() {
      return {
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
          api.me.register(this.email, this.pwd, this.confirmPwd).then(()=>{
            this.registered = true
          }).catch((err)=>{
            this.alreadyLoggedIn = err.response.data === "already logged in"
            if (!this.alreadyLoggedIn) {
              this.registerErr = err.response.data
            } 
          })
        }
      }
    }
  }
</script>

<style scoped lang="scss">
div.root {
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