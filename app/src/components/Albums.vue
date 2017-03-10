<template>
    <table class="albumlist pure-table pure-table-horizontal pure-table-striped">
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
    </table>
</template>

<script>
    var $ = require('jquery');
    let _ = require('lodash');
    let Album = require('./../Album').default;

    module.exports = {
            data: function () {
                return {
                    albums: []
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
                  });
              }
            }
    };
</script>