package main

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/ForgeOfBit/Nihan/server/internal/config"
	"github.com/ForgeOfBit/Nihan/server/internal/database"
	"github.com/ForgeOfBit/Nihan/server/internal/handlers"
	"github.com/ForgeOfBit/Nihan/server/internal/middleware"
	"github.com/ForgeOfBit/Nihan/server/internal/models"
	"github.com/ForgeOfBit/Nihan/server/internal/services"
	"github.com/ForgeOfBit/Nihan/server/internal/ws"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	// Load configuration.
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Set Gin mode.
	gin.SetMode(cfg.Server.GinMode)

	// Connect to PostgreSQL.
	ctx := context.Background()
	pool, err := database.NewPostgresPool(ctx, cfg.Database.URL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer pool.Close()
	log.Println("Connected to PostgreSQL")

	// Run database migrations.
	if err := database.RunMigrations(ctx, pool, "migrations"); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}
	log.Println("Database migrations complete")

	// Initialize services.
	discService := services.NewDiscriminatorService(pool)
	authService := services.NewAuthService(pool, &cfg.JWT, discService)
	userService := services.NewUserService(pool)
	messageService := services.NewMessageService(pool)

	// Initialize WebSocket hub.
	hub := ws.NewHub()
	go hub.Run()

	// Initialize handlers.
	authHandler := handlers.NewAuthHandler(authService)
	userHandler := handlers.NewUserHandler(userService, discService)
	messageHandler := handlers.NewMessageHandler(messageService)
	keyHandler := handlers.NewKeyHandler(pool)
	friendHandler := handlers.NewFriendHandler(pool)

	// Set up Gin router.
	router := gin.Default()

	// Global middleware.
	router.Use(middleware.CORSMiddleware(cfg.CORS.AllowedOrigins))
	router.Use(middleware.RateLimitMiddleware(cfg.RateLimit.RPS, cfg.RateLimit.Burst))

	// Health check endpoint.
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "ok",
			"timestamp": time.Now().UTC(),
		})
	})

	// ── API Routes ──────────────────────────────────────────────────────
	api := router.Group("/api")
	{
		// Auth routes (public).
		auth := api.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
			auth.POST("/refresh", authHandler.Refresh)
			auth.GET("/me", middleware.AuthMiddleware(cfg.JWT.AccessSecret), authHandler.Me)
		}

		// User routes (authenticated).
		users := api.Group("/users")
		users.Use(middleware.AuthMiddleware(cfg.JWT.AccessSecret))
		{
			users.GET("/search", userHandler.Search)
			users.GET("/:id", userHandler.GetProfile)
			users.PATCH("/me", userHandler.UpdateProfile)
			users.PATCH("/me/discriminator", userHandler.ChangeDiscriminator)
		}

		// Key bundle routes (authenticated).
		keys := api.Group("/keys")
		keys.Use(middleware.AuthMiddleware(cfg.JWT.AccessSecret))
		{
			keys.POST("/bundle", keyHandler.UploadBundle)
			keys.GET("/bundle/:userId", keyHandler.GetBundle)
		}

		// Friend routes (authenticated).
		friends := api.Group("/friends")
		friends.Use(middleware.AuthMiddleware(cfg.JWT.AccessSecret))
		{
			friends.POST("/request", friendHandler.SendRequest)
			friends.POST("/accept/:id", friendHandler.AcceptRequest)
			friends.GET("", friendHandler.ListFriends)
		}

		// Message routes (authenticated).
		messages := api.Group("/messages")
		messages.Use(middleware.AuthMiddleware(cfg.JWT.AccessSecret))
		{
			messages.GET("/:userId", messageHandler.GetHistory)
		}
	}

	// ── WebSocket Route ─────────────────────────────────────────────────
	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			origin := r.Header.Get("Origin")
			if origin == "" {
				return true
			}
			for _, allowed := range cfg.CORS.AllowedOrigins {
				if strings.EqualFold(origin, strings.TrimSpace(allowed)) {
					return true
				}
			}
			return false
		},
	}

	router.GET("/ws", func(c *gin.Context) {
		handleWebSocket(c, &upgrader, hub, pool, messageService, cfg.JWT.AccessSecret)
	})

	// ── Start HTTP Server ───────────────────────────────────────────────
	srv := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("Nihan server starting on port %s", cfg.Server.Port)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Server error: %v", err)
		}
	}()

	// ── Graceful Shutdown ───────────────────────────────────────────────
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// Set all online users to offline.
	_, _ = pool.Exec(ctx, "UPDATE users SET status = 'offline' WHERE status = 'online'")

	hub.Stop()

	shutdownCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited gracefully")
}

// handleWebSocket authenticates the WebSocket connection via a JWT token
// passed as a query parameter, upgrades the connection, and registers
// the client with the Hub.
func handleWebSocket(
	c *gin.Context,
	upgrader *websocket.Upgrader,
	hub *ws.Hub,
	pool *pgxpool.Pool,
	msgService *services.MessageService,
	accessSecret string,
) {
	// Authenticate via query parameter token for WebSocket connections.
	tokenString := c.Query("token")
	if tokenString == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "token query parameter is required"})
		return
	}

	claims := &middleware.JWTClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(accessSecret), nil
	})
	if err != nil || !token.Valid {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
		return
	}

	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user ID in token"})
		return
	}

	// Upgrade HTTP to WebSocket.
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed for user %s: %v", userID, err)
		return
	}

	// Set user status to online.
	_, _ = pool.Exec(context.Background(), "UPDATE users SET status = 'online' WHERE id = $1", userID)

	// Create the client with a message handler.
	handler := newWSMessageHandler(hub, msgService, pool)
	client := ws.NewClient(hub, conn, userID, handler)
	hub.Register(client)

	go client.WritePump()
	go client.ReadPump()
}

// newWSMessageHandler returns a MessageHandler function that processes
// incoming WebSocket events.
func newWSMessageHandler(hub *ws.Hub, msgService *services.MessageService, pool *pgxpool.Pool) ws.MessageHandler {
	return func(senderID uuid.UUID, event ws.Event) {
		ctx := context.Background()

		switch event.Type {
		case "message.send":
			handleMessageSend(ctx, hub, msgService, senderID, event)

		case "message.read":
			handleMessageRead(ctx, hub, msgService, senderID, event)

		case "typing.start", "typing.stop":
			handleTyping(hub, senderID, event)

		case "user.status":
			handleUserStatus(ctx, pool, senderID, event)

		default:
			log.Printf("ws: unknown event type '%s' from user %s", event.Type, senderID)
		}
	}
}

// handleMessageSend stores an encrypted message and forwards it to the receiver.
func handleMessageSend(ctx context.Context, hub *ws.Hub, msgService *services.MessageService, senderID uuid.UUID, event ws.Event) {
	var req services.WSSendMessage
	if err := json.Unmarshal(event.Payload, &req); err != nil {
		log.Printf("ws: invalid message.send payload from %s: %v", senderID, err)
		return
	}

	msg, err := msgService.Send(ctx, senderID, models.SendMessageRequest{
		ReceiverID:   req.ReceiverID,
		Ciphertext:   req.Ciphertext,
		Nonce:        req.Nonce,
		EphemeralKey: req.EphemeralKey,
		MessageType:  req.MessageType,
	})
	if err != nil {
		log.Printf("ws: failed to save message from %s: %v", senderID, err)
		return
	}

	payload, _ := json.Marshal(msg)

	// Forward to receiver if online.
	hub.SendToUser(req.ReceiverID, ws.Event{
		Type:    "message.receive",
		Payload: payload,
	})

	// Acknowledge to sender.
	hub.SendToUser(senderID, ws.Event{
		Type:    "message.sent",
		Payload: payload,
	})
}

// handleMessageRead marks messages as read and notifies the original sender.
func handleMessageRead(ctx context.Context, hub *ws.Hub, msgService *services.MessageService, readerID uuid.UUID, event ws.Event) {
	var req struct {
		SenderID uuid.UUID `json:"sender_id"`
	}
	if err := json.Unmarshal(event.Payload, &req); err != nil {
		log.Printf("ws: invalid message.read payload from %s: %v", readerID, err)
		return
	}

	if err := msgService.MarkAsRead(ctx, readerID, req.SenderID); err != nil {
		log.Printf("ws: failed to mark messages as read: %v", err)
		return
	}

	// Notify the original sender.
	payload, _ := json.Marshal(map[string]interface{}{"reader_id": readerID})
	hub.SendToUser(req.SenderID, ws.Event{
		Type:    "message.read",
		Payload: payload,
	})
}

// handleTyping forwards typing indicators to the target user.
func handleTyping(hub *ws.Hub, senderID uuid.UUID, event ws.Event) {
	var req struct {
		ReceiverID uuid.UUID `json:"receiver_id"`
	}
	if err := json.Unmarshal(event.Payload, &req); err != nil {
		return
	}

	payload, _ := json.Marshal(map[string]interface{}{"user_id": senderID})
	hub.SendToUser(req.ReceiverID, ws.Event{
		Type:    event.Type,
		Payload: payload,
	})
}

// handleUserStatus updates the user's status in the database.
func handleUserStatus(ctx context.Context, pool *pgxpool.Pool, userID uuid.UUID, event ws.Event) {
	var req struct {
		Status string `json:"status"`
	}
	if err := json.Unmarshal(event.Payload, &req); err != nil {
		return
	}

	validStatuses := map[string]bool{
		"online": true, "offline": true, "idle": true, "dnd": true, "invisible": true,
	}
	if !validStatuses[req.Status] {
		return
	}

	_, err := pool.Exec(ctx, "UPDATE users SET status = $1 WHERE id = $2", req.Status, userID)
	if err != nil {
		log.Printf("ws: failed to update status for user %s: %v", userID, err)
	}
}
