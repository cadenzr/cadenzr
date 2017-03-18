<template>
  <div class="login">
    <div class="error">
      <p v-if="error"><span class="fa fa-warning"></span> {{ error }}</p>
    </div>
    <form method="post"
          class="pure-form pure-form-stacked"
          @submit.prevent="submit">
      <input type="text"
             class="form-control"
             placeholder="Username"
             v-model="username">
      <input type="password"
             class="form-control"
             placeholder="Password"
             v-model="password">
      <p>
        <input type="submit"
               value="Login"
               class="pure-button">
      </p>
    </form>
  
  </div>
</template>

<script lang="ts">
//import Auth from '../Auth'

import Api from './../Api';
    import Vue from 'vue';


interface Login extends Vue {
  username: string;
  password: string;
  error: string;
}

export default {
  name: 'login',
  data() {
    return {
      username: '',
      password: '',
      error: '',
    }
  },
  mounted: function() {
  },
  methods: {
    submit: function() {
      let self = this;

      Api.authenticate(self.username, self.password)
      .then(() => {
        self.$router.push('/albums');
        self.$forceUpdate();
      })
      .catch((reason) => {
        self.error = reason.message;
        self.$forceUpdate();
      });
      //this.$parent.auth.login(this, credentials, '/albums')
    }
  }

} as Vue.ComponentOptions<Login>;
</script>