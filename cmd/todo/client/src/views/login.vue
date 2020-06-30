<template>
  <div class="root">
    <h1>login</h1>
    <input v-model="email" placeholder="email" @blur="validate" @keydown.enter="login">
    <span v-if="!emailIsValid" class="err">email is not valid</span>
    <input v-model="pwd" placeholder="pwd" type="password" @blur="validate" @keydown.enter="login">
    <span v-if="pwdErr.length > 0" class="err">{{pwdErr}}</span>
    <button @click="login">login</button>
    <a href="/#/register">register</a>
  </div>
</template>

<script>
  import api from '@/api'
  import router from '@/router'
  export default {
    name: 'login',
    data: function() {
      return {
        emailIsValid: true,
        email: "",
        pwdErr: "",
        pwd: ""
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
        return this.emailIsValid && this.pwdErr.length === 0
      },
      login: function(){
        if (this.validate()) {
          api.user.login(this.email, this.pwd).then(()=>{
            router.push('/lists')
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