package utils

import "sync"

// Menyimpan daftar "telinga" (channel) para Admin yang sedang online
var (
	AdminClients = make(map[chan string]bool)
	Mu           sync.Mutex
)

// Fungsi untuk menyiarkan pesan ke semua Admin yang terhubung
func BroadcastEmergency(msg string) {
	Mu.Lock()
	defer Mu.Unlock()
	for clientChan := range AdminClients {
		// Kirim pesan ke channel masing-masing admin
		clientChan <- msg 
	}
}