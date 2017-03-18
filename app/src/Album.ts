import * as _ from 'lodash';

import Song from './Song';
import env from './env';

class Album {
    id: number;
    songs: Array<Song>;
    cover: string;

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

    getArtist(): string {
        if (this.songs.length === 0) {
            return '';
        }

        if (this.songs[0].artist) {
            return this.songs[0].artist;
        }

        return '';
    }

    getCoverUrl(): string {
        if (!this.cover) {
            return '';
        }

        return env.backend + this.cover;
    }

    private hasSongs(): boolean {
        return this.songs.length > 0;
    }
}

export default Album;