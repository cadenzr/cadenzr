<template>
<div v-if="show" class="single-album">
    <div class="album-meta pure-g">
        <div class="album-meta-cover pure-u-4-24">
            <img :src="album.getSongs()[0].cover">
        </div>
        
        <div class="album-meta-info pure-u-20-24">
            <h1>{{album.name}}</h1>
            <h2>{{album.getSongs()[0].artist}} <span>{{album.year}}</span></h2>
        </div>
    </div>
    
    <table v-if="show" class="playlist">
        <thead>
            <tr>
                <th>#</th>
                <th><a v-on:click="toggleSort('name')">Title</a></th>
                <th><a v-on:click="toggleSort('artist')">Artist</a></th>
                <th><a v-on:click="toggleSort('album')">Album</a></th>
                <th><a v-on:click="toggleSort('year')">Year</a></th>
            </tr>
        </thead>
        <tbody>
            <tr v-for="(song, $index) in sortedSongs" v-on:click="play(song)" v-bind:class="{ playing: AudioPlayer.isCurrentSong(song) }">
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
    let Album = require('./../Album').default;
    let AudioPlayer = require('./../AudioPlayer').default;

    module.exports = {
            data: function () {
                return {
                    album: new Album(),
                    show: false,
                    AudioPlayer: AudioPlayer,
                    sortOrder: 'asc',
                    sortKey: 'name',
                }
            },
            computed: {
                sortedSongs: function() {
                    return _.orderBy(this.album.getSongs(), [this.sortKey], [this.sortOrder]);
                }
            },
            mounted () {
                this.loadSongs();
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
              loadSongs: function(){
                  let self = this
                  this.$http.get( "/albums/" + self.$route.params.id).then(response => {
                      let data = response.body
                      data.songs = _.map(data.songs, (song) => {
                          return new Song(song);
                      });

                      self.album = new Album(data);
                      self.show = true;
                  });
              },
              play: function(song){
                  // Assume that if user plays a song in this album he wants to play the whole album.
                  AudioPlayer.setQueue(this.album.getSongs());
                  AudioPlayer.setCurrentSong(song);
                  AudioPlayer.reload()
                  .then(() => {
                      AudioPlayer.play();
                  });
              }
            }
    };
</script>