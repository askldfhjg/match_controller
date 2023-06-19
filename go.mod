module match_controller

go 1.15

require (
	github.com/golang/protobuf v1.4.3
	github.com/micro/micro/v3 v3.3.0
	github.com/rs/cors v1.9.0 // indirect
	google.golang.org/protobuf v1.26.0-rc.1
)

// This can be removed once etcd becomes go gettable, version 3.4 and 3.5 is not,
// see https://github.com/etcd-io/etcd/issues/11154 and https://github.com/etcd-io/etcd/issues/11931.
replace google.golang.org/grpc => google.golang.org/grpc v1.26.0

replace github.com/micro/micro/v3 v3.3.0 => github.com/askldfhjg/micro/v3 v3.5.0
