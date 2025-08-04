# Host Monitor Go

Ein Cross-Platform System-Monitoring-Tool, das Systemmetriken sammelt und an einen Seq-Server sendet.

## Features

- **Cross-Platform**: Unterstützt Linux, Windows und macOS
- **Systemmetriken**: CPU, Memory, Disk, Netzwerk und TCP-Verbindungen
- **Windows Service**: Kann als Windows-Service installiert werden
- **Seq-Integration**: Sendet strukturierte Logs an Seq-Server
- **Konfigurierbar**: Anpassbare Überwachungsintervalle und Seq-URL

## Installation

### Binaries herunterladen

Lade die neueste Version aus den [Releases](../../releases) herunter.

### Aus Quellcode kompilieren

```bash
git clone https://github.com/USERNAME/host-monitor-go.git
cd host-monitor-go
go build -o host-monitor .
```

### Cross-Platform Build

```bash
# Alle Plattformen
./build.sh

# Mit spezifischer Version
VERSION=1.0.0 ./build.sh
```

## Verwendung

### Grundlegende Verwendung

```bash
# Standard-Ausführung (15s Intervall, localhost:5341)
./host-monitor

# Mit angepassten Parametern
./host-monitor --seq-url http://your-seq-server:5341 --interval 30s

# Debug-Modus (Ausgabe auf Konsole)
./host-monitor --debug
```

### Windows Service

```bash
# Service installieren
./host-monitor.exe --install

# Service mit anderem Namen installieren
./host-monitor.exe --install --service-name MyMonitor

# Service deinstallieren
./host-monitor.exe --uninstall

# Service-Status prüfen
sc query SystemMonitor
```

## Parameter

| Parameter | Beschreibung | Standard |
|-----------|--------------|----------|
| `--seq-url` | URL des Seq-Servers | `http://localhost:5341` |
| `--interval` | Überwachungsintervall | `15s` |
| `--debug` | Debug-Modus (Konsolen-Ausgabe) | `false` |
| `--install` | Windows Service installieren | - |
| `--uninstall` | Windows Service deinstallieren | - |
| `--service-name` | Name des Windows Service | `SystemMonitor` |

## Überwachte Metriken

### CPU
- CPU-Auslastung in Prozent
- Plattform-spezifische Implementierung

### Memory
- Gesamter verfügbarer Speicher
- Verwendeter Speicher
- Freier Speicher
- Auslastung in Prozent

### Disk
- Gesamter Speicherplatz
- Verwendeter Speicherplatz
- Freier Speicherplatz
- Auslastung in Prozent

### Netzwerk
- Bytes gesendet/empfangen
- Pakete gesendet/empfangen
- Übertragungsraten

### TCP-Verbindungen
- Anzahl aktiver TCP-Verbindungen

## Entwicklung

### Voraussetzungen

- Go 1.21 oder höher
- Für Windows Service: Windows-Entwicklungsumgebung

### Build-Befehle

```bash
# Entwicklung
go run . --debug

# Standard-Build
go build -o host-monitor .

# Optimierter Build
go build -ldflags "-w -s" -o host-monitor .

# Tests ausführen
go test ./...

# Code formatieren
go fmt ./...

# Statische Analyse
go vet ./...
```

### Architektur

- **main.go**: Hauptanwendungslogik und Metriken-Sammlung
- **service_windows.go**: Windows Service-Implementation
- **service_stub.go**: Plattform-Stubs für Nicht-Windows-Systeme
- **cpu_*.go**: Plattform-spezifische CPU-Monitoring-Implementierungen

## Dependencies

- [gopsutil](https://github.com/shirou/gopsutil): Cross-platform System-Monitoring-Library

## Lizenz

[MIT License](LICENSE)