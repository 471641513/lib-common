package beanstalkd

import (
	"context"
	"github.com/opay-org/lib-common/xlog"
	"sync"
	"time"

	"github.com/kr/beanstalk"
)

type AddrList struct {
	Addrs []string `toml:"addrs"`
}

type ConsumerConfig struct {
	Addrs  []string `toml:"addrs"`
	Tube   string   `toml:"tube"`
	Worker int      `toml:"worker"`
}

type Consumer struct {
	c        *ConsumerConfig
	msgQueue chan []byte
	handler  func([]byte)
	ctx      context.Context
	cancel   context.CancelFunc
	wg       *sync.WaitGroup
}

func (c *Consumer) run() {
	c.wg.Add(len(c.c.Addrs) + c.c.Worker)

	for _, addr := range c.c.Addrs {
		go c.receive(addr)
	}

	for i := 0; i < c.c.Worker; i++ {
		go c.handle()
	}
}

func (c *Consumer) Stop() {
	c.cancel()
	c.wg.Wait()
}

func (c *Consumer) handle() {
	defer c.wg.Done()

	for {
		select {
		case <-c.ctx.Done():
			if len(c.msgQueue) > 0 {
				xlog.Info(" send remaining msgs in queue len=%v", len(c.msgQueue))
				for {
					if len(c.msgQueue) == 0 {
						return
					}
					timeout := time.After(time.Second)
					select {
					case msg := <-c.msgQueue:
						c.handler(msg)
					case <-timeout:
						return
					}
				}
			}
			return
		case msg := <-c.msgQueue:
			c.handler(msg)
		}
	}
}

func (c *Consumer) receive(addr string) {
	defer c.wg.Done()

	var (
		tubeSet *beanstalk.TubeSet
		err     error
	)

	for {
		select {
		case <-c.ctx.Done():
			return
		default:
			if tubeSet == nil {
				tubeSet, err = newTubeSet(addr, c.c.Tube)

				if err != nil {
					xlog.Warn("can't connect beanstalkd server | addr: %s | tube: %s | error: %s", addr, c.c.Tube, err)
					time.Sleep(time.Second)
					continue
				}
			}

			id, body, err := tubeSet.Reserve(5 * time.Second)

			if err != nil {
				if e, ok := err.(beanstalk.ConnError); ok && e.Err == beanstalk.ErrTimeout {
					continue
				}

				xlog.Error("can't reserve job | addr: %s | tube: %s | error: %s", addr, c.c.Tube, err)

				tubeSet = nil
				time.Sleep(3 * time.Second)
				continue
			}

			c.msgQueue <- body

			if err := tubeSet.Conn.Delete(id); err != nil {
				xlog.Error("can't delete job | addr: %s | tube: %s | id: %d | error: %s", addr, c.c.Tube, id, err)
			}
		}
	}

}

func newTubeSet(addr string, tube string) (*beanstalk.TubeSet, error) {
	conn, err := beanstalk.Dial("tcp", addr)

	if err != nil {
		return nil, err
	}

	return beanstalk.NewTubeSet(conn, tube), nil
}

func NewConsumer(c *ConsumerConfig, handler func([]byte)) *Consumer {
	consumer := &Consumer{
		c:        c,
		msgQueue: make(chan []byte, 32),
		handler:  handler,
		wg:       &sync.WaitGroup{},
	}

	consumer.ctx, consumer.cancel = context.WithCancel(context.Background())
	consumer.run()
	return consumer
}

func NewConsumers(handlers map[string]func([]byte), cfg map[string]*ConsumerConfig) (consumers map[string]*Consumer) {
	consumers = make(map[string]*Consumer)
	for name, handler := range handlers {
		if c := cfg[name]; c != nil {
			consumers[name] = NewConsumer(c, handler)
		}
	}
	return
}
