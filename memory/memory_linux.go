// +build linux

package memory

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

// Get memory statistics
func Get() (*Stats, error) {
	// Reference: man 5 proc, Documentation/filesystems/proc.txt in Linux source code
	file, err := os.Open("/proc/meminfo")
	if err != nil {
		return nil, err
	}
	defer file.Close()
	return collectMemoryStats(file)
}

// Stats represents memory statistics for linux
type Stats struct {
	Total, Used, Buffers, Cached, Free, Available, Active, Inactive,
	SwapTotal, SwapUsed, SwapCached, SwapFree uint64
}

func collectMemoryStats(out io.Reader) (*Stats, error) {
	scanner := bufio.NewScanner(out)
	var memory Stats
	memStats := map[string]*uint64{
		"MemTotal":     &memory.Total,
		"MemFree":      &memory.Free,
		"MemAvailable": &memory.Available,
		"Buffers":      &memory.Buffers,
		"Cached":       &memory.Cached,
		"Active":       &memory.Active,
		"Inactive":     &memory.Inactive,
		"SwapCached":   &memory.SwapCached,
		"SwapTotal":    &memory.SwapTotal,
		"SwapFree":     &memory.SwapFree,
	}
	for scanner.Scan() {
		line := scanner.Text()
		i := strings.IndexRune(line, ':')
		if i < 0 {
			continue
		}
		if ptr := memStats[line[:i]]; ptr != nil {
			val := strings.TrimSpace(strings.TrimRight(line[i+1:], "kB"))
			if v, err := strconv.ParseUint(val, 10, 64); err == nil {
				*ptr = v * 1024
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scan error for /proc/meminfo: %s", err)
	}

	memory.SwapUsed = memory.SwapTotal - memory.SwapFree
	// This calculation may be changed in the future.
	memory.Used = memory.Total - memory.Free - memory.Buffers - memory.Cached
	return &memory, nil
}
