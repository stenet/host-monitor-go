//go:build windows

package main

import (
	"syscall"
	"unsafe"
)

// Windows-specific functions for CPU time retrieval
var (
	kernel32           = syscall.NewLazyDLL("kernel32.dll")
	procGetSystemTimes = kernel32.NewProc("GetSystemTimes")
)

func getCPUStatsWindows() CPUStats {
	var idleTime, kernelTime, userTime uint64

	// Try to get CPU times using Windows API
	if getSystemTimes(&idleTime, &kernelTime, &userTime) {
		// Windows returns times in 100-nanosecond intervals since January 1, 1601
		// Convert to nanoseconds
		const hundredNanosPerNano = 100

		// On Windows, kernel time includes idle time, so we need to subtract it
		// to get the actual kernel time used
		actualKernelTime := kernelTime - idleTime
		totalTime := actualKernelTime + userTime + idleTime

		return CPUStats{
			IdleTime:  idleTime * hundredNanosPerNano,
			TotalTime: totalTime * hundredNanosPerNano,
		}
	}

	// Fallback: return empty stats
	return CPUStats{}
}

func getSystemTimes(idleTime, kernelTime, userTime *uint64) bool {
	var idle, kernel, user syscall.Filetime

	ret, _, _ := procGetSystemTimes.Call(
		uintptr(unsafe.Pointer(&idle)),
		uintptr(unsafe.Pointer(&kernel)),
		uintptr(unsafe.Pointer(&user)),
	)

	if ret == 0 {
		return false
	}

	// Convert FILETIME to uint64 (100-nanosecond intervals since January 1, 1601)
	*idleTime = uint64(idle.HighDateTime)<<32 + uint64(idle.LowDateTime)
	*kernelTime = uint64(kernel.HighDateTime)<<32 + uint64(kernel.LowDateTime)
	*userTime = uint64(user.HighDateTime)<<32 + uint64(user.LowDateTime)

	return true
}

// Stub functions for other platforms
func getCPUStatsLinux() CPUStats {
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
