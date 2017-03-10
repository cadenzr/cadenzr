<template>
<div>
    <div class="album-meta pure-g">
        <div class="album-meta-cover pure-u-4-24">
            <img src="http://www.interactivepixel.net/env/jhap2wp/data/default_artwork/music_ph.png">
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
                <th>Title</th>
                <th>Artist</th>
                <th>Album</th>
                <th>Year</th>
            </tr>
        </thead>
        <tbody>
            <tr v-for="(song, $index) in album.getSongs()" v-on:click="play($index)">
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
                }
            },
            mounted () {
                this.loadSongs();
            },
            methods: {
              loadSongs: function(){
                  let self = this
                  $.getJSON( "/albums/" + self.$route.params.id, function(data) {
                      data.songs = _.map(data.songs, (song) => {
                          return new Song(song);
                      });

                      self.album = new Album(data);
                      self.show = true;
                  });
              },
              play: function(index){
                  this.album.setIndex(index);
                  // Assume that if user plays a song in this album he wants to play the whole album.
                  AudioPlayer.setProvider(this.album);
                  AudioPlayer.restartCurrent()
                  .then(() => {
                      AudioPlayer.play();
                  });
              }
            }
    };
</script>