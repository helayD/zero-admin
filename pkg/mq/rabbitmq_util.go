package mq

import (
	"fmt"
	"sync"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/zeromicro/go-zero/core/logx"
)

// RabbitMQ 配置
type RabbitMQ struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	MqUrl   string
	mu      sync.Mutex // 保护 conn 重连过程并发安全
}

// openChannel 获取 channel；若底层 connection 已被服务端关闭（idle timeout、重启、网络抖动），
// 自动重连后再返回新 channel。修复 sms-rpc 等长驻服务遇到 504 "channel/connection is not open" 的死链问题。
func (r *RabbitMQ) openChannel() (*amqp.Channel, error) {
	if r.MqUrl == "" {
		return nil, fmt.Errorf("rabbitmq connection is not initialized")
	}

	r.mu.Lock()
	if r.conn == nil || r.conn.IsClosed() {
		conn, err := amqp.Dial(r.MqUrl)
		if err != nil {
			r.mu.Unlock()
			logx.Errorf("rabbitmq 重连失败: %+v", err)
			return nil, fmt.Errorf("rabbitmq 重连失败: %w", err)
		}
		// 关闭旧连接（如果还存在）
		if r.conn != nil {
			_ = r.conn.Close()
		}
		r.conn = conn
		logx.Info("rabbitmq 连接重建成功")
	}
	conn := r.conn
	r.mu.Unlock()

	return conn.Channel()
}

// NewRabbitMQ 创建结构体实例
func NewRabbitMQ(MqUrl string) *RabbitMQ {
	return &RabbitMQ{MqUrl: MqUrl}
}

// Destroy 断开channel 和 connection
func (r *RabbitMQ) Destroy() {
	if r.channel != nil {
		if err := r.channel.Close(); err != nil {
			return
		}
	}

	if r.conn != nil {
		if err := r.conn.Close(); err != nil {
			return
		}
	}
}

// NewRabbitMQSimple 创建简单模式下RabbitMQ实例
func NewRabbitMQSimple(MqUrl string) *RabbitMQ {
	// 创建RabbitMQ实例
	rabbitmq := NewRabbitMQ(MqUrl)
	var err error
	// 获取connection
	rabbitmq.conn, err = amqp.Dial(rabbitmq.MqUrl)
	if err != nil {
		logx.Errorf("rabbitmq获取connection失败：%s:%+v", "failed to connect rabbitmq!", err)
		panic(err)
	}
	// 获取channel
	rabbitmq.channel, err = rabbitmq.conn.Channel()
	if err != nil {
		logx.Errorf("rabbitmq获取channel失败：%s:%+v", "failed to open a channel", err)
		panic(err)
	}

	logx.Info("rabbitmq连接成功")
	return rabbitmq
}

// PublishSimple 直接模式队列生产
func (r *RabbitMQ) PublishSimple(queueName string, message []byte) error {
	channel, err := r.openChannel()
	if err != nil {
		logx.Errorf("rabbitmq获取channel失败：%s:%+v", "failed to open a channel", err)
		return fmt.Errorf("rabbitmq获取channel失败: %v", err)
	}
	defer func() {
		_ = channel.Close()
	}()

	// 1.申请队列，如果队列不存在会自动创建，存在则跳过创建
	_, err = channel.QueueDeclare(
		queueName,
		// 是否持久化
		true,
		// 是否自动删除
		false,
		// 是否具有排他性
		false,
		// 是否阻塞处理
		false,
		// 额外的属性
		nil,
	)
	if err != nil {
		logx.Errorf("rabbitmq申请队列：%s失败, 错误消息: %+v", queueName, err)
		return fmt.Errorf("rabbitmq申请队列失败: %v", err)
	}
	// 调用channel 发送消息到队列中
	return channel.Publish(
		"",
		queueName,
		// 如果为true，根据自身exchange类型和routeKey规则无法找到符合条件的队列会把消息返还给发送者
		false,
		// 如果为true，当exchange发送消息到队列后发现队列上没有消费者，则会把消息返还给发送者
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        message,
		})
}

// ConsumeSimple simple 模式下消费者
func (r *RabbitMQ) ConsumeSimple(queueName string, handler func([]byte)) {
	channel, err := r.openChannel()
	if err != nil {
		logx.Errorf("rabbitmq获取channel失败：%s:%+v", "failed to open a channel", err)
		panic(err)
	}

	// 1.申请队列，如果队列不存在会自动创建，存在则跳过创建
	q, err := channel.QueueDeclare(
		queueName,
		// 是否持久化
		true,
		// 是否自动删除
		false,
		// 是否具有排他性
		false,
		// 是否阻塞处理
		false,
		// 额外的属性
		nil,
	)
	if err != nil {
		_ = channel.Close()
		logx.Errorf("rabbitmq申请队列：%s失败, 错误消息: %+v", queueName, err)
		panic(err)
	}

	// 接收消息
	msgs, err := channel.Consume(
		q.Name, // queue
		// 用来区分多个消费者
		"", // consumer
		// 是否自动应答
		true, // auto-ack
		// 是否独有
		false, // exclusive
		// 设置为true，表示 不能将同一个Connection中生产者发送的消息传递给这个Connection中 的消费者
		false, // no-local
		// 列是否阻塞
		false, // no-wait
		nil,   // args
	)
	if err != nil {
		_ = channel.Close()
		logx.Errorf("rabbitmq队列：%s接收消息失败, 错误消息: %+v", queueName, err)
		panic(err)
	}

	forever := make(chan bool)
	// 启用协程处理消息
	go func() {
		defer func() {
			_ = channel.Close()
		}()
		for d := range msgs {
			// 消息逻辑处理
			// log.Printf("Received a message: %s", d.Body)
			handler(d.Body)

		}
	}()

	logx.Infof("queue: %s, consumers %d Waiting for messages ...", queueName, q.Consumers)
	<-forever

}

// ConsumeSimpleWithAck simple 模式下消费者（手动 ACK）
// handler 返回 error 时消息会被 NACK 并重新入队，成功时 ACK
func (r *RabbitMQ) ConsumeSimpleWithAck(queueName string, handler func([]byte) error) {
	channel, err := r.openChannel()
	if err != nil {
		logx.Errorf("rabbitmq获取channel失败：%s:%+v", "failed to open a channel", err)
		panic(err)
	}

	q, err := channel.QueueDeclare(
		queueName,
		true,  // 持久化
		false, // 自动删除
		false, // 排他性
		false, // 阻塞
		nil,
	)
	if err != nil {
		_ = channel.Close()
		logx.Errorf("rabbitmq申请队列：%s失败, 错误消息: %+v", queueName, err)
		panic(err)
	}

	msgs, err := channel.Consume(
		q.Name,
		"",
		false, // auto-ack=false（手动模式）
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		_ = channel.Close()
		logx.Errorf("rabbitmq队列：%s接收消息失败, 错误消息: %+v", queueName, err)
		panic(err)
	}

	forever := make(chan bool)
	go func() {
		defer func() {
			_ = channel.Close()
		}()
		for d := range msgs {
			if err := handler(d.Body); err != nil {
				logx.Errorf("消息处理失败，重新入队: %s, err: %v", d.Body, err)
				_ = d.Nack(false, true)
			} else {
				_ = d.Ack(false)
			}
		}
	}()

	logx.Infof("queue(manual-ack): %s, consumers %d Waiting for messages ...", queueName, q.Consumers)
	<-forever
}

// SendDelayMessage 发送延时取消消息
func (r *RabbitMQ) SendDelayMessage(exchange, queueName, key string, message []byte, delayMinutes int) error {
	channel, err := r.openChannel()
	if err != nil {
		logx.Errorf("rabbitmq获取channel失败：%s:%+v", "failed to open a channel", err)
		return fmt.Errorf("rabbitmq获取channel失败: %v", err)
	}
	defer func() {
		_ = channel.Close()
	}()

	// 声明延时交换机
	err = channel.ExchangeDeclare(
		exchange,            // 交换机名称
		"x-delayed-message", // 类型（需要延时插件）
		true,                // 持久化
		false,               // 自动删除
		false,               // 内部
		false,               // 不等待
		amqp.Table{
			"x-delayed-type": "direct",
		},
	)
	if err != nil {
		logx.Errorf("声明延时交换机失败：%+v", err)
		return fmt.Errorf("声明延时交换机失败: %v", err)
	}

	// 声明订单取消队列
	_, err = channel.QueueDeclare(
		queueName, // 队列名称
		true,      // 持久化
		false,     // 自动删除
		false,     // 排他性
		false,     // 不等待
		nil,
	)
	if err != nil {
		logx.Errorf("声明队列失败：%+v", err)
		return fmt.Errorf("声明队列失败: %v", err)
	}

	// 绑定队列到交换机
	err = channel.QueueBind(
		queueName, // 队列名
		key,       // 路由键
		exchange,  // 交换机
		false,
		nil,
	)
	if err != nil {
		logx.Errorf("绑定队列失败：%+v", err)
		return fmt.Errorf("绑定队列失败: %v", err)
	}

	err = channel.Publish(
		exchange, // 交换机
		key,      // 路由键
		false,    // 强制
		false,    // 立即
		amqp.Publishing{
			ContentType:  "application/json",
			Body:         message,
			DeliveryMode: amqp.Persistent,
			Headers: amqp.Table{
				"x-delay": int64(delayMinutes * 60 * 1000), // 延时时间（毫秒）
			},
		},
	)
	if err != nil {
		return fmt.Errorf("发送延时消息失败: %v", err)
	}

	return nil
}

// SendMessage 发送消息（交换机类型由 caller 指定）
func (r *RabbitMQ) SendMessage(exchange, exchangeType, queueName, key string, message []byte) error {
	channel, err := r.openChannel()
	if err != nil {
		logx.Errorf("rabbitmq获取channel失败：%s:%+v", "failed to open a channel", err)
		return fmt.Errorf("rabbitmq获取channel失败: %v", err)
	}
	defer func() {
		_ = channel.Close()
	}()

	// 声明交换机
	err = channel.ExchangeDeclare(
		exchange,     // 交换机名称
		exchangeType, // 类型（direct/topic/fanout）
		true,         // 持久化
		false,        // 自动删除
		false,        // 内部
		false,        // 不等待
		nil,          // 其他属性
	)
	if err != nil {
		logx.Errorf("声明交换机失败：%+v", err)
		return fmt.Errorf("声明交换机失败: %v", err)
	}

	// 声明队列
	_, err = channel.QueueDeclare(
		queueName, // 队列名称
		true,      // 持久化
		false,     // 自动删除
		false,     // 排他性
		false,     // 不等待
		nil,
	)
	if err != nil {
		logx.Errorf("声明队列失败：%+v", err)
		return fmt.Errorf("声明队列失败: %v", err)
	}

	// 绑定队列到交换机
	err = channel.QueueBind(
		queueName, // 队列名
		key,       // 路由键
		exchange,  // 交换机
		false,
		nil,
	)
	if err != nil {
		logx.Errorf("绑定队列失败：%+v", err)
		return fmt.Errorf("绑定队列失败: %v", err)
	}

	err = channel.Publish(
		exchange, // 交换机
		key,      // 路由键
		false,    // 强制
		false,    // 立即
		amqp.Publishing{
			ContentType:  "application/json",
			Body:         message,
			DeliveryMode: amqp.Persistent,
		},
	)
	if err != nil {
		return fmt.Errorf("发送消息失败: %v", err)
	}

	return nil
}

// ConsumeTopicQueue topic 交换机模式下的消费者
// queueName: 队列名
// exchange: 交换机名
// routingKey: 路由键（支持通配符，如 pms.product.*.key）
// handler: 消息处理函数
func (r *RabbitMQ) ConsumeTopicQueue(queueName, exchange, routingKey string, handler func([]byte)) {
	channel, err := r.openChannel()
	if err != nil {
		logx.Errorf("rabbitmq获取channel失败：%s:%+v", "failed to open a channel", err)
		panic(err)
	}

	// 声明 topic 类型交换机
	err = channel.ExchangeDeclare(
		exchange, // 交换机名称
		"topic",  // 类型
		true,     // 持久化
		false,    // 自动删除
		false,    // 内部
		false,    // 不等待
		nil,      // 其他属性
	)
	if err != nil {
		_ = channel.Close()
		logx.Errorf("声明topic交换机 %s 失败: %+v", exchange, err)
		panic(err)
	}

	// 声明队列
	q, err := channel.QueueDeclare(
		queueName,
		true,  // 持久化
		false, // 自动删除
		false, // 排他性
		false, // 不阻塞
		nil,
	)
	if err != nil {
		_ = channel.Close()
		logx.Errorf("声明队列 %s 失败: %+v", queueName, err)
		panic(err)
	}

	// 绑定队列到 topic 交换机（使用通配符路由键）
	err = channel.QueueBind(
		queueName,  // 队列名
		routingKey, // 路由键（支持 * 和 # 通配符）
		exchange,   // 交换机
		false,
		nil,
	)
	if err != nil {
		_ = channel.Close()
		logx.Errorf("绑定队列 %s 到交换机 %s 失败(routingKey=%s): %+v", queueName, exchange, routingKey, err)
		panic(err)
	}

	// 接收消息
	msgs, err := channel.Consume(
		q.Name, // queue
		"",     // consumer tag
		true,   // auto-ack（当前实现使用 auto-ack=true，后续 Story 7.4 可改进为手动 ACK）
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,
	)
	if err != nil {
		_ = channel.Close()
		logx.Errorf("队列 %s 接收消息失败: %+v", queueName, err)
		panic(err)
	}

	logx.Infof("topic consumer 已启动: queue=%s, exchange=%s, routingKey=%s", queueName, exchange, routingKey)

	forever := make(chan bool)
	go func() {
		defer func() {
			_ = channel.Close()
		}()
		for d := range msgs {
			handler(d.Body)
		}
	}()

	<-forever
}
