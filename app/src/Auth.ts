import {router} from './main'
import * as Vue from 'vue';
import {jwt_decode} from 'jwt-decode';

export default {

    // authentication status
    authenticated: false,
    ready:         false,
    user:          false,

    // Send a request to the login URL and save the returned JWT
    login(context: any, creds: any, redirect: string) {

        context.$http.post('/login', creds).then((response : any) => {
        
                // get body data
                localStorage.setItem('user', JSON.stringify(response.data));
            
                (<any>Vue).http.headers.common['Authorization'] = 'Bearer ' + this.user.token;
                
                this.authenticated = true
                this.user = response.data
                
                // Redirect to a specified route
                if (redirect) {
                    router.push(redirect)
                }
                

        
          }, (response : any) => {
            // error callback
            console.log("error");
            console.log(response);
          });

    },
    
    // Check if JWT found in local storage
    checkLocalStorage() {
        
        if (localStorage.user) {
            this.user = JSON.parse(localStorage.user);     
            
            if (this.jwtValid(this.user.token)) {
                // Valid token
                (<any>Vue).http.headers.common['Authorization'] = 'Bearer ' + this.user.token;
                this.authenticated = true;
            }
            else {
                // Expired token
                console.log("JWT expired");
                this.logout();
            }
            
            this.ready = true;
        }
        
    },
    
    // Check if JWT expired.
    jwtValid(token: string) {
        let decoded = jwt_decode(token);
        
        return (decoded.exp >= Date.now() / 1000);
    },

    // To log out
    logout() {
        localStorage.removeItem('user');
        this.authenticated = false;
        router.push('/login')
    }
}