// Package network helps with multiplayer support
package network

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"net"
	"sync"
	"time"
)

// --- Network Types ---

// messageType defines the type of network message being sent
type messageType int

const (
	// msgConnect is sent when a player connects to a game
	msgConnect messageType = iota
	// msgDisconnect is sent when a player disconnects from a game
	msgDisconnect
	// msgGameState is sent to sync game state between players
	msgGameState
	// msgPlayerInput is sent when a player performs an input action
	msgPlayerInput
	// msgPing is sent to check connection status
	msgPing
	// msgPong is sent in response to a ping
	msgPong
)

// networkMessage represents a message sent over the network
type networkMessage struct {
	Type     messageType `json:"type"`
	PlayerID string      `json:"player_id"`
	Data     []byte      `json:"data"`
}

// networkRole defines whether this instance is a server or client
type networkRole int

const (
	// RoleServer indicates this instance is hosting the game
	RoleServer networkRole = iota
	// RoleClient indicates this instance is connecting to a host
	RoleClient
)

// Config holds configuration for network functionality
type Config struct {
	Role       networkRole // Whether this instance is a server or client
	Address    string      // Address to connect to (for client) or listen on (for server)
	Port       int         // Port to use for connection
	PlayerID   string      // Unique identifier for this player
	BufferSize int         // Size of message buffer
	GameName   string      // Name of the game (for display purposes)
}

// defaultNetworkConfig returns a default network configuration
func defaultNetworkConfig() *Config {
	return &Config{
		Role:       RoleServer,
		Address:    "localhost",
		Port:       8080,
		PlayerID:   fmt.Sprintf("player-%d", time.Now().UnixNano()%10000),
		BufferSize: 100,
		GameName:   "PIGO8 Game",
	}
}

// --- Network Manager ---

// Manager handles all networking functionality
type Manager struct {
	config *Config
	// UDP specific fields
	udpConn    *net.UDPConn            // UDP connection for both server and client
	serverAddr *net.UDPAddr            // Server address (used by clients)
	clients    map[string]*net.UDPAddr // Map of connected clients by player ID
	lastHeard  map[string]time.Time    // Last time we heard from each client
	// Message handling
	incomingMsgs chan networkMessage
	outgoingMsgs chan networkMessage
	// Callbacks
	onConnect     func(playerID string)
	onDisconnect  func(playerID string)
	onGameState   func(playerID string, data []byte)
	onPlayerInput func(playerID string, data []byte)
	// State
	mutex             sync.Mutex
	connectionLost    bool
	waitingForPlayers bool
	networkError      string
	done              chan struct{}
	stopOnce          sync.Once
	workers           sync.WaitGroup
	// Heartbeat
	heartbeatTicker   *time.Ticker
	heartbeatInterval time.Duration
}

var (
	// Global network manager instance
	networkManager *Manager
	networkMutex   sync.Mutex
)

// --- Network Initialization ---

// InitNetwork initializes the networking system with the given configuration
func InitNetwork(config *Config) error {
	// Save existing callbacks if we have a network manager
	var onConnect func(string)
	var onDisconnect func(string)
	var onGameState func(string, []byte)
	var onPlayerInput func(string, []byte)
	var previousManager *Manager

	networkMutex.Lock()
	if networkManager != nil {
		// Save the existing callbacks
		onConnect = networkManager.onConnect
		onDisconnect = networkManager.onDisconnect
		onGameState = networkManager.onGameState
		onPlayerInput = networkManager.onPlayerInput

		// Log the callback status
		log.Printf("Preserving existing callbacks: onConnect=%v, onDisconnect=%v, onGameState=%v, onPlayerInput=%v",
			onConnect != nil, onDisconnect != nil, onGameState != nil, onPlayerInput != nil)

		previousManager = networkManager
		networkManager = nil
	}
	networkMutex.Unlock()

	if previousManager != nil {
		previousManager.shutdownAndWait()
	}

	// Create a new network manager
	manager := &Manager{
		config:            config,
		incomingMsgs:      make(chan networkMessage, config.BufferSize),
		outgoingMsgs:      make(chan networkMessage, config.BufferSize),
		clients:           make(map[string]*net.UDPAddr),
		lastHeard:         make(map[string]time.Time),
		waitingForPlayers: config.Role == RoleServer, // Server starts waiting for players
		heartbeatInterval: 2 * time.Second,           // Send heartbeat every 2 seconds
		done:              make(chan struct{}),
		// Restore the callbacks
		onConnect:     onConnect,
		onDisconnect:  onDisconnect,
		onGameState:   onGameState,
		onPlayerInput: onPlayerInput,
	}

	// Start network processing in background
	manager.startWorker(manager.processMessages)

	// Start server or client based on role
	var err error
	if config.Role == RoleServer {
		err = manager.startServer()
	} else {
		err = manager.connectToServer()
	}
	if err != nil {
		manager.shutdownAndWait()
		return err
	}

	networkMutex.Lock()
	networkManager = manager
	networkMutex.Unlock()
	return nil
}

// ShutdownNetwork closes all network connections and stops processing
func ShutdownNetwork() {
	networkMutex.Lock()
	manager := networkManager
	networkManager = nil
	networkMutex.Unlock()

	if manager != nil {
		manager.shutdownAndWait()
	}
}

func (nm *Manager) startWorker(fn func()) {
	nm.workers.Add(1)
	go func() {
		defer nm.workers.Done()
		fn()
	}()
}

func (nm *Manager) isStopped() bool {
	select {
	case <-nm.done:
		return true
	default:
		return false
	}
}

func (nm *Manager) shutdownAndWait() {
	nm.stopOnce.Do(func() {
		close(nm.done)

		if nm.heartbeatTicker != nil {
			nm.heartbeatTicker.Stop()
		}

		if nm.udpConn != nil {
			if err := nm.udpConn.Close(); err != nil && !errors.Is(err, net.ErrClosed) {
				log.Printf("Error closing UDP connection: %v", err)
			}
		}
	})

	nm.workers.Wait()

	nm.mutex.Lock()
	nm.clients = make(map[string]*net.UDPAddr)
	nm.lastHeard = make(map[string]time.Time)
	nm.mutex.Unlock()
}

// --- Server Functions ---

// startServer initializes a UDP server that listens for client messages
func (nm *Manager) startServer() error {
	// Parse the UDP address to listen on
	addr := net.JoinHostPort(nm.config.Address, fmt.Sprintf("%d", nm.config.Port))
	log.Printf("Starting UDP server on %s...", addr)

	// Resolve the UDP address
	udpAddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		nm.networkError = fmt.Sprintf("Failed to resolve UDP address: %v", err)
		log.Printf("Failed to resolve UDP address %s: %v", addr, err)
		return fmt.Errorf("failed to resolve UDP address: %v", err)
	}

	// Create a UDP listener
	udpConn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		nm.networkError = fmt.Sprintf("Server start failed: %v", err)
		log.Printf("Failed to start UDP server on %s: %v", addr, err)
		return fmt.Errorf("failed to start UDP server: %v", err)
	}

	nm.udpConn = udpConn
	log.Printf("UDP Server successfully started on %s (local IP: %s)", addr, getLocalIP())

	// Start the heartbeat ticker
	nm.heartbeatTicker = time.NewTicker(nm.heartbeatInterval)
	nm.startWorker(func() {
		for {
			select {
			case <-nm.done:
				return
			case <-nm.heartbeatTicker.C:
				nm.sendHeartbeats()
			}
		}
	})

	// Start receiving messages in background
	nm.startWorker(nm.receiveMessages)
	return nil
}

// sendHeartbeats sends heartbeat messages to all connected clients
func (nm *Manager) sendHeartbeats() {
	nm.mutex.Lock()
	defer nm.mutex.Unlock()

	// Skip if no clients
	if len(nm.clients) == 0 {
		return
	}

	// Create heartbeat message
	heartbeatMsg := networkMessage{
		Type:     msgPing,
		PlayerID: nm.config.PlayerID,
		Data:     []byte{},
	}

	// Encode the message
	data, err := json.Marshal(heartbeatMsg)
	if err != nil {
		log.Printf("Error encoding heartbeat message: %v", err)
		return
	}

	// Send to all clients
	for playerID, addr := range nm.clients {
		_, err := nm.udpConn.WriteToUDP(data, addr)
		if err != nil {
			log.Printf("Error sending heartbeat to %s: %v", playerID, err)
		}
	}
}

// receiveMessages handles incoming UDP messages
func (nm *Manager) receiveMessages() {
	log.Printf("Starting to receive UDP messages...")

	// Buffer for incoming messages
	buffer := make([]byte, 4096)

	for {
		// Read from UDP connection
		n, addr, err := nm.udpConn.ReadFromUDP(buffer)
		if err != nil {
			if nm.isStopped() || errors.Is(err, net.ErrClosed) {
				// Normal shutdown
				return
			}
			log.Printf("Error reading UDP message: %v", err)
			continue
		}

		// Process the message
		dataCopy := make([]byte, n)
		copy(dataCopy, buffer[:n])
		addrCopy := &net.UDPAddr{IP: append([]byte(nil), addr.IP...), Port: addr.Port, Zone: addr.Zone}
		nm.startWorker(func() {
			nm.handleUDPMessage(dataCopy, addrCopy)
		})
	}
}

// handleUDPMessage processes a UDP message from a client
func (nm *Manager) handleUDPMessage(data []byte, addr *net.UDPAddr) {
	// Check if we have valid data
	if len(data) == 0 {
		log.Printf("Received empty UDP message, ignoring")
		return
	}

	// Decode the message
	var msg networkMessage
	if err := json.Unmarshal(data, &msg); err != nil {
		log.Printf("Error decoding UDP message: %v", err)
		return
	}

	// Validate the message
	if msg.Type < msgConnect || msg.Type > msgPong {
		log.Printf("Received UDP message with invalid type: %v, ignoring", msg.Type)
		return
	}

	// Update the last heard time for this client
	if msg.PlayerID != "" {
		nm.mutex.Lock()
		nm.lastHeard[msg.PlayerID] = time.Now()

		// If this is a new client, add them to our clients map
		if _, exists := nm.clients[msg.PlayerID]; !exists && msg.Type == msgConnect {
			nm.clients[msg.PlayerID] = addr
			nm.waitingForPlayers = false
			log.Printf("New client connected: %s from %s", msg.PlayerID, addr.String())

			// Notify about the connection
			if nm.onConnect != nil {
				nm.onConnect(msg.PlayerID)
			}
		}
		nm.mutex.Unlock()
	}

	// Process the message based on its type
	log.Printf("Processing UDP message: type=%v, playerID=%s", msg.Type, msg.PlayerID)

	// IMPORTANT: Get direct references to the callbacks to avoid race conditions
	var onGameState func(string, []byte)
	var onPlayerInput func(string, []byte)
	var onConnect func(string)
	var onDisconnect func(string)

	// Lock to safely access the callbacks
	nm.mutex.Lock()
	onGameState = nm.onGameState
	onPlayerInput = nm.onPlayerInput
	onConnect = nm.onConnect
	onDisconnect = nm.onDisconnect
	nm.mutex.Unlock()

	// Log the callback status for debugging
	log.Printf("Callback status: onGameState=%v, onPlayerInput=%v, onConnect=%v, onDisconnect=%v",
		onGameState != nil, onPlayerInput != nil, onConnect != nil, onDisconnect != nil)

	switch msg.Type {
	case msgConnect:
		log.Printf("Received connect message from %s", msg.PlayerID)
		// Already handled connection above, but also call the callback
		if onConnect != nil {
			onConnect(msg.PlayerID)
		}
	case msgDisconnect:
		log.Printf("Received disconnect message from %s", msg.PlayerID)
		// Handle client disconnect
		nm.handleClientDisconnect(msg.PlayerID)
	case msgGameState:
		log.Printf("Received game state message from %s, data size: %d bytes", msg.PlayerID, len(msg.Data))
		// Forward game state to the appropriate handler
		if onGameState != nil {
			log.Printf("Calling game state handler with data size: %d bytes", len(msg.Data))
			onGameState(msg.PlayerID, msg.Data)
		} else {
			log.Printf("Warning: No game state handler registered")
		}
	case msgPlayerInput:
		log.Printf("Received player input message from %s, data size: %d bytes", msg.PlayerID, len(msg.Data))
		// Forward player input to the appropriate handler
		if onPlayerInput != nil {
			log.Printf("Calling player input handler with data size: %d bytes", len(msg.Data))
			onPlayerInput(msg.PlayerID, msg.Data)
		} else {
			log.Printf("Warning: No player input handler registered")
		}
	case msgPing:
		// Respond with a pong
		log.Printf("Received ping from %s, sending pong", msg.PlayerID)
		nm.sendPong(msg.PlayerID, addr)
	case msgPong:
		log.Printf("Received pong from %s", msg.PlayerID)
		// Just update the last heard time (already done above)
	default:
		log.Printf("Received unknown message type: %v", msg.Type)
	}

	// We'll still forward the message to the channel for other processing
	// but we won't rely on it for callbacks anymore
	select {
	case <-nm.done:
		return
	case nm.incomingMsgs <- msg:
		// Message sent successfully
	default:
		// Channel is full, log and continue
		log.Printf("Warning: incoming message channel is full, dropping message")
	}
}

// handleClientDisconnect handles a client disconnection
func (nm *Manager) handleClientDisconnect(playerID string) {
	nm.mutex.Lock()
	delete(nm.clients, playerID)
	delete(nm.lastHeard, playerID)
	if len(nm.clients) == 0 {
		nm.waitingForPlayers = true
	}
	nm.mutex.Unlock()

	// Notify about the disconnection
	if nm.onDisconnect != nil {
		nm.onDisconnect(playerID)
	}

	log.Printf("Client disconnected: %s", playerID)
}

// sendPong sends a pong response to a ping
func (nm *Manager) sendPong(playerID string, addr *net.UDPAddr) {
	// Create pong message
	pongMsg := networkMessage{
		Type:     msgPong,
		PlayerID: nm.config.PlayerID,
		Data:     []byte{},
	}

	// Encode the message
	data, err := json.Marshal(pongMsg)
	if err != nil {
		log.Printf("Error encoding pong message: %v", err)
		return
	}

	// Use different send methods depending on role
	if nm.config.Role == RoleServer {
		// Server uses WriteToUDP to send to specific client
		_, err = nm.udpConn.WriteToUDP(data, addr)
	} else {
		// Client uses Write to send to the pre-connected server
		_, err = nm.udpConn.Write(data)
	}

	if err != nil {
		log.Printf("Error sending pong to %s: %v", playerID, err)
	}
}

// --- Client Functions ---

// connectToServer connects to a game server using UDP
func (nm *Manager) connectToServer() error {
	// Resolve the server address
	serverAddr := net.JoinHostPort(nm.config.Address, fmt.Sprintf("%d", nm.config.Port))
	log.Printf("Attempting to connect to UDP server at %s...", serverAddr)

	// Resolve the UDP address
	udpAddr, err := net.ResolveUDPAddr("udp", serverAddr)
	if err != nil {
		nm.mutex.Lock()
		nm.connectionLost = true
		nm.networkError = fmt.Sprintf("Failed to resolve UDP address: %v", err)
		nm.mutex.Unlock()

		log.Printf("Failed to resolve UDP address %s: %v", serverAddr, err)
		return fmt.Errorf("failed to resolve UDP address: %v", err)
	}

	// Create a local address to bind to (any available port)
	localAddr, err := net.ResolveUDPAddr("udp", "0.0.0.0:0")
	if err != nil {
		nm.mutex.Lock()
		nm.connectionLost = true
		nm.networkError = fmt.Sprintf("Failed to resolve local UDP address: %v", err)
		nm.mutex.Unlock()

		log.Printf("Failed to resolve local UDP address: %v", err)
		return fmt.Errorf("failed to resolve local UDP address: %v", err)
	}

	// Create a UDP connection
	udpConn, err := net.DialUDP("udp", localAddr, udpAddr)
	if err != nil {
		nm.mutex.Lock()
		nm.connectionLost = true
		nm.networkError = fmt.Sprintf("Connection failed: %v", err)
		nm.mutex.Unlock()

		log.Printf("Failed to connect to UDP server at %s: %v", serverAddr, err)
		return fmt.Errorf("failed to connect to UDP server: %v", err)
	}

	nm.udpConn = udpConn
	nm.serverAddr = udpAddr
	log.Printf("Successfully connected to UDP server at %s", serverAddr)

	// Send connect message
	connectMsg := networkMessage{
		Type:     msgConnect,
		PlayerID: nm.config.PlayerID,
		Data:     []byte{},
	}

	// Encode the message
	data, err := json.Marshal(connectMsg)
	if err != nil {
		if closeErr := nm.udpConn.Close(); closeErr != nil {
			log.Printf("Error closing UDP connection after encoding error: %v", closeErr)
		}
		return fmt.Errorf("failed to encode connect message: %v", err)
	}

	// Send the connect message
	_, err = nm.udpConn.Write(data)
	if err != nil {
		if closeErr := nm.udpConn.Close(); closeErr != nil {
			log.Printf("Error closing UDP connection after send error: %v", closeErr)
		}
		return fmt.Errorf("failed to send connect message: %v", err)
	}

	// Start receiving messages
	nm.startWorker(nm.receiveMessages)
	return nil
}

// Note: The receiveMessages function is already defined above
// This is a duplicate that was part of the old TCP implementation
// The UDP version is already implemented at line 270

// --- Message Processing ---

// processMessages handles all incoming and outgoing messages
func (nm *Manager) processMessages() {
	for {
		select {
		case <-nm.done:
			return
		case msg := <-nm.incomingMsgs:
			nm.handleIncomingMessage(msg)
		case msg := <-nm.outgoingMsgs:
			nm.sendMessage(msg)
		}
	}
}

// handleIncomingMessage processes an incoming network message
func (nm *Manager) handleIncomingMessage(msg networkMessage) {
	switch msg.Type {
	case msgGameState:
		if nm.onGameState != nil {
			nm.onGameState(msg.PlayerID, msg.Data)
		}
	case msgPlayerInput:
		if nm.onPlayerInput != nil {
			nm.onPlayerInput(msg.PlayerID, msg.Data)
		}
	case msgPing:
		// Respond with pong
		select {
		case <-nm.done:
			return
		case nm.outgoingMsgs <- networkMessage{
			Type:     msgPong,
			PlayerID: nm.config.PlayerID,
			Data:     msg.Data, // Echo the timestamp
		}:
		}
	}
}

// sendMessage sends a message to the appropriate destination using UDP
func (nm *Manager) sendMessage(msg networkMessage) {
	// Log the message being sent
	log.Printf("Sending message: type=%v, playerID=%s, dataSize=%d", msg.Type, msg.PlayerID, len(msg.Data))

	// First encode the message to JSON
	data, err := json.Marshal(msg)
	if err != nil {
		log.Printf("Error encoding message: %v", err)
		return
	}

	if nm.config.Role == RoleServer {
		// Server sends to specific client or broadcasts
		if msg.PlayerID == "all" {
			// Broadcast to all clients
			nm.mutex.Lock()
			clientCount := len(nm.clients)
			if clientCount == 0 {
				log.Printf("No clients to broadcast to")
			} else {
				log.Printf("Broadcasting to %d clients", clientCount)
			}

			for playerID, addr := range nm.clients {
				_, err := nm.udpConn.WriteToUDP(data, addr)
				if err != nil {
					log.Printf("Error sending message to client %s: %v", playerID, err)
				} else {
					log.Printf("Successfully sent message to client %s", playerID)
				}
			}
			nm.mutex.Unlock()
		} else {
			// Send to specific client
			nm.mutex.Lock()
			if addr, ok := nm.clients[msg.PlayerID]; ok {
				_, err := nm.udpConn.WriteToUDP(data, addr)
				if err != nil {
					log.Printf("Error sending message to client %s: %v", msg.PlayerID, err)
				} else {
					log.Printf("Successfully sent message to client %s", msg.PlayerID)
				}
			} else {
				log.Printf("Client %s not found in clients map", msg.PlayerID)
			}
			nm.mutex.Unlock()
		}
	} else {
		// Client always sends to server using Write (not WriteToUDP)
		// For client, we already have the server address set as the remote address
		if nm.udpConn != nil {
			_, err := nm.udpConn.Write(data)
			if err != nil {
				log.Printf("Error sending message to server: %v", err)
			} else {
				log.Printf("Successfully sent message to server")
			}
		} else {
			log.Printf("UDP connection is nil, cannot send message")
		}
	}
}

// IsNetworkInitialized returns whether the network system has been initialized
func IsNetworkInitialized() bool {
	networkMutex.Lock()
	defer networkMutex.Unlock()
	return networkManager != nil
}

// IsServer returns whether this instance is running as a server
func IsServer() bool {
	networkMutex.Lock()
	defer networkMutex.Unlock()
	return networkManager != nil && networkManager.config.Role == RoleServer
}

// IsClient returns whether this instance is running as a client
func IsClient() bool {
	networkMutex.Lock()
	defer networkMutex.Unlock()
	return networkManager != nil && networkManager.config.Role == RoleClient
}

// IsConnectionLost returns whether the connection to the server has been lost
func IsConnectionLost() bool {
	networkMutex.Lock()
	defer networkMutex.Unlock()
	if networkManager == nil {
		return false
	}

	networkManager.mutex.Lock()
	defer networkManager.mutex.Unlock()
	return networkManager.connectionLost
}

// GetNetworkError returns any network error message
func GetNetworkError() string {
	networkMutex.Lock()
	defer networkMutex.Unlock()
	if networkManager == nil {
		return ""
	}

	networkManager.mutex.Lock()
	defer networkManager.mutex.Unlock()
	return networkManager.networkError
}

// IsWaitingForPlayers returns whether the server is waiting for players to connect
func IsWaitingForPlayers() bool {
	networkMutex.Lock()
	defer networkMutex.Unlock()
	if networkManager == nil {
		return false
	}

	networkManager.mutex.Lock()
	defer networkManager.mutex.Unlock()
	return networkManager.waitingForPlayers
}

// getGameName returns the name of the multiplayer game
func getGameName() string {
	networkMutex.Lock()
	defer networkMutex.Unlock()
	if networkManager == nil {
		return ""
	}

	return networkManager.config.GameName
}

// getPlayerID returns the ID of the local player
func getPlayerID() string {
	networkMutex.Lock()
	defer networkMutex.Unlock()
	if networkManager == nil {
		return ""
	}
	return networkManager.config.PlayerID
}

// GetConnectedPlayers returns a list of connected player IDs (server only)
func GetConnectedPlayers() []string {
	networkMutex.Lock()
	defer networkMutex.Unlock()

	if networkManager == nil || networkManager.config.Role != RoleServer {
		return []string{}
	}

	networkManager.mutex.Lock()
	defer networkManager.mutex.Unlock()

	players := make([]string, 0, len(networkManager.clients))
	for id := range networkManager.clients {
		players = append(players, id)
	}
	return players
}

// --- Callback Registration ---

// SetOnConnectCallback sets the function to call when a player connects
func SetOnConnectCallback(callback func(playerID string)) {
	networkMutex.Lock()
	defer networkMutex.Unlock()

	if networkManager != nil {
		networkManager.onConnect = callback
	}
}

// SetOnDisconnectCallback sets the function to call when a player disconnects
func SetOnDisconnectCallback(callback func(playerID string)) {
	networkMutex.Lock()
	defer networkMutex.Unlock()

	if networkManager != nil {
		networkManager.onDisconnect = callback
	}
}

// SetOnGameStateCallback sets the function to call when game state is received
func SetOnGameStateCallback(callback func(playerID string, data []byte)) {
	networkMutex.Lock()
	defer networkMutex.Unlock()

	if networkManager != nil {
		networkManager.onGameState = callback
	}
}

// SetOnPlayerInputCallback sets the function to call when player input is received
func SetOnPlayerInputCallback(callback func(playerID string, data []byte)) {
	networkMutex.Lock()
	defer networkMutex.Unlock()

	log.Printf("Setting player input callback: %v", callback != nil)
	if networkManager != nil {
		networkManager.onPlayerInput = callback
		log.Printf("Player input callback set on network manager: %v", networkManager.onPlayerInput != nil)
	} else {
		log.Printf("WARNING: Network manager is nil, callback will be lost")
	}
}

// AreCallbacksRegistered checks if callbacks are registered
func AreCallbacksRegistered() bool {
	networkMutex.Lock()
	defer networkMutex.Unlock()

	if networkManager == nil {
		log.Printf("Network manager is nil, callbacks cannot be registered")
		return false
	}

	hasGameState := networkManager.onGameState != nil
	hasPlayerInput := networkManager.onPlayerInput != nil
	hasConnect := networkManager.onConnect != nil
	hasDisconnect := networkManager.onDisconnect != nil

	log.Printf("Callback status: onGameState=%v, onPlayerInput=%v, onConnect=%v, onDisconnect=%v",
		hasGameState, hasPlayerInput, hasConnect, hasDisconnect)

	return hasGameState && hasPlayerInput && hasConnect && hasDisconnect
}

// ForceRegisterCallbacks directly sets callbacks on the network manager
// This is a last resort to ensure callbacks are registered
func ForceRegisterCallbacks(
	onGameState func(string, []byte),
	onPlayerInput func(string, []byte),
	onConnect func(string),
	onDisconnect func(string)) {

	networkMutex.Lock()
	defer networkMutex.Unlock()

	if networkManager == nil {
		log.Printf("Cannot force register callbacks: network manager is nil")
		return
	}

	log.Printf("Forcing callback registration directly on network manager")
	networkManager.onGameState = onGameState
	networkManager.onPlayerInput = onPlayerInput
	networkManager.onConnect = onConnect
	networkManager.onDisconnect = onDisconnect

	log.Printf("Forced callback registration complete")
}

// --- Message Sending ---

// SendGameState sends the current game state to all players or a specific player
func SendGameState(data []byte, targetPlayerID string) {
	networkMutex.Lock()
	defer networkMutex.Unlock()

	if networkManager == nil {
		return
	}

	select {
	case <-networkManager.done:
		return
	case networkManager.outgoingMsgs <- networkMessage{
		Type:     msgGameState,
		PlayerID: targetPlayerID, // "all" for broadcast
		Data:     data,
	}:
	}
}

// SendPlayerInput sends player input to the server
func SendPlayerInput(data []byte) {
	networkMutex.Lock()
	defer networkMutex.Unlock()

	if networkManager == nil || networkManager.config.Role != RoleClient {
		return
	}

	select {
	case <-networkManager.done:
		return
	case networkManager.outgoingMsgs <- networkMessage{
		Type:     msgPlayerInput,
		PlayerID: networkManager.config.PlayerID,
		Data:     data,
	}:
	}
}

// ParseNetworkArgs parses command line arguments for network configuration
// This is a helper function to standardize network command line arguments
func ParseNetworkArgs() *Config {
	config := defaultNetworkConfig()

	// Define command line flags
	var role string
	flag.StringVar(&role, "role", "server", "Role: server or client")
	flag.StringVar(&config.Address, "connect", "localhost", "Server address to connect to (client only)")
	flag.IntVar(&config.Port, "port", 8080, "Port to use for connection")
	flag.StringVar(&config.GameName, "name", config.GameName, "Name of the multiplayer game")

	// Parse flags
	flag.Parse()

	// Set role based on flag
	if role == "client" {
		config.Role = RoleClient
	} else {
		config.Role = RoleServer
	}

	return config
}

// getLocalIP returns the local IP address that can be used for network connections
// It finds the first non-loopback IPv4 address on an active network interface
func getLocalIP() string {
	// Get all network interfaces
	interfaces, err := net.Interfaces()
	if err != nil {
		return "error getting IP"
	}

	// Look for a non-loopback, IPv4 address
	for _, iface := range interfaces {
		// Skip loopback and interfaces that are down
		if iface.Flags&net.FlagLoopback != 0 || iface.Flags&net.FlagUp == 0 {
			continue
		}

		// Get addresses for this interface
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}

		// Look for IPv4 addresses
		for _, addr := range addrs {
			var ip net.IP

			// Extract IP from address
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}

			// Skip non-IPv4 and loopback addresses
			if ip == nil || ip.IsLoopback() {
				continue
			}

			// Convert to IPv4 if needed
			ip = ip.To4()
			if ip == nil {
				continue // Not an IPv4 address
			}

			// Found a valid IP address
			return ip.String()
		}
	}

	return "IP not found"
}

// DrawNetworkStatus draws the current network status on the screen
// This is a helper function to standardize network status display
func DrawNetworkStatus() {
	networkMutex.Lock()
	defer networkMutex.Unlock()

	if networkManager == nil {
		return
	}

	networkManager.mutex.Lock()
	defer networkManager.mutex.Unlock()

	// Display network error if any
	if networkManager.networkError != "" {
		log.Println("Network Error: ", networkManager.networkError)
		return
	}

	// Display waiting message if waiting for players
	if networkManager.waitingForPlayers {
		if networkManager.config.Role == RoleServer {
			log.Printf("Network: Waiting for player to join ... your IP is: %v\n", getLocalIP())
		} else {
			log.Printf("Network: Connecting to server...\n")
		}
		return
	}

	// Display role information
	if networkManager.config.Role == RoleServer {
		log.Println("Network: Server mode")
	} else {
		log.Println("Network: Client Mode")
	}
}
