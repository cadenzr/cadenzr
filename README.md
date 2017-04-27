Cadenzr
=======

[![Build Status](https://travis-ci.org/cadenzr/cadenzr.svg?branch=master)](https://travis-ci.org/cadenzr/cadenzr)

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

    $ cd app; npm install; npm run build

Back in the main Cadenzr folder (`$GOPATH/src/github.com/cadenzr/cadenzr`), build the project and run the web service:

    $ go-bindata-assetfs -prefix="app" -ignore="\\.DS_Store" app/dist/...
    $ go build
    $ ./cadenzr

Your webserver will then run on port `8080` (default username is `admin`, leave password empty).
Copy `config.json.example` to `config.json`, to configure everything. (don't forget to remove the comments, otherwise the JSON is invalid.)


Web interface development
----------

Get the source code as described in previous chapter.

The `app/src` directory contains the source code for the web interface.

First you should start the go backend:

    $ ./cadenzr

Then run webpack in development mode:

    $ npm run dev

If your backend is not running on 127.0.0.1:8080 you should modify `app/config/custom.env.js` to suit your needs.

    $ cp ./custom.env.js.example ./custom.env.js
    // Change the target in proxyTable to your backend.


Authors
-------

[Mathias Beke](https://denbeke.be)  
Timo Truyts
