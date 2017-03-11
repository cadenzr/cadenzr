<template>
    <div v-if="show" class="albumlist pure-g">
        <div v-for="(album, $index) in albums" class="pure-u-1-4">
            <div class="album-container">
                <router-link :to="{ path: album.link }">
                    <div class="album">
                        <div class="album-cover" :style="{ 'background-image': 'url(' + album.getCover() + ')' }">
                            
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

<script>
    var $ = require('jquery');
    let _ = require('lodash');
    let Album = require('./../Album').default;

    module.exports = {
            data: function () {
                return {
                    albums: [],
                    show: false,
                    sortKey: 'name',
                    sortOrder: 'asc',
                }
            },
            mounted () {
                this.loadAlbums();
            },
            computed: {
                sortedAlbums: function() {
                    return _.orderBy(this.albums, [this.sortKey], [this.sortOrder]);
                }
            },
            methods: {
                toggleSort: function(key) {
                    this.sortKey = key;
                    if(this.sortOrder === 'asc') {
                        this.sortOrder = 'desc';
                    } else {
                        this.sortOrder = 'asc';
                    }
                },
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