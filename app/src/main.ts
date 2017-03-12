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


import './AudioPlayer';
import Song from './Song';
import {events as AudioPlayerEvents} from './AudioPlayer';

import Notifier from './Notifier';
import PubSub from './PubSub';

PubSub.subscribe(AudioPlayerEvents.SongChanged, (song: Song) => {
    Notifier.notify('Playing song: ' + song.name);
});


export var testAuth = Auth;

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
    console.log(to);
    console.log(testAuth);
    if (to.meta.requiresAuth && !testAuth.authenticated) {
        // if route requires auth and user isn't authenticated
        console.log("Not logged in");
        next('/login')
    } else {
        console.log("Logged in");
        next()
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
    },
    data: function() {
        return { user: {} };
    },
    computed: {
        auth: function() {
            return testAuth;
        }
    },
    methods: {
        checkLocalStorage: function() {
            console.log(localStorage)
            if (localStorage.user) {
                console.log(localStorage.user)
                this.user = JSON.parse(localStorage.user);
                Vue.http.headers.common['Authorization'] = 'Bearer ' + this.user.token;
                testAuth.authenticated = true;
            }
        },
        logout: function() {
            this.user = {};
            Auth.logout();
        }
    },
    mounted: function() {
        testAuth.test = "gataap";
        console.log("mounted");
        this.checkLocalStorage();
    }
});

Vue.http.options.emulateJSON = true;

