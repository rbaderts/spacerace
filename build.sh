#!/bin/sh -x

mkdir build

#protoc-gen-gogo --gofast_out=./core core/messages.proto

echo $GOPATH
#protoc-gen-gogo \
#protoc --gogofaster_out -I=. -I=../../..  --go_out=./ core/messages.proto
#protoc-gen-gogo -I=. -I=../../..  --go_out=./ core/messages.proto

##/	(protoc -I=. -I=../../../../../ -I=../../protobuf/ --gogo_out=. example.proto)


#protoc --proto_path=core --go_out=core  core/messages.proto
#protoc --proto_path=core --js_out=import_style=commonjs:src/js core/messages.proto
#protoc  -I=core -I=../../..  --js_out=import_style=commonjs,binary:src/js core/messages.proto github.com/gogo/protobuf/gogoproto/gogo.proto

#protoc  -I=core -I=../../..  --js_out=import_style=commonjs,binary:src/js core/messages.proto

#./node_modules/protobufjs/bin/pbjs  -t static-module -w commonjs -o src/js/messages_pb.js core/messages.proto


pushd core;flatc -g msg/gamestate.fbs;popd
pushd core;flatc -s -o ../web/js msg/gamestate.fbs;popd

pushd core;flatc -g msg/playercommands.fbs;popd
pushd core;flatc -s -o ../web/js msg/playercommands.fbs;popd

#pushd core;flatc -g msg/playerupdate.fbs;popd
#pushd core;flatc -s -o ../src/js msg/playerupdate.fbs;popd
#pushd core;flatc -g msg/playerinitialize.fbs;popd
#pushd core;flatc -s -o ../src/js msg/playerinitialize.fbs;popd

npm run build


#tsc src/js/game.js 
#browserify src/js/game.js > resources/js/game_bundle.js

#go build -race
#browserify web/js/game.js > web/js/game_bundle.js
#browserify web/js/world_map.js > web/js/world_map_bundle.js


packr build

#go build 

if [ "$1" == "-docker" ]; then
    echo "Docker"
    GOOS=linux go build 
    docker build -t rickbadertscher/spacerace -f Dockerfile.scratch .
fi

#GOOS=linux go build 
#docker build -t rickbadertscher/spacerace -f Dockerfile.scratch .


