Cadenzr
=======

*Self-hosted web app for music streaming*

![](https://cloud.githubusercontent.com/assets/3856745/24114228/bd9e6512-0d9f-11e7-8d4f-4645cc802d35.png)


About
-----

Cadenzr is a webplatform which allows you to create your own streaming service.  
Install Cadenzr, put some music on it, and enjoy your own music wherever you go!


Installing
----------

Get the source code in your `$GOPATH`:

    $ go get github.com/cadenzr/cadenzr
    
Then go to the `app` directory and build all assets for the front-end:
 
    $ cd app; npm run build

Back in the main Cadenzr folder (`$GOPATH/src/github.com/cadenzr/cadenzr`), build the project and run the web service:

    $ go build
    $ ./cadenzr

Your webserver will then run on port `8080` (default username is `admin`, leave password empty).
Copy `config.json.example` to `config.json`, to configure everything. (don't forget to remove the comments, otherwise the JSON is invalid.)


Authors
-------

[Mathias Beke](https://denbeke.be)  
Timo Truyts