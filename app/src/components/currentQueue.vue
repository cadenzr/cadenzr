<template>
    <div class="current-queue">
        <table v-if="show"
               class="playlist">
            <thead>
                <tr>
                    <th>#</th>
                    <th><a>Title</a></th>
                    <th class="sm-hide"><a>Artist</a></th>
                    <th class="md-hide"><a>Album</a></th>
                    <th class="md-hide"><a>Year</a></th>
                </tr>
            </thead>
            <tbody>
                <tr v-for="(song, $index) in queue"
                    v-on:click="play(song)"
                    v-bind:class="{ playing: player.isCurrentSong(song) }">
                    <td>{{$index+1}}</td>
                    <td>{{song.name}}</td>
                    <td class="sm-hide">{{song.artist}}</td>
                    <td class="md-hide">{{song.album}}</td>
                    <td class="md-hide">{{song.year}}</td>
                </tr>
            </tbody>
        </table>
    
    </div>
</template>

<script lang="ts">
    import * as $ from 'jquery';
    import * as _ from 'lodash';
    import Api from './../Api';
    import Song from './../Song';
    import Playlist from './../Playlist';
    import PubSub from './../PubSub';
    import {events as AudioPlayerEvents} from '@/AudioPlayer';
    import player from '@/AudioPlayer';
    import Vue from 'vue';

    interface CurrentQueue extends Vue {
                    queue: Array<any>;
                    show: boolean;
                    sortOrder: string;
                    sortKey: string;
                    subscriptions: Array<any>;
    }


    export default {
        name: 'current-queue',
            data: function () {
                return {
                    queue: player.getQueue(),
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
                let self = this;

                self.subscriptions.push(PubSub.subscribe(AudioPlayerEvents.QueueChanged, (queue:Array<any>) => {
                    self.queue = queue;
                    self.$forceUpdate();
                }));

                self.subscriptions.push(PubSub.subscribe(AudioPlayerEvents.SongChanged, (song:Song) => {
                    self.$forceUpdate();
                }));

                self.show = true;
            },
            beforeDestroy () {
                _.forEach(this.subscriptions, (s:any) => {
                    PubSub.unsubscribe(s);
                });
            },
            methods: {
                toggleSort: function(key:string) {
                    this.sortKey = key;
                    if(this.sortOrder === 'asc') {
                        this.sortOrder = 'desc';
                    } else {
                        this.sortOrder = 'asc';
                    }
                },
              play: function(song:Song){
                  player.setCurrentSong(song);
                  player.reload()
                  .then(() => {
                      player.play();
                  });
              }
            }
    } as Vue.ComponentOptions<CurrentQueue>;
</script>