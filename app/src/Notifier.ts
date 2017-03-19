import * as _ from 'lodash';

// typescript does not recognize Notification.
declare var Notification: any;

class Notifier {
    constructor() {
        if(this.nativeSupport()) {
            this.notification = Notification;
        }
    }

    notify(title: string, options?: any) {
        let doNotification = () => {
            new Notification(title, options);
        };

        if(!this.hasPermission()) {
            this.requestPermission()
            .then(() => {
                doNotification();
            });

            return;
        }

        doNotification();
    }

    requestPermission() : Promise<void> {
        if(!this.nativeSupport()) {
            return new Promise<void>((r, reject) => {
                reject();
            });
        }

        var self = this;
        return new Promise<void>((resolve, reject) => {
            self.notification.requestPermission((permission:string) => {
                if(permission === 'granted') {
                    resolve();
                } else {
                    reject();
                }
            });
        });
    }

    private nativeSupport() : boolean {
        return "Notification" in window;
    }

    private hasPermission() : boolean {
        return Notification.permission === 'granted';
    }



    // Use any because typescript does not recognize Notification correctly.
    private notification: any = null;
}

export default new Notifier();