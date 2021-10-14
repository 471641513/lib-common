package kafka_wrapper

import (
	"context"
	"runtime/debug"
	"sync"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/xutils/lib-common/local_context"
	"github.com/xutils/lib-common/utils"
	"github.com/xutils/lib-common/xlog"
)

type KafkaConsumerConfig struct {
	Brokers string   `toml:"brokers"`
	Topics  []string `toml:"topics"`
	Group   string   `toml:"group"`
	Worker  int      `toml:"worker"`
}

type KafkaConsumer struct {
	ctx    *local_context.LocalContext
	cancel context.CancelFunc

	consumer *kafka.Consumer
	conf     KafkaConsumerConfig
	wg       *sync.WaitGroup
	msgQueue chan []byte
	callback func(data []byte)
}

func NewKafkaConsumer(
	conf KafkaConsumerConfig,
	callback func(data []byte)) (*KafkaConsumer, error) {
	ctx := local_context.NewLocalContext()
	cc, err := kafka.NewConsumer(
		&kafka.ConfigMap{
			"bootstrap.servers": conf.Brokers,
			"group.id":          conf.Group,
		})
	if err != nil {
		return nil, err
	}

	err = cc.SubscribeTopics(conf.Topics, nil)
	if err != nil {
		return nil, err
	}
	consumer := &KafkaConsumer{
		ctx:      ctx,
		consumer: cc,
		callback: callback,
		conf:     conf,
		wg:       &sync.WaitGroup{},
		msgQueue: make(chan []byte, 32),
	}

	consumer.ctx.Context, consumer.cancel = context.WithCancel(context.Background())
	//xlog.Debug(" %s|| kafka consumer inited||conf=%v", consumer.ctx.LogId(), utils.MustString(conf))
	return consumer, nil
}

func (consumer *KafkaConsumer) Start() {
	consumer.run()
	xlog.Info(" %s|| kafka consumer started||conf=%v", consumer.ctx.LogId(), utils.MustString(consumer.conf))
}

func (consumer *KafkaConsumer) Stop() {
	err := consumer.consumer.Unsubscribe()
	if err != nil {
		xlog.Error(" %s||failed to Unsubscribe kafka consumer safely||err=%v", consumer.ctx.LogId(), err)
	}
	consumer.cancel()
	consumer.wg.Wait()
	xlog.Info(" %s|| kafka consumer stopped", consumer.ctx.LogId())
	err = consumer.consumer.Close()
	if err != nil {
		xlog.Error(" %s||failed to close kafka consumer safely||err=%v", consumer.ctx.LogId(), err)
	}
}

func (consumer *KafkaConsumer) msgCallback(data []byte) {
	defer func() {
		if e := recover(); e != nil {
			xlog.Fatal("panic=%v||\n%s", e, debug.Stack())
		}
	}()
	consumer.callback(data)
}

func (consumer *KafkaConsumer) run() {
	for i := 0; i < consumer.conf.Worker; i++ {
		go consumer.consumerWorker(i)
	}
	go consumer.receive()
}

func (consumer *KafkaConsumer) consumerWorker(idx int) {
	xlog.Info(" %s||consumer-%d start", consumer.ctx.LogId(), idx)
	consumer.wg.Add(1)
	defer func() {
		xlog.Info(" %s||consumer-%d stoped", consumer.ctx.LogId(), idx)
		consumer.wg.Done()
	}()
	for {
		select {
		case <-consumer.ctx.Done():
			// check remaining
			remain := len(consumer.msgQueue)
			if remain > 0 {
				xlog.Info("deal with remaining kafka msg||len=%v", remain)
				timeout := time.After(time.Second)
				select {
				case msg := <-consumer.msgQueue:
					consumer.msgCallback(msg)
					continue
				case <-timeout:
					return
				}
			}
			return
		case msg := <-consumer.msgQueue:
			consumer.msgCallback(msg)
		}
	}
}

func (consumer *KafkaConsumer) receive() {
	consumer.wg.Add(1)
	defer func() {
		consumer.wg.Done()
	}()

	xlog.Debug("consumer receive started")
	for {
		select {
		case <-consumer.ctx.Done():
			return
		default:
			msg, err := consumer.consumer.ReadMessage(5 * time.Second)
			//xlog.Debug("msg=%v||err=%v", msg, err)
			if err != nil {
				if kafkaErr, ok := err.(kafka.Error); ok {
					if kafkaErr.Code() == kafka.ErrTimedOut {
						continue
					}
				}
				xlog.Warn(" %v|| failed to read msg||err=%v",
					consumer.ctx.LogId(), err)
				time.Sleep(3 * time.Second)
				continue
			}
			consumer.msgQueue <- msg.Value
		}
	}
}
