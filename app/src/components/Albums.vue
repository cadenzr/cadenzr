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
                        self.albums = data;
                        self.albums.forEach(function(album) {
                            console.log(album.id);
                            album.link = "albums/" + album.id;
                        });
                  });
              }
            }
    };
</script>