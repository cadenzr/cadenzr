<template>
    <div v-if="show"
         class="current-queue">
        <span v-on:click="playPlaylist()">Play this playlist</span>
    
        <table class="playlist">
            <thead>
                <tr>
                    <th>#</th>
                    <th><a>Title</a></th>
                    <th><a>Artist</a></th>
                    <th><a>Album</a></th>
                    <th><a>Year</a></th>
                    <th></th>
                </tr>
            </thead>
            <tbody>
                <tr v-for="(song, $index) in playlist.getSongs()"
                    v-bind:class="{ playing: player.isCurrentSong(song) }">
                    <td>{{$index+1}}</td>
                    <td>{{song.name}}</td>
                    <td>{{song.artist}}</td>
                    <td>{{song.album}}</td>
                    <td>{{song.year}}</td>
                    <td><a v-on:click="removeSong(song);"><span class="fa fa-fw fa-times"></span></a></td>
                </tr>
            </tbody>
        </table>
    
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
    import {events as AudioPlayerEvents} from '@/AudioPlayer';
    import player from '@/AudioPlayer';

    interface PlaylistComponent extends Vue {
    playlist: Playlist;
    show: boolean;
    sortOrder: string;
    sortKey: string;
    subscriptions: Array<any>;
    }

    export default {
        name: 'playlist',
            data: function () {
                return {
                    playlist: new Playlist(),
                    show: false,
                    sortOrder: 'asc',
                    sortKey: 'name',
                    subscriptions: [],
                    player: player,
                }
            },
            computed: {

            },
            mounted () {
                let self = <any>this;
                self.fetchData();
            },
            watch: {
                '$route': 'fetchData'
            },
            route: {
                data() {
                    console.log('lol');
                },
            },
            beforeDestroy () {
                _.forEach(this.subscriptions, (s:any) => {
                    PubSub.unsubscribe(s);
                });
            },
            methods: {
                fetchData: function() {
                    let self = <any>this;
                    self.show = false;
                    Api.getPlaylist(self.$route.params.id).then((playlist:any) => {
                        self.playlist = new Playlist(playlist);
                        self.show = true;
                    });
                },
                toggleSort: function(key: string) {
                    this.sortKey = key;
                    if(this.sortOrder === 'asc') {
                        this.sortOrder = 'desc';
                    } else {
                        this.sortOrder = 'asc';
                    }
                },
                removeSong: function(song:Song) {
                    let self = this;
                    Api.deleteSongFromPlaylist(song, self.playlist)
                    .then(() => {
                        self.playlist.removeSong(song);
                        self.$forceUpdate();
                    });
                },
              playPlaylist: function(){
                  player.setQueue(this.playlist.songs);
                  if(this.playlist.songs.length > 0) {
                    player.setCurrentSong(this.playlist.songs[0]);
                    player.reload()
                    .then(() => {
                        player.play();
                    });
                  }
              }
            }
    } as Vue.ComponentOptions<PlaylistComponent>;
</script>