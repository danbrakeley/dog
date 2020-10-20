package main

import (
	"flag"
	"fmt"
	"math/rand"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/danbrakeley/dog"
)

var addr = flag.String("addr", "localhost:8080", "http service address")

func main() {
	flag.Parse()

	log, err := dog.Create(*addr)
	if err != nil {
		panic(err)
	}
	defer log.Close()
	log.SetMinLevel(dog.Transient)

	// setup a channel to take in client messages as they arrive
	chMsg := make(chan dog.WsMsg, 1024)
	dog.SetClientMsgHandler(log, func(m dog.WsMsg) {
		chMsg <- m
	})

	// setup a channel for commands to be run
	chCmd := make(chan command, 1024)

	// translate client messages to commands
	reNumber := regexp.MustCompile(`^([0-9]+)$`)
	rePrint := regexp.MustCompile(`^s ([0-9]+)$`)

	// start client message processor
	go func() {
		for {
			msg := <-chMsg
			strMsg := strings.TrimSpace(string(msg.Msg))

			// help
			if strMsg == "help" {
				sendHelp(chCmd)
				continue
			}

			// fatal
			if strMsg == "fatal" {
				log.Fatal("interrupting cow")
				continue
			}

			// just a number
			if reNumber.MatchString(strMsg) {
				n, err := strconv.ParseInt(strMsg, 10, 32)
				if err != nil {
					fmt.Printf("unable to parse number: %v", err)
					continue
				}
				addRandos(chCmd, int(n), false)
				continue
			}

			// "p <number>"
			m := rePrint.FindSubmatch(msg.Msg)
			if len(m) == 2 {
				n, err := strconv.ParseInt(string(m[1]), 10, 32)
				if err != nil {
					fmt.Printf("unable to parse number: %v", err)
					continue
				}
				addRandos(chCmd, int(n), true)
				continue
			}
		}
	}()

	chCmd <- command{Type: "info", Msg: "three..."}
	chCmd <- command{Type: "sleep", Dur: time.Second}
	chCmd <- command{Type: "info", Msg: "two..."}
	chCmd <- command{Type: "sleep", Dur: time.Second}
	chCmd <- command{Type: "info", Msg: "one..."}
	chCmd <- command{Type: "sleep", Dur: time.Second}
	chCmd <- command{Type: "transient", Msg: "this is a Transient log line"}
	chCmd <- command{Type: "verbose", Msg: "this is a Verbose log line"}
	chCmd <- command{Type: "info", Msg: "this is a Info log line"}
	chCmd <- command{Type: "warning", Msg: "this is a Warning log line"}
	chCmd <- command{Type: "error", Msg: "this is a Error log line"}
	sendHelp(chCmd)

	// run commands as they come in
	for cmd := range chCmd {
		switch cmd.Type {
		case "sleep":
			time.Sleep(cmd.Dur)
		case "transient":
			log.Transient(cmd.Msg, cmd.Fields...)
		case "verbose":
			log.Verbose(cmd.Msg, cmd.Fields...)
		case "info":
			log.Info(cmd.Msg, cmd.Fields...)
		case "warning":
			log.Warning(cmd.Msg, cmd.Fields...)
		case "error":
			log.Error(cmd.Msg, cmd.Fields...)
		case "fatal":
			log.Fatal(cmd.Msg, cmd.Fields...)
		}
	}

}

type command struct {
	Type   string
	Dur    time.Duration
	Msg    string
	Fields []dog.Fielder
}

func sendHelp(chCmd chan command) {
	chCmd <- command{Type: "transient", Msg: "------------------------------------------------------------"}
	chCmd <- command{Type: "warning", Msg: "------------------------------------------------------------"}
	chCmd <- command{Type: "info", Msg: "- Available Commands:"}
	chCmd <- command{Type: "info", Msg: "- <num> == send <num> random log lines"}
	chCmd <- command{Type: "info", Msg: "- s <num> == send <num> log lines, pausing randomly (simulate a real app)"}
	chCmd <- command{Type: "info", Msg: "- fatal == send a fatal log line (will kill server)"}
	chCmd <- command{Type: "info", Msg: "- help == send this message"}
	chCmd <- command{Type: "error", Msg: "------------------------------------------------------------"}
	chCmd <- command{Type: "verbose", Msg: "------------------------------------------------------------"}
}

func addRandos(chCmd chan command, n int, includeSleeps bool) {
	var rando int
	for i := 0; i < n; i++ {
		cmd := command{
			Msg: pick(msgContents),
		}

		count := 0
		rando = rand.Intn(40)
		switch {
		case rando < 5:
			count = 1
		case rando < 8:
			count = 2
		case rando < 11:
			count = 3
		case rando < 12:
			count = 4
		default:
		}

		for i := 0; i < count; i++ {
			var f dog.Fielder
			switch rand.Intn(7) {
			case 0:
				f = dog.Bool(pick(fieldNames)+"_b", rand.Intn(2) == 1)
			case 1:
				f = dog.Int(pick(fieldNames)+"_int", rand.Intn(10000))
			case 2:
				f = dog.Float32(pick(fieldNames)+"_f32", rand.Float32())
			case 3:
				f = dog.Float64(pick(fieldNames)+"_f64", rand.Float64())
			case 4:
				f = dog.Dur(pick(fieldNames)+"_dur", time.Duration(rand.Intn(59000))*time.Millisecond)
			case 5:
				f = dog.Time(pick(fieldNames)+"_time", time.Now().Add(-time.Duration(rand.Intn(1000)+1)*time.Minute))
			default:
				f = dog.String(pick(fieldNames)+"_str", pick(fieldNames))
			}
			cmd.Fields = append(cmd.Fields, f)
		}

		rando = rand.Intn(40)
		switch {
		case rando < 1:
			cmd.Type = "error"
		case rando < 4:
			cmd.Type = "warning"
		case rando < 6:
			cmd.Type = "verbose"
		default:
			cmd.Type = "info"
		}

		chCmd <- cmd

		if includeSleeps && rand.Intn(20) == 0 {
			chCmd <- command{
				Type: "sleep",
				Dur:  time.Duration(rand.Intn(2000)+200) * time.Millisecond,
			}
		}
	}
}

func pick(strs []string) string {
	return strs[rand.Intn(len(strs))]
}

var fieldNames = []string{
	"player_x", "player_v", "player_eta", "x", "y", "z", "loops", "count", "line", "source",
	"chips", "doors", "dimension", "auroa_of_importance", "interested_parties",
	"name", "address", "ssn", "homepage", "favorite_food_color", "file", "diskettes",
	"markers", "gold_coins", "gold", "g", "boomer_status", "true_gamer", "internal_organs",
}

var msgContents = []string{
	"Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum.",
	"What Is The Answer and Why?",
	"I Blame Your Mother",
	"What Are The Civilian Applications?",
	"Clear Air Turbulence",
	"Fate Amenable To Change",
	"Frank Exchange Of Views",
	"Anticipation Of A New Lover's Arrival, The",
	"Mounting plugin MinimapPlugin",
	"Using libcurl 7.55.1-DEV",
	"WinSock: version 1.1 (2.2), MaxSocks=32767, MaxUdp=65467",
	"OS: Windows 10 (Release 2004) (), CPU: Intel(R) Core(TM) i7-8700K CPU @ 3.70GHz, GPU: NVIDIA GeForce GTX 1080",
	"Presizing for max 2097152 objects, including 60000 objects not considered by GC, pre-allocating 9100000 bytes for permanent pool.",
	"Setting CVar [[s.ContinuouslyIncrementalGCWhileLevelsPendingPurge:0]]",
	"Setting CVar [[s.LevelStreamingActorsUpdateTimeLimit:1.5]]",
	"High frequency timer resolution =10.000000 MHz",
	"Texture pool is 5655 MB (70% of 8079 MB)",
	"Cooked Context: Using Shared Shader Library Global",
	"Failed to load Shared Shader Library: ControlRig and no native library supported.",
	"Opened pipeline cache after state change and enqueued 0 of 0 tasks for precompile.",
	"VoiceManager status is now: LoggedOut",
	"Marked systems 0000000010 as loaded. Loaded systems have changed from 0000000000 to 0000000010.",
	"Match State Changed from EnteringMap to WaitingToStart",
	"UAkGameplayStatics::PostEvent: No Event specified!",
	"Shutting down and abandoning module AnimationBudgetAllocator (114)",
	"Updating window title bar state: overlay mode, drag disabled, window buttons hidden, title bar hidden",
	"I am a firm believer in the people. If given the truth, they can be depended upon to meet any national crisis. The great point is to bring them the real facts.",
	"gpu_child_thread.cc(174)] Exiting GPU process due to errors during initialization",
	"waituntilobserver.cpp(148)] NOT IMPLEMENTED",
	"[.DisplayCompositor-0000027E94428560]RENDER WARNING: texture bound to texture unit 1 is not renderable. It maybe non-power-of-2 and have incompatible texture filtering.",
	`Verified acquired payload: PortalPrereqSetup at path: C:\ProgramData\Package Cache\.unverified\PortalPrereqSetup, moving to: C:\ProgramData\Package Cache\{F9C5C994-F6B9-4D75-B3E7-AD01B84073E9}v1.0.0.0\LauncherPrereqSetup_x64.msi.`,
	"Warning: Setting a new default format with a different version or profile after the global shared context is created may cause issues with context sharing.",
	"Animation Driver: using vsync: 6.94 ms",
	"libmpv_render: GL_SHADING_LANGUAGE_VERSION='OpenGL ES GLSL ES 3.00 (ANGLE 2.1.0.57ea533f79a7)'",
	"[Web] [Workers] Initializing...",
	"[Web] [Sources] Finished initialization",
	"Checking for possible conflicting running processes... (attempt 1)",
	"Collected all directories and file handles",
	"Distributing 13 actions to XGE",
	"Set ProjectVersion to 1.0.0.0. Version Checksum will be recalculated on next use.",
	"No Audio Capture implementations found. Audio input will be silent.",
	"The referenced script on this Behaviour (Game Object 'fiBackupSceneStorage') is missing!",
	"Error: name must not be falsy",
	"Setting up 6 worker threads for Enlighten.",
	`NullReferenceException: Object reference not set to an instance of an object
at GizmoScript.OnDestroy () [0x00027] in <55abb9e2b16e4a54b84ade9dcc6e7b40>:0`,
	"Unloading 280 unused Assets to reduce memory usage. Loaded Objects now: 54405.",
	"Shader Unsupported: 'Hidden/PostProcessing/Uber' - Pass '' has no vertex shader",
	"Serialization depth limit 7 exceeded at 'GridButton.Thumbnail'. There may be an object composition cycle in one or more of your serialized classes.",
	"failed to read application settings, reverting to default.",
	"Running benchmark, system scan included, and monitoring included",
	"foo",
	"Could not open service McShield for query, start and stop. McAfee may not be installed, or we don't have access.",
	"mbox yesno: You should not choose a root path that include spaces in directory names.  Proceed anyway?",
	"Application: Opened default audio device with 44.1khz / 16 bit stereo audio, 2048 sample size buffer",
	"UniverseClient: Client disconnecting...",
	"OpenGL version: '4.5.14008 Compatibility Profile Context 21.19.137.1' vendor: 'ATI Technologies Inc.' renderer: 'AMD Radeon R9 200 Series' shader: '4.50'",
	"Application: stopped gracefully",
	`PFRO Error: \??\C:\Program Files (x86)\Mozilla Maintenance Service\maintenanceservice_tmp.exe, |delete operation|, 0xc0000034`,
	`PFRO Error: \??\C:\Program Files\Mozilla Firefox\tobedeleted\mozd0c28138-bfd0-4572-a13f-7e4e96dbf32d, |delete operation|, 0xc000003a`,
	"Please run the Get-WindowsUpdateLog PowerShell command to convert ETW traces into a readable WindowsUpdate.log.",
	"Successfully Submitted Heartbeat Report",
	"No infection found.",
	"NetpDoDomainJoin: using new computer names",
	"[sti_ci.dll] GetOptionalDevicePropertyString, SetupDiGetDeviceProperty() failed. Err=0x490.",
	"**************** Started trace for Module: [wiaservc.dll] in Executable [svchost.exe] ProcessID: [4476] at 2020/09/30 10:24:13:485 ****************",
}
