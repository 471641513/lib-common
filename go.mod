module github.com/xutils/lib-common

go 1.16

require (
	github.com/BurntSushi/toml v0.3.1
	github.com/confluentinc/confluent-kafka-go v1.3.0
	github.com/go-redis/redis v6.15.7+incompatible
	github.com/golang/groupcache v0.0.0-20200121045136-8c9f03a8e57e
	github.com/golang/protobuf v1.5.2
	github.com/google/go-cmp v0.5.6 // indirect
	github.com/grpc-ecosystem/grpc-gateway v1.16.0
	github.com/jinzhu/gorm v1.9.12
	github.com/json-iterator/go v1.1.9
	github.com/kr/beanstalk v0.0.0-20180818045031-cae1762e4858
	github.com/onsi/ginkgo v1.12.0 // indirect
	github.com/onsi/gomega v1.9.0 // indirect
	github.com/prometheus/client_golang v1.5.0
	github.com/rs/xid v1.2.1
	github.com/smallnest/weighted v0.0.0-20200122032019-adf21c9b8bd1
	github.com/sony/sonyflake v1.0.0
	github.com/stretchr/testify v1.5.1
	golang.org/x/exp v0.0.0-20200224162631-6cc2880d07d6
	golang.org/x/net v0.0.0-20210405180319-a5a99cb37ef4
	google.golang.org/appengine v1.6.6 // indirect
	google.golang.org/genproto v0.0.0-20210903162649-d08c68adba83
	google.golang.org/grpc v1.40.0
	google.golang.org/protobuf v1.27.1
)

replace github.com/xutils/lib-common v0.0.0 => ./ // indirect
