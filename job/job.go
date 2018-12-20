package job

import (
	"context"

	"github.com/gogo/protobuf/proto"
	pb "github.com/swanky2009/goim/grpc/logic"
	"github.com/swanky2009/goim/job/g"
	"github.com/swanky2009/goim/job/g/conf"

	cluster "github.com/bsm/sarama-cluster"
)

// Job is push job.
type Job struct {
	//c        *conf.Config
	consumer *cluster.Consumer
	comet    *Comet
}

// New new a push job.
func New(c *conf.Config) *Job {
	j := &Job{
		//c:        c,
		consumer: newKafkaSub(c.Kafka),
		comet:    NewComet(c.Comet),
	}
	return j
}

func newKafkaSub(c *conf.Kafka) *cluster.Consumer {
	config := cluster.NewConfig()
	config.Consumer.Return.Errors = true
	config.Group.Return.Notifications = true
	consumer, err := cluster.NewConsumer(c.Brokers, c.Group, []string{c.Topic}, config)
	if err != nil {
		panic(err)
	}
	return consumer
}

// Close close resounces.
func (j *Job) Close() {
	if j.consumer != nil {
		j.consumer.Close()
	}
	if j.comet != nil {
		j.comet.Close()
	}
}

// Consume messages, watch signals
func (j *Job) Consume() {
	for {
		select {
		case err := <-j.consumer.Errors():
			g.Logger.Errorf("consumer error(%v)", err)
		case n := <-j.consumer.Notifications():
			g.Logger.Infof("consumer rebalanced(%v)", n)
		case msg, ok := <-j.consumer.Messages():
			if !ok {
				return
			}
			j.consumer.MarkOffset(msg, "")
			// process push message
			pushMsg := new(pb.PushMsg)
			if err := proto.Unmarshal(msg.Value, pushMsg); err != nil {
				g.Logger.Errorf("proto.Unmarshal(%v) error(%v)", msg, err)
				continue
			}
			if err := j.push(context.Background(), pushMsg); err != nil {
				g.Logger.Errorf("j.push(%v) error(%v)", pushMsg, err)
			}
			g.Logger.Infof("consume: %s/%d/%d/%s/%v", msg.Topic, msg.Partition, msg.Offset, msg.Key, pushMsg)
		}
	}
}
