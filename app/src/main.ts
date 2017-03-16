//import * as _ from 'lodash';
//import * as $ from 'jquery';
//import {test} from './test';
import * as Vue from 'vue';
import * as Router from 'vue-router';
import * as VueResource from 'vue-resource';
import * as Dropzone from 'vue2-dropzone';
import * as _ from 'lodash';

Vue.use(VueResource)
Vue.use(Router);
Vue.component( 'dropzone', Dropzone );

import * as AudioPlayerComponent from './components/AudioPlayer.vue';
import * as AlbumsComponent from './components/Albums.vue';
import * as AlbumComponent from './components/Album.vue';
import * as CurrentQueueComponent from './components/CurrentQueue.vue';
import * as LoginComponent from './components/Login.vue';
import * as SidebarComponent from './components/Sidebar.vue';
import * as PlaylistComponent from './components/Playlist.vue';
import * as UploadComponent from './components/Upload.vue';


import './AudioPlayer';
import Song from './Song';
import Api from './Api';
import {events as AudioPlayerEvents} from './AudioPlayer';
import {events as ApiEvents} from './Api';

import Notifier from './Notifier';
import PubSub from './PubSub';

PubSub.subscribe(AudioPlayerEvents.SongStarted, (song: Song) => {
    Notifier.notify('Playing song: ' + song.name);
});

PubSub.subscribe(AudioPlayerEvents.SongEnded, (song: Song) => {
    Api.incrementPlayed(song);
});

export var router = new Router({
    routes: [
          { path: '/', component: AlbumsComponent, meta: { requiresAuth: true } },
          { path: '/albums', component: AlbumsComponent, meta: { requiresAuth: true } },
          { path: '/albums/:id', component: AlbumComponent, meta: { requiresAuth: true } },
          { path: '/current-queue', component: CurrentQueueComponent, meta: { requiresAuth: true } },
          { path: '/login', component: LoginComponent },
          { path: '/playlists/:id', component: PlaylistComponent, meta: { requiresAuth: true } },
          { path: '/upload', component: UploadComponent, meta: { requiresAuth: true } },

    ],
});


router.beforeEach(function (to, from, next) {
    if(to.meta.requiresAuth && !Api.isAuthenticated()) {
        next('/login');
        return;
    }

    if(to.path === '/login' && Api.isAuthenticated()) {
        next('/');
        return;
    }

    return next();
})

var app = new Vue({
    el: '#app',
    router: router,
    components: {
        'audio-player': AudioPlayerComponent,
        'albums': AlbumsComponent,
        'album': AlbumComponent,
        'current-queue': CurrentQueueComponent,
        'Sidebar': SidebarComponent,
        'playlist': PlaylistComponent,
        'upload': UploadComponent,
    },
    mounted: function() {
        let self = this;
        (<any>this).subscriptions.push(PubSub.subscribe(ApiEvents.LoggedOut, () => {
            (<any>self).$router.push('/login');
        }));
    },
    beforeDestroy: () => {
        _.forEach(this.subscriptions, (s) => {
            PubSub.unsubscribe(s);
        });
    },
    data: function() {
        return { 
            user: {},
            subscriptions: [],
        };
    },
});
