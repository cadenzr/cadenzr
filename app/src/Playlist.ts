import * as _ from 'lodash';

import Song from './Song';

class Playlist {
    id: number;
    name: string;
    songs: Array<Song>;

    constructor(data: any) {
        this.songs = [];

        if(data) {
           _.assign(this, data);
        }
    }

    getSongs() : Array<Song> {
        return this.songs;
    }

    removeSong(song : Song) {
        _.remove(this.songs, {id: song.id});
    }

    private hasSongs() : boolean {
        return this.songs.length > 0;
    }
}

export default Playlist;