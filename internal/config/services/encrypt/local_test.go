package encrypt

import (
	"context"
	"testing"

	"github.com/thand-io/agent/internal/models"
)

func TestNewLocalEncryptionFromConfig(t *testing.T) {
	config := &models.BasicConfig{}
	enc := NewLocalEncryptionFromConfig(config)

	if enc == nil {
		t.Error("NewLocalEncryptionFromConfig returned nil")
	}
}

func TestLocalVault_Initialize(t *testing.T) {
	tests := []struct {
		name     string
		password string
		salt     string
		wantErr  bool
	}{
		{
			name:     "with custom password and salt",
			password: "my-secure-password-123",
			salt:     "my-custom-salt",
			wantErr:  false,
		},
		{
			name:     "with only password",
			password: "my-secure-password",
			salt:     "",
			wantErr:  false,
		},
		{
			name:     "with only salt",
			password: "",
			salt:     "my-salt",
			wantErr:  false,
		},
		{
			name:     "with empty config (uses defaults)",
			password: "",
			salt:     "",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &models.BasicConfig{}
			if tt.password != "" {
				config.SetKeyWithValue("password", tt.password)
			}
			if tt.salt != "" {
				config.SetKeyWithValue("salt", tt.salt)
			}

			enc := NewLocalEncryptionFromConfig(config)
			err := enc.Initialize()

			if (err != nil) != tt.wantErr {
				t.Errorf("Initialize() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestLocalVault_EncryptDecrypt(t *testing.T) {
	tests := []struct {
		name      string
		password  string
		salt      string
		plaintext string
	}{
		{
			name:      "simple text",
			password:  "test-password",
			salt:      "test-salt",
			plaintext: "Hello, World!",
		},
		{
			name:      "json data",
			password:  "secure-pass",
			salt:      "unique-salt",
			plaintext: `{"key": "value", "number": 123}`,
		},
		{
			name:      "unicode text",
			password:  "unicode-password",
			salt:      "unicode-salt",
			plaintext: "Hello 世界 🌍 мир",
		},
		{
			name:      "long text",
			password:  "long-password",
			salt:      "long-salt",
			plaintext: "This is a much longer piece of text that simulates a more realistic scenario where we might be encrypting larger payloads such as credentials, tokens, or configuration data.",
		},
		{
			name:      "special characters",
			password:  "special-pass!@#$",
			salt:      "special-salt!@#$",
			plaintext: "Text with special chars: !@#$%^&*()_+-=[]{}|;':\",./<>?",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &models.BasicConfig{}
			config.SetKeyWithValue("password", tt.password)
			config.SetKeyWithValue("salt", tt.salt)

			enc := NewLocalEncryptionFromConfig(config)
			if err := enc.Initialize(); err != nil {
				t.Fatalf("Initialize() error = %v", err)
			}

			ctx := context.Background()

			// Encrypt
			ciphertext, err := enc.Encrypt(ctx, []byte(tt.plaintext))
			if err != nil {
				t.Fatalf("Encrypt() error = %v", err)
			}

			if len(ciphertext) == 0 {
				t.Error("Encrypt() returned empty ciphertext")
			}

			// Ciphertext should be different from plaintext
			if string(ciphertext) == tt.plaintext {
				t.Error("Encrypt() returned plaintext as ciphertext")
			}

			// Decrypt
			decrypted, err := enc.Decrypt(ctx, ciphertext)
			if err != nil {
				t.Fatalf("Decrypt() error = %v", err)
			}

			if string(decrypted) != tt.plaintext {
				t.Errorf("Decrypt() = %q, want %q", string(decrypted), tt.plaintext)
			}
		})
	}
}

func TestLocalVault_EncryptErrors(t *testing.T) {
	config := &models.BasicConfig{}
	config.SetKeyWithValue("password", "test-password")
	config.SetKeyWithValue("salt", "test-salt")

	enc := NewLocalEncryptionFromConfig(config)
	if err := enc.Initialize(); err != nil {
		t.Fatalf("Initialize() error = %v", err)
	}

	ctx := context.Background()

	// Test empty plaintext
	_, err := enc.Encrypt(ctx, []byte{})
	if err == nil {
		t.Error("Encrypt() with empty plaintext should return error")
	}

	// Test nil plaintext
	_, err = enc.Encrypt(ctx, nil)
	if err == nil {
		t.Error("Encrypt() with nil plaintext should return error")
	}
}

func TestLocalVault_DecryptErrors(t *testing.T) {
	config := &models.BasicConfig{}
	config.SetKeyWithValue("password", "test-password")
	config.SetKeyWithValue("salt", "test-salt")

	enc := NewLocalEncryptionFromConfig(config)
	if err := enc.Initialize(); err != nil {
		t.Fatalf("Initialize() error = %v", err)
	}

	ctx := context.Background()

	tests := []struct {
		name       string
		ciphertext []byte
		wantErr    bool
	}{
		{
			name:       "empty ciphertext",
			ciphertext: []byte{},
			wantErr:    true,
		},
		{
			name:       "nil ciphertext",
			ciphertext: nil,
			wantErr:    true,
		},
		{
			name:       "invalid json",
			ciphertext: []byte("not json"),
			wantErr:    true,
		},
		{
			name:       "invalid base64 nonce",
			ciphertext: []byte(`{"nonce": "!!!invalid!!!", "ciphertext": "dGVzdA=="}`),
			wantErr:    true,
		},
		{
			name:       "invalid base64 ciphertext",
			ciphertext: []byte(`{"nonce": "dGVzdA==", "ciphertext": "!!!invalid!!!"}`),
			wantErr:    true,
		},
		{
			name:       "valid json but wrong data (decryption fails)",
			ciphertext: []byte(`{"nonce": "dGVzdHRlc3R0ZXN0", "ciphertext": "dGVzdA=="}`),
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := enc.Decrypt(ctx, tt.ciphertext)
			if (err != nil) != tt.wantErr {
				t.Errorf("Decrypt() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestLocalVault_DecryptWithWrongPassword(t *testing.T) {
	ctx := context.Background()

	// Encrypt with one password
	config1 := &models.BasicConfig{}
	config1.SetKeyWithValue("password", "password-one")
	config1.SetKeyWithValue("salt", "same-salt")

	enc1 := NewLocalEncryptionFromConfig(config1)
	if err := enc1.Initialize(); err != nil {
		t.Fatalf("Initialize() error = %v", err)
	}

	ciphertext, err := enc1.Encrypt(ctx, []byte("secret message"))
	if err != nil {
		t.Fatalf("Encrypt() error = %v", err)
	}

	// Try to decrypt with different password
	config2 := &models.BasicConfig{}
	config2.SetKeyWithValue("password", "password-two")
	config2.SetKeyWithValue("salt", "same-salt")

	enc2 := NewLocalEncryptionFromConfig(config2)
	if err := enc2.Initialize(); err != nil {
		t.Fatalf("Initialize() error = %v", err)
	}

	_, err = enc2.Decrypt(ctx, ciphertext)
	if err == nil {
		t.Error("Decrypt() with wrong password should return error")
	}
}

func TestLocalVault_DecryptWithWrongSalt(t *testing.T) {
	ctx := context.Background()

	// Encrypt with one salt
	config1 := &models.BasicConfig{}
	config1.SetKeyWithValue("password", "same-password")
	config1.SetKeyWithValue("salt", "salt-one")

	enc1 := NewLocalEncryptionFromConfig(config1)
	if err := enc1.Initialize(); err != nil {
		t.Fatalf("Initialize() error = %v", err)
	}

	ciphertext, err := enc1.Encrypt(ctx, []byte("secret message"))
	if err != nil {
		t.Fatalf("Encrypt() error = %v", err)
	}

	// Try to decrypt with different salt
	config2 := &models.BasicConfig{}
	config2.SetKeyWithValue("password", "same-password")
	config2.SetKeyWithValue("salt", "salt-two")

	enc2 := NewLocalEncryptionFromConfig(config2)
	if err := enc2.Initialize(); err != nil {
		t.Fatalf("Initialize() error = %v", err)
	}

	_, err = enc2.Decrypt(ctx, ciphertext)
	if err == nil {
		t.Error("Decrypt() with wrong salt should return error")
	}
}

func TestLocalVault_EncryptProducesUniqueOutput(t *testing.T) {
	config := &models.BasicConfig{}
	config.SetKeyWithValue("password", "test-password")
	config.SetKeyWithValue("salt", "test-salt")

	enc := NewLocalEncryptionFromConfig(config)
	if err := enc.Initialize(); err != nil {
		t.Fatalf("Initialize() error = %v", err)
	}

	ctx := context.Background()
	plaintext := []byte("same message")

	// Encrypt the same message multiple times
	ciphertext1, err := enc.Encrypt(ctx, plaintext)
	if err != nil {
		t.Fatalf("Encrypt() error = %v", err)
	}

	ciphertext2, err := enc.Encrypt(ctx, plaintext)
	if err != nil {
		t.Fatalf("Encrypt() error = %v", err)
	}

	// Each encryption should produce different output due to random nonce
	if string(ciphertext1) == string(ciphertext2) {
		t.Error("Encrypt() should produce unique output for the same plaintext")
	}

	// But both should decrypt to the same plaintext
	decrypted1, err := enc.Decrypt(ctx, ciphertext1)
	if err != nil {
		t.Fatalf("Decrypt() error = %v", err)
	}

	decrypted2, err := enc.Decrypt(ctx, ciphertext2)
	if err != nil {
		t.Fatalf("Decrypt() error = %v", err)
	}

	if string(decrypted1) != string(decrypted2) {
		t.Error("Both ciphertexts should decrypt to the same plaintext")
	}

	if string(decrypted1) != string(plaintext) {
		t.Errorf("Decrypted text = %q, want %q", string(decrypted1), string(plaintext))
	}
}

func TestLocalVault_Shutdown(t *testing.T) {
	config := &models.BasicConfig{}
	config.SetKeyWithValue("password", "test-password")
	config.SetKeyWithValue("salt", "test-salt")

	enc := NewLocalEncryptionFromConfig(config).(*localVault)
	if err := enc.Initialize(); err != nil {
		t.Fatalf("Initialize() error = %v", err)
	}

	// Verify key is set
	if len(enc.key) == 0 {
		t.Error("Key should be set after Initialize()")
	}

	// Shutdown should clear the key
	if err := enc.Shutdown(); err != nil {
		t.Fatalf("Shutdown() error = %v", err)
	}

	// Verify key is zeroed
	for i, b := range enc.key {
		if b != 0 {
			t.Errorf("Key byte %d should be 0 after Shutdown(), got %d", i, b)
		}
	}
}

func TestDeriveKey(t *testing.T) {
	tests := []struct {
		name     string
		password string
		salt     string
	}{
		{
			name:     "simple password and salt",
			password: "password",
			salt:     "salt",
		},
		{
			name:     "complex password",
			password: "C0mpl3x!@#$%^&*()_+Password",
			salt:     "unique-salt-value",
		},
		{
			name:     "unicode password",
			password: "пароль密码パスワード",
			salt:     "unicode-salt",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key := deriveKey(tt.password, tt.salt)

			// Key should be 32 bytes (256 bits)
			if len(key) != 32 {
				t.Errorf("deriveKey() key length = %d, want 32", len(key))
			}

			// Same inputs should produce same key
			key2 := deriveKey(tt.password, tt.salt)
			if string(key) != string(key2) {
				t.Error("deriveKey() should be deterministic")
			}

			// Different salt should produce different key
			key3 := deriveKey(tt.password, tt.salt+"-different")
			if string(key) == string(key3) {
				t.Error("deriveKey() should produce different keys for different salts")
			}

			// Different password should produce different key
			key4 := deriveKey(tt.password+"-different", tt.salt)
			if string(key) == string(key4) {
				t.Error("deriveKey() should produce different keys for different passwords")
			}
		})
	}
}

func TestLocalVault_RoundTrip_WithDefaultConfig(t *testing.T) {
	// Test with empty config (uses default values)
	config := &models.BasicConfig{}

	enc := NewLocalEncryptionFromConfig(config)
	if err := enc.Initialize(); err != nil {
		t.Fatalf("Initialize() error = %v", err)
	}

	ctx := context.Background()
	plaintext := []byte("testing with default config")

	ciphertext, err := enc.Encrypt(ctx, plaintext)
	if err != nil {
		t.Fatalf("Encrypt() error = %v", err)
	}

	decrypted, err := enc.Decrypt(ctx, ciphertext)
	if err != nil {
		t.Fatalf("Decrypt() error = %v", err)
	}

	if string(decrypted) != string(plaintext) {
		t.Errorf("Decrypt() = %q, want %q", string(decrypted), string(plaintext))
	}
}

func TestLocalVault_BinaryData(t *testing.T) {
	config := &models.BasicConfig{}
	config.SetKeyWithValue("password", "binary-test")
	config.SetKeyWithValue("salt", "binary-salt")

	enc := NewLocalEncryptionFromConfig(config)
	if err := enc.Initialize(); err != nil {
		t.Fatalf("Initialize() error = %v", err)
	}

	ctx := context.Background()

	// Test with binary data including null bytes
	binaryData := []byte{0x00, 0x01, 0x02, 0xFF, 0xFE, 0x00, 0x7F, 0x80}

	ciphertext, err := enc.Encrypt(ctx, binaryData)
	if err != nil {
		t.Fatalf("Encrypt() error = %v", err)
	}

	decrypted, err := enc.Decrypt(ctx, ciphertext)
	if err != nil {
		t.Fatalf("Decrypt() error = %v", err)
	}

	if string(decrypted) != string(binaryData) {
		t.Errorf("Decrypt() binary data mismatch")
	}
}
