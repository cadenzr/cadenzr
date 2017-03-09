//import * as _ from 'lodash';
//import * as $ from 'jquery';
//import {test} from './test';
import * as Vue from 'vue';
import * as Router from 'vue-router';
Vue.use(Router);

import * as AudioPlayer from './components/AudioPlayer.vue';
import * as Albums from './components/Albums.vue';
import * as Album from './components/Album.vue';


var router = new Router({
    routes: [
          { path: '/', component: Albums },
          { path: '/albums', component: Albums },
          { path: '/albums/:id', component: Album }
    ],
});

var app = new Vue({
    el: '#app',
    router: router,
    components: {
        'audio-player': AudioPlayer,
        'albums': Albums,
        'album': Album,
    }
});

