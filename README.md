Build & Run 
-----------

* Install and setup latest golang 
* Install protobuf for you platform
    
      (OSX)
      
      % brew install protobuf

* Install gogoproto extendsion

      % go get github.com/gogo/protobuf/proto
      % go get github.com/gogo/protobuf/jsonpb
      % go get github.com/gogo/protobuf/protoc-gen-gogo
      % go get github.com/gogo/protobuf/gogoproto


`% ./build.sh` <br>
`% ./spacerace server` <br>
`localhost:8080` <br>



