<template>
    <div class="albumlist pure-g">
        <div v-for="(album, $index) in albums" class="pure-u-1-4">
            <div class="album-container">
                <router-link :to="{ path: album.link }">
                    <div class="album">
                        <div class="album-cover" :style="{ 'background-image': 'url(' + album.getSongs()[0].cover + ')' }">
                            
                        </div>
                        <div class="album-meta">
                            <div class="album-meta-info pure-u-20-24">
                                <h1>{{album.name}}</h1>
                                <h2>{{album.getSongs()[0].artist}} <span>{{album.year}}</span></h2>
                            </div>
                        </div>
                    </div>
                </router-link>
            </div>
        </div>
        
        <!--<table v-if="show" class="albumlist">
            <thead>
                <tr>
                    <th>#</th>
                    <th>Album</th>
                    <th>Year</th>
                </tr>
            </thead>
            <tbody>
                <tr v-for="(album, $index) in albums">
                    
                    <td>
                        <router-link :to="{ path: album.link }">{{$index+1}}</router-link>
                    </td>
                    <td>
                        <router-link :to="{ path: album.link }">{{album.name}}</router-link>
                    </td>
                    <td>
                        <router-link :to="{ path: album.link }">{{album.year}}</router-link>
                    </td>
                </tr>
                
            </tbody>
        </table>-->
        
    </div>
   
</template>

<script>
    var $ = require('jquery');
    let _ = require('lodash');
    let Album = require('./../Album').default;

    module.exports = {
            data: function () {
                return {
                    albums: [],
                    show: false,
                }
            },
            mounted () {
                this.loadAlbums();
            },
            methods: {
              loadAlbums: function(){
                  let self = this
                  $.getJSON( "/albums", function(data) {
                        self.albums = _.map(data, (album) => {
                            album.link = 'albums/' + album.id;
                            return new Album(album);
                        });

                        self.show = true;
                  });
              }
            }
    };
</script>