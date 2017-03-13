import {router} from './main'

export default {

    // authentication status
    authenticated: false,
    ready:         false,

    // Send a request to the login URL and save the returned JWT

    login(context, creds, redirect) {
        /*
        context.$http.post('/login', creds, (data) => {
            localStorage.setItem('user', JSON.stringify(data))

            this.authenticated = true
            context.$root.user = data

            // Redirect to a specified route
            if (redirect) {
                router.go(redirect)
            }

            console.log("post");

            }).error((errors) => {
                context.errors = errors;
            })
            */
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

    // To log out
    logout: function() {
        localStorage.removeItem('user');
        this.authenticated = false;
        router.go('/login')
    }
}