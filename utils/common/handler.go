package common

import (
	"fmt"
	"os"
	"os/signal"

	log "github.com/Sirupsen/logrus"
	"github.com/aporeto-inc/trireme"
	"github.com/aporeto-inc/trireme/cmd/remoteenforcer"
	"github.com/aporeto-inc/trireme/cmd/systemdutil"
	"github.com/aporeto-inc/trireme/enforcer"
	"github.com/aporeto-inc/trireme/monitor"
	"github.com/aporeto-inc/trireme/monitor/cliextractor"
	"github.com/aporeto-inc/trireme/monitor/dockermonitor"
	"github.com/aporeto-inc/trireme/processmon"
)

// KillContainerOnError defines if the Container is getting killed if the policy Application resulted in an error
const KillContainerOnError = true

// ProcessArgs handles all commands options for trireme
func ProcessArgs(arguments map[string]interface{}, processor enforcer.PacketProcessor) (err error) {

	if arguments["enforce"].(bool) {
		// Run enforcer and exit
		return remoteenforcer.LaunchRemoteEnforcer(processor)
	}

	if arguments["run"].(bool) || arguments["<cgroup>"] != nil {
		// Execute a command or process a cgroup cleanup and exit
		return systemdutil.ExecuteCommand(arguments)
	}

	if !arguments["daemon"].(bool) {
		log.Error("Invalid parameters")
		return fmt.Errorf("Invalid parameters")
	}

	// Trireme Daemon Commands
	processDaemonArgs(arguments, processor)
	return nil
}

// processDaemonArgs is responsible for creating a trireme daemon
func processDaemonArgs(arguments map[string]interface{}, processor enforcer.PacketProcessor) {

	var t trireme.Trireme
	var m monitor.Monitor
	var rm monitor.Monitor
	var err error
	var customExtractor dockermonitor.DockerMetadataExtractor

	// Setup external processors
	ExternalProcessor = processor

	// Setup incoming args
	processmon.GlobalCommandArgs = arguments

	if arguments["--swarm"].(bool) {
		log.WithFields(log.Fields{
			"Package":   "main",
			"Extractor": "Swarm",
		}).Info("Using Docker Swarm extractor")
		customExtractor = SwarmExtractor
	} else if arguments["--extractor"].(bool) {
		extractorfile := arguments["<metadatafile>"].(string)
		log.WithFields(log.Fields{
			"Package":   "main",
			"Extractor": extractorfile,
		}).Info("Using custom extractor")
		customExtractor, err = cliextractor.NewExternalExtractor(extractorfile)
		if err != nil {
			log.Fatalf("External metadata extractor cannot be accessed: %s", err)
		}
	}

	targetNetworks := []string{"172.17.0.0/24", "10.0.0.0/8"}
	if len(arguments["--target-networks"].([]string)) > 0 {
		log.WithFields(log.Fields{
			"Package":         "main",
			"target networks": arguments["--target-networks"].([]string),
		}).Info("Target Networks")
		targetNetworks = arguments["--target-networks"].([]string)
	}

	if !arguments["--hybrid"].(bool) {
		remote := arguments["--remote"].(bool)
		if arguments["--usePKI"].(bool) {
			keyFile := arguments["--keyFile"].(string)
			certFile := arguments["--certFile"].(string)
			caCertFile := arguments["--caCert"].(string)
			log.WithFields(log.Fields{
				"Package":      "main",
				"key-file":     keyFile,
				"cert-file":    certFile,
				"ca-cert-file": caCertFile,
			}).Info("Setting up trireme with PKI")
			t, m = TriremeWithPKI(keyFile, certFile, caCertFile, targetNetworks, &customExtractor, remote, KillContainerOnError)
		} else {
			log.Info("Setting up trireme with PSK")
			t, m = TriremeWithPSK(targetNetworks, &customExtractor, remote, KillContainerOnError)
		}
	} else { // Hybrid mode
		t, m, rm = HybridTriremeWithPSK(targetNetworks, &customExtractor, KillContainerOnError)
		if rm == nil {
			log.Fatalln("Failed to create remote monitor for hybrid")
		}
	}

	if t == nil {
		log.Fatalln("Failed to create Trireme")
	}

	if m == nil {
		log.Fatalln("Failed to create Monitor")
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	// Start services
	if err := t.Start(); err != nil {
		log.Fatalln("Failed to start Trireme")
	}

	if err := m.Start(); err != nil {
		log.Fatalln("Failed to start monitor")
	}

	if rm != nil {
		if err := rm.Start(); err != nil {
			log.Fatalln("Failed to start remote monitor")
		}
	}

	// Wait for Ctrl-C
	<-c

	fmt.Println("Bye!")
	m.Stop() // nolint
	t.Stop() // nolint
	if rm != nil {
		rm.Stop() // nolint
	}
}
