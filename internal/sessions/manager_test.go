package sessions

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/thand-io/agent/internal/models"
)

func TestLoginServer_GetSessions(t *testing.T) {
	sessions := map[string]models.LocalSession{
		"provider1": {Version: 1, Expiry: time.Now().Add(1 * time.Hour)},
		"provider2": {Version: 1, Expiry: time.Now().Add(2 * time.Hour)},
	}

	ls := LoginServer{
		Version:   "1.0",
		Timestamp: time.Now(),
		Sessions:  sessions,
	}

	result := ls.GetSessions()

	if len(result) != 2 {
		t.Errorf("Expected 2 sessions, got %d", len(result))
	}

	if _, ok := result["provider1"]; !ok {
		t.Error("Expected provider1 to be in sessions")
	}

	if _, ok := result["provider2"]; !ok {
		t.Error("Expected provider2 to be in sessions")
	}
}

func TestLoginServer_GetFirstActiveSession(t *testing.T) {
	tests := []struct {
		name           string
		sessions       map[string]models.LocalSession
		providers      []string
		expectError    bool
		expectProvider string
	}{
		{
			name:        "empty sessions returns error",
			sessions:    map[string]models.LocalSession{},
			providers:   nil,
			expectError: true,
		},
		{
			name: "finds active session",
			sessions: map[string]models.LocalSession{
				"provider1": {Version: 1, Expiry: time.Now().Add(1 * time.Hour)},
			},
			providers:      nil,
			expectError:    false,
			expectProvider: "provider1",
		},
		{
			name: "expired session returns error",
			sessions: map[string]models.LocalSession{
				"provider1": {Version: 1, Expiry: time.Now().Add(-1 * time.Hour)},
			},
			providers:   nil,
			expectError: true,
		},
		{
			name: "filters by provider list",
			sessions: map[string]models.LocalSession{
				"provider1": {Version: 1, Expiry: time.Now().Add(1 * time.Hour)},
				"provider2": {Version: 1, Expiry: time.Now().Add(1 * time.Hour)},
			},
			providers:      []string{"provider2"},
			expectError:    false,
			expectProvider: "provider2",
		},
		{
			name: "provider not in filter list returns error",
			sessions: map[string]models.LocalSession{
				"provider1": {Version: 1, Expiry: time.Now().Add(1 * time.Hour)},
			},
			providers:   []string{"provider2"},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ls := LoginServer{
				Version:   "1.0",
				Timestamp: time.Now(),
				Sessions:  tt.sessions,
			}

			providerName, session, err := ls.GetFirstActiveSession(tt.providers...)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if session == nil {
				t.Error("Expected session, got nil")
				return
			}

			if len(tt.expectProvider) != 0 && providerName != tt.expectProvider {
				t.Errorf("Expected provider %s, got %s", tt.expectProvider, providerName)
			}
		})
	}
}

func TestSessionManager_AddAndGetSession(t *testing.T) {
	manager := &SessionManager{
		Servers: make(map[string]LoginServer),
	}

	loginServer := "test.example.com"
	provider := "test-provider"
	session := models.LocalSession{
		Version: 1,
		Expiry:  time.Now().Add(1 * time.Hour),
		Session: "test-session-token",
	}

	// Setup temp directory for session files
	tmpDir := setupTempSessionDir(t)
	defer os.RemoveAll(tmpDir)

	err := manager.AddSession(loginServer, provider, session)
	if err != nil {
		t.Fatalf("Failed to add session: %v", err)
	}

	retrievedSession, err := manager.GetSession(loginServer, provider)
	if err != nil {
		t.Fatalf("Failed to get session: %v", err)
	}

	if retrievedSession.Version != session.Version {
		t.Errorf("Expected version %d, got %d", session.Version, retrievedSession.Version)
	}

	if retrievedSession.Session != session.Session {
		t.Errorf("Expected session token %s, got %s", session.Session, retrievedSession.Session)
	}
}

func TestSessionManager_GetSession_NotFound(t *testing.T) {
	manager := &SessionManager{
		Servers: make(map[string]LoginServer),
	}

	loginServer := "test.example.com"

	_, err := manager.GetSession(loginServer, "nonexistent-provider")
	if err == nil {
		t.Error("Expected error for non-existent session, got nil")
	}
}

func TestSessionManager_RemoveSession(t *testing.T) {
	manager := &SessionManager{
		Servers: make(map[string]LoginServer),
	}

	loginServer := "test.example.com"
	provider := "test-provider"
	session := models.LocalSession{
		Version: 1,
		Expiry:  time.Now().Add(1 * time.Hour),
		Session: "test-session-token",
	}

	// Setup temp directory for session files
	tmpDir := setupTempSessionDir(t)
	defer os.RemoveAll(tmpDir)

	// Add session first
	err := manager.AddSession(loginServer, provider, session)
	if err != nil {
		t.Fatalf("Failed to add session: %v", err)
	}

	// Verify session exists
	_, err = manager.GetSession(loginServer, provider)
	if err != nil {
		t.Fatalf("Session should exist before removal: %v", err)
	}

	// Remove session
	err = manager.RemoveSession(loginServer, provider)
	if err != nil {
		t.Fatalf("Failed to remove session: %v", err)
	}

	// Verify session is removed
	_, err = manager.GetSession(loginServer, provider)
	if err == nil {
		t.Error("Session should not exist after removal")
	}
}

func TestSessionManager_GetFirstActiveSession(t *testing.T) {
	manager := &SessionManager{
		Servers: make(map[string]LoginServer),
	}

	loginServer := "test.example.com"

	// Setup temp directory for session files
	tmpDir := setupTempSessionDir(t)
	defer os.RemoveAll(tmpDir)

	// Add an active session
	activeSession := models.LocalSession{
		Version: 1,
		Expiry:  time.Now().Add(1 * time.Hour),
		Session: "active-session-token",
	}

	err := manager.AddSession(loginServer, "active-provider", activeSession)
	if err != nil {
		t.Fatalf("Failed to add session: %v", err)
	}

	providerName, session, err := manager.GetFirstActiveSession(loginServer)
	if err != nil {
		t.Fatalf("Failed to get first active session: %v", err)
	}

	if providerName != "active-provider" {
		t.Errorf("Expected provider 'active-provider', got '%s'", providerName)
	}

	if session == nil {
		t.Error("Expected session, got nil")
	}
}

func TestSessionManager_GetLoginServer(t *testing.T) {
	manager := &SessionManager{
		Servers: make(map[string]LoginServer),
	}

	tests := []struct {
		name        string
		loginServer string
		expectError bool
	}{
		{
			name:        "valid login server",
			loginServer: "https://test.example.com",
			expectError: false,
		},
		{
			name:        "simple hostname",
			loginServer: "test.example.com",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server, err := manager.GetLoginServer(tt.loginServer)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if server == nil {
				t.Error("Expected server, got nil")
			}
		})
	}
}

func TestSessionManager_GetLastTimestamp(t *testing.T) {
	manager := &SessionManager{
		Servers: make(map[string]LoginServer),
	}

	loginServer := "test.example.com"
	provider := "test-provider"
	expiry := time.Now().Add(1 * time.Hour)
	session := models.LocalSession{
		Version: 1,
		Expiry:  expiry,
		Session: "test-session-token",
	}

	// Setup temp directory for session files
	tmpDir := setupTempSessionDir(t)
	defer os.RemoveAll(tmpDir)

	err := manager.AddSession(loginServer, provider, session)
	if err != nil {
		t.Fatalf("Failed to add session: %v", err)
	}

	// Test getting timestamp for specific provider
	timestamp := manager.GetLastTimestamp(loginServer, provider)
	if timestamp == nil {
		t.Fatal("Expected timestamp, got nil")
	}

	// Timestamps should be within a small delta
	if timestamp.Unix() != expiry.Unix() {
		t.Errorf("Expected timestamp %v, got %v", expiry, *timestamp)
	}
}

func TestSessionManager_GetLastTimestamp_NoProvider(t *testing.T) {
	manager := &SessionManager{
		Servers: make(map[string]LoginServer),
	}

	loginServer := "test.example.com"

	// Setup temp directory for session files
	tmpDir := setupTempSessionDir(t)
	defer os.RemoveAll(tmpDir)

	session := models.LocalSession{
		Version: 1,
		Expiry:  time.Now().Add(1 * time.Hour),
		Session: "test-session-token",
	}

	err := manager.AddSession(loginServer, "provider1", session)
	if err != nil {
		t.Fatalf("Failed to add session: %v", err)
	}

	// Test getting timestamp for login server (no specific provider)
	timestamp := manager.GetLastTimestamp(loginServer, "")
	if timestamp == nil {
		t.Fatal("Expected timestamp, got nil")
	}
}

func TestSessionManager_CreateLoginServer(t *testing.T) {
	manager := &SessionManager{
		Servers: make(map[string]LoginServer),
	}

	loginServer := "new-server.example.com"

	// Initially, server should not exist
	if _, ok := manager.Servers[loginServer]; ok {
		t.Error("Server should not exist initially")
	}

	// Create the login server
	manager.createLoginServer(loginServer)

	// Server should exist now
	server, ok := manager.Servers[loginServer]
	if !ok {
		t.Fatal("Server should exist after creation")
	}

	if server.Version != "1.0" {
		t.Errorf("Expected version '1.0', got '%s'", server.Version)
	}

	if server.Sessions == nil {
		t.Error("Sessions map should be initialized")
	}

	// Calling createLoginServer again should not overwrite existing server
	existingTimestamp := server.Timestamp
	manager.createLoginServer(loginServer)

	if manager.Servers[loginServer].Timestamp != existingTimestamp {
		t.Error("Existing server should not be overwritten")
	}
}

func TestSessionManager_LoadAndCommit(t *testing.T) {
	manager := &SessionManager{
		Servers: make(map[string]LoginServer),
	}

	loginServer := "test-load.example.com"

	// Setup temp directory for session files
	tmpDir := setupTempSessionDir(t)
	defer os.RemoveAll(tmpDir)

	session := models.LocalSession{
		Version: 1,
		Expiry:  time.Now().Add(1 * time.Hour),
		Session: "test-session-token",
	}

	// Add and commit a session
	err := manager.AddSession(loginServer, "provider1", session)
	if err != nil {
		t.Fatalf("Failed to add session: %v", err)
	}

	// Create a new manager and load the session
	newManager := &SessionManager{
		Servers: make(map[string]LoginServer),
	}

	err = newManager.Load(loginServer)
	if err != nil {
		t.Fatalf("Failed to load session: %v", err)
	}

	// Verify the session was loaded
	loadedSession, err := newManager.GetSession(loginServer, "provider1")
	if err != nil {
		t.Fatalf("Failed to get loaded session: %v", err)
	}

	if loadedSession.Session != session.Session {
		t.Errorf("Expected session token %s, got %s", session.Session, loadedSession.Session)
	}
}

func TestSessionManager_Load_EmptyFile(t *testing.T) {
	manager := &SessionManager{
		Servers: make(map[string]LoginServer),
	}

	loginServer := "empty-file-test.example.com"

	// Setup temp directory for session files
	tmpDir := setupTempSessionDir(t)
	defer os.RemoveAll(tmpDir)

	// Create an empty session file
	sessionFile := filepath.Join(tmpDir, loginServer+".yaml")
	_, err := os.Create(sessionFile)
	if err != nil {
		t.Fatalf("Failed to create empty file: %v", err)
	}

	err = manager.Load(loginServer)
	if err != nil {
		t.Fatalf("Failed to load empty file: %v", err)
	}

	// Verify default LoginServer was created
	server, ok := manager.Servers[loginServer]
	if !ok {
		t.Fatal("Server should exist after loading empty file")
	}

	if server.Version != "1.0" {
		t.Errorf("Expected version '1.0', got '%s'", server.Version)
	}
}

func TestSessionManager_AwaitRefresh_ContextCancellation(t *testing.T) {
	manager := &SessionManager{
		Servers: make(map[string]LoginServer),
	}

	loginServer := "test.example.com"

	// Setup temp directory for session files
	tmpDir := setupTempSessionDir(t)
	defer os.RemoveAll(tmpDir)

	// Create a context that will be cancelled immediately
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	result := manager.AwaitRefresh(ctx, loginServer)

	if result != nil {
		t.Error("Expected nil result when context is cancelled")
	}
}

func TestSessionManager_AwaitProviderRefresh_ContextCancellation(t *testing.T) {
	manager := &SessionManager{
		Servers: make(map[string]LoginServer),
	}

	loginServer := "test.example.com"
	provider := "test-provider"

	// Setup temp directory for session files
	tmpDir := setupTempSessionDir(t)
	defer os.RemoveAll(tmpDir)

	// Add a session first
	session := models.LocalSession{
		Version: 1,
		Expiry:  time.Now().Add(1 * time.Hour),
		Session: "test-session-token",
	}
	err := manager.AddSession(loginServer, provider, session)
	if err != nil {
		t.Fatalf("Failed to add session: %v", err)
	}

	// Create a context that will be cancelled immediately
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	result := manager.AwaitProviderRefresh(ctx, loginServer, provider)

	if result != nil {
		t.Error("Expected nil result when context is cancelled")
	}
}

func TestGetSessionManager(t *testing.T) {
	manager := GetSessionManager()

	if manager == nil {
		t.Fatal("Expected session manager, got nil")
	}

	// Calling again should return the same instance
	manager2 := GetSessionManager()

	if manager != manager2 {
		t.Error("GetSessionManager should return the same instance")
	}
}

// Helper function to setup a temporary session directory for testing
func setupTempSessionDir(t *testing.T) string {
	t.Helper()

	tmpDir, err := os.MkdirTemp("", "thand-sessions-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}

	// Override the session path for testing
	originalPath := SESSION_MANAGER_PATH
	SESSION_MANAGER_PATH = tmpDir

	t.Cleanup(func() {
		SESSION_MANAGER_PATH = originalPath
	})

	return tmpDir
}
