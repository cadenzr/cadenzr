<template>
<div class="audio-player">
    <div class="pure-g">
        <div class="pure-u-4-24 playback-controls">
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
        </div>
        <div class="pure-u-16-24 current-song">
            <input ref="timeSlider" class="time-slider" type="range" style="width: 100%; display: block;" min="0" v-bind:max="duration">
            <p v-if="currentSong" class="song">{{currentSong.name}}</p>
            <p v-if="currentSong" class="artist">{{currentSong.artist}}</p>
        </div>
        <div class="pure-u-4-24 volume-controls">
            <input type="range"  class="volume-slider"  v-model="volume" min="0" max="100">
        </div>
    </div>                  
</div>
</template>

<script>
    let PubSub = require('./../PubSub').default;
    let AudioPlayer = require('./../AudioPlayer').default;
    let AudioPlayerEvents = require('./../AudioPlayer').events;

    module.exports = {
            data: function () {
                return {
                    playing: false,
                    volume: AudioPlayer.getVolume(),
                    currentTime: 0,
                    duration: 0,
                    currentSong: null,
                }
            },
            mounted: function () {
                let self = this;
                let isSeeking = false;

                PubSub.subscribe(AudioPlayerEvents.SongChanged, (song) => {
                    self.currentSong = song;
                    self.duration = song.duration;
                    self.$forceUpdate();
                });

                PubSub.subscribe(AudioPlayerEvents.TimeChanged, (time) => {
                    if(isSeeking) {
                        // Just return so the slider does not jump back.
                        return;
                    }

                    self.currentTime = time;
                    self.$refs.timeSlider.value = self.currentTime;          
                });

                PubSub.subscribe(AudioPlayerEvents.Pause, () => {
                    self.playing = false;
                    self.$forceUpdate(); 
                });

                PubSub.subscribe(AudioPlayerEvents.Play, () => {
                    self.playing = true;
                    self.$forceUpdate(); 
                });

                PubSub.subscribe(AudioPlayerEvents.VolumeChanged, (volume) => {
                    self.volume = volume;
                    self.$forceUpdate(); 
                });

                self.$watch('volume', () => {
                    AudioPlayer.setVolume(self.volume);
                });

                self.$refs.timeSlider.addEventListener('input', () => {
                    isSeeking = true;
                });

                self.$refs.timeSlider.addEventListener('change', () => {
                    isSeeking = false;
                    AudioPlayer.seek(self.$refs.timeSlider.value);
                });

            },
            methods: {
                next: function() {
                    AudioPlayer.next();
                },
                prev: function() {
                    AudioPlayer.prev();
                },
                play: function() {
                    AudioPlayer.play();
                },
                pause: function() {
                    AudioPlayer.pause();
                },
            }
    };
</script>