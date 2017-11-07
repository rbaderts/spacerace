#!/bin/sh -x

mkdir build

#protoc-gen-gogo --gofast_out=./core core/messages.proto

echo $GOPATH
#//protoc-gen-gogo \
protoc \
     -I=. -I=../../..  --gogoslick_out=./ core/messages.proto

##/	(protoc -I=. -I=../../../../../ -I=../../protobuf/ --gogo_out=. example.proto)


#protoc --proto_path=core --go_out=core  core/messages.proto
#protoc --proto_path=core --js_out=import_style=commonjs:src/js core/messages.proto
protoc  -I=core -I=../../..  --js_out=import_style=commonjs,binary:src/js core/messages.proto github.com/gogo/protobuf/gogoproto/gogo.proto 


npm run build

browserify resources/js/game.js > resources/js/game_bundle.js



#go build -race

go build 

if [ "$1" == "-docker" ]; then
    echo "Docker"
    GOOS=linux go build 
    docker build -t rickbadertscher/spacerace -f Dockerfile.scratch .
fi

#GOOS=linux go build 
#docker build -t rickbadertscher/spacerace -f Dockerfile.scratch .


