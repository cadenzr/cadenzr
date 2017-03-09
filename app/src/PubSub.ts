import * as _ from 'lodash';

class PubSub {
    constructor() {
        this.subscribers = {};
        this.id = 1;
    }

    subscribe(topic: string, subscriber: Function) : any {
        if(!this.subscribers.hasOwnProperty(topic)) {
            this.subscribers[topic] = {};
        }

        let id = this.id;
        this.id++;
        this.subscribers[topic][id] = subscriber;
        return {id: id, topic: topic};
    }

    publish(topic: any, data?: any) {
        if(!this.subscribers.hasOwnProperty(topic)) {
            this.subscribers[topic] = {};
        }

        _.forEach(this.subscribers[topic], (subscriber: Function) => {
            subscriber(data);
        })
    }

    unsubscribe(token: any) {
        if(!this.subscribers.hasOwnProperty(token.topic)) {
            return;
        }

        if(!this.subscribers[token.topic].hasOwnProperty(token.id)) {
            return;
        }

        delete this.subscribers[token.topic][token.id];
        if(this.subscribers[token.topic].length === 0 ) {
            delete this.subscribers[token.topic];
        }
    }

    private id: number;
    private subscribers : any;
}

export default new PubSub();