<template>
  <div class="root">
    <h1>reset pwd</h1>
    <input
      ref="email"
      autocomplete="email"
      v-model="email"
      placeholder="email"
      @blur="validate"
      @keydown.enter="resetPwd"
    />
    <span v-if="!emailIsValid" class="err">email is not valid</span>
    <button @click="resetPwd">reset</button>
    <a href="/#/login">login</a>
    <span v-if="showCheckEmailsMsg">please check your emails for reset pwd</span>
  </div>
</template>

<script>
export default {
  name: "resetPwd",
  data: function () {
    return {
      emailIsValid: true,
      email: "",
      showCheckEmailsMsg: false
    };
  },
  methods: {
    validate: function () {
      if (this.email.length > 0) {
        this.emailIsValid = /^.+@.+\..+$/.test(this.email);
      }
      return this.emailIsValid;
    },
    resetPwd: function () {
      if (this.validate()) {
        this.$api.user.resetPwd(this.email).then(() => {
          this.showCheckEmailsMsg = true
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