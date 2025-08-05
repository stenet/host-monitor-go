# Host Monitor

Ein Cross-Platform System-Monitoring-Tool, das Systemmetriken sammelt und an einen Seq-Server sendet.

## Features

- **Cross-Platform**: Unterstützt Linux, Windows und macOS
- **Container-Ready**: Docker-Unterstützung mit Host-Monitoring
- **Systemmetriken**: CPU, Memory, Disk, Netzwerk und TCP-Verbindungen
- **Windows Service**: Kann als Windows-Service installiert werden
- **Seq-Integration**: Sendet strukturierte Logs an Seq-Server im CLEF-Format
- **Konfigurierbar**: Anpassbare Überwachungsintervalle und Seq-URL

## Installation

### Binaries herunterladen

Lade die neueste Version aus den [Releases](../../releases) herunter.

### Aus Quellcode kompilieren

```bash
git clone https://github.com/stenet/host-monitor-go.git
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
# Standard-Ausführung (15s Intervall, http://seq:5341)
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
sc query HostMonitor
```

### Docker Container

```bash
# Mit Standard-Konfiguration
docker run -v /:/host:ro stefanheim/host-monitor:latest

# Mit angepassten Parametern
docker run -v /:/host:ro -e SEQ_URL=http://seq-server:5341 stefanheim/host-monitor:latest
```

### Docker Compose

```yaml
version: '3.8'

services:
  host-monitor:
    image: stefanheim/host-monitor:latest
    container_name: host-monitor
    restart: unless-stopped
    pid: host                    # Zugriff auf Host-Prozesse
    network_mode: host           # Zugriff auf Host-Netzwerk-Interfaces
    volumes:
      - /:/host:ro               # Host-Filesystem für Monitoring (read-only)
      - /proc:/host/proc:ro      # Host-Prozess-Informationen
      - /sys:/host/sys:ro        # Host-System-Informationen
    environment:
      - SEQ_URL=http://seq:5341
      - INTERVAL=15s
```

## Parameter

| Parameter | Beschreibung | Standard |
|-----------|--------------|----------|
| `--seq-url` | URL des Seq-Servers | `http://seq:5341` |
| `--interval` | Überwachungsintervall | `15s` |
| `--debug`, `-d` | Debug-Modus (Konsolen-Ausgabe) | `false` |
| `--install` | Windows Service installieren | - |
| `--uninstall` | Windows Service deinstallieren | - |
| `--service-name` | Name des Windows Service | `HostMonitor` |

### Umgebungsvariablen

| Variable | Beschreibung | Standard |
|----------|--------------|----------|
| `SEQ_URL` | URL des Seq-Servers | `http://seq:5341` |
| `INTERVAL` | Überwachungsintervall | `15s` |

## Konfigurationsdatei (Optional)

Eine optionale `config.json` Datei kann im gleichen Verzeichnis wie die Anwendung erstellt werden, um zusätzliche Überwachungsoptionen zu konfigurieren:

```json
{
  "disk": "/custom/path",
  "processes": [
    "nginx",
    "mysql",
    "redis"
  ]
}
```

### Konfigurationsoptionen

| Option | Beschreibung | Standard |
|--------|--------------|----------|
| `disk` | Pfad zur zu überwachenden Disk/Partition | `/` (Linux/macOS) oder `C:\` (Windows) |
| `processes` | Liste von Prozessnamen zur Überwachung | Keine (keine Prozessüberwachung) |

### Prozessüberwachung

- **Processes_Not_Running_Count**: Anzahl der nicht laufenden Prozesse (0 wenn keine Prozesse konfiguriert)
- **Processes_Not_Running**: Array mit Namen der nicht laufenden Prozesse (nur wenn welche fehlen)
- Prozessnamen sind case-insensitive
- Auf Windows wird `.exe` automatisch ignoriert

## Überwachte Metriken

### CPU
- CPU-Auslastung in Prozent
- Plattform-spezifische Implementierung

### Memory
- Verwendeter Speicher in MB
- Auslastung in Prozent

### Disk
- Freier Speicherplatz in GB
- Auslastung in Prozent
- Konfigurierbare Disk/Partition über config.json

### Netzwerk
- Übertragungsraten in Bytes pro Sekunde (RX/TX)
- Ohne Loopback-Interfaces

### TCP-Verbindungen
- Anzahl aktiver TCP-Verbindungen

### Prozesse (Optional)
- Anzahl der nicht laufenden konfigurierten Prozesse
- Liste der nicht laufenden Prozesse

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