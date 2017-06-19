import * as $ from "jquery";
import PubSub from './PubSub';
import * as jwt_decode from 'jwt-decode';
import Song from './Song';
import Album from './Album';
import Playlist from './Playlist';

import * as _ from 'lodash';

let events = {
    Authenticated: 'Api:Authenticated',
    LoggedOut: 'Api:LoggedOut',
};

type User = {
    id: number;
    username: string;
};

class Api {
    constructor() {

    }

    authenticate(username: string, password: string): Promise<void> {
        let p = new Promise<void>((resolve, reject) => {
            $.ajax({
                method: "post",
                dataType: 'json',
                url: this.apiEndpoint + 'login',
                data: { username: username, password: password },
            })
                .then((response) => {
                    this.storeToken(response.token);
                    PubSub.publish(events.Authenticated);
                    resolve();
                })
                .fail((response) => {
                    console.log('Api::authenticate Authentication failed.');
                    if (response.responseJSON) {
                        reject(response.responseJSON);
                    } else {
                        reject({ 'message': 'Something is wrong on the server.' });
                    }
                });
        });

        return p;
    }

    getAlbums(): Promise<any> {
        let p = new Promise<any>((resolve, reject) => {
            $.ajax({
                method: "get",
                dataType: 'json',
                url: this.apiEndpoint + 'albums',
                beforeSend: (xhr) => {
                    this.setToken(xhr);
                },
            })
                .then((response) => {
                    resolve(response);
                })
                .fail((response) => {
                    console.log('Api::getAlbums failed.');
                    this.checkUnauthorized(response);
                    if (response.responseJSON) {
                        reject(response.responseJSON);
                    } else {
                        reject({ 'message': 'Something is wrong on the server.' });
                    }
                });
        });

        return p;
    }

    getAlbum(id: number): Promise<any> {
        let p = new Promise<any>((resolve, reject) => {
            $.ajax({
                method: "get",
                dataType: 'json',
                url: this.apiEndpoint + 'albums/' + id.toString(),
                beforeSend: (xhr) => {
                    this.setToken(xhr);
                },
            })
                .then((response) => {
                    resolve(response);
                })
                .fail((response) => {
                    console.log('Api::getAlbum failed.');
                    this.checkUnauthorized(response);
                    if (response.responseJSON) {
                        reject(response.responseJSON);
                    } else {
                        reject({ 'message': 'Something is wrong on the server.' });
                    }
                });
        });

        return p;
    }
    
    
    getArtists(): Promise<any> {
        let p = new Promise<any>((resolve, reject) => {
            $.ajax({
                method: "get",
                dataType: 'json',
                url: this.apiEndpoint + 'artists',
                beforeSend: (xhr) => {
                    this.setToken(xhr);
                },
            })
                .then((response) => {
                    resolve(response);
                })
                .fail((response) => {
                    console.log('Api::getArtists failed.');
                    this.checkUnauthorized(response);
                    if (response.responseJSON) {
                        reject(response.responseJSON);
                    } else {
                        reject({ 'message': 'Something is wrong on the server.' });
                    }
                });
        });
    
        return p;
    }

    getPlaylists(): Promise<any> {
        let p = new Promise<any>((resolve, reject) => {
            $.ajax({
                method: "get",
                dataType: 'json',
                url: this.apiEndpoint + 'playlists',
                beforeSend: (xhr) => {
                    this.setToken(xhr);
                },
            })
                .then((response) => {
                    response.data = _.map(response.data, (playlist: any) => {
                        playlist.songs = _.map(playlist.songs, (song) => {
                            return new Song(song);
                        });

                        return new Playlist(playlist);
                    })

                    resolve(response);
                })
                .fail((response) => {
                    console.log('Api::getPlaylists failed.');
                    this.checkUnauthorized(response);
                    if (response.responseJSON) {
                        reject(response.responseJSON);
                    } else {
                        reject({ 'message': 'Something is wrong on the server.' });
                    }
                });
        });

        return p;
    }


    getPlaylist(id: number): Promise<any> {
        let p = new Promise<any>((resolve, reject) => {
            $.ajax({
                method: "get",
                dataType: 'json',
                url: this.apiEndpoint + 'playlists/' + id,
                beforeSend: (xhr) => {
                    this.setToken(xhr);
                },
            })
                .then((response) => {
                    response.songs = _.map(response.songs, (song) => {
                        return new Song(song);
                    });

                    resolve(new Playlist(response));
                })
                .fail((response) => {
                    console.log('Api::getPlaylist failed.');
                    this.checkUnauthorized(response);
                    if (response.responseJSON) {
                        reject(response.responseJSON);
                    } else {
                        reject({ 'message': 'Something is wrong on the server.' });
                    }
                });
        });

        return p;
    }

    addSongsToPlaylist(songs: any, playlist: Playlist): Promise<any> {
        // TODO song type can be Song | Array<Song>
        if (!Array.isArray(songs)) {
            songs = [songs];
        }

        let ids = _.map(songs, (song: any) => { return song.id; });

        let p = new Promise<any>((resolve, reject) => {
            $.ajax({
                method: "POST",
                dataType: 'json',
                url: this.apiEndpoint + 'playlists/' + playlist.id + '/songs',
                data: { songs: ids },
                beforeSend: (xhr) => {
                    this.setToken(xhr);
                },
            })
                .then((response) => {
                    resolve();
                })
                .fail((response) => {
                    console.log('Api::addSongToPlaylist failed.');
                    this.checkUnauthorized(response);
                    if (response.responseJSON) {
                        reject(response.responseJSON);
                    } else {
                        reject({ 'message': 'Something is wrong on the server.' });
                    }
                });
        });

        return p;
    }

    createPlaylist(playlist: Playlist): Promise<any> {
        let p = new Promise<any>((resolve, reject) => {
            $.ajax({
                method: "POST",
                dataType: 'json',
                url: this.apiEndpoint + 'playlists',
                data: { name: playlist.name, },
                beforeSend: (xhr) => {
                    this.setToken(xhr);
                },
            })
                .then((response) => {
                    resolve(new Playlist(response));
                })
                .fail((response) => {
                    console.log('Api::createPlaylist failed.');
                    this.checkUnauthorized(response);
                    if (response.responseJSON) {
                        reject(response.responseJSON);
                    } else {
                        reject({ 'message': 'Something is wrong on the server.' });
                    }
                });
        });

        return p;
    }

    deletePlaylist(playlist: Playlist): Promise<any> {
        let p = new Promise<any>((resolve, reject) => {
            $.ajax({
                method: "DELETE",
                url: this.apiEndpoint + 'playlists/' + playlist.id,
                beforeSend: (xhr) => {
                    this.setToken(xhr);
                },
            })
                .then((response) => {
                    resolve(response);
                })
                .fail((response) => {
                    console.log('Api::deletePlaylist failed: ' + JSON.stringify(response));
                    this.checkUnauthorized(response);
                    if (response.responseJSON) {
                        reject(response.responseJSON);
                    } else {
                        reject({ 'message': 'Something is wrong on the server.' });
                    }
                });
        });

        return p;
    }


    deleteSongFromPlaylist(song: Song, playlist: Playlist): Promise<any> {
        let p = new Promise<any>((resolve, reject) => {
            $.ajax({
                method: "DELETE",
                url: this.apiEndpoint + 'playlists/' + playlist.id + '/songs/' + song.id,
                beforeSend: (xhr) => {
                    this.setToken(xhr);
                },
            })
                .then((response) => {
                    resolve(response);
                })
                .fail((response) => {
                    console.log('Api::deleteSongFromPlaylist failed: ' + JSON.stringify(response));
                    this.checkUnauthorized(response);
                    if (response.responseJSON) {
                        reject(response.responseJSON);
                    } else {
                        reject({ 'message': 'Something is wrong on the server.' });
                    }
                });
        });

        return p;
    }

    incrementPlayed(s: Song): Promise<any> {
        let p = new Promise<any>((resolve, reject) => {
            $.ajax({
                method: "post",
                dataType: 'json',
                url: this.apiEndpoint + 'songs/' + s.id.toString() + '/played',
                beforeSend: (xhr) => {
                    this.setToken(xhr);
                },
            })
                .then((response) => {
                    resolve(response);
                })
                .fail((response) => {
                    console.log('Api::incrementPlayed failed.');
                    this.checkUnauthorized(response);
                    if (response.responseJSON) {
                        reject(response.responseJSON);
                    } else {
                        reject({ 'message': 'Something is wrong on the server.' });
                    }
                });
        });

        return p;
    }

    scan(): Promise<any> {
        let p = new Promise<any>((resolve, reject) => {
            $.ajax({
                method: "post",
                url: this.apiEndpoint + 'scan',
                beforeSend: (xhr) => {
                    this.setToken(xhr);
                },
            })
                .then((response) => {
                    resolve(response);
                })
                .fail((response) => {
                    console.log('Api::scan failed.');
                    this.checkUnauthorized(response);
                    if (response.responseJSON) {
                        reject(response.responseJSON);
                    } else {
                        reject({ 'message': 'Something is wrong on the server.' });
                    }
                });
        });

        return p;
    }

    getMe(): Promise<User> {
        let p = new Promise<User>((resolve, reject) => {
            let token = this.retrieveToken();
            if (token === null) {
                reject();
                return;
            }

            let decoded = jwt_decode(token);
            return resolve({
                id: decoded.id,
                username: decoded.username,
            });
        });

        return p;
    }

    logout() {
        localStorage.removeItem('api.token');
        PubSub.publish(events.LoggedOut);
    }

    // Check if JWT expired.
    private jwtValid(token: string): boolean {
        let decoded = jwt_decode(token);
        return (decoded.exp >= (Date.now() / 1000));
    }

    isAuthenticated(): boolean {
        let token = this.retrieveToken();
        if (token === null) {
            return false;
        }

        return this.retrieveToken() !== null && this.jwtValid(token);
    }

    private storeToken(token: string) {
        localStorage.setItem('api.token', token);
    }

    retrieveToken(): string | null {
        return localStorage.getItem('api.token');
    }

    private setToken(xhr: JQueryXHR) {
        xhr.setRequestHeader('Authorization', 'Bearer ' + this.retrieveToken());
    }

    private checkUnauthorized(response: any) {
        if (response.status === 401) {
            this.logout();
        }
    }


    public apiEndpoint: string = '/api/';
}

export { events };
let api = new Api();
export { api };
export default api;