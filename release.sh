#!/bin/sh

# Remove old build files
os=`uname`
arch=`uname -m`
#version=`git tag -l --points-at HEAD`
version="0.1-beta"


echo "Compiling JS/CSS assets for front-end"
cd app/
npm install
npm run build
cd ..

go-bindata-assetfs -prefix="app" -ignore="\\.DS_Store" app/dist/...
xgo --targets="" --dest="release" --out="cadenzr-${version}" .

cd release

for f in *
do
    fExt="${f%.*}"
    echo "Building release: ${fExt}"

    name="${fExt}"

    echo "Creating directories"
    mkdir archive
    mkdir ./archive/$name
    mkdir ./archive/$name/app
    mkdir ./archive/$name/images
    mkdir ./archive/$name/media

    echo "Copying text files"
    cp ../README.md ./archive/$name/
    cp ../LICENSE ./archive/$name/
    cp ../config.json.example ./archive/$name/

    #echo "Compiling Go back-end"
    #go build
    echo "Copying build files"
    cp $f ./archive/$name/cadenzr
    #cp -r ../app/dist ./archive/$name/app/

    echo "Creating archive"
    tar -zcf ./archive/$name.tar.gz ./archive/$name

done