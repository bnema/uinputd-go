package styles

import "github.com/charmbracelet/lipgloss"

// Nerd Font symbols using Unicode
const (
	// Status symbols
	IconSuccess  = "\uf00c" // nf-fa-check
	IconError    = "\uf00d" // nf-fa-times
	IconWarning  = "\uf071" // nf-fa-exclamation_triangle
	IconInfo     = "\uf05a" // nf-fa-info_circle
	IconQuestion = "\uf059" // nf-fa-question_circle

	// Action symbols
	IconLock      = "\uf023" // nf-fa-lock
	IconUnlock    = "\uf09c" // nf-fa-unlock
	IconInstall   = "\uf019" // nf-fa-download
	IconUninstall = "\uf1f8" // nf-fa-trash
	IconConfig    = "\uf013" // nf-fa-cog
	IconKey       = "\uf084" // nf-fa-key

	// System symbols
	IconDaemon   = "\uf013" // nf-fa-cog
	IconService  = "\uf233" // nf-fa-server
	IconTerminal = "\uf120" // nf-fa-terminal
	IconKeyboard = "\uf11c" // nf-fa-keyboard

	// Network/Connection
	IconConnected  = "\uf0c1" // nf-fa-link
	IconDisconnect = "\uf127" // nf-fa-unlink
	IconSocket     = "\uf1e6" // nf-fa-plug

	// Progress
	IconSpinner = "\uf110" // nf-fa-spinner
	IconCheck   = "\uf00c" // nf-fa-check
	IconCross   = "\uf00d" // nf-fa-times

	// Misc
	IconBullet     = "\u2022" // bullet point
	IconArrowRight = "\uf061" // nf-fa-arrow_right
	IconArrowDown  = "\uf063" // nf-fa-arrow_down
)

// Color styles
var (
	SuccessStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("10")).Bold(true) // Green
	ErrorStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Bold(true)  // Red
	WarningStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("11")).Bold(true) // Yellow
	InfoStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("12"))            // Blue
	DimStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))             // Gray
	BoldStyle    = lipgloss.NewStyle().Bold(true)
)

// Helper functions for common patterns
func Success(msg string) string {
	return SuccessStyle.Render(IconSuccess) + " " + msg
}

func Error(msg string) string {
	return ErrorStyle.Render(IconError) + " " + msg
}

func Warning(msg string) string {
	return WarningStyle.Render(IconWarning) + " " + msg
}

func Info(msg string) string {
	return InfoStyle.Render(IconInfo) + " " + msg
}

func Dim(msg string) string {
	return DimStyle.Render(msg)
}

func Bold(msg string) string {
	return BoldStyle.Render(msg)
}

// Section header
func Section(title string) string {
	return "\n" + BoldStyle.Render(title) + "\n"
}

// List item
func ListItem(msg string) string {
	return "  " + DimStyle.Render(IconBullet) + " " + msg
}

// Step (numbered)
func Step(num int, msg string) string {
	return "  " + InfoStyle.Render(string(rune('0'+num))) + ". " + msg
}
