//go:build windows

package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/mgr"
)

type windowsService struct {
	debug    bool
	seqURL   string
	interval time.Duration
}

func (ws *windowsService) Execute(args []string, r <-chan svc.ChangeRequest, s chan<- svc.Status) (bool, uint32) {
	s <- svc.Status{State: svc.StartPending}

	// Start monitoring in separate goroutine
	stopCh := make(chan struct{})
	go ws.runMonitoring(stopCh)

	s <- svc.Status{State: svc.Running, Accepts: svc.AcceptStop | svc.AcceptShutdown}

	for c := range r {
		switch c.Cmd {
		case svc.Interrogate:
			s <- c.CurrentStatus
		case svc.Stop, svc.Shutdown:
			s <- svc.Status{State: svc.StopPending}
			close(stopCh)
			return false, 0
		default:
			continue
		}
	}

	return false, 0
}

func (ws *windowsService) runMonitoring(stopCh <-chan struct{}) {
	hostname := getHostname()

	prevNetStats := getNetworkStats()
	prevCPUStats := getCPUStats()
	prevTime := time.Now()

	ticker := time.NewTicker(ws.interval)
	defer ticker.Stop()

	for {
		select {
		case <-stopCh:
			return
		case <-ticker.C:
			currNetStats := getNetworkStats()
			currCPUStats := getCPUStats()
			currTime := time.Now()

			timeDiff := currTime.Sub(prevTime).Seconds()
			metrics := collectMetrics(hostname, prevNetStats, currNetStats, prevCPUStats, currCPUStats, timeDiff)

			if ws.debug {
				printDebugMetrics(metrics)
			} else {
				sendToSeq(ws.seqURL, metrics)
			}

			prevNetStats = currNetStats
			prevCPUStats = currCPUStats
			prevTime = currTime
		}
	}
}

func runAsWindowsService(debug bool, seqURL string, interval time.Duration) {
	service := &windowsService{
		debug:    debug,
		seqURL:   seqURL,
		interval: interval,
	}

	err := svc.Run("SystemMonitor", service)
	if err != nil {
		log.Fatalf("Service lief nicht: %v", err)
	}
}

func isWindowsService() bool {
	isIntSess, err := svc.IsAnInteractiveSession()
	if err != nil {
		return false
	}
	return !isIntSess
}

func installWindowsService(serviceName, seqURL string, interval time.Duration, debug bool) {
	exePath, err := os.Executable()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fehler beim Ermitteln des Programmpfads: %v\n", err)
		return
	}

	m, err := mgr.Connect()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fehler beim Verbinden zum Service Manager: %v\n", err)
		return
	}
	defer m.Disconnect()

	s, err := m.OpenService(serviceName)
	if err == nil {
		s.Close()
		fmt.Printf("Service '%s' ist bereits installiert\n", serviceName)
		return
	}

	// Build service arguments with all parameters
	serviceArgs := fmt.Sprintf("%s --seq-url \"%s\" --interval %s", exePath, seqURL, interval)
	if debug {
		serviceArgs += " --debug"
	}

	s, err = m.CreateService(serviceName, serviceArgs, mgr.Config{
		DisplayName: "System Monitor Service",
		Description: "Sammelt Systemmetriken und sendet sie an Seq",
		StartType:   mgr.StartAutomatic,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fehler beim Erstellen des Service: %v\n", err)
		return
	}
	defer s.Close()

	fmt.Printf("Service '%s' erfolgreich installiert mit Seq-URL: %s\n", serviceName, seqURL)
	fmt.Printf("Interval: %s, Debug-Modus: %t\n", interval, debug)
	fmt.Println("Starten Sie den Service mit: sc start", serviceName)
}

func uninstallWindowsService(serviceName string) {
	m, err := mgr.Connect()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fehler beim Verbinden zum Service Manager: %v\n", err)
		return
	}
	defer m.Disconnect()

	s, err := m.OpenService(serviceName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Service '%s' nicht gefunden: %v\n", serviceName, err)
		return
	}
	defer s.Close()

	status, err := s.Query()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fehler beim Abfragen des Service-Status: %v\n", err)
		return
	}

	if status.State != svc.Stopped {
		fmt.Printf("Service '%s' wird gestoppt...\n", serviceName)
		_, err = s.Control(svc.Stop)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Fehler beim Stoppen des Service: %v\n", err)
			return
		}

		timeout := time.Now().Add(30 * time.Second)
		for status.State != svc.Stopped {
			if time.Now().After(timeout) {
				fmt.Fprintf(os.Stderr, "Timeout beim Stoppen des Service\n")
				return
			}
			time.Sleep(300 * time.Millisecond)
			status, err = s.Query()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Fehler beim Abfragen des Service-Status: %v\n", err)
				return
			}
		}
	}

	err = s.Delete()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fehler beim Löschen des Service: %v\n", err)
		return
	}

	fmt.Printf("Service '%s' erfolgreich deinstalliert\n", serviceName)
}
