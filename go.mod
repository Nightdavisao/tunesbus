module tunesbus

go 1.26.4

require (
	github.com/ammario/weakmap v0.3.1-0.20240402193103-1887e5a51caf
	github.com/charmbracelet/log v1.0.0
	github.com/go-ole/go-ole v1.3.0
	github.com/godbus/dbus/v5 v5.2.2
	github.com/pelletier/go-toml/v2 v2.4.3
	github.com/quarckster/go-mpris-server v1.0.3
)

require (
	github.com/aymanbagabas/go-osc52/v2 v2.0.1 // indirect
	github.com/charmbracelet/colorprofile v0.2.3-0.20250311203215-f60798e515dc // indirect
	github.com/charmbracelet/lipgloss v1.1.0 // indirect
	github.com/charmbracelet/x/ansi v0.8.0 // indirect
	github.com/charmbracelet/x/cellbuf v0.0.13-0.20250311204145-2c3ea96c31dd // indirect
	github.com/charmbracelet/x/term v0.2.1 // indirect
	github.com/go-logfmt/logfmt v0.6.1 // indirect
	github.com/lucasb-eyer/go-colorful v1.2.0 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/mattn/go-runewidth v0.0.16 // indirect
	github.com/muesli/termenv v0.16.0 // indirect
	github.com/rivo/uniseg v0.4.7 // indirect
	github.com/xo/terminfo v0.0.0-20220910002029-abceb7e1c41e // indirect
	golang.org/x/exp v0.0.0-20231006140011-7918f672742d // indirect
	golang.org/x/sys v0.30.0 // indirect
)

replace github.com/godbus/dbus/v5 => ./internal/dbus

replace github.com/quarckster/go-mpris-server => ./internal/mpris-server
