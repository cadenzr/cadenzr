import {router} from './main'
import * as Vue from 'vue';

export default {

    // authentication status
    authenticated: false,
    ready:         false,
    user:          undefined,

    // Send a request to the login URL and save the returned JWT

    login(context, creds, redirect) {

        context.$http.post('/login', creds).then(response => {
        
                // get body data
                localStorage.setItem('user', JSON.stringify(response.data))
                console.log(response.data)
            
                
                this.authenticated = true
                context.$root.user = response.data
                
                // Redirect to a specified route
                alert(redirect);
                if (redirect) {
                    console.log(redirect);
                    router.push(redirect)
                }
                

        
          }, response => {
            // error callback
            console.log("error");
            console.log(response);
          });

    },
    
    checkLocalStorage() {
        
        //console.log(localStorage)
        if (localStorage.user) {
            
            
            this.user = JSON.parse(localStorage.user);
            
            
            if (this.jwtValid(this.user.token)) {
                // Valid token
                Vue.http.headers.common['Authorization'] = 'Bearer ' + this.user.token;
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
    
    jwtValid(token) {
        let jwt_decode = require('jwt-decode');
        var decoded = jwt_decode(token);
        console.log(decoded);
        
        return (decoded.exp >= Date.now() / 1000);
    },

    // To log out
    logout: function() {
        localStorage.removeItem('user');
        this.authenticated = false;
        router.go('/login')
    }
}