package consumer

import (
	"time"
)

type DelayI interface {
	DelayConsumer(delay time.Duration) error
	//GetDelay() time.Duration
}

//type Delay struct {
//	consumer *kafka.Consumer
//	ctx      context.Context
//	logger   *zap.Logger
//	delay    time.Duration
//}
//
//func NewDelay(ctx context.Context, consumer *kafka.Consumer, l *zap.Logger, delay time.Duration) Delay {
//	return Delay{
//		consumer: consumer,
//		ctx:      ctx,
//		logger:   l,
//		delay:    delay,
//	}
//}
//
//func (d Delay) DelayConsumer() error {
//	if d.delay <= time.Second {
//		return nil
//	}
//	partitions, err := d.consumer.Assignment()
//	if err != nil {
//		return err
//	}
//	err = d.consumer.Pause(partitions)
//	if err != nil {
//		return err
//	}
//	time.Sleep(d.delay - time.Second)
//	select {
//	case <-d.ctx.Done():
//		d.logger.Warn("service already stopped")
//		return nil
//	default:
//		partitions, err = d.consumer.Assignment()
//		if err != nil {
//			return err
//		}
//		err = d.consumer.Resume(partitions)
//		if err != nil {
//			return err
//		}
//	}
//
//	return nil
//}
//
//func (d Delay) GetDelay() time.Duration {
//	return d.delay
//}
