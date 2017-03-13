<!-- src/components/Login.vue -->
<template>
  <div class="login">
    <p class="error" v-if="error">
      <p>{{ error }}</p>
    </p>
    <form method="post" class="pure-form pure-form-stacked" @submit.prevent="submit">
        <input
          type="text"
          class="form-control"
          placeholder="Username"
          v-model="credentials.username"
        >
        <input
          type="password"
          class="form-control"
          placeholder="Password"
          v-model="credentials.password"
        >
      <p><input type="submit" value="Login" class="pure-button"></p>
    </form>
    
  </div>
</template>

<script>
//import Auth from '../Auth'

export default {
  data() {
    return {
      // We need to initialize the component with any
      // properties that will be used in it
      credentials: {
        username: '',
        password: ''
      },
      error: ''
    }
  },
  methods: {
    submit: function() {
      var credentials = {
        username: this.credentials.username,
        password: this.credentials.password
      }
      this.$parent.auth.login(this, credentials, '/albums')
    }
  }

}
</script>