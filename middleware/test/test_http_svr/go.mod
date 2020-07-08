module test_http_svr

go 1.12

require (
	github.com/grpc-ecosystem/grpc-gateway v1.14.6

	github.com/opay-org/lib-common v0.0.0
	github.com/stretchr/testify v1.4.0
	golang.org/x/net v0.0.0-20200625001655-4c5254603344
	google.golang.org/grpc v1.30.0
	test_proto v0.0.0
)

replace test_proto v0.0.0 => ../test_proto

replace github.com/opay-org/lib-common v0.0.0 => ../../../
