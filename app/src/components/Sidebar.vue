<template>
  <div class="sidebar">
    <!--<h1>Cadenzr</h1>-->
    
    <div class="logo">
        
    </div>
    
    <nav>
        <ul>
            <li>
                <router-link :to="{ path: '/albums' }">
                    <span class="fa fa-fw fa-caret-square-o-right"></span> Albums
                </router-link>
            </li>
            <li>
                <router-link :to="{ path: '/' }">
                    <span class="fa fa-fw fa-microphone"></span> Artists
                </router-link>
            </li>
            <li>
                <router-link :to="{ path: '/' }">
                    <span class="fa fa-fw fa-play-circle-o"></span> Playing Now
                </router-link>
            </li>
        </ul>
    </nav>
    
    
    <div class="settings">
        <nav>
            <ul v-if="login">
                <li>
                    <span class="fa fa-fw fa-user-circle-o"></span> {{name}}
                </li>
                <li>
                    <a @click="logout">
                        <span class="fa fa-fw fa-sign-out"></span> Logout
                    </a>
                </li>
            </ul>
        </nav>
    </div>
    
    
  </div>
</template>

<script>
    
let PubSub = require('./../PubSub').default;
let AuthEvents = require('./../Auth').events;

export default {
  data() {
      console.log(this.$parent.auth.authenticated);
    return {
      // We need to initialize the component with any
      // properties that will be used in it
      login: this.$parent.auth.authenticated,
      name: this.$parent.auth.name,
      subscriptions: [],
    }
  },
  methods: {
      logout: function() {
          this.$parent.auth.logout();
      }
  },
  mounted: function () {
      let self = this;
  
      self.subscriptions.push(PubSub.subscribe(AuthEvents.LoggedIn, () => {
          self.login = true;
      }));
      self.subscriptions.push(PubSub.subscribe(AuthEvents.LoggedOut, () => {
          self.login = false;
      }));
  },
  beforeDestroy: function () {
      _.forEach(this.subscriptions, (s) => {
          PubSub.unsubscribe(s);
      });
  },

}
</script>