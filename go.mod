module match_controller

go 1.15

require (
	github.com/bmizerany/assert v0.0.0-20160611221934-b7ed37b82869 // indirect
	github.com/desertbit/timer v0.0.0-20180107155436-c41aec40b27f // indirect
	github.com/golang/protobuf v1.4.3
	github.com/micro/micro/v3 v3.3.0
	github.com/rs/cors v1.9.0 // indirect
	google.golang.org/protobuf v1.26.0-rc.1
)

// This can be removed once etcd becomes go gettable, version 3.4 and 3.5 is not,
// see https://github.com/etcd-io/etcd/issues/11154 and https://github.com/etcd-io/etcd/issues/11931.
replace google.golang.org/grpc => google.golang.org/grpc v1.26.0

replace github.com/micro/micro/v3 => github.com/wolfplus2048/micro/v3 v3.3.1-0.20220817101940-c66d492da5fa
