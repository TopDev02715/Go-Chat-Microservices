package chat

import (
	"time"

	"github.com/Shopify/sarama"
	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-kafka/v2/pkg/kafka"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/ThreeDotsLabs/watermill/message/router/middleware"
	"github.com/minghsu0107/go-random-chat/pkg/common"
	"github.com/minghsu0107/go-random-chat/pkg/config"
)

var (
	logger = watermill.NewStdLogger(
		false,
		false,
	)
)

func NewKafkaPublisher(config *config.Config) (message.Publisher, error) {
	kafkaPublisher, err := kafka.NewPublisher(
		kafka.PublisherConfig{
			Brokers:   common.GetServerAddrs(config.Chat.Kafka.Addrs),
			Marshaler: kafka.DefaultMarshaler{},
		},
		logger,
	)
	if err != nil {
		return nil, err
	}

	return kafkaPublisher, nil
}

func NewKafkaSubscriber(config *config.Config) (message.Subscriber, error) {
	saramaConfig := sarama.NewConfig()
	saramaConfig.Consumer.Fetch.Default = 1024 * 1024
	saramaConfig.Consumer.Offsets.AutoCommit.Enable = true
	saramaConfig.Consumer.Offsets.AutoCommit.Interval = 1 * time.Second

	kafkaSubscriber, err := kafka.NewSubscriber(
		kafka.SubscriberConfig{
			Brokers:       common.GetServerAddrs(config.Chat.Kafka.Addrs),
			Unmarshaler:   kafka.DefaultMarshaler{},
			ConsumerGroup: watermill.NewUUID(),
			InitializeTopicDetails: &sarama.TopicDetail{
				NumPartitions:     1,
				ReplicationFactor: 2,
			},
			OverwriteSaramaConfig: saramaConfig,
		},
		logger,
	)
	if err != nil {
		return nil, err
	}

	return kafkaSubscriber, nil
}

func NewMessageRouter() (*message.Router, error) {
	router, err := message.NewRouter(message.RouterConfig{}, logger)
	if err != nil {
		return nil, err
	}

	router.AddMiddleware(
		middleware.CorrelationID,
		middleware.Timeout(time.Second*15),
		middleware.Recoverer,
	)
	return router, nil
}
