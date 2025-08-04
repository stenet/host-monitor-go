package main

import (
    "bytes"
    "encoding/json"
    "flag"
    "fmt"
    "io"
    "net/http"
    "os"
    "runtime"
    "strings"
    "time"

    "github.com/shirou/gopsutil/v3/disk"
    "github.com/shirou/gopsutil/v3/mem"
    "github.com/shirou/gopsutil/v3/net"
)

type SystemMetrics struct {
    Timestamp       string  `json:"@t"`
    MessageTemplate string  `json:"@mt"`
    Application     string  `json:"Application"`
    Hostname        string  `json:"Hostname"`
    CPUPercent      float64 `json:"CPU_Percent"`
    MemoryPercent   float64 `json:"Memory_Percent"`
    MemoryMB        float64 `json:"Memory_MB"`
    DiskPercent     float64 `json:"Disk_Percent"`
    DiskFreeGB      float64 `json:"Disk_Free_GB"`
    NetworkRXBPS    uint64  `json:"Network_RX_BPS"`
    NetworkTXBPS    uint64  `json:"Network_TX_BPS"`
    TCPConnections  int     `json:"TCP_Connections"`
}

type NetworkStats struct {
    RXBytes uint64
    TXBytes uint64
}

type CPUStats struct {
    IdleTime  uint64 // in nanoseconds
    TotalTime uint64 // in nanoseconds
}

func main() {
    // Command line flags
    debug := flag.Bool("debug", false, "Enable debug mode")
    flag.BoolVar(debug, "d", false, "Enable debug mode (shorthand)")
    seqURL := flag.String("seq-url", getEnvWithDefault("SEQ_URL", "http://seq:5341"), "Seq server URL")
    
    intervalDefault := getEnvWithDefault("INTERVAL", "15s")
    intervalParsed, err := time.ParseDuration(intervalDefault)
    if err != nil {
        fmt.Fprintf(os.Stderr, "Fehler beim Parsen der INTERVAL-Umgebungsvariable: %v\n", err)
        intervalParsed = 15 * time.Second
    }
    interval := flag.Duration("interval", intervalParsed, "Monitoring interval")

    // Windows service flags
    installService := flag.Bool("install", false, "Install as Windows service")
    uninstallService := flag.Bool("uninstall", false, "Uninstall Windows service")
    serviceName := flag.String("service-name", "SystemMonitor", "Windows service name")

    flag.Parse()

    // Handle Windows service installation/uninstallation
    if runtime.GOOS == "windows" {
        if *installService {
            if *seqURL == "" {
                fmt.Fprintf(os.Stderr, "Fehler: Seq-URL ist verpflichtend für Service-Installation. Verwenden Sie --seq-url.\n")
                os.Exit(1)
            }
            installWindowsService(*serviceName, *seqURL, *interval, *debug)
            return
        }
        if *uninstallService {
            uninstallWindowsService(*serviceName)
            return
        }

        // Check if running as Windows service
        if isWindowsService() {
            runAsWindowsService(*debug, *seqURL, *interval)
            return
        }
    }

    // Get hostname
    hostname := getHostname()

    // Initial measurements
    prevNetStats := getNetworkStats()
    prevCPUStats := getCPUStats()
    prevTime := time.Now()

    for {
        time.Sleep(*interval)

        // Current measurements
        currNetStats := getNetworkStats()
        currCPUStats := getCPUStats()
        currTime := time.Now()

        // Calculate time difference
        timeDiff := currTime.Sub(prevTime).Seconds()

        // Get system metrics
        metrics := collectMetrics(hostname, prevNetStats, currNetStats, prevCPUStats, currCPUStats, timeDiff)

        if *debug {
            printDebugMetrics(metrics)
        } else {
            sendToSeq(*seqURL, metrics)
        }

        // Update previous values
        prevNetStats = currNetStats
        prevCPUStats = currCPUStats
        prevTime = currTime
    }
}

func collectMetrics(hostname string, prevNet, currNet NetworkStats, prevCPU, currCPU CPUStats, timeDiff float64) SystemMetrics {
    // CPU usage - calculate percentage over time interval
    var cpuUsage float64

    // Overflow protection: ensure current values are greater than previous
    if currCPU.IdleTime >= prevCPU.IdleTime && currCPU.TotalTime >= prevCPU.TotalTime {
        idleDiff := currCPU.IdleTime - prevCPU.IdleTime
        totalDiff := currCPU.TotalTime - prevCPU.TotalTime

        if totalDiff > 0 {
            cpuUsage = 100.0 - (float64(idleDiff)*100.0)/float64(totalDiff)

            // Clamp CPU usage to reasonable bounds
            if cpuUsage < 0 {
                cpuUsage = 0
            } else if cpuUsage > 100 {
                cpuUsage = 100
            }
        }
    }

    // Memory usage
    memInfo, err := mem.VirtualMemory()
    var memPercent, memMB float64
    if err == nil {
        memPercent = memInfo.UsedPercent
        memMB = float64(memInfo.Used) / 1024 / 1024
    }

    // Disk usage (root filesystem)
    var diskPercent, diskFreeGB float64
    var diskPath string
    if runtime.GOOS == "windows" {
        diskPath = "C:\\"
    } else {
        diskPath = "/"
    }

    diskInfo, err := disk.Usage(diskPath)
    if err == nil {
        diskPercent = diskInfo.UsedPercent
        diskFreeGB = float64(diskInfo.Free) / 1024 / 1024 / 1024
    }

    // Network I/O rates (bytes per second)
    var netRXBPS, netTXBPS uint64
    if timeDiff > 0 && currNet.RXBytes >= prevNet.RXBytes && currNet.TXBytes >= prevNet.TXBytes {
        rxDiff := currNet.RXBytes - prevNet.RXBytes
        txDiff := currNet.TXBytes - prevNet.TXBytes
        netRXBPS = uint64(float64(rxDiff) / timeDiff)
        netTXBPS = uint64(float64(txDiff) / timeDiff)
    }

    // TCP connections (simplified - actual count may vary by OS)
    tcpConns := getTCPConnectionCount()

    return SystemMetrics{
        Timestamp:       time.Now().Format(time.RFC3339),
        MessageTemplate: "System Metrics from {Hostname}",
        Application:     "Monitor",
        Hostname:        hostname,
        CPUPercent:      cpuUsage,
        MemoryPercent:   memPercent,
        MemoryMB:        memMB,
        DiskPercent:     diskPercent,
        DiskFreeGB:      diskFreeGB,
        NetworkRXBPS:    netRXBPS,
        NetworkTXBPS:    netTXBPS,
        TCPConnections:  tcpConns,
    }
}

func getCPUStats() CPUStats {
    if runtime.GOOS == "linux" {
        return getCPUStatsLinux()
    } else if runtime.GOOS == "windows" {
        return getCPUStatsWindows()
    } else if runtime.GOOS == "darwin" {
        return getCPUStatsMacOS()
    }
    // Fallback for other systems
    return CPUStats{}
}

func getNetworkStats() NetworkStats {
    netStats, err := net.IOCounters(false)
    if err != nil || len(netStats) == 0 {
        return NetworkStats{}
    }

    var totalRX, totalTX uint64
    for _, stat := range netStats {
        // Skip loopback interfaces
        if stat.Name != "lo" && stat.Name != "Loopback Pseudo-Interface 1" {
            totalRX += stat.BytesRecv
            totalTX += stat.BytesSent
        }
    }

    return NetworkStats{
        RXBytes: totalRX,
        TXBytes: totalTX,
    }
}

func getTCPConnectionCount() int {
    connections, err := net.Connections("tcp")
    if err != nil {
        return 0
    }
    return len(connections)
}

func printDebugMetrics(metrics SystemMetrics) {
    fmt.Println("===== System Metrics =====")
    fmt.Printf("Timestamp: %s\n", metrics.Timestamp)
    fmt.Printf("Hostname: %s\n", metrics.Hostname)
    fmt.Printf("CPU Usage: %.2f%%\n", metrics.CPUPercent)
    fmt.Printf("Memory Usage: %.2f%%\n", metrics.MemoryPercent)
    fmt.Printf("Memory Usage: %.2f MB\n", metrics.MemoryMB)
    fmt.Printf("Disk Usage: %.2f%%\n", metrics.DiskPercent)
    fmt.Printf("Disk Free: %.2f GB\n", metrics.DiskFreeGB)
    fmt.Printf("Network RX: %d Bytes/s\n", metrics.NetworkRXBPS)
    fmt.Printf("Network TX: %d Bytes/s\n", metrics.NetworkTXBPS)
    fmt.Printf("TCP Connections: %d\n", metrics.TCPConnections)
    fmt.Println("==========================")
}

func sendToSeq(seqURL string, metrics SystemMetrics) {
    jsonData, err := json.Marshal(metrics)
    if err != nil {
        fmt.Fprintf(os.Stderr, "%s - Fehler beim JSON-Encoding: %v\n",
            time.Now().Format(time.RFC3339), err)
        return
    }

    resp, err := http.Post(seqURL+"/ingest/clef", "application/json", bytes.NewBuffer(jsonData))
    if err != nil {
        fmt.Fprintf(os.Stderr, "%s - Fehler beim Senden an Seq: %v\n",
            time.Now().Format(time.RFC3339), err)
        return
    }
    defer resp.Body.Close()

    if resp.StatusCode < 200 || resp.StatusCode >= 300 {
        body, _ := io.ReadAll(resp.Body)
        fmt.Fprintf(os.Stderr, "%s - Fehler beim Senden an Seq (HTTP %d): %s\n",
            time.Now().Format(time.RFC3339), resp.StatusCode, string(body))
    }
}

func getEnvWithDefault(key, defaultValue string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return defaultValue
}

// getHostname ermittelt den Hostname, bevorzugt aus /etc/hostname für Docker-Container
func getHostname() string {
    // Erst versuchen, aus /etc/hostname zu lesen (für Docker-Container)
    if content, err := os.ReadFile("/host/etc/hostname"); err == nil {
        hostname := strings.TrimSpace(string(content))
        if hostname != "" {
            return hostname
        }
    }

    // Fallback auf os.Hostname()
    if hostname, err := os.Hostname(); err == nil {
        return hostname
    }

    return "unknown"
}
