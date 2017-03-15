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
                <router-link :to="{ path: '/artists' }">
                    <span class="fa fa-fw fa-microphone"></span> Artists
                </router-link>
            </li>
            <li v-on:drop="drop" v-on:dragover="dragover">
                <router-link :to="{ path: '/current-queue' }"  >
                    <span class="fa fa-fw fa-play-circle-o"></span> Playing Now
                </router-link>
            </li>
        </ul>
    </nav>
    
    
    <div class="settings">
        <nav>
            <ul v-if="login">
                <li>
                    <span class="fa fa-fw fa-user-circle-o"></span> {{me.username}}
                </li>
                <li>
                    <a @click="logout">
                        <span class="fa fa-fw fa-sign-out"></span> Logout
                    </a>
                </li>
                <li>
                    <a v-if="!scanning" @click="scan">
                        <span class="fa fa-fw fa-refresh"></span> Scan
                    </a>
                    <span v-if="scanning">
                        <span class="fa fa-fw fa-spinner fa-spin"></span> Scanning...
                    </span>
                </li>
            </ul>
        </nav>
    </div>
    
    
  </div>
</template>

<script>
    
let PubSub = require('./../PubSub').default;
let Api = require('./../Api').default;
let ApiEvents = require('./../Api').events;
let Song = require('./../Song').default;
let _ = require('lodash');
let AudioPlayer = require('./../AudioPlayer').default;

export default {
  data() {
    return {
      // We need to initialize the component with any
      // properties that will be used in it
      login: Api.isAuthenticated(),
      me: {},
      subscriptions: [],
      scanning: false,
    }
  },
  methods: {
      drop: function(e) {
          let songs = e.dataTransfer.getData('songs');
          if(songs) {
              songs = JSON.parse(songs);
              songs = _.map(songs, (song) => {
                  return new Song(song);
              });

            AudioPlayer.setQueue(songs);
            AudioPlayer.reload()
            .then(() => {
                AudioPlayer.play();
            });
          }

      },
      dragover: function(e) {
          e.preventDefault();
      },
      logout: function() {
          //this.$parent.auth.logout();
          Api.logout();
      },
      scan: function() {
          let self = this;
          self.scanning = true;
          Api.scan()
          .then(() => {
              self.scanning = false;
          })
          .catch(() => {
              self.scanning = false;
          });
      },
  },
  mounted: function () {
        let self = this;
        self.subscriptions.push(PubSub.subscribe(ApiEvents.Authenticated, () => {
            self.login = true;
            Api.getMe()
            .then((me) => {
                self.me = me;
            });
            self.$forceUpdate();
        }));
        self.subscriptions.push(PubSub.subscribe(ApiEvents.LoggedOut, () => {
            self.login = false;
            self.$forceUpdate();
        }));

        Api.getMe()
        .then((me) => {
            self.me = me;
        });
  },
  beforeDestroy: function () {
      _.forEach(this.subscriptions, (s) => {
          PubSub.unsubscribe(s);
      });
  },

}
</script>