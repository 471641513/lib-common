all: install

GOPATH:=$(CURDIR)

LINUXARG=GOOS=linux GOARCH=amd64

BUILDARG=-ldflags " -s -X main.buildtime=`date '+%Y-%m-%d_%H:%M:%S'` -X main.githash=`git rev-parse HEAD`"

export GO111MODULE=on
export GOPATH
export GW_PATH=./test_proto

p:
	#git submodule update --init
	#git submodule update --remote
	protoc -I. --gofast_out=plugins=grpc:${GW_PATH} test.proto
	protoc -I. --grpc-gateway_out=logtostderr=true:${GW_PATH} test.proto

	find ${GW_PATH} -name "*.pb.gw.go" | xargs sed -i -e '/Register.*HandlerServer(/s/) error {/, interceptor grpc.UnaryServerInterceptor) error {/g'
	find ${GW_PATH} -name "*.pb.gw.go" | xargs sed -i -e '/local_request_.*_0.*pathParams)/s/pathParams)/pathParams, interceptor)/g'
	find ${GW_PATH} -name "*.pb.gw.go" | xargs sed -i -e '/local_request_.*_0.*{/s/) (/, interceptor grpc.UnaryServerInterceptor) (/g'
	find ${GW_PATH} -name "*.pb.gw.go" | xargs ssed -i -e '/func local_request_.*/s/\(func local_request_\)\(.*\)\(_0.*\)/\1\2\3\n\thandlerWrapper := _\2_Handler/'
	find ${GW_PATH} -name "*.pb.gw.go" | xargs ssed -i -e '/msg, err := server\./s/msg, err := server\.\(.*\)(\(.*\),\(.*\))/rsp, err := handlerWrapper(server, \2, dec, interceptor)/'
	find ${GW_PATH} -name "*.pb.gw.go" | xargs ssed -i -e '/handlerWrapper/{$!{N;s/\(protoReq \)\(.*\)/\1\2\n\tdec := func(ptri interface{}) error {\n \
			ptr, _ := ptri.(*\2)\n \
			*(ptr) = protoReq\n \
			return nil\n\t}/;ty;P;D;:y}}'
	find  ${GW_PATH} -name "*.pb.gw.go" | xargs ssed -i -e '/handlerWrapper(/a\\tmsg := rsp.(proto.Message)'

