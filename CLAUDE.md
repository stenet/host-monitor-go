# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Projekt-Überblick

Dies ist ein System-Monitoring-Tool in Go, das Systemmetriken sammelt und an einen Seq-Server sendet. Das Tool überwacht CPU, Memory, Disk, Netzwerk und TCP-Verbindungen auf Linux, Windows und macOS.

## Build-Befehle

```bash
# Standard-Build für alle Plattformen
./build.sh

# Build mit spezifischer Version
VERSION=1.2.0 ./build.sh

# Build für aktuelle Plattform
go build -o host-monitor .

# Build mit Optimierungen (ohne Debug-Infos)
go build -ldflags "-w -s" -o host-monitor .
```

## Windows Service Installation

```bash
# Service installieren (Windows)
./host-monitor.exe --install

# Service mit anderem Namen installieren
./host-monitor.exe --install --service-name MyMonitor

# Service deinstallieren
./host-monitor.exe --uninstall

# Service Status überprüfen
sc query HostMonitor

# Service manuell starten/stoppen
sc start HostMonitor
sc stop HostMonitor
```

## Entwicklungs-Befehle

```bash
# Programm im Debug-Modus ausführen
go run . -d
go run . --debug

# Mit angepassten Parametern ausführen
go run . --seq-url http://localhost:5341 --interval 30s

# Dependencies aktualisieren
go mod tidy
go mod download

# Code formatieren
go fmt ./...

# Statische Code-Analyse
go vet ./...
```

## Architektur

### Kernkomponenten

1. **main.go**: Enthält die gesamte Anwendungslogik
   - `SystemMetrics` struct: Datenstruktur für alle gesammelten Metriken
   - `main()`: Initialisiert Flags, startet Monitoring-Loop oder Service
   - `collectMetrics()`: Zentrale Funktion zum Sammeln aller Metriken
   - Plattform-spezifische CPU-Funktionen für Linux, Windows und macOS

2. **service_windows.go**: Windows Service Implementation
   - `windowsService` struct: Service Handler für Windows
   - Service Installation/Deinstallation
   - Service-spezifisches Monitoring mit Lifecycle-Management

3. **service_stub.go**: Plattform-Stubs für Nicht-Windows-Systeme
   - Leere Implementierungen für Linux/macOS Builds

### Plattform-Unterstützung

- **Linux**: Liest CPU-Stats aus `/proc/stat`
- **Windows**: Verwendet Windows API `GetSystemTimes`
- **macOS**: Nutzt Mach system calls und sysctl

### Externe Dependencies

- `github.com/shirou/gopsutil/v3`: Cross-platform System-Monitoring-Library
  - Verwendet für Memory, Disk, Network und TCP-Connection Metriken

### Datenfluss

1. Programm startet mit konfigurierbarem Intervall (default: 15s)
2. Initial-Messungen für CPU und Netzwerk werden erfasst
3. In jedem Intervall:
   - Neue Messungen werden genommen
   - Differenzen werden berechnet (für Rate-basierte Metriken)
   - Metriken werden gesammelt und strukturiert
   - Im Debug-Modus: Ausgabe auf Console
   - Im Normal-Modus: JSON-Post an Seq-Server

### Seq-Integration

- Sendet Metriken im CLEF-Format an Seq (`/ingest/clef` Endpoint)
- Strukturierte Logs mit `@t` (Timestamp) und `@mt` (Message Template)
- Fehlerbehandlung mit deutschen Fehlermeldungen

## Code-Konventionen

- Fehlerbehandlung erfolgt meist durch Fallback auf Null-Werte statt Panic
- Deutsche Fehlermeldungen in der Konsole
- Englische Feldnamen in JSON/Structs
- CGO wird für macOS CPU-Metriken verwendet