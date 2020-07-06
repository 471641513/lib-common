module .

go 1.12

require (
	github.com/opay-org/lib-common v0.0.0
	test_http_svr v0.0.0
	test_proto v0.0.0
)

replace github.com/opay-org/lib-common v0.0.0 => ../../

replace test_proto v0.0.0 => ./test_proto

replace test_http_svr v0.0.0 => ./test_http_svr
