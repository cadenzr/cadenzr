<template>
    <div v-if="show"
         class="single-album">
        <div class="album-meta pure-g">
            <div class="album-meta-cover pure-u-4-24">
                <img :src="album.getSongs()[0].getCoverUrl()">
            </div>
    
            <div class="album-meta-info pure-u-20-24">
                <h1>{{album.name}}</h1>
                <h2>{{album.getSongs()[0].artist}} <span>{{album.year}}</span></h2>
            </div>
        </div>
    
        <div class="album-songs">
            <table v-if="show"
                   class="playlist">
                <thead>
                    <tr>
                        <th>#</th>
                        <th><a v-on:click="toggleSort('name')">Title</a></th>
                        <th class="sm-hide"><a v-on:click="toggleSort('artist')">Artist</a></th>
                        <th class="md-hide"><a v-on:click="toggleSort('album')">Album</a></th>
                        <th class="md-hide"><a v-on:click="toggleSort('year')">Year</a></th>
                        <th class="md-hide"><a v-on:click="toggleSort('played')">Plays</a></th>
                    </tr>
                </thead>
                <tbody>
                    <tr v-for="(song, $index) in sortedSongs"
                        v-on:click="play(song)"
                        v-bind:class="{ playing: player.isCurrentSong(song) }">
                        <td>{{$index+1}}</td>
                        <td>{{song.name}}</td>
                        <td class="sm-hide">{{song.artist}}</td>
                        <td class="md-hide">{{song.album}}</td>
                        <td class="md-hide">{{song.year}}</td>
                        <td class="md-hide">{{song.played}}</td>
                    </tr>
                </tbody>
            </table>
    
        </div>
    
    </div>
</template>

<script lang="ts">
    import * as $ from 'jquery';
    import * as _ from 'lodash';
    import Api from './../Api';
    import Song from './../Song';
    import Album from './../Album';
    import Playlist from './../Playlist';
    import PubSub from './../PubSub';
    import {events as AudioPlayerEvents} from './../AudioPlayer';
    import player from '@/AudioPlayer';
    import Vue from 'vue';

    interface AlbumComponent extends Vue {
                    album: Album;
                    show: boolean;
                    sortOrder: string;
                    sortKey: string;
    }


    export default {
        name: 'album',
            data: function () {
                return {
                    album: new Album(),
                    show: false,
                    sortOrder: 'asc',
                    sortKey: 'name',
                    player: player,
                }
            },
            computed: {
                sortedSongs: function() {
                    return _.orderBy(this.album.getSongs(), [this.sortKey], [this.sortOrder]);
                }
            },
            mounted () {
                (<any>this).loadSongs();
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
              loadSongs: function(){
                  let self = (<any>this);
                  Api.getAlbum(self.$route.params.id).then((album:Album) => {
                      album.songs = _.map(album.songs, (song:Song) => {
                          return new Song(song);
                      });

                      self.album = new Album(album);
                      self.show = true;
                  });
              },
              play: function(song:Song){
                  // Assume that if user plays a song in this album he wants to play the whole album.
                  player.setQueue(this.album.getSongs());
                  player.setCurrentSong(song);
                  player.reload()
                  .then(() => {
                      player.play();
                  });
              }
            }
    } as Vue.ComponentOptions<AlbumComponent>;
</script>