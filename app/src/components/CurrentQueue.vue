<template>
<div>    
    <table v-if="show" class="playlist">
        <thead>
            <tr>
                <th>#</th>
                <th><a>Title</a></th>
                <th><a>Artist</a></th>
                <th><a>Album</a></th>
                <th><a>Year</a></th>
            </tr>
        </thead>
        <tbody>
            <tr v-for="(song, $index) in queue" v-on:click="play(song)" v-bind:class="{ playing: AudioPlayer.isCurrentSong(song) }">
                <td>{{$index+1}}</td>
                <td>{{song.name}}</td>
                <td>{{song.artist}}</td>
                <td>{{song.album}}</td>
                <td>{{song.year}}</td>
            </tr>
        </tbody>
    </table>
    
</div>
</template>

<script>
    let $ = require('jquery');
    let _ = require('lodash');
    let Song = require('./../Song').default;
    let PubSub = require('./../PubSub').default;
    let AudioPlayer = require('./../AudioPlayer').default;
    let AudioPlayerEvents = require('./../AudioPlayer').events;

    module.exports = {
            data: function () {
                return {
                    queue: AudioPlayer.getQueue(),
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

                self.subscriptions.push(PubSub.subscribe(AudioPlayerEvents.QueueChanged, (queue) => {
                    self.queue = queue;
                    self.$forceUpdate();
                }));

                self.subscriptions.push(PubSub.subscribe(AudioPlayerEvents.SongChanged, (song) => {
                    self.$forceUpdate();
                }));

                self.show = true;
            },
            beforeDestroy () {
                _.forEach(this.subscriptions, (s) => {
                    PubSub.unsubscribe(s);
                });
            },
            methods: {
                toggleSort: function(key) {
                    this.sortKey = key;
                    if(this.sortOrder === 'asc') {
                        this.sortOrder = 'desc';
                    } else {
                        this.sortOrder = 'asc';
                    }
                },
              play: function(song){
                  this.AudioPlayer.setCurrentSong(song);
                  this.AudioPlayer.reload()
                  .then(() => {
                      this.AudioPlayer.play();
                  });
              }
            }
    };
</script>