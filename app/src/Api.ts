import * as $ from "jquery";
import {Promise} from 'es6-promise';
import PubSub from './PubSub';
import * as jwt_decode from 'jwt-decode';

let events = {
    Authenticated: 'Api:Authenticated',
    LoggedOut: 'Api:LoggedOut',
};

class Api {
    constructor() {

    }

    authenticate(username: string, password: string) : Promise<void> {
        let p = new Promise<void>((resolve, reject) => {
            $.ajax({
                method: "post",
                dataType: 'json',
                url: '/login',
                data: {username: username, password: password},
            })
            .then((response) => {
                this.storeToken(response.token);
                PubSub.publish(events.Authenticated);
                resolve();
            })
            .fail((response) => {
                console.log('Api::authenticate Authentication failed.');
                if(response.responseJSON) {
                    reject(response.responseJSON);
                } else {
                    reject({'message': 'Something is wrong on the server.'});
                }
            });
        });

        return p;
    }

    getAlbums() : Promise<any> {
        let p = new Promise<any>((resolve, reject) => {
            $.ajax({
                method: "get",
                dataType: 'json',
                url: this.endpoint + 'albums',
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
                if(response.responseJSON) {
                    reject(response.responseJSON);
                } else {
                    reject({'message': 'Something is wrong on the server.'});
                }
            });
        });

        return p;
    }

    getAlbum(id: number) : Promise<any> {
        let p = new Promise<any>((resolve, reject) => {
            $.ajax({
                method: "get",
                dataType: 'json',
                url: this.endpoint + 'albums/' + id.toString(),
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
                if(response.responseJSON) {
                    reject(response.responseJSON);
                } else {
                    reject({'message': 'Something is wrong on the server.'});
                }
            });
        });

        return p;
    }

    getMe() : Promise<any> {
        let p = new Promise<any>((resolve, reject) => {
            let decoded = jwt_decode(this.retrieveToken());
            return resolve({
                username: decoded.username,
            });
        });

        return p;
    }

    logout() {
        localStorage.removeItem('api.token');
        PubSub.publish(events.LoggedOut);
    }

    isAuthenticated() : boolean {
        return this.retrieveToken() !== null;
    }

    private storeToken(token: string) {
        localStorage.setItem('api.token', token);
    }

    private retrieveToken() : string {
        return localStorage.getItem('api.token');
    }

    private setToken(xhr: JQueryXHR) {
        xhr.setRequestHeader('Authorization', 'Bearer ' + this.retrieveToken());
    }

    private checkUnauthorized(response: any) {
        if(response.status === 401) {
            this.logout();
        }
    }


    private endpoint: string = '/api/';
}

export {events};
export default new Api();