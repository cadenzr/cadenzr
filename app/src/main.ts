//import * as _ from 'lodash';
//import * as $ from 'jquery';
//import {test} from './test';
import * as Vue from 'vue';
import * as Router from 'vue-router';
import * as VueResource from 'vue-resource';
Vue.use(VueResource)
Vue.use(Router);

// authentication service
import Auth from './Auth';

import * as AudioPlayerComponent from './components/AudioPlayer.vue';
import * as AlbumsComponent from './components/Albums.vue';
import * as AlbumComponent from './components/Album.vue';
import * as CurrentQueueComponent from './components/CurrentQueue.vue';
import * as LoginComponent from './components/Login.vue';
import * as SidebarComponent from './components/Sidebar.vue';


import './AudioPlayer';
import Song from './Song';
import {events as AudioPlayerEvents} from './AudioPlayer';

import Notifier from './Notifier';
import PubSub from './PubSub';

PubSub.subscribe(AudioPlayerEvents.SongChanged, (song: Song) => {
    Notifier.notify('Playing song: ' + song.name);
});


export var authentication = Auth;

export var router = new Router({
    routes: [
          { path: '/', component: AlbumsComponent, meta: { requiresAuth: true } },
          { path: '/albums', component: AlbumsComponent, meta: { requiresAuth: true } },
          { path: '/albums/:id', component: AlbumComponent, meta: { requiresAuth: true } },
          { path: '/current-queue', component: CurrentQueueComponent, meta: { requiresAuth: true } },
          { path: '/login', component: LoginComponent }

    ],
});


router.beforeEach(function (to, from, next) {
    
    if(authentication.ready) {
        if (to.meta.requiresAuth && !authentication.authenticated) {
            // if route requires auth and user isn't authenticated
            next('/login')
        } else {
            next()
        }
    }
    else {        
        // Wait until auth is initialized
        new Promise(function(resolve, reject) {
            authentication.checkLocalStorage();
            resolve("checkLocalStorage");
        }).then(function() {
            if (to.meta.requiresAuth && !authentication.authenticated) {
                // if route requires auth and user isn't authenticated
                next('/login')
            } else {
                next()
            }
        });
    }  
    
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
    },
    data: function() {
        return { user: {} };
    },
    computed: {
        auth: function() {
            return authentication;
        }
    },
});

(<any>Vue).http.options.emulateJSON = true;

