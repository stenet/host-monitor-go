//go:build linux

package main

import (
    "bytes"
    "os"
    "strconv"
)

/*
#include <unistd.h>

long get_linux_clock_ticks() {
    return sysconf(_SC_CLK_TCK);
}
*/
import "C"

func getCPUStatsLinux() CPUStats {
    data, err := os.ReadFile("/proc/stat")
    if err != nil {
        return CPUStats{}
    }

    lines := bytes.Split(data, []byte("\n"))
    if len(lines) == 0 {
        return CPUStats{}
    }

    // Parse first line: "cpu  user nice system idle iowait irq softirq steal guest guest_nice"
    fields := bytes.Fields(lines[0])
    if len(fields) < 8 || string(fields[0]) != "cpu" {
        return CPUStats{}
    }

    var values [10]uint64
    var totalTicks uint64

    // Parse numeric values starting from field 1 (skip "cpu")
    for i := 1; i < len(fields) && i < 11; i++ {
        if val, err := strconv.ParseUint(string(fields[i]), 10, 64); err == nil {
            values[i-1] = val
            totalTicks += val
        }
    }

    // Idle time is field 4 (index 3 in our values array)
    idleTicks := values[3]

    // Get actual tick rate from Linux kernel
    ticksPerSecond := int(C.get_linux_clock_ticks())
    if ticksPerSecond <= 0 {
        ticksPerSecond = 100 // fallback
    }

    // Convert ticks to nanoseconds
    const nanosPerSecond = 1000000000
    nanosPerTick := nanosPerSecond / uint64(ticksPerSecond)

    return CPUStats{
        IdleTime:  idleTicks * nanosPerTick,
        TotalTime: totalTicks * nanosPerTick,
    }
}

// Stub functions for other platforms
func getCPUStatsWindows() CPUStats {
    return CPUStats{}
}

func getCPUStatsMacOS() CPUStats {
    return CPUStats{}
}

func getCPUStatsMacOSSyscall() CPUStats {
    return CPUStats{}
}

func getMacOSTickRate() int {
    return 100
}

