package gametime

import (
	"math"
	"time"

	"github.com/beevik/ntp"
)

// GetNetworkTime retrieves the real calibrated time from an NTP server.
func GetNetworkTime(server string, timeout time.Duration) (time.Time, error) {
	resp, err := ntp.QueryWithOptions(server, ntp.QueryOptions{
		Timeout: timeout,
	})
	if err != nil {
		return time.Time{}, err
	}
	// Add ClockOffset to the current local time to get the calibrated network time
	calibratedTime := time.Now().Add(resp.ClockOffset)
	return calibratedTime, nil
}

// IsSystemTimeValid checks if the local system time matches the trusted network time.
// Returns true if the difference is below the tolerance threshold.
func IsSystemTimeValid(systemTime, ntpTime time.Time, tolerance time.Duration) bool {
	diff := systemTime.Sub(ntpTime)
	return math.Abs(diff.Seconds()) <= tolerance.Seconds()
}
