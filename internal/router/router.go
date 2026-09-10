package router

import (
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

func InitRouter(broker *realtime.RedisBroker, wsManager *ws.Manager) *gin.Engine {
	r := gin.New()
	r.Use(gin.Logger())
	r.Use(middleware.ErrorMiddleware())
	r.GET("/ping", func(ctx *gin.Context) {
		response.Success(ctx, gin.H{
			"msg": "pong",
		})
	})

	messageRepo :=
		repository.NewMessageRepository(
			database.DB,
		)

	userRepo := repository.NewUserRepository()
	friendRepo := repository.NewFriendREpository()
	teamRepo := repository.NewTeamRepository()
	channelRepo := repository.NewChannelRepository()
	legacyMessageRepo := repository.NewChannelMessageRepository()
	readRepo := repository.NewChannelReadRepository()
	conversation := repository.NewConversationRepository()

	conversationService := service.NewConversationService(conversation, friendRepo)

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
	realtimeService := service.NewRealtimeService(
		channelRepo,
		teamRepo,
	)
	messageService := service.NewChannelMessageService(
		messageRepo,
		realtimeService,
		conversationService,
	)
	readService := service.NewChannelReadService(
		readRepo,
		legacyMessageRepo,
		realtimeService,
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
		broker,
	)

	wsController := controller.NewWSController(
		wsManager,
		wsMessageHandler,
	)

	messageController := controller.NewChannelMessageController(
		messageService,
	)
	readController := controller.NewChannelReadController(
		readService,
	)
	conversationController := controller.NewConversationController(conversationService)

	api := r.Group("/api")

	user := api.Group("/users")
	{
		user.POST("/register", userController.Register)
		user.POST("/login", userController.Login)

	}

	auth := api.Group("")
	auth.Use(middleware.AuthMiddleware())
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
		}
	}

	return r
}
