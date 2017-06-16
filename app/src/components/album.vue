<template>
    <div v-if="show"
         class="single-album">
        <div class="album-meta pure-g">
            <div class="album-meta-cover pure-u-4-24">
                <img :src="album.getCoverUrl()">
            </div>
    
            <div class="album-meta-info pure-u-20-24">
                <h1>{{album.name}}</h1>
                <h2>{{album.getSongs()[0].artist}} <span>{{album.year}}</span></h2>
                <h2>
                    <a :href="downloadUrl" data-balloon="Download album" data-balloon-pos="up"><span class="fa fa-download"></span></a> 
                    <a :href="downloadPlaylistUrl" data-balloon="Download album stream file (.m3u8)" data-balloon-pos="up"><span class="fa fa-list"></span></a>
                </h2>
            </div>
        </div>
    
        <div class="album-songs">
            <table v-if="show"
                   class="playlist">
                <thead>
                    <tr>
                        <th><a v-on:click="toggleSort('track')">#</a></th>
                        <th><a v-on:click="toggleSort('name')">Title</a></th>
                        <th class="sm-hide"><a v-on:click="toggleSort('artist')">Artist</a></th>
                        <th class="md-hide"><a v-on:click="toggleSort('album')">Album</a></th>
                        <th class="md-hide"><a v-on:click="toggleSort('year')">Year</a></th>
                    </tr>
                </thead>
                <tbody>
                    <tr v-for="(song, $index) in sortedSongs"
                        v-on:click="play(song)"
                        v-bind:class="{ playing: player.isCurrentSong(song) }">
                        <td>{{song.track}}</td>
                        <td>{{song.name}}</td>
                        <td class="sm-hide">{{song.artist}}</td>
                        <td class="md-hide">{{song.album}}</td>
                        <td class="md-hide">{{song.year}}</td>
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
                    sortKey: 'track', // sort on track number by default
                    player: player,
                    downloadUrl: '',
                }
            },
            computed: {
                sortedSongs: function() {
                    console.log(this.album);
                    return _.orderBy(this.album.getSongs(), [this.sortKey, 'track'], [this.sortOrder]);
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
                      self.downloadUrl = Api.apiEndpoint + 'albums/' + album.id.toString() + '/download?token=' + Api.retrieveToken();
                      self.downloadPlaylistUrl = Api.apiEndpoint + 'albums/' + album.id.toString() + '/playlist.m3u8?token=' + Api.retrieveToken();
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