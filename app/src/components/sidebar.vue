<template>
    <div class="sidebar" v-bind:class="{ active: isActive }">
        
        <a class="toggle-hide" @click="toggleNav"><span class="fa fa-fw fa-close"></span></a>
        <a class="toggle-show" @click="toggleNav"><span class="fa fa-fw fa-bars"></span></a>
        
        <!--<h1>Cadenzr</h1>-->
    
        <div class="logo">
    
        </div>
    
        <nav>
            <ul>
                <li>
                    <router-link :to="{ path: '/albums' }" v-on:click.native="toggleNav">
                        <span class="fa fa-fw fa-caret-square-o-right"></span> Albums
                    </router-link>
                </li>
                <li>
                    <router-link :to="{ path: '/artists' }" v-on:click.native="toggleNav">
                        <span class="fa fa-fw fa-microphone"></span> Artists
                    </router-link>
                </li>
    
                <li v-on:drop="dropQueue"
                    v-on:dragover="dragover">
                    <router-link :to="{ path: '/current-queue' }" v-on:click.native="toggleNav">
                        <span class="fa fa-fw fa-play-circle-o"></span> Playing Now
                    </router-link>
                </li>
    
                <li>
                    <span class="fa fa-fw fa-list"></span> Playlists <a v-on:click="showAddPlaylist = !showAddPlaylist;"><span class="fa fa-fw" v-bind:class="{'fa-plus': !showAddPlaylist, 'fa-times': showAddPlaylist}"></span></a>
    
                </li>
    
                <li v-if="showAddPlaylist">
                    <form class="pure-form">
                        <span class="fa fa-fw"></span>
                        <input ref="playlistName"
                               v-model="playlistName"
                               v-on:keyup.enter="createPlaylist()"
                               v-on:keyup.esc="showAddPlaylist = false;"
                               type="text"
                               class="">
                    </form>
                </li>
    
            </ul>
    
            <ul class="playlists">
                <li v-on:dragover="dragover"
                    v-on:drop="dropPlaylist(playlist, $event)"
                    v-for="playlist in playlists">
                    <span class="fa fa-fw"></span>
                    <router-link :to="{ path: '/playlists/' + playlist.id }" v-on:click.native="toggleNav">
                        <span class="fa fa-fw fa-music"></span> {{playlist.name}}
                    </router-link>
                    <span class="fa fa-fw fa-times"
                          v-on:click="deletePlaylist(playlist)"></span>
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
                        <a v-if="!scanning"
                           @click="scan">
                            <span class="fa fa-fw fa-refresh"></span> Scan
                        </a>
                        <span v-if="scanning">
                            <span class="fa fa-fw fa-spinner fa-spin"></span> Scanning...
                        </span>
                    </li>
                    <li>
                        <router-link :to="{ path: '/upload' }" v-on:click.native="toggleNav">
                            <span class="fa fa-fw fa-upload"></span> Upload
                        </router-link>
                    </li>
                </ul>
            </nav>
        </div>
    
    </div>
</template>

<script lang="ts">
    
    import Vue from 'vue';
    
    import * as $ from 'jquery';
    import * as _ from 'lodash';
    import Api from './../Api';
    import Song from './../Song';
    import Playlist from './../Playlist';
    import PubSub from './../PubSub';
    import {events as AudioPlayerEvents} from './../AudioPlayer';
    import {events as ApiEvents} from './../Api';
    import AudioPlayer from './../AudioPlayer';

        interface Sidebar extends Vue {
      login: boolean;
      me: any;
      subscriptions: Array<any>;
      scanning: boolean;
      playlists: Array<Playlist>;
      showAddPlaylist: boolean;
      playlistName: string;
      isActive: boolean;
    }

export default {
    name: 'sidebar',
  data() {
    return {
      // We need to initialize the component with any
      // properties that will be used in it
      login: Api.isAuthenticated(),
      me: {},
      subscriptions: [],
      scanning: false,
      playlists: [],
      showAddPlaylist: false,
      playlistName: '',
      isActive: false,
    }
  },
  methods: {
      toggleNav: function() {
        this.isActive = !this.isActive;
      },
      dropQueue: function(e:any) {
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
      dropPlaylist: function(playlist:any, e:any) {
          let songs = e.dataTransfer.getData('songs');
          if(songs) {
              songs = JSON.parse(songs);
              songs = _.map(songs, (song:Song) => {
                  return new Song(song);
              });

            Api.addSongsToPlaylist(songs, playlist);
          }

      },
      dragover: function(e:any) {
          e.preventDefault();
      },
      logout: function() {
          //this.$parent.auth.logout();
          Api.logout();
          this.isActive = false;
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
      createPlaylist: function() {
          let self = this;
          let playlist = new Playlist();
          playlist.name = self.playlistName;
          Api.createPlaylist(playlist)
          .then((playlist : any) => {
              self.playlists.push(playlist);
          });

          self.playlistName = '';
          self.showAddPlaylist = false;
      },

      deletePlaylist: function(playlist:any) {
          let self = this;

          Api.deletePlaylist(playlist)
          .then(() => {
              _.remove(self.playlists, {id: playlist.id});
              // Required because the array is mutated.
              self.$forceUpdate();
          });
      },
  },
  mounted: function () {
        let self = <any>this;
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

        self.$watch('showAddPlaylist', () => {
            if(self.showAddPlaylist) {
                self.$refs.playlistName.focus();
            }
        });

        Api.getMe()
        .then((me) => {
            self.me = me;
        });

        Api.getPlaylists()
        .then((playlists) => {
            self.playlists = playlists;
        });
  },
  beforeDestroy: function () {
      _.forEach(this.subscriptions, (s) => {
          PubSub.unsubscribe(s);
      });
  },

} as Vue.ComponentOptions<Sidebar>;

</script>