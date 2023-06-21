module match_controller

go 1.15

require (
	github.com/askldfhjg/match_apis/match_frontend/proto v0.0.0-20230621021020-57881bae15e6
	github.com/askldfhjg/match_apis/match_process/proto v0.0.0-20230621021020-57881bae15e6
	github.com/golang/protobuf v1.4.3
	github.com/gomodule/redigo v1.8.9
	github.com/micro/micro/v3 v3.3.0
	github.com/pkg/errors v0.9.1
	github.com/rs/cors v1.9.0 // indirect
	google.golang.org/protobuf v1.26.0-rc.1
)

// This can be removed once etcd becomes go gettable, version 3.4 and 3.5 is not,
// see https://github.com/etcd-io/etcd/issues/11154 and https://github.com/etcd-io/etcd/issues/11931.
//replace google.golang.org/grpc => google.golang.org/grpc v1.26.0

replace github.com/micro/micro/v3 v3.3.0 => github.com/askldfhjg/micro/v3 v3.5.0
