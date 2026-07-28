package gametime

import (
	"testing"
	"time"
)

func TestIsSystemTimeValid(t *testing.T) {
	now := time.Now()
	// Selisih 10 detik (valid dalam toleransi 1 menit)
	if !IsSystemTimeValid(now, now.Add(10*time.Second), 1*time.Minute) {
		t.Error("expected system time to be valid (10s diff within 1m tolerance)")
	}

	// Selisih 2 jam (tidak valid dalam toleransi 5 menit)
	if IsSystemTimeValid(now, now.Add(2*time.Hour), 5*time.Minute) {
		t.Error("expected system time to be invalid (2h diff exceeds 5m tolerance)")
	}
}

func TestGetNetworkTime(t *testing.T) {
	// Coba ambil waktu ntp
	ntpTime, err := GetNetworkTime("pool.ntp.org", 3*time.Second)
	if err != nil {
		t.Logf("skipping NTP server test (offline or query timed out: %v)", err)
		return
	}

	if ntpTime.IsZero() {
		t.Error("expected network time to be non-zero")
	}
}
