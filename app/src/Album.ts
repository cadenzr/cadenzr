import * as _ from 'lodash';

import {Provider} from './AudioPlayer';
import Song from './Song';

class Album implements Provider {
    songs: Array<Song>;
    cover: string;

    constructor(data: any) {
        this.songs = [];

        if(data) {
           _.assign(this, data);
        }

        this.currentIndex = 0;
    }

    setSongs(songs: Array<Song>) {
        this.songs = songs;

        if(!this.hasSongs()) {
            return;
        }

        this.setIndex(0);
    }

    getSongs() : Array<Song> {
        return this.songs;
    }

    getArtist() : string {
        if(this.songs.length === 0) {
            return '';
        }

        if(this.songs[0].artist) {
            return this.songs[0].artist;
        }

        return '';
    }

    getCover() : string {
        if(!this.cover) {
            return '';
        }

        return this.cover;
    }

    nextSong() : Song {
        if(!this.hasSongs()) {
            return null;
        }

        this.setIndex(this.getIndex()+1);
        return this.songs[this.getIndex()];
    }

    prevSong() : Song {
        if(!this.hasSongs()) {
            return null;
        }

        this.setIndex(this.getIndex()-1);
        return this.songs[this.getIndex()];
    }

    currentSong() : Song {
        if(!this.hasSongs()) {
            return null;
        }

        return this.songs[this.getIndex()];
    }


    private hasSongs() : boolean {
        return this.songs.length > 0;
    }

    /**
     * Sets and clamps index to [0-songs.length).
     * @param index 
     */
    setIndex(index: number) {
        if(!this.hasSongs()) {
            return;
        }

        if(index < 0) {
            index = this.songs.length - 1;
        } else {
            index = index % this.songs.length;
        }

        this.currentIndex = index;
    }

    private getIndex() : number {
        return this.currentIndex;
    }

    private currentIndex: number;
}

export default Album;