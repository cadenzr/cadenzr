<template>
    <table v-if="show" class="albumlist">
        <thead>
            <tr>
                <th>#</th>
                <th><a href="#" v-on:click="toggleSort('name')">Album</a></th>
                <th><a href="#" v-on:click="toggleSort('year')">Year</a></th>
            </tr>
        </thead>
        <tbody>
            <tr v-for="(album, $index) in sortedAlbums">
                
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