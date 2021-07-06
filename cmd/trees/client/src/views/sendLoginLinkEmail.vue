<template>
  <div class="root">
    <h1>send login link email</h1>
    <input
      ref="email"
      v-model="email"
      placeholder="email"
      @blur="validate"
      @keydown.enter="sendLoginLinkEmail"
    />
    <span v-if="!emailIsValid" class="err">email is not valid</span>
    <button @click="sendLoginLinkEmail">send login link</button>
    <a href="/#/login">login</a>
    <span v-if="showCheckEmailsMsg"
      >please check your emails for login link</span
    >
  </div>
</template>

<script>
export default {
  name: "sendLoginLinkEmail",
  data: function () {
    return {
      emailIsValid: true,
      email: "",
      showCheckEmailsMsg: false,
    };
  },
  methods: {
    validate: function () {
      if (this.email.length > 0) {
        this.emailIsValid = /^.+@.+\..+$/.test(this.email);
      }
      return this.emailIsValid;
    },
    sendLoginLinkEmail: function () {
      if (this.validate()) {
        this.$api.user.sendLoginLinkEmail(this.email).then(() => {
          this.showCheckEmailsMsg = true;
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