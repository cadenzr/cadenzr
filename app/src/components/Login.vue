<template>
  <div class="login">
    <div class="error">
      <p v-if="error"><span class="fa fa-warning"></span> {{ error }}</p>
    </div>
    <form method="post" class="pure-form pure-form-stacked" @submit.prevent="submit">
        <input
          type="text"
          class="form-control"
          placeholder="Username"
          v-model="username"
        >
        <input
          type="password"
          class="form-control"
          placeholder="Password"
          v-model="password"
        >
      <p><input type="submit" value="Login" class="pure-button"></p>
    </form>
    
  </div>
</template>

<script>
//import Auth from '../Auth'

let Api = require('./../Api').default;
let router = require('./../main').router;

export default {
  data() {
    return {
      username: '',
      password: '',
      error: '',
    }
  },
  mounted: function() {
    let self = this;
    let username = localStorage.getItem('login.username');
    let password = localStorage.getItem('login.password');
    if(username && password) {
      Api.authenticate(username, password)
      .then(() => {
        self.$router.go('/albums');
      })
      .catch((reason) => {
        self.error = 'Automatic login failed. Reason: ' + reason.message;
        localStorage.removeItem('login.username');
        localStorage.removeItem('login.password');
        self.$forceUpdate();
      });
    }
  },
  methods: {
    submit: function() {
      let self = this;

      Api.authenticate(self.username, self.password)
      .then(() => {
        return Api.getMe();
      })
      .then((me) => {
        localStorage.setItem('login.username', self.username);
        localStorage.setItem('login.password', self.password);

        self.$router.go('/albums');
        //self.$forceUpdate();
      })
      .catch((reason) => {
        self.error = reason.message;
        self.$forceUpdate();
      });
      //this.$parent.auth.login(this, credentials, '/albums')
    }
  }

}
</script>