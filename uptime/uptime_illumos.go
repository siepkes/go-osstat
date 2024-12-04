//go:build illumos || solaris
// +build illumos solaris

package uptime

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

func get() (time.Duration, error) {
	cmd := exec.CommandContext(ctx, "kstat", "-p", "unix:0:system_misc:boot_time")
	out, err := cmd.StdoutPipe()
	if err != nil {
		return 0, err
	}
	if err := cmd.Start(); err != nil {
		return 0, err
	}

	scanner := bufio.NewScanner(out)
	var bootTimeStr string
	for scanner.Scan() {
		line := scanner.Text()
		// Expected stdout: "unix:0:system_misc:boot_time  <timestamp>"
		parts := strings.Fields(line)
		if len(parts) == 2 {
			bootTimeStr = parts[1]
			break
		}
	}
	if err := scanner.Err(); err != nil {
		return 0, err
	}
	if err := cmd.Wait(); err != nil {
		return 0, err
	}

	// Validate and parse boot time
	if bootTimeStr == "" {
		return 0, fmt.Errorf("could not find boot_time in output")
	}
	bootTime, err := strconv.ParseInt(bootTimeStr, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid boot_time value: %v", err)
	}

	// Calculate total uptime
	now := time.Now().Unix()
	uptimeSeconds := now - bootTime
	if uptimeSeconds < 0 {
		return 0, errors.New("boot_time is in the future")
	}
	return time.Duration(uptimeSeconds) * time.Second, nil
}
