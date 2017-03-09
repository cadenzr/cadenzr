<template>
<ol>
    <li v-for="song in songs"><a v-on:click="play(song)">{{song.name}}</a></li>
</ol>
</template>

<script>
    var $ = require('jquery');

    module.exports = {
            data: function () {
                return {
                    songs: []
                }
            },
            mounted () {
                this.loadSongs();
            },
            methods: {
              loadSongs: function(){
                  self = this
                  $.getJSON( "./albums/" + self.$route.params.id + "/songs", function(data) {
                      self.songs = data;
                      var count = 0;
                      self.songs.forEach(function(song) {
                          song.index = count;
                          count++;
                      })
                  });
              },
              play: function(song){
                  self = this;
                  console.log(song.stream_location);
                  //app.$refs.player.song_stream = song.stream_location;
                  self.$root.$refs.player.playSong(self.songs, song.index);
              }
            }
    };
</script>