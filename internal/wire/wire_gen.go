// Code generated by Wire. DO NOT EDIT.

//go:generate wire
//go:build !wireinject
// +build !wireinject

package wire

import (
	"github.com/minghsu0107/go-random-chat/pkg/chat"
	"github.com/minghsu0107/go-random-chat/pkg/common"
	"github.com/minghsu0107/go-random-chat/pkg/config"
	"github.com/minghsu0107/go-random-chat/pkg/infra"
	"github.com/minghsu0107/go-random-chat/pkg/uploader"
	"github.com/minghsu0107/go-random-chat/pkg/web"
)

// Injectors from wire.go:

func InitializeWebServer(name string) (*common.Server, error) {
	configConfig, err := config.NewConfig()
	if err != nil {
		return nil, err
	}
	engine := web.NewGinServer(name)
	httpServer := web.NewHttpServer(name, configConfig, engine)
	router := web.NewRouter(httpServer)
	infraCloser := web.NewInfraCloser()
	observibilityInjector := common.NewObservibilityInjector(configConfig)
	server := common.NewServer(name, router, infraCloser, observibilityInjector)
	return server, nil
}

func InitializeChatServer(name string) (*common.Server, error) {
	configConfig, err := config.NewConfig()
	if err != nil {
		return nil, err
	}
	observibilityInjector := common.NewObservibilityInjector(configConfig)
	engine := chat.NewGinServer(name, configConfig)
	melodyMatchConn := chat.NewMelodyMatchConn()
	melodyChatConn := chat.NewMelodyChatConn(configConfig)
	universalClient, err := infra.NewRedisClient(configConfig)
	if err != nil {
		return nil, err
	}
	redisCache := infra.NewRedisCache(universalClient)
	userRepo := chat.NewRedisUserRepo(redisCache)
	subscriber, err := infra.NewKafkaSubscriber(configConfig)
	if err != nil {
		return nil, err
	}
	matchSubscriber, err := chat.NewMatchSubscriber(melodyMatchConn, userRepo, subscriber)
	if err != nil {
		return nil, err
	}
	messageSubscriber, err := chat.NewMessageSubscriber(subscriber, melodyChatConn)
	if err != nil {
		return nil, err
	}
	idGenerator, err := chat.NewSonyFlake()
	if err != nil {
		return nil, err
	}
	userService := chat.NewUserService(userRepo, idGenerator)
	publisher, err := infra.NewKafkaPublisher(configConfig)
	if err != nil {
		return nil, err
	}
	messageRepo := chat.NewMessageRepo(configConfig, redisCache, publisher)
	messageService := chat.NewMessageService(messageRepo, userRepo, idGenerator)
	matchingRepo := chat.NewMatchingRepo(redisCache, publisher)
	channelRepo := chat.NewRedisChannelRepo(redisCache)
	matchingService := chat.NewMatchingService(matchingRepo, channelRepo, idGenerator)
	channelService := chat.NewChannelService(channelRepo, userRepo)
	httpServer := chat.NewHttpServer(configConfig, observibilityInjector, engine, melodyMatchConn, melodyChatConn, matchSubscriber, messageSubscriber, userService, messageService, matchingService, channelService)
	router := chat.NewRouter(httpServer)
	infraCloser := chat.NewInfraCloser()
	server := common.NewServer(name, router, infraCloser, observibilityInjector)
	return server, nil
}

func InitializeUploaderServer(name string) (*common.Server, error) {
	configConfig, err := config.NewConfig()
	if err != nil {
		return nil, err
	}
	engine := uploader.NewGinServer(name)
	httpServer := uploader.NewHttpServer(name, configConfig, engine)
	router := uploader.NewRouter(httpServer)
	infraCloser := uploader.NewInfraCloser()
	observibilityInjector := common.NewObservibilityInjector(configConfig)
	server := common.NewServer(name, router, infraCloser, observibilityInjector)
	return server, nil
}
