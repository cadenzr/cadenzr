<template>
<table class="playlist pure-table pure-table-horizontal pure-table-striped">
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
        <tr v-for="song in songs" v-on:click="play(song)">
            <td>{{song.index + 1}}</td>
            <td>{{song.name}}</td>
            <td>{{song.artist}}</td>
            <td>{{song.album}}</td>
            <td>{{song.year}}</td>
        </tr>
    </tbody>
</table>
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