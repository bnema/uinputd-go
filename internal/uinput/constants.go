package uinput

// Linux input event constants from <linux/input.h>

// Event types
const (
	EvSyn = 0x00 // Synchronization events
	EvKey = 0x01 // Key/button events
	EvRel = 0x02 // Relative axes (mouse movement)
	EvAbs = 0x03 // Absolute axes (touchscreen)
)

// Synchronization event codes
const (
	SynReport = 0 // Marks end of event batch
)

// Key event values
const (
	KeyRelease = 0 // Key released
	KeyPress   = 1 // Key pressed
	KeyRepeat  = 2 // Key held (auto-repeat)
)

// Common key codes (from <linux/input-event-codes.h>)
const (
	KeyReserved   = 0
	KeyEsc        = 1
	Key1          = 2
	Key2          = 3
	Key3          = 4
	Key4          = 5
	Key5          = 6
	Key6          = 7
	Key7          = 8
	Key8          = 9
	Key9          = 10
	Key0          = 11
	KeyMinus      = 12
	KeyEqual      = 13
	KeyBackspace  = 14
	KeyTab        = 15
	KeyQ          = 16
	KeyW          = 17
	KeyE          = 18
	KeyR          = 19
	KeyT          = 20
	KeyY          = 21
	KeyU          = 22
	KeyI          = 23
	KeyO          = 24
	KeyP          = 25
	KeyLeftBrace  = 26
	KeyRightBrace = 27
	KeyEnter      = 28
	KeyLeftCtrl   = 29
	KeyA          = 30
	KeyS          = 31
	KeyD          = 32
	KeyF          = 33
	KeyG          = 34
	KeyH          = 35
	KeyJ          = 36
	KeyK          = 37
	KeyL          = 38
	KeySemicolon  = 39
	KeyApostrophe = 40
	KeyGrave      = 41
	KeyLeftShift  = 42
	KeyBackslash  = 43
	KeyZ          = 44
	KeyX          = 45
	KeyC          = 46
	KeyV          = 47
	KeyB          = 48
	KeyN          = 49
	KeyM          = 50
	KeyComma      = 51
	KeyDot        = 52
	KeySlash      = 53
	KeyRightShift = 54
	KeyLeftAlt    = 56
	KeySpace      = 57
	KeyCapsLock   = 58
	Key102ND      = 86  // Extra key on non-US keyboards (< > |)
	KeyRightAlt   = 100 // AltGr
	KeyRightCtrl  = 97
)

// Device name and ID
const (
	DeviceName = "uinputd-virtual-keyboard"
	BusVirtual = 0x06 // BUS_VIRTUAL
	VendorID   = 0x1234
	ProductID  = 0x5678
	Version    = 1
)

// uinput ioctl constants
const (
	UI_SET_EVBIT  = 0x40045564
	UI_SET_KEYBIT = 0x40045565
	UI_DEV_CREATE = 0x5501
	UI_DEV_DESTROY = 0x5502
	UI_DEV_SETUP  = 0x405c5503
)
