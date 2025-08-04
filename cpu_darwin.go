//go:build darwin

package main

/*
#include <mach/mach.h>
#include <mach/processor_info.h>
#include <mach/mach_host.h>
#include <sys/sysctl.h>

typedef struct processor_cpu_load_info processor_cpu_load_info_data_t;

int get_cpu_load_info(processor_cpu_load_info_data_t *info, mach_msg_type_number_t *count) {
    mach_port_t host_port = mach_host_self();
    processor_info_array_t cpu_load_info;
    mach_msg_type_number_t cpu_msg_count;

    kern_return_t error = host_processor_info(host_port, PROCESSOR_CPU_LOAD_INFO,
                                            &cpu_msg_count, &cpu_load_info, count);

    if (error != KERN_SUCCESS) {
        return 0;
    }

    // Copy first CPU's info (we'll aggregate if needed)
    if (cpu_msg_count > 0) {
        processor_cpu_load_info_data_t *cpu_info = (processor_cpu_load_info_data_t *)cpu_load_info;
        *info = cpu_info[0];
    }

    return 1;
}

int get_macos_cpu_ticks(unsigned long long *ticks) {
    size_t size = sizeof(unsigned long long) * 4;
    if (sysctlbyname("kern.cp_time", ticks, &size, NULL, 0) != 0) {
        return 0;
    }
    return 1;
}

int get_macos_tick_rate() {
    struct clockinfo clockinfo;
    size_t size = sizeof(clockinfo);
    if (sysctlbyname("kern.clockrate", &clockinfo, &size, NULL, 0) != 0) {
        return 0;
    }
    return clockinfo.hz;
}
*/
import "C"

func getCPUStatsMacOS() CPUStats {
    // Try syscall approach first (more reliable for aggregated data)
    if stats := getCPUStatsMacOSSyscall(); stats.TotalTime > 0 {
        return stats
    }

    // Fallback: use Mach system calls
    var cpuInfo C.processor_cpu_load_info_data_t
    var count C.mach_msg_type_number_t

    // Get CPU load info using Mach system calls
    if getCPULoadInfo(&cpuInfo, &count) {
        // macOS provides CPU ticks in different states
        user := uint64(cpuInfo.cpu_ticks[C.CPU_STATE_USER])
        system := uint64(cpuInfo.cpu_ticks[C.CPU_STATE_SYSTEM])
        idle := uint64(cpuInfo.cpu_ticks[C.CPU_STATE_IDLE])
        nice := uint64(cpuInfo.cpu_ticks[C.CPU_STATE_NICE])

        totalTicks := user + system + idle + nice

        // Get actual tick rate from macOS kernel
        ticksPerSecond := getMacOSTickRate()
        if ticksPerSecond <= 0 {
            ticksPerSecond = 100 // fallback
        }

        // Convert ticks to nanoseconds
        const nanosPerSecond = 1000000000
        nanosPerTick := nanosPerSecond / uint64(ticksPerSecond)

        return CPUStats{
            IdleTime:  idle * nanosPerTick,
            TotalTime: totalTicks * nanosPerTick,
        }
    }

    return CPUStats{}
}

func getCPUStatsMacOSSyscall() CPUStats {
    // This implementation uses kern.cp_time which provides aggregate CPU stats
    // This is more accurate than the Mach calls for total system CPU usage

    var cpuTicks [4]uint64 // user, nice, system, idle

    if getMacOSCPUTicks(cpuTicks[:]) {
        user := cpuTicks[0]
        nice := cpuTicks[1]
        system := cpuTicks[2]
        idle := cpuTicks[3]

        totalTicks := user + nice + system + idle

        // Get actual tick rate from macOS kernel
        ticksPerSecond := getMacOSTickRate()
        if ticksPerSecond <= 0 {
            ticksPerSecond = 100 // fallback
        }

        // Convert ticks to nanoseconds
        const nanosPerSecond = 1000000000
        nanosPerTick := nanosPerSecond / uint64(ticksPerSecond)

        return CPUStats{
            IdleTime:  idle * nanosPerTick,
            TotalTime: totalTicks * nanosPerTick,
        }
    }

    return CPUStats{}
}

func getCPULoadInfo(info *C.processor_cpu_load_info_data_t, count *C.mach_msg_type_number_t) bool {
    result := C.get_cpu_load_info(info, count)
    return result != 0
}

func getMacOSCPUTicks(ticks []uint64) bool {
    if len(ticks) < 4 {
        return false
    }

    var cTicks [4]C.ulonglong
    result := C.get_macos_cpu_ticks(&cTicks[0])

    if result != 0 {
        for i := 0; i < 4; i++ {
            ticks[i] = uint64(cTicks[i])
        }
        return true
    }

    return false
}

func getMacOSTickRate() int {
    return int(C.get_macos_tick_rate())
}

// Stub functions for other platforms  
func getCPUStatsLinux() CPUStats {
    return CPUStats{}
}

func getCPUStatsWindows() CPUStats {
    return CPUStats{}
}

