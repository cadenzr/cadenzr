//import * as _ from 'lodash';
//import * as $ from 'jquery';
//import {test} from './test';
import * as Vue from 'vue';
import * as Router from 'vue-router';
Vue.use(Router);

import * as AudioPlayerComponent from './components/AudioPlayer.vue';
import * as AlbumsComponent from './components/Albums.vue';
import * as AlbumComponent from './components/Album.vue';
import * as CurrentQueueComponent from './components/CurrentQueue.vue';


import './AudioPlayer';
import Song from './Song';
import {events as AudioPlayerEvents} from './AudioPlayer';

import Notifier from './Notifier';
import PubSub from './PubSub';

PubSub.subscribe(AudioPlayerEvents.SongChanged, (song: Song) => {
    Notifier.notify('Playing song: ' + song.name);
});

var router = new Router({
    routes: [
          { path: '/', component: AlbumsComponent },
          { path: '/albums', component: AlbumsComponent },
          { path: '/albums/:id', component: AlbumComponent },
          { path: '/current-queue', component: CurrentQueueComponent }

    ],
});

var app = new Vue({
    el: '#app',
    router: router,
    components: {
        'audio-player': AudioPlayerComponent,
        'albums': AlbumsComponent,
        'album': AlbumComponent,
        'current-queue': CurrentQueueComponent,
    }
});

