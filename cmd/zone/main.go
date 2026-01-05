package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/rs/zerolog/log"

	"github.com/hp-mmo/backend/internal/character"
	"github.com/hp-mmo/backend/internal/pkg/config"
	"github.com/hp-mmo/backend/internal/pkg/db"
	"github.com/hp-mmo/backend/internal/pkg/logger"
	"github.com/hp-mmo/backend/internal/zone"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Configure properly in production
	},
}

// ZoneServer handles WebSocket connections for a zone
type ZoneServer struct {
	zoneID      string
	shard       string
	players     map[string]*zone.Player
	playersMu   sync.RWMutex
	charRepo    *character.Repository
	broadcast   chan zone.Message
	tickRate    time.Duration
}

// NewZoneServer creates a new zone server
func NewZoneServer(zoneID, shard string, charRepo *character.Repository, tickRateMs int) *ZoneServer {
	return &ZoneServer{
		zoneID:    zoneID,
		shard:     shard,
		players:   make(map[string]*zone.Player),
		charRepo:  charRepo,
		broadcast: make(chan zone.Message, 100),
		tickRate:  time.Duration(tickRateMs) * time.Millisecond,
	}
}

// HandleWebSocket handles WebSocket connections
func (s *ZoneServer) HandleWebSocket(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Error().Err(err).Msg("Failed to upgrade connection")
		return
	}
	defer conn.Close()

	sessionID := uuid.New().String()
	log.Info().Str("session_id", sessionID).Msg("New WebSocket connection")

	// Read messages
	for {
		_, data, err := conn.ReadMessage()
		if err != nil {
			log.Error().Err(err).Str("session_id", sessionID).Msg("Read error")
			s.handleDisconnect(sessionID)
			break
		}

		var msg zone.Message
		if err := json.Unmarshal(data, &msg); err != nil {
			log.Error().Err(err).Msg("Failed to unmarshal message")
			s.sendError(conn, "", "INVALID_MESSAGE", "Invalid message format")
			continue
		}

		s.handleMessage(conn, sessionID, &msg)
	}
}

// handleMessage handles incoming WebSocket messages
func (s *ZoneServer) handleMessage(conn *websocket.Conn, sessionID string, msg *zone.Message) {
	switch msg.MsgType {
	case zone.MsgTypeConnect:
		s.handleConnect(conn, sessionID, msg)
	case zone.MsgTypeHeartbeat:
		s.handleHeartbeat(conn, sessionID, msg)
	case zone.MsgTypeEnterWorld:
		s.handleEnterWorld(conn, sessionID, msg)
	default:
		log.Warn().Str("msg_type", string(msg.MsgType)).Msg("Unknown message type")
		s.sendError(conn, msg.MsgID, "UNKNOWN_MSG_TYPE", "Unknown message type")
	}
}

// handleConnect handles CONNECT message
func (s *ZoneServer) handleConnect(conn *websocket.Conn, sessionID string, msg *zone.Message) {
	var payload zone.ConnectPayload
	if err := json.Unmarshal(msg.Payload, &payload); err != nil {
		s.sendError(conn, msg.MsgID, "INVALID_PAYLOAD", "Invalid connect payload")
		return
	}

	// Parse character ID
	characterID, err := uuid.Parse(payload.CharacterID)
	if err != nil {
		s.sendError(conn, msg.MsgID, "INVALID_CHARACTER_ID", "Invalid character ID")
		return
	}

	// Load character from database
	char, err := s.charRepo.GetCharacterByID(context.Background(), characterID)
	if err != nil {
		s.sendError(conn, msg.MsgID, "CHARACTER_NOT_FOUND", "Character not found")
		return
	}

	// Load stats
	stats, err := s.charRepo.GetLatestStats(context.Background(), characterID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to load stats")
		s.sendError(conn, msg.MsgID, "INTERNAL_ERROR", "Failed to load character stats")
		return
	}

	// Create player
	player := &zone.Player{
		SessionID:   sessionID,
		CharacterID: characterID,
		Name:        char.Name,
		House:       char.House,
		Grade:       char.Grade,
		Position: zone.Vector3{
			X: char.PositionX,
			Y: char.PositionY,
			Z: char.PositionZ,
		},
		Rotation: zone.Rotation{
			Yaw: char.RotationYaw,
		},
		Stats: zone.Stats{
			HealthCurrent:  stats.HealthCurrent,
			HealthMax:      stats.HealthMax,
			ManaCurrent:    stats.ManaCurrent,
			ManaMax:        stats.ManaMax,
			StaminaCurrent: stats.StaminaCurrent,
			StaminaMax:     stats.StaminaMax,
		},
		ConnectedAt:  time.Now(),
		LastActivity: time.Now(),
	}

	// Add player to server
	s.playersMu.Lock()
	s.players[sessionID] = player
	s.playersMu.Unlock()

	// Send CONNECT_ACK
	ackPayload := zone.ConnectAckPayload{
		SessionID: sessionID,
		Character: zone.CharacterInfo{
			ID:    characterID.String(),
			Name:  char.Name,
			House: char.House,
			Grade: char.Grade,
		},
		Zone: zone.ZoneInfo{
			ID:    s.zoneID,
			Shard: s.shard,
		},
		ServerTime:          time.Now().UnixMilli(),
		TickRateMs:          50,
		HeartbeatIntervalMs: 10000,
	}

	s.sendMessage(conn, zone.MsgTypeConnectAck, msg.MsgID, true, ackPayload)
	log.Info().Str("character", char.Name).Str("session_id", sessionID).Msg("Player connected")
}

// handleHeartbeat handles HEARTBEAT message
func (s *ZoneServer) handleHeartbeat(conn *websocket.Conn, sessionID string, msg *zone.Message) {
	var payload zone.HeartbeatPayload
	if err := json.Unmarshal(msg.Payload, &payload); err != nil {
		return
	}

	serverTime := time.Now().UnixMilli()
	latency := serverTime - payload.ClientTime

	ackPayload := zone.HeartbeatAckPayload{
		ServerTime: serverTime,
		LatencyMs:  latency,
	}

	s.sendMessage(conn, zone.MsgTypeHeartbeatAck, msg.MsgID, true, ackPayload)

	// Update activity
	s.playersMu.Lock()
	if player, exists := s.players[sessionID]; exists {
		player.LastActivity = time.Now()
	}
	s.playersMu.Unlock()
}

// handleEnterWorld handles ENTER_WORLD message
func (s *ZoneServer) handleEnterWorld(conn *websocket.Conn, sessionID string, msg *zone.Message) {
	s.playersMu.RLock()
	player, exists := s.players[sessionID]
	s.playersMu.RUnlock()

	if !exists {
		s.sendError(conn, msg.MsgID, "UNAUTHORIZED", "Player not connected")
		return
	}

	// Build nearby players list
	nearbyPlayers := []zone.PlayerInfo{}
	s.playersMu.RLock()
	for sid, p := range s.players {
		if sid != sessionID {
			nearbyPlayers = append(nearbyPlayers, zone.PlayerInfo{
				CharacterID: p.CharacterID.String(),
				Name:        p.Name,
				House:       p.House,
				Grade:       p.Grade,
				Position:    p.Position,
				Rotation:    p.Rotation,
				VisibleEquipment: map[string]string{},
				Effects:     []string{},
			})
		}
	}
	s.playersMu.RUnlock()

	// Send WORLD_SNAPSHOT
	snapshotPayload := zone.WorldSnapshotPayload{
		Zone: zone.ZoneInfo{
			ID:    s.zoneID,
			Shard: s.shard,
		},
		Self: zone.PlayerState{
			CharacterID: player.CharacterID.String(),
			Position:    player.Position,
			Rotation:    player.Rotation,
			Stats:       player.Stats,
			Flags:       []string{},
			Cooldowns:   map[string]int64{},
			Effects:     []string{},
		},
		NearbyPlayers: nearbyPlayers,
	}

	s.sendMessage(conn, zone.MsgTypeWorldSnapshot, msg.MsgID, true, snapshotPayload)

	// Broadcast presence to other players
	presencePayload := zone.PresenceUpdatePayload{
		EventType: "player_entered",
		Player: zone.PlayerInfo{
			CharacterID: player.CharacterID.String(),
			Name:        player.Name,
			House:       player.House,
			Grade:       player.Grade,
			Position:    player.Position,
			Rotation:    player.Rotation,
			VisibleEquipment: map[string]string{},
			Effects:     []string{},
		},
	}

	// TODO: Broadcast to all other players (would need connection tracking)
	_ = presencePayload

	log.Info().Str("character", player.Name).Msg("Player entered world")
}

// handleDisconnect handles player disconnection
func (s *ZoneServer) handleDisconnect(sessionID string) {
	s.playersMu.Lock()
	player, exists := s.players[sessionID]
	if exists {
		delete(s.players, sessionID)
		log.Info().Str("character", player.Name).Msg("Player disconnected")
	}
	s.playersMu.Unlock()
}

// sendMessage sends a message to the client
func (s *ZoneServer) sendMessage(conn *websocket.Conn, msgType zone.MessageType, refMsgID string, success bool, payload interface{}) {
	payloadBytes, _ := json.Marshal(payload)

	msg := zone.Message{
		MsgType:   msgType,
		MsgID:     uuid.New().String(),
		RefMsgID:  refMsgID,
		Timestamp: time.Now().UnixMilli(),
		Success:   success,
		Payload:   payloadBytes,
	}

	data, _ := json.Marshal(msg)
	conn.WriteMessage(websocket.TextMessage, data)
}

// sendError sends an error message
func (s *ZoneServer) sendError(conn *websocket.Conn, refMsgID, code, message string) {
	msg := zone.Message{
		MsgType:   zone.MsgTypeError,
		MsgID:     uuid.New().String(),
		RefMsgID:  refMsgID,
		Timestamp: time.Now().UnixMilli(),
		Success:   false,
		Error: &zone.ErrorPayload{
			Code:    code,
			Message: message,
		},
	}

	data, _ := json.Marshal(msg)
	conn.WriteMessage(websocket.TextMessage, data)
}

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load configuration")
	}

	// Initialize logger
	logger.Init(cfg.Logging.Level, cfg.Logging.Format)
	log.Info().Str("zone_id", cfg.Zone.ID).Str("shard", cfg.Zone.Shard).Msg("Starting HP MMO Zone Server")

	// Connect to PostgreSQL
	database, err := db.Connect(
		cfg.GetDBConnectionString(),
		cfg.DB.MaxOpenConns,
		cfg.DB.MaxIdleConns,
	)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to database")
	}
	defer db.Close(database)

	// Initialize repository
	charRepo := character.NewRepository(database)

	// Create zone server
	zoneServer := NewZoneServer(cfg.Zone.ID, cfg.Zone.Shard, charRepo, cfg.Zone.TickRateMs)

	// Setup Gin
	if cfg.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	router.Use(gin.Recovery())

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "healthy",
			"service": "zone",
			"zone_id": cfg.Zone.ID,
			"shard":   cfg.Zone.Shard,
			"players": len(zoneServer.players),
		})
	})

	// WebSocket endpoint
	router.GET("/ws", zoneServer.HandleWebSocket)

	// Create HTTP server
	srv := &http.Server{
		Addr:           fmt.Sprintf(":%d", cfg.Server.ZonePort),
		Handler:        router,
		ReadTimeout:    30 * time.Second,
		WriteTimeout:   30 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	// Start server in goroutine
	go func() {
		log.Info().Int("port", cfg.Server.ZonePort).Msg("Zone server listening")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("Failed to start server")
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info().Msg("Shutting down zone server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal().Err(err).Msg("Server forced to shutdown")
	}

	log.Info().Msg("Zone server exited successfully")
}
