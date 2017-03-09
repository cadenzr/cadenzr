<template>
<ul>
    <li v-for="album in albums">
        <router-link :to="{ path: album.link }">
            {{album.name}}
        </router-link>
    </li>
</ul>
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
                  self = this
                  $.getJSON( "./albums", function(data) {
                        self.albums = _.map(data, (album) => {
                            album.link = 'albums/' + album.id;
                            return new Album(album);
                        });
                  });
              }
            }
    };
</script>