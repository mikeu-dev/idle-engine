package gametime

import (
	"math"
	"time"

	"github.com/beevik/ntp"
)

// GetNetworkTime mengambil waktu nyata terkalibrasi dari server NTP.
func GetNetworkTime(server string, timeout time.Duration) (time.Time, error) {
	resp, err := ntp.QueryWithOptions(server, ntp.QueryOptions{
		Timeout: timeout,
	})
	if err != nil {
		return time.Time{}, err
	}
	// Tambahkan ClockOffset ke waktu lokal saat ini untuk mendapatkan waktu jaringan terkalibrasi
	calibratedTime := time.Now().Add(resp.ClockOffset)
	return calibratedTime, nil
}

// IsSystemTimeValid memeriksa apakah waktu lokal sistem cocok dengan waktu jaringan terpercaya.
// Mengembalikan true jika selisihnya berada di bawah nilai toleransi.
func IsSystemTimeValid(systemTime, ntpTime time.Time, tolerance time.Duration) bool {
	diff := systemTime.Sub(ntpTime)
	return math.Abs(diff.Seconds()) <= tolerance.Seconds()
}
