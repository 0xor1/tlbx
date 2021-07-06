<template>
  <div class="root">
    <h1>login</h1>
    <input
      ref="email"
      v-model="email"
      placeholder="email"
      @blur="validate"
      @keydown.enter="login"
    />
    <span v-if="!emailIsValid" class="err">email is not valid</span>
    <input
      v-model="pwd"
      placeholder="pwd"
      type="password"
      @blur="validate"
      @keydown.enter="login"
    />
    <span v-if="pwdErr.length > 0" class="err">{{ pwdErr }}</span>
    <button @click="login">login</button>
    <a href="/#/register">register</a>
    <span><a href="/#/sendLoginLinkEmail">send login link</a></span>
    <span class="small"><a href="/#/resetPwd">forgot pwd?</a></span>
  </div>
</template>

<script>
export default {
  name: "login",
  data: function () {
    return {
      emailIsValid: true,
      email: "",
      pwdErr: "",
      pwd: "",
    };
  },
  methods: {
    validate: function () {
      if (this.email.length > 0) {
        this.emailIsValid = /^.+@.+\..+$/.test(this.email);
      }
      if (this.pwd.length > 0) {
        if (this.pwd.length < 8) {
          this.pwdErr = "pwd must be 8 at least 8 characters long";
        } else if (!/[0-9]/.test(this.pwd)) {
          this.pwdErr = "pwd must contain a number";
        } else if (!/[a-z]/.test(this.pwd)) {
          this.pwdErr = "pwd must contain a lowercase letter";
        } else if (!/[A-Z]/.test(this.pwd)) {
          this.pwdErr = "pwd must contain an uppercase letter";
        } else if (!/[\W]/.test(this.pwd)) {
          this.pwdErr = "pwd must contain an non alphernumeric character";
        } else {
          this.pwdErr = "";
        }
      }
      return this.emailIsValid && this.pwdErr.length === 0;
    },
    login: function () {
      if (this.validate()) {
        this.$api.user.login(this.email, this.pwd).then((me) => {
          this.$u.rtr.goto(`/host/${me.id}/projects`);
        });
      }
    },
  },
  mounted() {
    this.$nextTick(() => {
      this.$refs.email.focus();
    });
  },
};
</script>

<style scoped lang="scss">
div.root {
  & > * {
    display: block;
    margin-bottom: 5px;
  }
  button,
  a {
    display: inline;
    margin-right: 15px;
  }
  .small {
    font-size: 0.75pc;
  }
}
</style>