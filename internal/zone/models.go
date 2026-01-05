package zone

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// MessageType represents WebSocket message types
type MessageType string

const (
	// Client → Server
	MsgTypeConnect          MessageType = "CONNECT"
	MsgTypeResumeSession    MessageType = "RESUME_SESSION"
	MsgTypeHeartbeat        MessageType = "HEARTBEAT"
	MsgTypeEnterWorld       MessageType = "ENTER_WORLD"
	MsgTypeMoveIntent       MessageType = "MOVE_INTENT"
	MsgTypeInteractIntent   MessageType = "INTERACT_INTENT"
	MsgTypeCombatActionIntent MessageType = "COMBAT_ACTION_INTENT"
	MsgTypePurchaseItem     MessageType = "PURCHASE_ITEM"
	MsgTypeSellItem         MessageType = "SELL_ITEM"
	MsgTypeEquipItem        MessageType = "EQUIP_ITEM"
	MsgTypeUnequipItem      MessageType = "UNEQUIP_ITEM"

	// Server → Client
	MsgTypeConnectAck       MessageType = "CONNECT_ACK"
	MsgTypeHeartbeatAck     MessageType = "HEARTBEAT_ACK"
	MsgTypeWorldSnapshot    MessageType = "WORLD_SNAPSHOT"
	MsgTypePresenceUpdate   MessageType = "PRESENCE_UPDATE"
	MsgTypeStateDiff        MessageType = "STATE_DIFF"
	MsgTypePositionCorrection MessageType = "POSITION_CORRECTION"
	MsgTypeInventoryUpdate  MessageType = "INVENTORY_UPDATE"
	MsgTypeCurrencyUpdate   MessageType = "CURRENCY_UPDATE"
	MsgTypeError            MessageType = "ERROR"
	MsgTypeKick             MessageType = "KICK"
)

// Message represents a WebSocket message
type Message struct {
	MsgType   MessageType     `json:"msg_type"`
	MsgID     string          `json:"msg_id"`
	RefMsgID  string          `json:"ref_msg_id,omitempty"`
	Timestamp int64           `json:"timestamp"`
	Success   bool            `json:"success,omitempty"`
	Payload   json.RawMessage `json:"payload,omitempty"`
	Error     *ErrorPayload   `json:"error,omitempty"`
}

// ErrorPayload represents error information
type ErrorPayload struct {
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
}

// ConnectPayload represents CONNECT message payload
type ConnectPayload struct {
	ConnectionToken string `json:"connection_token"`
	CharacterID     string `json:"character_id"`
	ClientVersion   string `json:"client_version"`
	Platform        string `json:"platform"`
}

// ConnectAckPayload represents CONNECT_ACK message payload
type ConnectAckPayload struct {
	SessionID string       `json:"session_id"`
	Character CharacterInfo `json:"character"`
	Zone      ZoneInfo     `json:"zone"`
	ServerTime int64       `json:"server_time"`
	TickRateMs int         `json:"tick_rate_ms"`
	HeartbeatIntervalMs int `json:"heartbeat_interval_ms"`
}

// CharacterInfo represents basic character info
type CharacterInfo struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	House string `json:"house"`
	Grade int    `json:"grade"`
}

// ZoneInfo represents zone information
type ZoneInfo struct {
	ID    string `json:"id"`
	Shard string `json:"shard"`
}

// EnterWorldPayload represents ENTER_WORLD message payload
type EnterWorldPayload struct {
	ZoneID            string   `json:"zone_id"`
	RequestedPosition *Vector3 `json:"requested_position,omitempty"`
}

// WorldSnapshotPayload represents WORLD_SNAPSHOT message payload
type WorldSnapshotPayload struct {
	Zone          ZoneInfo     `json:"zone"`
	Self          PlayerState  `json:"self"`
	NearbyPlayers []PlayerInfo `json:"nearby_players"`
}

// PlayerState represents detailed player state
type PlayerState struct {
	CharacterID string            `json:"character_id"`
	Position    Vector3           `json:"position"`
	Rotation    Rotation          `json:"rotation"`
	Stats       Stats             `json:"stats"`
	Flags       []string          `json:"flags"`
	Cooldowns   map[string]int64  `json:"cooldowns"`
	Effects     []string          `json:"effects"`
}

// PlayerInfo represents nearby player information
type PlayerInfo struct {
	CharacterID       string            `json:"character_id"`
	Name              string            `json:"name"`
	House             string            `json:"house"`
	Grade             int               `json:"grade"`
	Position          Vector3           `json:"position"`
	Rotation          Rotation          `json:"rotation"`
	VisibleEquipment  map[string]string `json:"visible_equipment"`
	Effects           []string          `json:"effects"`
}

// HeartbeatPayload represents HEARTBEAT message payload
type HeartbeatPayload struct {
	ClientTime int64 `json:"client_time"`
}

// HeartbeatAckPayload represents HEARTBEAT_ACK message payload
type HeartbeatAckPayload struct {
	ServerTime int64 `json:"server_time"`
	LatencyMs  int64 `json:"latency_ms"`
}

// PresenceUpdatePayload represents PRESENCE_UPDATE message payload
type PresenceUpdatePayload struct {
	EventType string     `json:"event_type"` // player_entered, player_left
	Player    PlayerInfo `json:"player"`
}

// Vector3 represents a 3D position
type Vector3 struct {
	X float32 `json:"x"`
	Y float32 `json:"y"`
	Z float32 `json:"z"`
}

// Rotation represents rotation
type Rotation struct {
	Yaw float32 `json:"yaw"`
}

// Stats represents character stats
type Stats struct {
	HealthCurrent int `json:"health_current"`
	HealthMax     int `json:"health_max"`
	ManaCurrent   int `json:"mana_current"`
	ManaMax       int `json:"mana_max"`
	StaminaCurrent int `json:"stamina_current"`
	StaminaMax    int `json:"stamina_max"`
}

// Player represents a connected player
type Player struct {
	SessionID   string
	CharacterID uuid.UUID
	Name        string
	House       string
	Grade       int
	Position    Vector3
	Rotation    Rotation
	Stats       Stats
	ConnectedAt time.Time
	LastActivity time.Time
}

// PurchaseItemPayload represents PURCHASE_ITEM message payload
type PurchaseItemPayload struct {
	ItemDefID string `json:"item_def_id"`
	Quantity  int    `json:"quantity"`
	VendorID  string `json:"vendor_id"`
}

// SellItemPayload represents SELL_ITEM message payload
type SellItemPayload struct {
	InventoryItemID string `json:"inventory_item_id"`
	Quantity        int    `json:"quantity"`
}

// EquipItemPayload represents EQUIP_ITEM message payload
type EquipItemPayload struct {
	InventoryItemID string `json:"inventory_item_id"`
	Slot            string `json:"slot"`
}

// UnequipItemPayload represents UNEQUIP_ITEM message payload
type UnequipItemPayload struct {
	Slot string `json:"slot"`
}

// InventoryUpdatePayload represents INVENTORY_UPDATE message payload
type InventoryUpdatePayload struct {
	Action   string      `json:"action"` // added, removed, updated
	ItemDefID string     `json:"item_def_id"`
	Quantity int         `json:"quantity"`
	Data     interface{} `json:"data,omitempty"`
}

// CurrencyUpdatePayload represents CURRENCY_UPDATE message payload
type CurrencyUpdatePayload struct {
	Currencies map[string]int64 `json:"currencies"`
}
