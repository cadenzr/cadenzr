<template>
<div class="audio-player">
                    <input ref="timeSlider" type="range" style="width: 100%; display: block;" min="0" v-bind:max="duration">

                    <a class="prev" v-on:click="prev">
                        <span class="fa fa-step-backward"></span>
                    </a>
                    <a class="play"  v-on:click="play" v-if="!playing">
                        <span class="fa fa-play"></span>
                    </a>
                    <a class="pause"  v-on:click="pause" v-if="playing">
                        <span class="fa fa-pause"></span>
                    </a>
                    <a class="next"  v-on:click="next">
                        <span class="fa fa-step-forward"></span>
                    </a>
                    <audio ref="audioplayer" v-on:ended="next" controls>
                        <source :src="song_stream" type="audio/mpeg">
                        Your browser does not support the audio tag.
                    </audio>
                    <input type="range" v-model="volume" min="0" max="100">
                    <p>
                       <span v-if="index >= 0">{{songs[index].name}}</span>
                    </p>
</div>
</template>

<script>
    module.exports = {
            data: function () {
                return {
                    song_stream: "",
                    index: -1,
                    songs: [],
                    playing: false,
                    volume: 20,
                    currentTime: 0,
                    duration: 0,
                }
            },
            mounted: function () {
                var self = this;
                var timeupdate = true;
                this.$refs.audioplayer.addEventListener("timeupdate", function() {
                    if(!timeupdate) {
                        return;
                    }

                    self.currentTime = self.$refs.audioplayer.currentTime;
                    self.$refs.timeSlider.value = self.currentTime;
                });

                this.$refs.timeSlider.addEventListener("input", function() {
                    timeupdate = false;
                });

                this.$refs.timeSlider.addEventListener("change", function() {
                    self.currentTime = self.$refs.timeSlider.value;
                    self.$refs.audioplayer.currentTime = self.currentTime;
                    timeupdate = true;
                });

                this.$refs.audioplayer.addEventListener("loadeddata", function() {
                    self.duration = self.$refs.audioplayer.duration;
                });

                this.$watch('index', function () {
                    this.song_stream = this.songs[this.index].stream_location;
                    this.$refs.audioplayer.load()
                    this.play()
                });

              this.$watch('volume', function () {
                  this.$refs.audioplayer.volume = this.volume / 100;
              });


            },
            methods: {
                playSong: function(songs, index) {
                    this.songs = songs;
                    this.index = index;
                },
                next: function() {
                    this.index = (this.index + 1) % this.songs.length;
                },
                prev: function() {
                    this.index = (((this.index - 1) % this.songs.length) + this.songs.length) % this.songs.length; // shitty module since JS doesn't like mod of negative numbers
                },
                play: function() {
                    this.$refs.audioplayer.play()
                    this.playing = true;
                },
                pause: function() {
                    this.$refs.audioplayer.pause()
                    this.playing = false;
                },
            }
    };
</script>