import * as _ from 'lodash';

class Song {
    id: number;
    album: string;
    artist: string;
    genre: string;
    mime: string;
    name: string;
    stream_location: string;
    year: number;
    duration: number;
    cover: string;

    constructor(data: any) {
        if (data) {
            _.assign(this, data);
        }
    }

    getCoverUrl(): string {
        return this.cover;
    }


}

export default Song;