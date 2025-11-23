package uinput

import (
	"context"
	"fmt"
	"os"
	"sync"
	"unsafe"

	"github.com/bnema/uinputd-go/internal/logger"
	"golang.org/x/sys/unix"
)

// Device represents a virtual uinput keyboard device.
type Device struct {
	fd   *os.File
	mu   sync.Mutex
	name string
}

// New creates and initializes a new virtual keyboard device.
// This opens /dev/uinput and configures it as a keyboard.
func New(ctx context.Context) (*Device, error) {
	log := logger.LogFromCtx(ctx)
	log.Info("creating virtual keyboard device", "name", DeviceName)

	// Open /dev/uinput
	fd, err := os.OpenFile("/dev/uinput", os.O_WRONLY|unix.O_NONBLOCK, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to open /dev/uinput: %w (do you have permissions?)", err)
	}

	d := &Device{
		fd:   fd,
		name: DeviceName,
	}

	// Setup device capabilities and create the virtual device
	if err := d.setup(ctx); err != nil {
		d.Close()
		return nil, fmt.Errorf("device setup failed: %w", err)
	}

	log.Info("virtual keyboard device created successfully")
	return d, nil
}

// setup configures the uinput device with keyboard capabilities.
func (d *Device) setup(ctx context.Context) error {
	log := logger.LogFromCtx(ctx)

	// Enable key events (EV_KEY)
	if err := d.ioctl(UI_SET_EVBIT, uintptr(EvKey)); err != nil {
		return fmt.Errorf("set EV_KEY: %w", err)
	}

	// Enable synchronization events (EV_SYN)
	if err := d.ioctl(UI_SET_EVBIT, uintptr(EvSyn)); err != nil {
		return fmt.Errorf("set EV_SYN: %w", err)
	}

	// Enable all keyboard keys (KEY_RESERVED to KEY_MAX)
	// Note: KEY_MAX is typically 0x2ff (767)
	for key := uint16(KeyReserved); key <= 767; key++ {
		if err := d.ioctl(UI_SET_KEYBIT, uintptr(key)); err != nil {
			// Some keys might fail, log but continue
			log.Debug("failed to enable key", "keycode", key, "error", err)
		}
	}

	// Configure device setup structure
	// This uses UI_DEV_SETUP ioctl (kernel >= 4.5)
	setup := uiSetup{
		ID: inputID{
			Bustype: BusVirtual,
			Vendor:  VendorID,
			Product: ProductID,
			Version: Version,
		},
		FFEffectsMax: 0,
	}
	copy(setup.Name[:], d.name)

	// Write setup structure
	if err := d.ioctlSetup(&setup); err != nil {
		return fmt.Errorf("UI_DEV_SETUP: %w", err)
	}

	// Create the device (makes it available to the system)
	if err := d.ioctl(UI_DEV_CREATE, 0); err != nil {
		return fmt.Errorf("UI_DEV_CREATE: %w", err)
	}

	return nil
}

// Close destroys the virtual device and closes the file descriptor.
func (d *Device) Close() error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.fd == nil {
		return nil
	}

	// Destroy the device
	if err := d.ioctl(UI_DEV_DESTROY, 0); err != nil {
		// Log but don't fail - we still want to close the fd
		fmt.Fprintf(os.Stderr, "warning: UI_DEV_DESTROY failed: %v\n", err)
	}

	// Close file descriptor
	err := d.fd.Close()
	d.fd = nil
	return err
}

// ioctl performs an ioctl system call on the device.
func (d *Device) ioctl(req, arg uintptr) error {
	_, _, errno := unix.Syscall(unix.SYS_IOCTL, d.fd.Fd(), req, arg)
	if errno != 0 {
		return errno
	}
	return nil
}

// ioctlSetup performs UI_DEV_SETUP ioctl with uiSetup structure.
func (d *Device) ioctlSetup(setup *uiSetup) error {
	_, _, errno := unix.Syscall(
		unix.SYS_IOCTL,
		d.fd.Fd(),
		uintptr(UI_DEV_SETUP),
		uintptr(unsafe.Pointer(setup)),
	)
	if errno != 0 {
		return errno
	}
	return nil
}

// uiSetup is the structure for UI_DEV_SETUP ioctl.
// See: <linux/uinput.h> struct uinput_setup
type uiSetup struct {
	ID           inputID
	Name         [80]byte
	FFEffectsMax uint32
}

// inputID identifies the device (bustype, vendor, product, version).
// See: <linux/input.h> struct input_id
type inputID struct {
	Bustype uint16
	Vendor  uint16
	Product uint16
	Version uint16
}
