<template>
    <div v-if="show"
         class="albumlist pure-g">
        <div v-for="(album, $index) in albums"
             class="pure-u-1-2 pure-u-md-1-3 pure-u-lg-1-4">
            <div v-on:dragstart="dragstart"
                 class="album-container"
                 draggable="true"
                 :data-album-index="$index">
                <router-link :to="{ path: album.link }">
                    <div class="album">
                        <div class="album-cover"
                             :style="{ 'background-image': 'url(' + album.getCoverUrl() + ')' }">
                            <div class="album-play">
                                <div class="album-play-button"
                                     @click.prevent="playAlbum(album)">
                                    <span class="fa fa-fw fa-play"></span>
                                </div>
                            </div>
                        </div>
                        <div class="album-meta">
                            <div class="album-meta-info pure-u-20-24">
                                <h1>{{album.name}}</h1>
                                <h2>{{album.getArtist()}} <span>{{album.year}}</span></h2>
                            </div>
                        </div>
                    </div>
                </router-link>
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
    import AudioPlayerEvents from './../AudioPlayer';
    import AudioPlayer from './../AudioPlayer';
    import env from '@/env';
    import Vue from 'vue';

    interface Albums extends Vue {
                    albums: Array<Album>;
                    show: boolean;
                    sortOrder: string;
                    sortKey: string;
    }

    export default {
        name: 'albums',
            data: function () {
                return {
                    albums: [],
                    show: false,
                    sortKey: 'name',
                    sortOrder: 'asc',
                    AudioPlayer: AudioPlayer,
                }
            },
            mounted () {
                (<any>this).loadAlbums();
            },
            computed: {
                sortedAlbums: function() {
                    return _.orderBy(this.albums, [this.sortKey], [this.sortOrder]);
                }
            },
            methods: {
                dragstart: function(e : any) {
                    let index = e.srcElement.getAttribute('data-album-index');
                    let album = this.albums[index];
                    let img = new Image();
                    img.src = 'data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAACAAAAAgCAMAAABEpIrGAAAAw1BMVEUAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAABAcKsgAAAAQHRSTlMAAQIDBAUGCAsMDxARFBcbHCkqNTw9RktQVF9naG11gIKDiIyOkpedoKKlpqqttbq8wcXR19nc3uLo6/Hz9/v9cP1/IAAAANFJREFUOMvFktcSgjAQRUMwIlbA3nsXEXtn//+rnFCUkuCb3qfsnjPJThKEuBG1SpJPpdLcAthxaLZmghMShYLSOsE7YUHU+nfwJygo9FjgC+QJECvk4NeCNW3HCLeuimmDLRybBcFtMASzlvE1wsKyLAUbYSF4s/8WBBwnqL0rwHPW5ghkw38Lux7BF+HqExIuw3KasHbYOrRq0OIw9mYwPL6XaZ2K/nJtcAN4TIr2m6IFcyaM38uLTxARI50PX7E4EoYe1wliJ1PXz5d1I8+EL7ggW9U/YokyAAAAAElFTkSuQmCC';
                    e.dataTransfer.setData('songs', JSON.stringify(album.getSongs()));
                    e.dataTransfer.setDragImage(img, 0, 0);
                },
                toggleSort: function(key: string) {
                    this.sortKey = key;
                    if(this.sortOrder === 'asc') {
                        this.sortOrder = 'desc';
                    } else {
                        this.sortOrder = 'asc';
                    }
                },
                loadAlbums: function(){
                    let self = this
                    Api.getAlbums()
                    .then(albums => {
                        self.albums = _.map(albums, (album : any) => {
                            album.songs = _.map(album.songs, (song) => {
                                return new Song(song);
                            });

                            album.link = 'albums/' + album.id;

                            return new Album(album);
                        });

                        self.show = true;
                    });
                },
                playAlbum: function(album : Album) {
                    AudioPlayer.setQueue(album.getSongs());
                    AudioPlayer.reload()
                    .then(() => {
                        AudioPlayer.play();
                    });
                },
            }
    } as Vue.ComponentOptions<Albums>;
</script>