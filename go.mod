module tunesbus

go 1.26.4

require (
	github.com/go-ole/go-ole v1.3.0
	github.com/godbus/dbus/v5 v5.2.2
	github.com/quarckster/go-mpris-server v1.0.3
)

require golang.org/x/sys v0.27.0 // indirect

replace github.com/godbus/dbus/v5 => ../godbus-dbus
