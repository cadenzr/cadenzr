import {Promise} from 'es6-promise';
import * as _ from 'lodash';

import Song from './Song';
import PubSub from './PubSub';

let events = {
    SongChanged: 'AudioPlayer:song-changed',
    VolumeChanged: 'AudioPlayer:volume-changed',
    TimeChanged: 'AudioPlayer:time-changed',
    Play: 'AudioPlayer:Play',
    Pause: 'AudioPlayer:pause',
    QueueChanged: 'AudioPlayer:QueueChanged'
};

interface Provider {
    nextSong() : Song;
    prevSong() : Song;
    currentSong() : Song;
}

class NullProvider implements Provider {
    nextSong() : Song {
        return null;
    }

    prevSong() : Song {
        return null;
    }

    currentSong() : Song {
        return null;
    }
}

class AudioPlayer {
    constructor() {
        let self = this;

        this.provider = new NullProvider();
        this.currentQueue = [];
        this.currentIndex = null;

        this.audioEl = document.createElement('audio');
        this.audioEl.style.display = 'none';

        this.audioEl.addEventListener('timeupdate', () => {
            PubSub.publish(events.TimeChanged, self.audioEl.currentTime);
        });

        this.audioEl.addEventListener('volumechange', () => {
            PubSub.publish(events.VolumeChanged, self.audioEl.volume * 100);
        });

        this.audioEl.addEventListener('ended', () => {
            self.next();
        });

        this.setVolume(30);

        document.body.appendChild(this.audioEl);
    }

    /**
     * Get the current playing song.
     */
    currentSong() : Song {
        if(this.currentIndex === null) {
            return null;
        }

        return this.currentQueue[this.currentIndex];
    }

    /**
     * Check if a song is the current song. Check is done by id.
     * 
     * @param s Check if this song is the current song.
     */
    isCurrentSong(s: Song | number) : boolean {
        let cs = this.currentSong();
        if(cs === null) {
            return false;
        }

        let id = 0;
        if(s instanceof Song) {
            id = s.id;
        } else {
            id = s;
        }

        return id === cs.id;
    }

    play() {
        this.audioEl.play();
        this.audioEl.onplaying = () => {
            PubSub.publish(events.Play, this.provider.currentSong());
        };
    }

    pause() {
        this.audioEl.pause();
        PubSub.publish(events.Pause, this.provider.currentSong());
    }

    next() : Promise<void> {
        if(this.currentIndex === null) {
            return Promise.resolve();
        }

        this.currentIndex++;
        if(this.currentIndex >= this.currentQueue.length) {
            this.currentIndex = 0;
        }

        return this.reload()
        .then(() => {
            this.play();
        });
    }

    prev() : Promise<void> {
        if(this.currentIndex === null) {
            return Promise.resolve();
        }

        this.currentIndex--;
        if(this.currentIndex < 0) {
            this.currentIndex = this.currentQueue.length-1;
        }

        return this.reload()
        .then(() => {
            this.play();
        });
    }

    setQueue(queue: Array<Song>) {
        this.currentQueue = queue;
        if(this.currentQueue.length === 0) {
            this.currentIndex = null;
        } else {
            this.currentIndex = 0;
        }

        PubSub.publish(events.QueueChanged, this.currentQueue);
    }

    getQueue() : Array<Song> {
        return this.currentQueue;
    }

    /**
     * Set the current song. Does not yet start playing.
     * The song should be in the queue.
     * 
     * @param s 
     */
    setCurrentSong(s: Song) {
        let index = _.findIndex(this.currentQueue, (s2) => { return s2.id === s.id; });
        if(index === -1) {
            return Promise.resolve();
        }

        this.currentIndex = index;
    }

    /**
     * 
     * @param s Seek to time. Should be between 0-totalDuration.
     */
    seek(s: number) {
        s = Math.min(Math.max(0, s), this.audioEl.duration);

        this.audioEl.currentTime = s;
    }

    /**
     * 
     * @param s Seek to a percentage of the audio. Between 0-100;
     */
    seekFromPercentage(s: number) {
        s = Math.min(Math.max(0, s), 100);
        s = s / 100;

        this.seek(this.audioEl.duration * s);
    }

    /**
     * 
     * @param v Number between 0-100.
     */
    setVolume(v: number) {
        v = Math.min(Math.max(0, v), 100);
        v = v / 100;

        this.audioEl.volume = v;
    }

    getVolume() : number {
        return this.audioEl.volume * 100;
    }

    reload() : Promise<void> {
        if(this.currentIndex === null) {
            return Promise.resolve();
        }

        let s = this.currentSong()
        return this.loadSong(s, true);
    }

    private loadSong(s: Song, publish: boolean) : Promise<any> {
        var self = this;

        let p = new Promise<any>((resolve) => {
            self.audioEl.ondurationchange = () => {
                s.duration = self.audioEl.duration;

                if(publish) {
                    PubSub.publish(events.SongChanged, s);
                }

                resolve();
            };
        });

        self.audioEl.src = 'songs/' + s.id + '/stream';
        self.audioEl.load(); // Required so events are being generated.

        return p;
    }

    private audioEl: HTMLMediaElement
    private provider: Provider;
    private currentQueue: Array<Song>;
    private currentIndex: number;
}

export {events};
export {Provider};
export default new AudioPlayer();