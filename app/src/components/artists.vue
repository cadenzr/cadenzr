<template>
    <div v-if="show"
         class="artistlist pure-g">
        
        <div class="artists-index pure-u-1-4">
            <table v-if="show">
                <thead>
                </thead>
                <tbody>
                    <tr v-for="(artist, $index) in artists">
                        <td></td>
                        <td>{{artist.name}}</td>
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
    import Album from './../Album';
    import Artist from './../Artist';
    import Song from './../Song';
    import Playlist from './../Playlist';
    import PubSub from './../PubSub';
    import AudioPlayerEvents from './../AudioPlayer';
    import AudioPlayer from './../AudioPlayer';
    import Vue from 'vue';

    interface Artists extends Vue {
                    artists: Array<Artist>;
                    show: boolean;
                    sortOrder: string;
                    sortKey: string;
    }

    export default {
        name: 'artists',
            data: function () {
                return {
                    artists: [],
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
                    return _.orderBy(this.artists, [this.sortKey], [this.sortOrder]);
                }
            },
            methods: {
                /*
                dragstart: function(e : any) {
                    let index = e.srcElement.getAttribute('data-album-index');
                    let album = this.albums[index];
                    let img = new Image();
                    img.src = 'data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAACAAAAAgCAMAAABEpIrGAAAAw1BMVEUAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAABAcKsgAAAAQHRSTlMAAQIDBAUGCAsMDxARFBcbHCkqNTw9RktQVF9naG11gIKDiIyOkpedoKKlpqqttbq8wcXR19nc3uLo6/Hz9/v9cP1/IAAAANFJREFUOMvFktcSgjAQRUMwIlbA3nsXEXtn//+rnFCUkuCb3qfsnjPJThKEuBG1SpJPpdLcAthxaLZmghMShYLSOsE7YUHU+nfwJygo9FjgC+QJECvk4NeCNW3HCLeuimmDLRybBcFtMASzlvE1wsKyLAUbYSF4s/8WBBwnqL0rwHPW5ghkw38Lux7BF+HqExIuw3KasHbYOrRq0OIw9mYwPL6XaZ2K/nJtcAN4TIr2m6IFcyaM38uLTxARI50PX7E4EoYe1wliJ1PXz5d1I8+EL7ggW9U/YokyAAAAAElFTkSuQmCC';
                    e.dataTransfer.setData('songs', JSON.stringify(album.getSongs()));
                    e.dataTransfer.setDragImage(img, 0, 0);
                },
                */
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
                    Api.getArtists()
                    .then(response => {
                        self.artists = _.map(response.data, (artist : any) => {
                            artist.albums = _.map(artist.albums, (album : any) => {
                                album.songs = _.map(album.songs, (song : any) => {
                                    return new Song(song);
                                });
                                return new Album(album);
                            });
                            return new Artist(artist);
                        });

                        self.show = true;
                        console.log(self.artists)
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
    } as Vue.ComponentOptions<Artists>;
</script>