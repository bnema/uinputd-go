package server

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/bnema/uinputd-go/internal/config"
	"github.com/bnema/uinputd-go/internal/layouts"
	layoutMocks "github.com/bnema/uinputd-go/internal/layouts/mocks"
	"github.com/bnema/uinputd-go/internal/protocol"
	"github.com/bnema/uinputd-go/internal/uinput"
	uinputMocks "github.com/bnema/uinputd-go/internal/uinput/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// newTestServer creates a server instance for testing with mocked dependencies
func newTestServer(device uinput.DeviceInterface, registry layouts.RegistryInterface) *Server {
	return &Server{
		cfg: &config.Config{
			Layout: "us",
			Performance: config.PerformanceConfig{
				CharDelayMs:   10,
				StreamDelayMs: 50,
			},
		},
		device:   device,
		registry: registry,
	}
}

func TestHandleType(t *testing.T) {
	tests := []struct {
		name          string
		payload       protocol.TypePayload
		setupMocks    func(*uinputMocks.MockDeviceInterface, *layoutMocks.MockRegistryInterface, *layoutMocks.MockLayout)
		expectedError bool
	}{
		{
			name: "simple text typing",
			payload: protocol.TypePayload{
				Text:   "Hi",
				Layout: "us",
			},
			setupMocks: func(device *uinputMocks.MockDeviceInterface, registry *layoutMocks.MockRegistryInterface, layout *layoutMocks.MockLayout) {
				registry.On("Get", "us").Return(layout, nil)

				// Mock H (uppercase, needs Shift)
				layout.On("CharToKeySequence", mock.Anything, 'H').Return([]layouts.KeySequence{
					{Keycode: 35, Modifier: layouts.ModShift},
				}, nil)

				// Mock i (lowercase)
				layout.On("CharToKeySequence", mock.Anything, 'i').Return([]layouts.KeySequence{
					{Keycode: 23, Modifier: layouts.ModNone},
				}, nil)

				// Expect key presses
				device.On("SendKeyWithModifier", mock.Anything, uint16(uinput.KeyLeftShift), uint16(35)).Return(nil)
				device.On("SendKey", mock.Anything, uint16(23)).Return(nil)
			},
			expectedError: false,
		},
		{
			name: "layout not found",
			payload: protocol.TypePayload{
				Text:   "Test",
				Layout: "invalid",
			},
			setupMocks: func(device *uinputMocks.MockDeviceInterface, registry *layoutMocks.MockRegistryInterface, layout *layoutMocks.MockLayout) {
				registry.On("Get", "invalid").Return(nil, assert.AnError)
			},
			expectedError: true,
		},
		{
			name: "use default layout when not specified",
			payload: protocol.TypePayload{
				Text:   "a",
				Layout: "",
			},
			setupMocks: func(device *uinputMocks.MockDeviceInterface, registry *layoutMocks.MockRegistryInterface, layout *layoutMocks.MockLayout) {
				registry.On("Get", "us").Return(layout, nil) // Should use default from config

				layout.On("CharToKeySequence", mock.Anything, 'a').Return([]layouts.KeySequence{
					{Keycode: 30, Modifier: layouts.ModNone},
				}, nil)

				device.On("SendKey", mock.Anything, uint16(30)).Return(nil)
			},
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mocks
			mockDevice := uinputMocks.NewMockDeviceInterface(t)
			mockRegistry := layoutMocks.NewMockRegistryInterface(t)
			mockLayout := layoutMocks.NewMockLayout(t)

			// Setup mock expectations
			tt.setupMocks(mockDevice, mockRegistry, mockLayout)

			// Create server with mocks
			server := newTestServer(mockDevice, mockRegistry)

			// Marshal payload
			payloadBytes, _ := json.Marshal(tt.payload)

			// Call handleType
			err := server.handleType(context.Background(), payloadBytes)

			// Assert
			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestHandleStream(t *testing.T) {
	tests := []struct {
		name          string
		payload       protocol.StreamPayload
		setupMocks    func(*uinputMocks.MockDeviceInterface, *layoutMocks.MockRegistryInterface, *layoutMocks.MockLayout)
		expectedError bool
		verifyDelays  bool
	}{
		{
			name: "stream with custom delays",
			payload: protocol.StreamPayload{
				Text:      "Hi",
				Layout:    "us",
				CharDelay: 20,
				DelayMs:   100,
			},
			setupMocks: func(device *uinputMocks.MockDeviceInterface, registry *layoutMocks.MockRegistryInterface, layout *layoutMocks.MockLayout) {
				registry.On("Get", "us").Return(layout, nil)

				// Mock character sequences
				layout.On("CharToKeySequence", mock.Anything, 'H').Return([]layouts.KeySequence{
					{Keycode: 35, Modifier: layouts.ModShift},
				}, nil)
				layout.On("CharToKeySequence", mock.Anything, 'i').Return([]layouts.KeySequence{
					{Keycode: 23, Modifier: layouts.ModNone},
				}, nil)

				// Expect key presses
				device.On("SendKeyWithModifier", mock.Anything, uint16(uinput.KeyLeftShift), uint16(35)).Return(nil)
				device.On("SendKey", mock.Anything, uint16(23)).Return(nil)
			},
			expectedError: false,
			verifyDelays:  true,
		},
		{
			name: "stream with config default delays",
			payload: protocol.StreamPayload{
				Text:      "a",
				Layout:    "us",
				CharDelay: 0, // Should use config default (10ms)
				DelayMs:   0, // Should use config default (50ms)
			},
			setupMocks: func(device *uinputMocks.MockDeviceInterface, registry *layoutMocks.MockRegistryInterface, layout *layoutMocks.MockLayout) {
				registry.On("Get", "us").Return(layout, nil)

				layout.On("CharToKeySequence", mock.Anything, 'a').Return([]layouts.KeySequence{
					{Keycode: 30, Modifier: layouts.ModNone},
				}, nil)

				device.On("SendKey", mock.Anything, uint16(30)).Return(nil)
			},
			expectedError: false,
		},
		{
			name: "stream multiple words with spaces",
			payload: protocol.StreamPayload{
				Text:      "hi yo",
				Layout:    "us",
				CharDelay: 5,
				DelayMs:   10,
			},
			setupMocks: func(device *uinputMocks.MockDeviceInterface, registry *layoutMocks.MockRegistryInterface, layout *layoutMocks.MockLayout) {
				registry.On("Get", "us").Return(layout, nil)

				// Mock all characters
				layout.On("CharToKeySequence", mock.Anything, 'h').Return([]layouts.KeySequence{
					{Keycode: 35, Modifier: layouts.ModNone},
				}, nil)
				layout.On("CharToKeySequence", mock.Anything, 'i').Return([]layouts.KeySequence{
					{Keycode: 23, Modifier: layouts.ModNone},
				}, nil)
				layout.On("CharToKeySequence", mock.Anything, 'y').Return([]layouts.KeySequence{
					{Keycode: 21, Modifier: layouts.ModNone},
				}, nil)
				layout.On("CharToKeySequence", mock.Anything, 'o').Return([]layouts.KeySequence{
					{Keycode: 24, Modifier: layouts.ModNone},
				}, nil)
				layout.On("CharToKeySequence", mock.Anything, ' ').Return([]layouts.KeySequence{
					{Keycode: 57, Modifier: layouts.ModNone},
				}, nil)

				// Expect key presses for both words and space
				device.On("SendKey", mock.Anything, mock.Anything).Return(nil)
			},
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mocks
			mockDevice := uinputMocks.NewMockDeviceInterface(t)
			mockRegistry := layoutMocks.NewMockRegistryInterface(t)
			mockLayout := layoutMocks.NewMockLayout(t)

			// Setup mock expectations
			tt.setupMocks(mockDevice, mockRegistry, mockLayout)

			// Create server with mocks
			server := newTestServer(mockDevice, mockRegistry)

			// Marshal payload
			payloadBytes, _ := json.Marshal(tt.payload)

			// Measure execution time if verifying delays
			start := time.Now()

			// Call handleStream
			err := server.handleStream(context.Background(), payloadBytes)

			elapsed := time.Since(start)

			// Assert
			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)

				// Verify delays were applied (rough check)
				if tt.verifyDelays && tt.payload.CharDelay > 0 {
					// With "Hi" (2 chars) and CharDelay=20ms, should take at least 40ms
					expectedMin := time.Duration(len(tt.payload.Text)*tt.payload.CharDelay) * time.Millisecond
					assert.GreaterOrEqual(t, elapsed, expectedMin, "Delays should be applied")
				}
			}
		})
	}
}

func TestHandleKey(t *testing.T) {
	tests := []struct {
		name          string
		payload       protocol.KeyPayload
		setupMocks    func(*uinputMocks.MockDeviceInterface)
		expectedError bool
	}{
		{
			name: "plain key press",
			payload: protocol.KeyPayload{
				Keycode:  28,
				Modifier: "",
			},
			setupMocks: func(device *uinputMocks.MockDeviceInterface) {
				device.On("SendKey", mock.Anything, uint16(28)).Return(nil)
			},
			expectedError: false,
		},
		{
			name: "key with shift modifier",
			payload: protocol.KeyPayload{
				Keycode:  30,
				Modifier: "shift",
			},
			setupMocks: func(device *uinputMocks.MockDeviceInterface) {
				device.On("SendKeyWithModifier", mock.Anything, uint16(uinput.KeyLeftShift), uint16(30)).Return(nil)
			},
			expectedError: false,
		},
		{
			name: "key with ctrl modifier",
			payload: protocol.KeyPayload{
				Keycode:  46,
				Modifier: "ctrl",
			},
			setupMocks: func(device *uinputMocks.MockDeviceInterface) {
				device.On("SendKeyWithModifier", mock.Anything, uint16(uinput.KeyLeftCtrl), uint16(46)).Return(nil)
			},
			expectedError: false,
		},
		{
			name: "key with altgr modifier",
			payload: protocol.KeyPayload{
				Keycode:  30,
				Modifier: "altgr",
			},
			setupMocks: func(device *uinputMocks.MockDeviceInterface) {
				device.On("SendKeyWithModifier", mock.Anything, uint16(uinput.KeyRightAlt), uint16(30)).Return(nil)
			},
			expectedError: false,
		},
		{
			name: "invalid modifier",
			payload: protocol.KeyPayload{
				Keycode:  30,
				Modifier: "invalid",
			},
			setupMocks: func(device *uinputMocks.MockDeviceInterface) {
				// No expectations - should error before calling device
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock
			mockDevice := uinputMocks.NewMockDeviceInterface(t)

			// Setup mock expectations
			tt.setupMocks(mockDevice)

			// Create server with mock
			server := newTestServer(mockDevice, layoutMocks.NewMockRegistryInterface(t))

			// Marshal payload
			payloadBytes, _ := json.Marshal(tt.payload)

			// Call handleKey
			err := server.handleKey(context.Background(), payloadBytes)

			// Assert
			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestHandlePing(t *testing.T) {
	server := newTestServer(
		uinputMocks.NewMockDeviceInterface(t),
		layoutMocks.NewMockRegistryInterface(t),
	)

	err := server.handlePing(context.Background())
	assert.NoError(t, err)
}

func TestHandleCommand(t *testing.T) {
	tests := []struct {
		name        string
		commandType protocol.CommandType
		expectError bool
	}{
		{
			name:        "type command",
			commandType: protocol.CommandType_Type,
			expectError: false,
		},
		{
			name:        "stream command",
			commandType: protocol.CommandType_Stream,
			expectError: false,
		},
		{
			name:        "key command",
			commandType: protocol.CommandType_Key,
			expectError: false,
		},
		{
			name:        "ping command",
			commandType: protocol.CommandType_Ping,
			expectError: false,
		},
		{
			name:        "unknown command",
			commandType: "unknown",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDevice := uinputMocks.NewMockDeviceInterface(t)
			mockRegistry := layoutMocks.NewMockRegistryInterface(t)
			mockLayout := layoutMocks.NewMockLayout(t)

			server := newTestServer(mockDevice, mockRegistry)

			// Setup mocks for valid commands
			if !tt.expectError {
				switch tt.commandType {
				case protocol.CommandType_Type, protocol.CommandType_Stream:
					mockRegistry.On("Get", "us").Return(mockLayout, nil)
					mockLayout.On("CharToKeySequence", mock.Anything, mock.Anything).Return([]layouts.KeySequence{
						{Keycode: 30, Modifier: layouts.ModNone},
					}, nil).Maybe()
					mockDevice.On("SendKey", mock.Anything, mock.Anything).Return(nil).Maybe()
				case protocol.CommandType_Key:
					mockDevice.On("SendKey", mock.Anything, mock.Anything).Return(nil).Maybe()
				}
			}

			cmd := &protocol.Command{
				Type: tt.commandType,
				Payload: json.RawMessage(`{
					"text": "a",
					"layout": "us",
					"keycode": 28
				}`),
			}

			err := server.handleCommand(context.Background(), cmd)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
