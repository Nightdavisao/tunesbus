//go:build windows

package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"sync"
	"syscall"
	"time"
	"tunesbus/internal/itunes"
	"tunesbus/internal/wine"

	"github.com/ammario/weakmap"

	"github.com/charmbracelet/log"

	"github.com/go-ole/go-ole"
	"github.com/quarckster/go-mpris-server/pkg/events"
	"github.com/quarckster/go-mpris-server/pkg/server"
	"github.com/quarckster/go-mpris-server/pkg/types"
)

type WeakTrackArtworkCache struct {
	store weakmap.Map[int64, string]
}

type CmdArguments struct {
	diagnosticsOnly *bool
}

type OnceGroup struct {
	quitOnce        sync.Once
	mprisStartOnce  sync.Once
	initialSyncOnce sync.Once
}

type MainState struct {
	arguments CmdArguments
	config    *ProgramConfig
	tunesDisp *ole.IDispatch

	mux  sync.RWMutex
	sync OnceGroup

	currentMetadata *types.Metadata
	playbackState   PlaybackState

	server       *server.Server
	mprisHandler *events.EventHandler

	comSink *itunes.COMEventSink

	ticker *time.Ticker
	quit   chan struct{}

	artworkCache WeakTrackArtworkCache
}

func (state *MainState) ensureMprisStarted() {
	state.sync.mprisStartOnce.Do(func() {
		go state.startServingBus(state.server)
	})
}

func (state *MainState) waitForMprisReady(timeout time.Duration) bool {
	if state.server == nil {
		return false
	}
	if state.server.Conn != nil {
		return true
	}
	if timeout <= 0 {
		select {
		case <-state.server.Ready():
			return true
		default:
			return false
		}
	}

	timer := time.NewTimer(timeout)
	defer timer.Stop()

	select {
	case <-state.server.Ready():
		return true
	case <-state.quit:
		return false
	case <-timer.C:
		return state.server.Conn != nil
	}
}

func (state *MainState) emitInitialMprisState() {
	state.sync.initialSyncOnce.Do(func() {
		state.mprisHandler.Player.OnAll()
	})
}

func (state *MainState) startServingBus(s *server.Server) {
	log.Info("starting MPRIS server...")
	err := s.Listen()

	if err != nil {
		state.QuitSafely(err, "startMprisServer failed, quitting")
	}
}

func (state *MainState) startTicker() {
	optionsTicker := time.NewTicker(2 * time.Second)
	defer optionsTicker.Stop()

	for {
		select {
		case <-state.quit:
			return
		case <-state.ticker.C:
			if !state.waitForMprisReady(0) {
				continue
			}
			state.emitPlaybackChanges(state.refreshPlaybackState(false))
		case <-optionsTicker.C:
			if !state.waitForMprisReady(0) {
				continue
			}
			state.emitPlaybackChanges(state.refreshPlaybackState(true))
		}
	}
}

func (state *MainState) ParseArgs() {
	debugModePtr := flag.Bool("debug", false, "Enable debug logging")

	state.arguments.diagnosticsOnly = flag.Bool("diagnostics", false, "Show diagnostics and quit.")
	flag.Parse()

	if *debugModePtr {
		log.SetLevel(log.DebugLevel)
		return
	}
}

const TITLE_WINDOW = "tunesbus"

func main() {
	runtime.LockOSThread()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt, syscall.SIGTERM)

	state := &MainState{
		ticker:          time.NewTicker(500 * time.Millisecond),
		currentMetadata: &types.Metadata{},
		artworkCache: WeakTrackArtworkCache{
			store: weakmap.Map[int64, string]{},
		},
		quit: make(chan struct{}),
	}
	go func() {
		<-sigs
		state.QuitSafely(nil, "")
	}()

	var err error
	
	err = ParseConfigFile()
	if err != nil {
		log.Error("failed to parse config file, will use defaults", "error", err)
	}
	state.config = programConfig
	state.ParseArgs()


	if *state.arguments.diagnosticsOnly {
		text := ""

		wineVersion, err := wine.GetWineVersion()
		if err != nil {
			wine.ErrorMessageBox(
				TITLE_WINDOW,
				fmt.Sprintf("Error on getting Wine version: %v\n", err),
			)
			return
		}
		text = text + fmt.Sprintf("Wine version: %s\n", wineVersion)

		wineBuild, err := wine.GetWineBuild()
		if err != nil {
			wine.ErrorMessageBox(
				TITLE_WINDOW,
				fmt.Sprintf("Error on getting Wine build ID: %v\n", err),
			)
			return
		}
		text = text + fmt.Sprintf("Wine build: %s\n", wineBuild)

		tmpDir, err := wine.UnixTmpDirAsDosPath()
		if err != nil {
			wine.ErrorMessageBox(
				TITLE_WINDOW,
				fmt.Sprintf("Error on getting temporary Unix directory: %v\n", err),
			)
			return
		}
		text = text + fmt.Sprintf("Temporary Unix directory (as DOS): %s\n", tmpDir)
		text = text + fmt.Sprintf("WINEPREFIX env: %s\n", wine.GetWinePrefix())

		wine.InfoMessageBox(TITLE_WINDOW, text)
		return
	}


	state.tunesDisp, err = itunes.NewTunesDispatch()
	if err != nil {
		state.QuitSafely(err, "failed to initialize dispatcher")
		return
	}

	busRoot := BusRoot{
		state: state,
	}

	busPlayer := BusPlayer{
		state: state,
	}

	state.server = server.NewServer(state.config.MPRIS.BusNameSuffix, busRoot, &busPlayer)
	state.mprisHandler = events.NewEventHandler(state.server)

	handler := &tunesEventHandler{
		state:   state,
		handler: state.mprisHandler,
	}

	state.comSink, err = itunes.NewCOMEventSink(state.tunesDisp, handler)
	if err != nil {
		state.QuitSafely(err, "something failed when setting up the event sink")
	}

	go state.startTicker()

	err = state.comSink.ListenEvents()
	if err != nil {
		state.QuitSafely(err, "failed to listen for COM events")
	}

	log.Info("the end")
}

func (state *MainState) QuitSafely(err error, message string) {
	if state.server.Conn != nil {
		err := state.server.Stop()
		if err != nil {
			log.Warn("failed to stop the dbus server. you might want to force kill wineserver with \"wineserver -k\"...")
			wine.WarningMessageBox(
				"tunesbus",
				"Failed to stop the dbus server. You might want to force kill wineserver with the command below.\n\n"+
					fmt.Sprintf("WINEPREFIX=%s wineserver -k", wine.GetWinePrefix()),
			)
		}
	}
	if state.comSink != nil {
		state.comSink.DisconnectObject()
	}

	code := 0

	if err != nil {
		code = 1
		defer os.Exit(code)

		if message != "" {
			log.Error(message, "error", err)
			wine.ErrorMessageBox("tunesbus", message)
		} else {
			log.Error("quitting because of critical error", "error", err)
		}
		return
	} else {
		log.Info("now quitting...")
		defer os.Exit(code)
	}

	if state.tunesDisp != nil {
		state.tunesDisp.Release()
	}
	state.sync.quitOnce.Do(func() {
		close(state.quit)
	})
}
