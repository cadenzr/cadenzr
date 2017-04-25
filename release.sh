#!/bin/sh

# Remove old build files
os=`uname`
arch=`uname -m`
version=`git tag -l --points-at HEAD`

name="cadenzr_${version}_${os}_${arch}"

echo "Cleaning old install"
rm -rf $name

echo "Creating directories"
mkdir $name
mkdir $name/app
mkdir $name/images
mkdir $name/media

echo "Copying text files"
cp README.md $name/
cp LICENSE $name/

echo "Compiling Go back-end"
go build
cp ./cadenzr ./$name/

echo "Compiling JS/CSS assets for front-end"
cd app/
npm run build
cd ..
cp -r ./app/dist ./$name/app/

echo "Creating archive"
tar -zcvf $name.tar.gz $name