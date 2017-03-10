import {Promise} from 'es6-promise';

import Song from './Song';
import PubSub from './PubSub';

let events = {
    SongChanged: 'AudioPlayer:song-changed',
    VolumeChanged: 'AudioPlayer:volume-changed',
    TimeChanged: 'AudioPlayer:time-changed',
    Play: 'AudioPlayer:Play',
    Pause: 'AudioPlayer:pause',
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

        this.audioEl = document.createElement('audio');
        this.audioEl.style.display = 'none';

        this.audioEl.addEventListener('timeupdate', () => {
            PubSub.publish(events.TimeChanged, self.audioEl.currentTime);
        });

        this.audioEl.addEventListener('volumechange', () => {
            PubSub.publish(events.VolumeChanged, self.audioEl.volume * 100);
        });

        this.setVolume(30);

        document.body.appendChild(this.audioEl);
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
        let s = this.provider.nextSong();
        if(s == null) {
            return;
        }

        return this.loadSong(s, true)
        .then(() => {
            this.play();
        });

    }

    prev() : Promise<void> {
        let s = this.provider.prevSong();
        if(s == null) {
            return;
        }

        return this.loadSong(s, true)
        .then(() => {
            this.play();
        });
    }

    /**
     * Force a refetch of the current song from the provider.
     * Useful if the user wants to play a song in an album for example.
     */
    restartCurrent() : Promise<void> {
        let s = this.provider.currentSong();
        if(s == null) {
            return;
        }

        return this.loadSong(s, true);
    }

    /**
     * 
     * @param p 
     */
    setProvider(p: Provider) {
        this.provider = p;
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

        self.audioEl.src = s.stream_location;
        self.audioEl.load(); // Required so events are being generated.

        return p;
    }

    private audioEl: HTMLMediaElement
    private provider: Provider;
}

export {events};
export {Provider};
export default new AudioPlayer();