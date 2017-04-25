import Vue from 'vue';
import VueRouter from 'vue-router';
import * as VueResource from 'vue-resource';
import * as Dropzone from 'vue2-dropzone';
import * as VueLazyload from 'vue-lazyload'
import * as _ from 'lodash';

import components from './components';
import * as Api from './Api';
import * as AudioPlayer from './AudioPlayer';
import Song from './Song';
import router from './router';

import './assets/sass/style.scss';

Vue.use(VueResource)
Vue.component('dropzone', Dropzone);
Vue.use(VueLazyload)


import Notifier from './Notifier';
import PubSub from './PubSub';

PubSub.subscribe(AudioPlayer.events.SongStarted, (song: Song) => {
    Notifier.notify('Playing song: ' + song.name);
});

PubSub.subscribe(AudioPlayer.events.SongEnded, (song: Song) => {
    Api.api.incrementPlayed(song);
});

var app = new Vue({
    el: '#app',
    router: router,
    components: {
        'album': components.albumComponent,
        'albums': components.albumsComponent,
        'audio-player': components.audioPlayerComponent,
        'current-queue': components.currentQueueComponent,
        'login': components.loginComponent,
        'playlist': components.playlistComponent,
        'sidebar': components.sidebarComponent,
        'upload': components.uploadComponent,
    },
    mounted: function () {
        let self = this;
        (<any>this).subscriptions.push(PubSub.subscribe(Api.events.LoggedOut, () => {
            (<any>self).$router.push('/login');
        }));
        if (!Api.api.isAuthenticated()) {
            (<any>self).$router.push('/login');
        }
    },
    beforeDestroy: () => {
        _.forEach(this.subscriptions, (s) => {
            PubSub.unsubscribe(s);
        });
    },
    data: function () {
        return {
            user: {},
            subscriptions: [],
        };
    },
});
