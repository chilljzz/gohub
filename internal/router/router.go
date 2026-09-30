package router

import (
	"fmt"
	"time"

	"github.com/chilljzz/gohub/internal/controller"
	"github.com/chilljzz/gohub/internal/database"
	"github.com/chilljzz/gohub/internal/middleware"
	"github.com/chilljzz/gohub/internal/realtime"
	"github.com/chilljzz/gohub/internal/repository"
	"github.com/chilljzz/gohub/internal/response"
	"github.com/chilljzz/gohub/internal/service"
	"github.com/chilljzz/gohub/internal/ws"
	"github.com/gin-gonic/gin"
)

func InitRouter(
	broker *realtime.RedisBroker,
	wsManager *ws.Manager,
	messageEventPublisher service.MessageEventPublisher,
) (*gin.Engine, error) {
	r := gin.New()

	if err := r.SetTrustedProxies(nil); err != nil {
		return nil, fmt.Errorf("ser trusted proxies: %w", err)
	}

	r.Use(middleware.RequestID())

	r.Use(middleware.AccessLogger())

	r.Use(middleware.SecurityHeaders())

	r.Use(middleware.ErrorMiddleware())

	r.GET("/ping", func(ctx *gin.Context) {
		response.Success(ctx, gin.H{
			"msg": "pong",
		})
	})

	db := database.DB
	messageRepo :=
		repository.NewMessageRepository(
			db,
		)

	userRepo := repository.NewUserRepository()
	friendRepo := repository.NewFriendREpository()
	teamRepo := repository.NewTeamRepository()
	channelRepo := repository.NewChannelRepository()
	presenceRepo :=
		repository.NewPresenceRepository(
			database.RedisClient,
		)
	// legacyMessageRepo := repository.NewChannelMessageRepository()
	// readRepo := repository.NewChannelReadRepository()

	realtimeService := service.NewRealtimeService(
		channelRepo,
		teamRepo,
	)
	presenceService :=
		service.NewPresenceService(
			presenceRepo,
			90*time.Second,
		)
	conversation := repository.NewConversationRepository(db)
	conversationReadRepo := repository.NewConversationReadRepository(db)

	conversationService := service.NewConversationService(
		conversation,
		friendRepo,
		realtimeService,
		presenceService,
	)

	friendService := service.NewFriendService(
		friendRepo,
		userRepo,
	)

	teamService := service.NewTeamService(
		teamRepo,
		userRepo,
	)

	channelService := service.NewChannelService(
		channelRepo,
		teamRepo,
	)

	messageService := service.NewChannelMessageService(
		messageRepo,
		realtimeService,
		conversationService,
		messageEventPublisher,
	)

	conversationMessageService :=
		service.NewConversationMessageService(
			messageRepo,
			conversationService,
			messageEventPublisher,
		)

	readService := service.NewChannelReadService(
		conversationReadRepo,
		messageRepo,
		realtimeService,
		conversationService,
	)

	directReceiptService := service.NewDirectReceiptService(
		conversationReadRepo,
		messageRepo,
		conversationService,
	)

	userController := controller.NewUserController()

	friendController := controller.NewFriendController(
		friendService,
	)

	teamController := controller.NewTeamController(
		teamService,
	)

	channelController := controller.NewChannelController(
		channelService,
	)

	wsMessageHandler := controller.NewWSMessageHandler(
		wsManager,
		realtimeService,
		messageService,
		conversationService,
		conversationMessageService,
		directReceiptService,
		broker,
	)

	wsController := controller.NewWSController(
		wsManager,
		wsMessageHandler,
		presenceService,
	)

	messageController := controller.NewChannelMessageController(
		messageService,
	)
	readController := controller.NewChannelReadController(
		readService,
	)
	conversationController := controller.NewConversationController(conversationService)

	conversationMessageController :=
		controller.NewConversationMessageController(
			conversationMessageService,
		)

	directReceiptController := controller.NewDirectReceiptController(
		directReceiptService,
	)

	loginLimiter := middleware.NewRateLimiter(
		database.RedisClient,
		"gohub:ratelimit:login",
		10,
		time.Minute,
	)
	registerLimiter := middleware.NewRateLimiter(
		database.RedisClient,
		"gohub:ratelimit:register",
		5,
		time.Minute,
	)

	apiLimiter := middleware.NewRateLimiter(
		database.RedisClient,
		"gohub:ratelimit:api",
		300,
		time.Minute,
	)
	// wsLimiter := middleware.NewRateLimiter(
	// 	database.RedisClient,
	// 	"gohub:ratelimit:ws",
	// 	20,
	// 	time.Minute,
	// )

	api := r.Group("/api")

	user := api.Group("/users")
	{
		user.POST("/register", registerLimiter.ByIP(), userController.Register)
		user.POST("/login", loginLimiter.ByIP(), userController.Login)

	}

	auth := api.Group("")
	auth.Use(middleware.AuthMiddleware(), apiLimiter.ByUser())
	{

		auth.GET("/ws", wsController.Connect)

		auth.GET("/users/me", userController.Me)
		auth.PUT("/users/me", userController.UpdateMe)
		friends := auth.Group("/friends")
		{
			friends.POST(
				"/requests",
				friendController.SendRequest,
			)
			friends.GET(
				"/requests",
				friendController.ListPendingRequests,
			)
			friends.POST(
				"/requests/:id/accept",
				friendController.AcceptRequest,
			)
			friends.GET(
				"",
				friendController.ListFriends,
			)
		}

		teams := auth.Group("/teams")
		{
			teams.POST("", teamController.Create)
			teams.GET("", teamController.ListMine)
			teams.GET("/:id/members", teamController.ListMembers)
			teams.POST("/:id/members", teamController.AddMember)

			channels := teams.Group("/:id/channels")
			{
				channels.POST("", channelController.Create)
				channels.GET("", channelController.List)
			}
		}
		auth.GET("/channels/:id/messages", messageController.ListRecent)
		auth.POST("/channels/:id/read", readController.MarkRead)
		auth.GET("/channels/:id/unread-count", readController.UnreadCount)
		auth.GET("/channels/:id/messages/sync", messageController.Sync)

		conversationGroup := auth.Group("/conversations")
		{
			conversationGroup.POST("/direct", conversationController.CreateDirect)

			conversationGroup.GET(
				"/:id/messages",
				conversationMessageController.ListRecent,
			)

			conversationGroup.GET(
				"/:id/messages/sync",
				conversationMessageController.Sync,
			)
			conversationGroup.GET(
				"",
				conversationController.List,
			)
			conversationGroup.GET(
				"/:id/receipt-state",
				directReceiptController.ReceiptState,
			)
		}
	}

	return r, nil
}
