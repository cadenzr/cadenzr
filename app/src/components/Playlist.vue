<template>
<div v-if="show" class="current-queue">
    <span v-on:click="playPlaylist()">Play this playlist</span>

    <table  class="playlist">
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
            <tr v-for="(song, $index) in playlist.getSongs()" v-bind:class="{ playing: AudioPlayer.isCurrentSong(song) }">
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

<script>
    let $ = require('jquery');
    let _ = require('lodash');
    let Api = require('./../Api').default;
    let Song = require('./../Song').default;
    let Playlist = require('./../Playlist').default;
    let PubSub = require('./../PubSub').default;
    let AudioPlayer = require('./../AudioPlayer').default;
    let AudioPlayerEvents = require('./../AudioPlayer').events;

    module.exports = {
            data: function () {
                return {
                    playlist: new Playlist(),
                    show: false,
                    AudioPlayer: AudioPlayer,
                    sortOrder: 'asc',
                    sortKey: 'name',
                    subscriptions: [],
                }
            },
            computed: {

            },
            mounted () {
                let self = this;
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
                _.forEach(this.subscriptions, (s) => {
                    PubSub.unsubscribe(s);
                });
            },
            methods: {
                fetchData: function() {
                    let self = this;
                    self.show = false;
                    Api.getPlaylist(self.$route.params.id).then(playlist => {
                        self.playlist = new Playlist(playlist);
                        self.show = true;
                    });
                },
                toggleSort: function(key) {
                    this.sortKey = key;
                    if(this.sortOrder === 'asc') {
                        this.sortOrder = 'desc';
                    } else {
                        this.sortOrder = 'asc';
                    }
                },
                removeSong: function(song) {
                    let self = this;
                    Api.deleteSongFromPlaylist(song, self.playlist)
                    .then(() => {
                        self.playlist.removeSong(song);
                        self.$forceUpdate();
                    });
                },
              playPlaylist: function(){
                  this.AudioPlayer.setQueue(this.playlist.songs);
                  if(this.playlist.songs.length > 0) {
                    this.AudioPlayer.setCurrentSong(this.playlist.songs[0]);
                    this.AudioPlayer.reload()
                    .then(() => {
                        this.AudioPlayer.play();
                    });
                  }
              }
            }
    };
</script>