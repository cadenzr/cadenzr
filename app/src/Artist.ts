import * as _ from 'lodash';

import Song from './Song';

class Artist {
    id: number;
    songs: Array<Song>;

    constructor(data?: any) {
        this.songs = [];

        if (data) {
            _.assign(this, data);
        }
    }

    setSongs(songs: Array<Song>) {
        this.songs = songs;

        if (!this.hasSongs()) {
            return;
        }
    }

    getSongs(): Array<Song> {
        return this.songs;
    }


    private hasSongs(): boolean {
        return this.songs.length > 0;
    }
}

export default Artist;