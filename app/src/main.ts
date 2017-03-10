//import * as _ from 'lodash';
//import * as $ from 'jquery';
//import {test} from './test';
import * as Vue from 'vue';
import * as Router from 'vue-router';
Vue.use(Router);

import * as AudioPlayerComponent from './components/AudioPlayer.vue';
import * as AlbumsComponent from './components/Albums.vue';
import * as AlbumComponent from './components/Album.vue';

import './AudioPlayer';

var router = new Router({
    routes: [
          { path: '/', component: AlbumsComponent },
          { path: '/albums', component: AlbumsComponent },
          { path: '/albums/:id', component: AlbumComponent }
    ],
});

var app = new Vue({
    el: '#app',
    router: router,
    components: {
        'audio-player': AudioPlayerComponent,
        'albums': AlbumsComponent,
        'album': AlbumComponent,
    }
});

