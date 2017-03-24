<template>
    <div class="audio-player">
        <div class="pure-g">
            <div class="pure-u-1 pure-u-md-4-24 playback-controls">
                <a class="prev"
                   v-on:click="prev">
                    <span class="fa fa-step-backward"></span>
                </a>
                <a class="play"
                   v-on:click="play"
                   v-if="!playing">
                    <span class="fa fa-play"></span>
                </a>
                <a class="pause"
                   v-on:click="pause"
                   v-if="playing">
                    <span class="fa fa-pause"></span>
                </a>
                <a class="next"
                   v-on:click="next">
                    <span class="fa fa-step-forward"></span>
                </a>
            </div>
            <div class="pure-u-1 pure-u-md-16-24 current-song">
                <progress-bar v-bind:played="played" ref="timeSlider"></progress-bar>
                <img v-if="currentSong"
                     class="cover"
                     :src="currentSong.getCoverUrl()"></img>
                <p v-if="currentSong"
                   class="song">{{currentSong.name}}</p>
                <p v-if="currentSong"
                   class="artist">{{currentSong.artist}}</p>
            </div>
            <div class="pure-u-1 pure-u-md-4-24 volume-controls">
                <input type="range"
                       class="volume-slider"
                       v-model="volume"
                       min="0"
                       max="100">
            </div>
        </div>
    </div>
</template>

<script lang="ts">
    import * as $ from 'jquery';
    import * as _ from 'lodash';
    import Api from './../Api';
    import Album from './../Album';
    import Song from './../Song';
    import Playlist from './../Playlist';
    import PubSub from './../PubSub';
    import {events as AudioPlayerEvents} from './../AudioPlayer';
    import AudioPlayer from './../AudioPlayer';
    import Vue from 'vue';
    import * as progressBar from './media-controls/progressBar.vue';


    interface AudioPlayerComponent extends Vue {
                    playing: boolean;
                    volume: number;
                    currentTime: number;
                    duration: number;
                    currentSong: Song|null;
                    subscriptions: Array<any>;
                    played: number;
    }

    export default {
        name: 'audio-player',
        components: {
            progressBar,
        },
            data: function () {
                return {
                    playing: false,
                    volume: AudioPlayer.getVolume(),
                    currentTime: 0,
                    duration: 0,
                    currentSong: null,
                    subscriptions: [],
                    played: 0,
                }
            },
            mounted: function () {
                let self = <any>this;
                let isSeeking = false;

                (<any>self).subscriptions.push(PubSub.subscribe(AudioPlayerEvents.SongChanged, (song:Song) => {
                    self.currentSong = song;
                    self.duration = song.duration;
                    self.$forceUpdate();
                }));

                (<any>self).subscriptions.push(PubSub.subscribe(AudioPlayerEvents.TimeChanged, (time:number) => {
                    if(isSeeking) {
                        // Just return so the slider does not jump back.
                        return;
                    }

                    //self.played = time / (<any>AudioPlayer).currentSong().duration;
                    self.currentTime = time;
                    self.$refs.timeSlider.value = self.currentTime;          
                }));

                (<any>self).subscriptions.push(PubSub.subscribe(AudioPlayerEvents.Pause, () => {
                    self.playing = false;
                    self.$forceUpdate(); 
                }));

                (<any>self).subscriptions.push(PubSub.subscribe(AudioPlayerEvents.Play, () => {
                    self.$refs.timeSlider.$emit('progress-bar-start', AudioPlayer.currentSong());
                    self.playing = true;
                    self.$forceUpdate(); 
                }));

                (<any>self).subscriptions.push(PubSub.subscribe(AudioPlayerEvents.VolumeChanged, (volume:number) => {
                    self.volume = volume;
                    self.$forceUpdate(); 
                }));

                (<any>self).$watch('volume', () => {
                    AudioPlayer.setVolume(self.volume);
                });

                /*(<any>self).$refs.timeSlider.addEventListener('input', () => {
                    isSeeking = true;
                });*/

                /*(<any>self).$refs.timeSlider.addEventListener('change', () => {
                    isSeeking = false;
                    AudioPlayer.seek((<any>self).$refs.timeSlider.value);
                });*/

            },
            beforeDestroy () {
                _.forEach((<any>this).subscriptions, (s:any) => {
                    PubSub.unsubscribe(s);
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
    } as Vue.ComponentOptions<AudioPlayerComponent>;
</script>