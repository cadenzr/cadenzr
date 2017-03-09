import {Provider} from './AudioPlayer';

class Song implements Provider {
    id: number;
    album: string;
    artist: string;
    genre: string;
    mime: string;
    name: string;
    stream_location: string;
    year: number;
    duration: number;

    constructor(data: any) {
        if(data) {
            (<any>Object).assign(this, data); 
        }
    }

    nextSong() : Song {
        return this;
    }

    prevSong() : Song {
        return this;
    }

    currentSong() : Song {
        return this;
    }
}

export default Song;