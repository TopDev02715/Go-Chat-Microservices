package chat

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/minghsu0107/go-random-chat/pkg/common"
	"github.com/minghsu0107/go-random-chat/pkg/config"
	log "github.com/sirupsen/logrus"
	metrics "github.com/slok/go-http-metrics/metrics/prometheus"
	prommiddleware "github.com/slok/go-http-metrics/middleware"
	ginmiddleware "github.com/slok/go-http-metrics/middleware/gin"
	"gopkg.in/olahol/melody.v1"
)

var (
	sessUidKey = "sessuid"
	sessCidKey = "sesscid"

	MelodyMatch MelodyMatchConn
	MelodyChat  MelodyChatConn
)

type MelodyMatchConn struct {
	*melody.Melody
}
type MelodyChatConn struct {
	*melody.Melody
}

type HttpServer struct {
	svr             *gin.Engine
	mm              MelodyMatchConn
	mc              MelodyChatConn
	httpPort        string
	httpServer      *http.Server
	matchSubscriber *MatchSubscriber
	msgSubscriber   *MessageSubscriber
	userSvc         UserService
	msgSvc          MessageService
	matchSvc        MatchingService
	chanSvc         ChannelService
}

func NewMelodyMatchConn() MelodyMatchConn {
	MelodyMatch = MelodyMatchConn{
		melody.New(),
	}
	return MelodyMatch
}

func NewMelodyChatConn(config *config.Config) MelodyChatConn {
	m := melody.New()
	m.Config.MaxMessageSize = config.Chat.Message.MaxSizeByte
	MelodyChat = MelodyChatConn{
		m,
	}
	return MelodyChat
}

func NewGinServer(name string, config *config.Config) *gin.Engine {
	common.InitLogging()

	svr := gin.New()
	svr.Use(gin.Recovery())
	svr.Use(common.LoggingMiddleware())
	svr.Use(common.MaxAllowed(config.Chat.Http.MaxConn))
	svr.Use(common.CORSMiddleware())

	mdlw := prommiddleware.New(prommiddleware.Config{
		Recorder: metrics.NewRecorder(metrics.Config{
			Prefix: name,
		}),
	})
	svr.Use(ginmiddleware.Handler("", mdlw))
	return svr
}

func NewHttpServer(config *config.Config, obsInjector *common.ObservibilityInjector, svr *gin.Engine, mm MelodyMatchConn, mc MelodyChatConn, matchSubscriber *MatchSubscriber, msgSubscriber *MessageSubscriber, userSvc UserService, msgSvc MessageService, matchSvc MatchingService, chanSvc ChannelService) common.HttpServer {
	initJWT(config)

	return &HttpServer{
		svr:             svr,
		mm:              mm,
		mc:              mc,
		httpPort:        config.Chat.Http.Port,
		matchSubscriber: matchSubscriber,
		msgSubscriber:   msgSubscriber,
		userSvc:         userSvc,
		msgSvc:          msgSvc,
		matchSvc:        matchSvc,
		chanSvc:         chanSvc,
	}
}

func initJWT(config *config.Config) {
	common.JwtSecret = config.Chat.JWT.Secret
	common.JwtExpirationSecond = config.Chat.JWT.ExpirationSecond
}

func (r *HttpServer) RegisterRoutes() {
	r.svr.GET("/api/match", r.Match)
	r.svr.GET("/api/chat", r.StartChat)

	userGroup := r.svr.Group("/api/user")
	{
		userGroup.POST("", r.CreateUser)
		userGroup.GET("/:uid/name", r.GetUserName)
	}
	usersGroup := r.svr.Group("/api/users")
	usersGroup.Use(common.JWTAuth())
	{
		usersGroup.GET("", r.GetChannelUsers)
		usersGroup.GET("/online", r.GetOnlineUsers)
	}
	channelGroup := r.svr.Group("/api/channel")
	channelGroup.Use(common.JWTAuth())
	{
		channelGroup.GET("/messages", r.ListMessages)
		channelGroup.DELETE("", r.DeleteChannel)
	}

	r.mm.HandleConnect(r.HandleMatchOnConnect)
	r.mm.HandleClose(r.HandleMatchOnClose)

	r.mc.HandleMessage(r.HandleChatOnMessage)
	r.mc.HandleConnect(r.HandleChatOnConnect)
	r.mc.HandleClose(r.HandleChatOnClose)
}

func (r *HttpServer) Run() {
	go func() {
		r.RegisterRoutes()
		addr := ":" + r.httpPort
		r.httpServer = &http.Server{
			Addr:    addr,
			Handler: common.NewOtelHttpHandler(r.svr, "chat_http"),
		}
		log.Infoln("http server listening on ", addr)
		err := r.httpServer.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()
	go func() {
		err := r.matchSubscriber.Run()
		if err != nil {
			log.Fatal(err)
		}
	}()
	go func() {
		err := r.msgSubscriber.Run()
		if err != nil {
			log.Fatal(err)
		}
	}()
}
func (r *HttpServer) GracefulStop(ctx context.Context) error {
	err := r.httpServer.Shutdown(ctx)
	if err != nil {
		return err
	}
	err = r.matchSubscriber.GracefulStop()
	if err != nil {
		return err
	}
	err = r.msgSubscriber.GracefulStop()
	if err != nil {
		return err
	}
	return nil
}

func response(c *gin.Context, httpCode int, err error) {
	message := err.Error()
	c.JSON(httpCode, common.ErrResponse{
		Message: message,
	})
}
