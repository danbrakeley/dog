package main

import (
	"flag"
	"math/rand"
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

	log.Info("first")
	time.Sleep(time.Second)
	log.Info("second")
	time.Sleep(time.Second)
	log.Info("third")
	time.Sleep(time.Second)
	log.Transient("this is a Transient log line")
	log.Verbose("this is a Verbose log line")
	log.Info("this is an Info log line")
	log.Warning("this is a Warning log line")
	log.Error("this is an Error log line")
	// log.Fatal("this is a Fatal log line")
	time.Sleep(time.Second)

	var rando int
	for {
		msg := pick(msgContents)

		fields := []dog.Fielder{}
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
			fields = append(fields, f)
		}

		rando = rand.Intn(40)
		switch {
		case rando < 1:
			log.Error(msg, fields...)
		case rando < 4:
			log.Warning(msg, fields...)
		case rando < 6:
			log.Verbose(msg, fields...)
		default:
			log.Info(msg, fields...)
		}

		if rand.Intn(20) != 0 {
			continue
		}
		time.Sleep(time.Duration(rand.Intn(2000)+200) * time.Millisecond)
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
