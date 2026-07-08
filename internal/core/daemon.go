package core

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Job represents a generic task dispatched across the mesh network.
type Job struct {
	ID      string `json:"id"`
	Type    string `json:"type"` // e.g. "exec"
	Payload string `json:"payload"`
}

// JobOutput represents a chunk of output streamed from a running job.
type JobOutput struct {
	Type     string `json:"type"` // "stdout", "stderr", "exit", "error"
	Content  string `json:"content"`
	ExitCode int    `json:"exit_code,omitempty"`
}

// DiscoveredPeer represents a peer node discovered over the local network.
type DiscoveredPeer struct {
	ID              string        `json:"device_id"`
	Name            string        `json:"device_name"`
	OS              string        `json:"os"`
	IP              string        `json:"ip"`
	Port            int           `json:"port"`
	Fingerprint     string        `json:"fingerprint"`
	Transport       string        `json:"transport"`
	Latency         string        `json:"latency"`
	Hardware        HardwareInfo  `json:"hardware"`
	OllamaInstalled bool          `json:"ollama_installed"`
	OllamaVersion   string        `json:"ollama_version"`
	Models          []OllamaModel `json:"models"`
	LastSeen        time.Time     `json:"-"`
	Protocol        int           `json:"protocol"`
	Version         string        `json:"version"`
	Capabilities    Capabilities  `json:"capabilities"`
}

// PairRequest represents pairing credentials sent during the handshake.
type PairRequest struct {
	DeviceID    string       `json:"device_id"`
	DeviceName  string       `json:"device_name"`
	OS          string       `json:"os"`
	Fingerprint string       `json:"fingerprint"`
	Transport   string       `json:"transport"`
	Certificate string       `json:"certificate"`
	Protocol    int          `json:"protocol"`
	Version     string       `json:"version"`
	Capabilities Capabilities `json:"capabilities"`
}

// PairResponse contains pairing results.
type PairResponse struct {
	Approved    bool         `json:"approved"`
	DeviceID    string       `json:"device_id"`
	DeviceName  string       `json:"device_name"`
	OS          string       `json:"os"`
	Fingerprint string       `json:"fingerprint"`
	Certificate string       `json:"certificate"`
	Protocol    int          `json:"protocol"`
	Version     string       `json:"version"`
	Capabilities Capabilities `json:"capabilities"`
}

// Daemon represents the background service coordination layer.
type Daemon struct {
	cfg           *Config
	peersMu       sync.Mutex
	discovered    map[string]DiscoveredPeer
	pendingPair   *PairRequest
	pendingIP     string
	pairDecideCh  chan bool
	pairDecision  bool
	peerTLSConfig *tls.Config
	hardware      HardwareInfo
	EventBus      *EventBus
}

// NewDaemon instantiates a new mesh networking background coordinator.
func NewDaemon(cfg *Config) (*Daemon, error) {
	d := &Daemon{
		cfg:          cfg,
		discovered:   make(map[string]DiscoveredPeer),
		pairDecideCh: make(chan bool),
		hardware: HardwareInfo{
			CPU:  "Profiling...",
			RAM:  "Profiling...",
			GPU:  "Profiling...",
			VRAM: "-",
			CUDA: "No",
			OS:   runtime.GOOS,
		},
		EventBus: NewEventBus(),
	}
	go func() {
		d.hardware = DetectHardware()
	}()
	return d, nil
}

// Start runs the UDP discovery broadcaster/listener and TCP TLS/mTLS control server.
func (d *Daemon) Start(localOnly bool) error {
	// 1. Setup TLS config for remote mTLS connections
	cert, err := tls.X509KeyPair([]byte(d.cfg.CertificatePEM), []byte(d.cfg.PrivateKeyPEM))
	if err != nil {
		return fmt.Errorf("failed to load TLS key pair: %w", err)
	}

	d.peerTLSConfig = &tls.Config{
		Certificates: []tls.Certificate{cert},
		ClientAuth:   tls.RequireAnyClientCert,
		MinVersion:   tls.VersionTLS13,
	}

	// 2. Start UDP Discovery Broadcaster & Listener
	go d.runUDPBroadcast()
	go d.runUDPListener()

	// 3. Start local TCP/TLS control API
	mux := http.NewServeMux()
	mux.HandleFunc("/nearby", d.handleLocalNearby)
	mux.HandleFunc("/pending", d.handleLocalPending)
	mux.HandleFunc("/respond", d.handleLocalRespond)
	mux.HandleFunc("/connect", d.handleLocalConnect)
	mux.HandleFunc("/devices", d.handleLocalDevices)
	mux.HandleFunc("/devices/rename", d.handleLocalDevicesRename)
	mux.HandleFunc("/devices/remove", d.handleLocalDevicesRemove)
	mux.HandleFunc("/devices/block", d.handleLocalDevicesBlock)
	mux.HandleFunc("/run", d.handleLocalRun)
	mux.HandleFunc("/chat", d.handleLocalChat)

	// Remote P2P handlers
	mux.HandleFunc("/remote/pair", d.handleRemotePair)
	mux.HandleFunc("/remote/run", d.handleRemoteRun)
	mux.HandleFunc("/remote/chat", d.handleRemoteChat)

	server := &http.Server{
		Addr:      "0.0.0.0:8911",
		Handler:   mux,
		TLSConfig: d.peerTLSConfig,
	}

	// Clean up stale discovered peers in background
	go d.cleanupDiscoveredPeersLoop()

	log.Printf("Orion Daemon started on port 8911")
	return server.ListenAndServeTLS("", "")
}

// verifyLocal verifies that request originated from localhost/loopback interface.
func (d *Daemon) verifyLocal(r *http.Request) bool {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return false
	}
	return host == "127.0.0.1" || host == "::1" || host == "localhost"
}

// verifyRemoteFingerprint verifies remote peer TLS cert against trusted devices list.
func (d *Daemon) verifyRemoteFingerprint(r *http.Request) bool {
	if r.TLS == nil || len(r.TLS.PeerCertificates) == 0 {
		return false
	}
	peerCert := r.TLS.PeerCertificates[0]
	peerHash := sha256.Sum256(peerCert.Raw)
	
	// Compute fingerprint prefix (e.g. 8F:E1:12:A2)
	parts := make([]string, 4)
	for i := 0; i < 4; i++ {
		parts[i] = fmt.Sprintf("%02X", peerHash[i])
	}
	peerFingerprint := strings.Join(parts, ":")

	if peerFingerprint == d.cfg.PublicKey {
		return true
	}

	// Reload config to get latest trusted lists
	cfg, err := LoadConfig()
	if err != nil {
		return false
	}

	for _, dev := range cfg.Devices {
		if dev.Fingerprint == peerFingerprint {
			// Check if blocked
			for _, blocked := range cfg.BlockedDevices {
				if blocked == dev.ID {
					return false
				}
			}
			return true
		}
	}
	return false
}

func getLocalIPAndTransport() (string, string) {
	interfaces, err := net.Interfaces()
	if err != nil {
		return "127.0.0.1", "Wi-Fi" // fallback
	}
	for _, iface := range interfaces {
		// Skip down or loopback interfaces
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			if ip == nil || ip.IsLoopback() {
				continue
			}
			ip4 := ip.To4()
			if ip4 == nil {
				continue
			}

			// Determine transport by interface name
			name := strings.ToLower(iface.Name)
			transport := "Wi-Fi" // default fallback
			if strings.Contains(name, "eth") || strings.Contains(name, "ethernet") || strings.Contains(name, "lan") || strings.Contains(name, "enp") || strings.Contains(name, "ens") {
				transport = "Ethernet"
			} else if strings.Contains(name, "wlan") || strings.Contains(name, "wifi") || strings.Contains(name, "wi-fi") || strings.Contains(name, "wireless") {
				transport = "Wi-Fi"
			} else if strings.Contains(name, "usb") || strings.Contains(name, "ndis") {
				transport = "USB"
			} else if strings.Contains(name, "blue") || strings.Contains(name, "bt") {
				transport = "Bluetooth"
			}
			return ip4.String(), transport
		}
	}
	return "127.0.0.1", "Wi-Fi"
}

func getLocalCapabilities(installed bool, hw HardwareInfo) Capabilities {
	isGPU := hw.GPU != "CPU Only"
	isCUDA := hw.CUDA == "Yes"
	isMetal := runtime.GOOS == "darwin" && (runtime.GOARCH == "arm64" || runtime.GOARCH == "amd64")
	isROCm := false
	if runtime.GOOS == "linux" && strings.Contains(strings.ToLower(hw.GPU), "amd") {
		isROCm = true
	}
	return Capabilities{
		Ollama:    installed,
		CUDA:      isCUDA,
		GPU:       isGPU,
		ROCm:      isROCm,
		Metal:     isMetal,
		Benchmark: true,
		Hardware:  true,
	}
}

func getLocalOllamaInfo() (bool, string, []OllamaModel) {
	client := &http.Client{Timeout: 300 * time.Millisecond}
	resp, err := client.Get("http://127.0.0.1:11434/api/tags")
	if err != nil {
		return false, "", nil
	}
	defer resp.Body.Close()

	// Get version
	version := "unknown"
	vResp, vErr := client.Get("http://127.0.0.1:11434/api/version")
	if vErr == nil {
		defer vResp.Body.Close()
		var vPayload struct {
			Version string `json:"version"`
		}
		if json.NewDecoder(vResp.Body).Decode(&vPayload) == nil {
			version = vPayload.Version
		}
	}

	// Get running models from /api/ps
	runningModels := make(map[string]bool)
	psResp, psErr := client.Get("http://127.0.0.1:11434/api/ps")
	if psErr == nil {
		defer psResp.Body.Close()
		var psPayload struct {
			Models []struct {
				Name string `json:"name"`
			} `json:"models"`
		}
		if json.NewDecoder(psResp.Body).Decode(&psPayload) == nil {
			for _, m := range psPayload.Models {
				runningModels[m.Name] = true
			}
		}
	}

	var payload struct {
		Models []struct {
			Name    string `json:"name"`
			Size    int64  `json:"size"`
			Details struct {
				Quantization string `json:"quantization_level"`
			} `json:"details"`
		} `json:"models"`
	}
	if json.NewDecoder(resp.Body).Decode(&payload) != nil {
		return true, version, nil
	}

	list := make([]OllamaModel, len(payload.Models))
	for i, m := range payload.Models {
		status := "Idle"
		if runningModels[m.Name] {
			status = "Running"
		}
		quant := m.Details.Quantization
		if quant == "" {
			quant = "-"
		}
		list[i] = OllamaModel{
			Name:         m.Name,
			Size:         m.Size,
			Quantization: quant,
			Status:       status,
		}
	}
	return true, version, list
}

// UDP Discovery implementation
func (d *Daemon) runUDPBroadcast() {
	addr, err := net.ResolveUDPAddr("udp", "255.255.255.255:8910")
	if err != nil {
		return
	}
	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		return
	}
	defer conn.Close()

	for {
		installed, version, models := getLocalOllamaInfo()
		_, transport := getLocalIPAndTransport()

		msg := DiscoveredPeer{
			ID:              d.cfg.DeviceID,
			Name:            d.cfg.DeviceName,
			OS:              runtime.GOOS,
			Port:            8911,
			Fingerprint:     d.cfg.PublicKey,
			Transport:       transport,
			Hardware:        d.hardware,
			OllamaInstalled: installed,
			OllamaVersion:   version,
			Models:          models,
			Protocol:        1,
			Version:         "0.1.0-beta",
			Capabilities:    getLocalCapabilities(installed, d.hardware),
		}
		data, err := json.Marshal(msg)
		if err == nil {
			conn.Write(data)
		}
		time.Sleep(3 * time.Second)
	}
}

func (d *Daemon) runUDPListener() {
	addr, err := net.ResolveUDPAddr("udp", "0.0.0.0:8910")
	if err != nil {
		return
	}
	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		return
	}
	defer conn.Close()

	var buf [2048]byte
	for {
		n, src, err := conn.ReadFromUDP(buf[:])
		if err != nil {
			continue
		}

		var peer DiscoveredPeer
		if err := json.Unmarshal(buf[:n], &peer); err != nil {
			continue
		}

		if peer.ID == d.cfg.DeviceID {
			continue // Skip self
		}

		peer.IP = src.IP.String()
		peer.LastSeen = time.Now()

		// Measure real latency
		startTime := time.Now()
		tcpDialer := net.Dialer{Timeout: 100 * time.Millisecond}
		tConn, tErr := tcpDialer.Dial("tcp", net.JoinHostPort(peer.IP, strconv.Itoa(peer.Port)))
		if tErr == nil {
			peer.Latency = fmt.Sprintf("%d ms", time.Since(startTime).Milliseconds())
			tConn.Close()
		} else {
			peer.Latency = "-"
		}

		d.peersMu.Lock()
		_, alreadyKnown := d.discovered[peer.ID]
		d.discovered[peer.ID] = peer
		d.peersMu.Unlock()

		if !alreadyKnown {
			d.EventBus.Publish(Event{
				Type:    EventDeviceOnline,
				Payload: peer,
			})
		}
	}
}

func (d *Daemon) cleanupDiscoveredPeersLoop() {
	for {
		time.Sleep(5 * time.Second)
		d.peersMu.Lock()
		for id, peer := range d.discovered {
			// Evict only if not seen for 2 minutes
			if time.Since(peer.LastSeen) > 2*time.Minute {
				delete(d.discovered, id)
				d.EventBus.Publish(Event{
					Type:    EventDeviceOffline,
					Payload: id,
				})
			}
		}
		d.peersMu.Unlock()
	}
}

// HTTP handlers for Local Loopback Control API
func (d *Daemon) handleLocalNearby(w http.ResponseWriter, r *http.Request) {
	if !d.verifyLocal(r) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	d.peersMu.Lock()
	peers := make([]DiscoveredPeer, 0, len(d.discovered))
	for _, p := range d.discovered {
		peers = append(peers, p)
	}
	d.peersMu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(peers)
}

func (d *Daemon) handleLocalPending(w http.ResponseWriter, r *http.Request) {
	if !d.verifyLocal(r) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if d.pendingPair == nil {
		json.NewEncoder(w).Encode(map[string]interface{}{"pending": false})
	} else {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"pending":     true,
			"device_name": d.pendingPair.DeviceName,
			"device_id":   d.pendingPair.DeviceID,
			"fingerprint": d.pendingPair.Fingerprint,
		})
	}
}

func (d *Daemon) handleLocalRespond(w http.ResponseWriter, r *http.Request) {
	if !d.verifyLocal(r) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	action := r.URL.Query().Get("action")
	if d.pendingPair == nil {
		http.Error(w, "No pending request", http.StatusBadRequest)
		return
	}

	if action == "accept" {
		d.pairDecision = true
		// Save trusted relationship
		cfg, err := LoadConfig()
		if err == nil {
			// Remove if already exists
			var clean []Device
			for _, dev := range cfg.Devices {
				if dev.ID != d.pendingPair.DeviceID {
					clean = append(clean, dev)
				}
			}
			clean = append(clean, Device{
				ID:             d.pendingPair.DeviceID,
				Name:           d.pendingPair.DeviceName,
				OS:             d.pendingPair.OS,
				IP:             d.pendingIP,
				Fingerprint:    d.pendingPair.Fingerprint,
				Transport:      d.pendingPair.Transport,
				Status:         "online",
				Certificate:    d.pendingPair.Certificate,
				Alias:          d.pendingPair.DeviceName,
				TrustTimestamp: time.Now().Format(time.RFC3339),
			})
			cfg.Devices = clean
			SaveConfig(cfg)
		}
		select {
		case d.pairDecideCh <- true:
		default:
		}
	} else {
		d.pairDecision = false
		select {
		case d.pairDecideCh <- false:
		default:
		}
	}

	d.pendingPair = nil
	w.WriteHeader(http.StatusOK)
}

func (d *Daemon) handleLocalConnect(w http.ResponseWriter, r *http.Request) {
	if !d.verifyLocal(r) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	targetIP := r.URL.Query().Get("ip")
	targetID := r.URL.Query().Get("id")

	if targetIP == "" {
		http.Error(w, "Missing target IP", http.StatusBadRequest)
		return
	}

	// Dial remote peer TLS
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, // skipping CA cert check because self-signed
	}
	client := &http.Client{Transport: tr, Timeout: 30 * time.Second}

	installed, _, _ := getLocalOllamaInfo()
	_, transport := getLocalIPAndTransport()

	reqPayload := PairRequest{
		DeviceID:     d.cfg.DeviceID,
		DeviceName:   d.cfg.DeviceName,
		OS:           runtime.GOOS,
		Fingerprint:  d.cfg.PublicKey,
		Transport:    transport,
		Certificate:  d.cfg.CertificatePEM,
		Protocol:     1,
		Version:      "0.1.0-beta",
		Capabilities: getLocalCapabilities(installed, d.hardware),
	}
	payloadBytes, _ := json.Marshal(reqPayload)

	resp, err := client.Post(fmt.Sprintf("https://%s:8911/remote/pair", targetIP), "application/json", strings.NewReader(string(payloadBytes)))
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to contact target: %v", err), http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		http.Error(w, "Rejected by remote host", http.StatusForbidden)
		return
	}

	var pairResp PairResponse
	if err := json.NewDecoder(resp.Body).Decode(&pairResp); err != nil || !pairResp.Approved {
		http.Error(w, "Connection rejected or malformed handshake", http.StatusForbidden)
		return
	}

	if pairResp.Protocol != 1 {
		http.Error(w, fmt.Sprintf("Protocol version mismatch (Local: 1, Remote: %d). Please upgrade Orion.", pairResp.Protocol), http.StatusBadRequest)
		return
	}

	// Save trusted device details
	cfg, err := LoadConfig()
	if err == nil {
		var clean []Device
		for _, dev := range cfg.Devices {
			if dev.ID != targetID {
				clean = append(clean, dev)
			}
		}
		clean = append(clean, Device{
			ID:             pairResp.DeviceID,
			Name:           pairResp.DeviceName,
			OS:             pairResp.OS,
			IP:             targetIP,
			Fingerprint:    pairResp.Fingerprint,
			Status:         "online",
			Certificate:    pairResp.Certificate,
			Alias:          pairResp.DeviceName,
			TrustTimestamp: time.Now().Format(time.RFC3339),
			Protocol:       pairResp.Protocol,
			Version:        pairResp.Version,
			Capabilities:   pairResp.Capabilities,
			Transport:      transport,
		})
		cfg.Devices = clean
		SaveConfig(cfg)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(pairResp)
}

func (d *Daemon) handleLocalDevices(w http.ResponseWriter, r *http.Request) {
	if !d.verifyLocal(r) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	cfg, _ := LoadConfig()
	d.peersMu.Lock()
	defer d.peersMu.Unlock()

	updated := make([]Device, len(cfg.Devices))
	for i, dev := range cfg.Devices {
		if p, ok := d.discovered[dev.ID]; ok {
			if time.Since(p.LastSeen) > 8*time.Second {
				dev.Status = "offline"
				dev.Latency = fmt.Sprintf("Last seen %d sec ago", int(time.Since(p.LastSeen).Seconds()))
			} else {
				dev.Status = "online"
				dev.Latency = p.Latency
			}
			dev.Hardware = p.Hardware
			dev.OllamaInstalled = p.OllamaInstalled
			dev.OllamaVersion = p.OllamaVersion
			dev.Models = p.Models
			dev.IP = p.IP
			dev.Transport = p.Transport
			dev.Protocol = p.Protocol
			dev.Version = p.Version
			dev.Capabilities = p.Capabilities
		} else {
			dev.Status = "offline"
			dev.Latency = "-"
		}
		updated[i] = dev
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updated)
}

func (d *Daemon) handleLocalDevicesRename(w http.ResponseWriter, r *http.Request) {
	if !d.verifyLocal(r) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	id := r.URL.Query().Get("id")
	alias := r.URL.Query().Get("alias")

	cfg, err := LoadConfig()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	for i, dev := range cfg.Devices {
		if dev.ID == id {
			cfg.Devices[i].Name = alias
			SaveConfig(cfg)
			w.WriteHeader(http.StatusOK)
			return
		}
	}
	http.Error(w, "Device not found", http.StatusNotFound)
}

func (d *Daemon) handleLocalDevicesRemove(w http.ResponseWriter, r *http.Request) {
	if !d.verifyLocal(r) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	id := r.URL.Query().Get("id")

	cfg, err := LoadConfig()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var clean []Device
	for _, dev := range cfg.Devices {
		if dev.ID != id {
			clean = append(clean, dev)
		}
	}
	cfg.Devices = clean
	SaveConfig(cfg)
	w.WriteHeader(http.StatusOK)
}

func (d *Daemon) handleLocalDevicesBlock(w http.ResponseWriter, r *http.Request) {
	if !d.verifyLocal(r) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	id := r.URL.Query().Get("id")

	cfg, err := LoadConfig()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	cfg.BlockedDevices = append(cfg.BlockedDevices, id)
	SaveConfig(cfg)
	w.WriteHeader(http.StatusOK)
}

// Remote Handlers
func (d *Daemon) handleRemotePair(w http.ResponseWriter, r *http.Request) {
	var req PairRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid pairing payload", http.StatusBadRequest)
		return
	}

	if req.Protocol != 1 {
		http.Error(w, fmt.Sprintf("Protocol version mismatch (Local: 1, Remote: %d). Please upgrade Orion.", req.Protocol), http.StatusBadRequest)
		return
	}

	host, _, _ := net.SplitHostPort(r.RemoteAddr)
	d.pendingIP = host
	d.pendingPair = &req

	// Block waiting for local consent decision (with 45s timeout)
	select {
	case approved := <-d.pairDecideCh:
		if approved {
			installed, _, _ := getLocalOllamaInfo()
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(PairResponse{
				Approved:     true,
				DeviceID:     d.cfg.DeviceID,
				DeviceName:   d.cfg.DeviceName,
				OS:           runtime.GOOS,
				Fingerprint:  d.cfg.PublicKey,
				Certificate:  d.cfg.CertificatePEM,
				Protocol:     1,
				Version:      "0.1.0-beta",
				Capabilities: getLocalCapabilities(installed, d.hardware),
			})
		} else {
			http.Error(w, "Rejected", http.StatusForbidden)
		}
	case <-time.After(45 * time.Second):
		d.pendingPair = nil
		http.Error(w, "Timeout waiting for consent", http.StatusRequestTimeout)
	}
}

func (d *Daemon) handleLocalRun(w http.ResponseWriter, r *http.Request) {
	if !d.verifyLocal(r) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	targetIP := r.URL.Query().Get("ip")
	command := r.URL.Query().Get("command")

	if targetIP == "" || command == "" {
		http.Error(w, "Missing target or command", http.StatusBadRequest)
		return
	}

	// Dial P2P node over strict mTLS client certificate configuration!
	cert, err := tls.X509KeyPair([]byte(d.cfg.CertificatePEM), []byte(d.cfg.PrivateKeyPEM))
	if err != nil {
		http.Error(w, "Failed to load identity keys", http.StatusInternalServerError)
		return
	}

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			Certificates:       []tls.Certificate{cert},
			InsecureSkipVerify: true, // Verification using pinned fingerprint callback below
			VerifyPeerCertificate: func(rawCerts [][]byte, verifiedChains [][]*x509.Certificate) error {
				if len(rawCerts) == 0 {
					return fmt.Errorf("no certs presented")
				}
				peerCert, _ := x509.ParseCertificate(rawCerts[0])
				hash := sha256.Sum256(peerCert.Raw)
				parts := make([]string, 4)
				for i := 0; i < 4; i++ {
					parts[i] = fmt.Sprintf("%02X", hash[i])
				}
				peerFingerprint := strings.Join(parts, ":")

				cfg, _ := LoadConfig()
				if peerFingerprint == cfg.PublicKey {
					return nil
				}
				for _, dev := range cfg.Devices {
					if dev.Fingerprint == peerFingerprint {
						return nil
					}
				}
				return fmt.Errorf("untrusted peer certificate fingerprint: %s", peerFingerprint)
			},
		},
	}

	client := &http.Client{Transport: tr, Timeout: 5 * time.Minute}

	jobPayload := Job{
		ID:      "job-1",
		Type:    "exec",
		Payload: command,
	}
	payloadBytes, _ := json.Marshal(jobPayload)

	resp, err := client.Post(fmt.Sprintf("https://%s:8911/remote/run", targetIP), "application/json", strings.NewReader(string(payloadBytes)))
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to run remote job: %v", err), http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	w.Header().Set("Content-Type", "application/json-stream")
	w.WriteHeader(http.StatusOK)

	flusher, flusherOk := w.(http.Flusher)
	decoder := json.NewDecoder(resp.Body)
	for {
		var output JobOutput
		if err := decoder.Decode(&output); err != nil {
			if err == io.EOF {
				break
			}
			break
		}
		json.NewEncoder(w).Encode(output)
		if flusherOk {
			flusher.Flush()
		}
	}
}

func (d *Daemon) handleRemoteRun(w http.ResponseWriter, r *http.Request) {
	// REQUIRE mTLS verified cert fingerprint
	if !d.verifyRemoteFingerprint(r) {
		http.Error(w, "Unauthorized Remote Connection", http.StatusUnauthorized)
		return
	}

	var job Job
	if err := json.NewDecoder(r.Body).Decode(&job); err != nil {
		http.Error(w, "Invalid job payload", http.StatusBadRequest)
		return
	}

	if job.Type != "exec" {
		http.Error(w, "Unsupported job type", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	flusher, flusherOk := w.(http.Flusher)
	writeChunk := func(out JobOutput) {
		json.NewEncoder(w).Encode(out)
		if flusherOk {
			flusher.Flush()
		}
	}

	args := strings.Fields(job.Payload)
	if len(args) == 0 {
		writeChunk(JobOutput{Type: "error", Content: "empty command"})
		return
	}

	cmd := exec.Command(args[0], args[1:]...)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		writeChunk(JobOutput{Type: "error", Content: err.Error()})
		return
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		writeChunk(JobOutput{Type: "error", Content: err.Error()})
		return
	}

	if err := cmd.Start(); err != nil {
		writeChunk(JobOutput{Type: "error", Content: err.Error()})
		return
	}

	var wg sync.WaitGroup
	wg.Add(2)

	readStream := func(r io.Reader, isStderr bool) {
		defer wg.Done()
		scanner := bufioNewScanner(r) // custom scanner to handle very long lines
		for scanner.Scan() {
			outType := "stdout"
			if isStderr {
				outType = "stderr"
			}
			writeChunk(JobOutput{
				Type:    outType,
				Content: scanner.Text(),
			})
		}
	}

	go readStream(stdout, false)
	go readStream(stderr, true)

	wg.Wait()
	exitCode := 0
	if err := cmd.Wait(); err != nil {
		exitCode = 1
		if exitError, ok := err.(*exec.ExitError); ok {
			exitCode = exitError.ExitCode()
		}
		writeChunk(JobOutput{Type: "error", Content: err.Error()})
	}

	writeChunk(JobOutput{
		Type:     "exit",
		ExitCode: exitCode,
	})
}

// bufioNewScanner builds a custom buffered scanner.
func bufioNewScanner(r io.Reader) *netScanner {
	return &netScanner{r: bufioNewReader(r)}
}

type netScanner struct {
	r   *bufioReader
	err error
	txt string
}

func (s *netScanner) Scan() bool {
	line, err := s.r.ReadLine()
	if err != nil {
		s.err = err
		return false
	}
	s.txt = string(line)
	return true
}

func (s *netScanner) Text() string {
	return s.txt
}

type bufioReader struct {
	r *bufio.Reader
}

func bufioNewReader(r io.Reader) *bufioReader {
	return &bufioReader{r: bufio.NewReader(r)}
}

func (b *bufioReader) ReadLine() ([]byte, error) {
	line, isPrefix, err := b.r.ReadLine()
	if err != nil {
		return nil, err
	}
	for isPrefix {
		var rest []byte
		rest, isPrefix, err = b.r.ReadLine()
		if err != nil {
			return nil, err
		}
		line = append(line, rest...)
	}
	return line, nil
}

func (d *Daemon) handleRemoteChat(w http.ResponseWriter, r *http.Request) {
	if !d.verifyRemoteFingerprint(r) {
		http.Error(w, "Unauthorized peer certificate", http.StatusUnauthorized)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}

	ollamaReq, err := http.NewRequest(http.MethodPost, "http://127.0.0.1:11434/api/chat", bytes.NewReader(bodyBytes))
	if err != nil {
		http.Error(w, "Failed to prepare request to local Ollama", http.StatusInternalServerError)
		return
	}
	ollamaReq.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 5 * time.Minute}
	ollamaResp, err := client.Do(ollamaReq)
	if err != nil {
		http.Error(w, "Local Ollama server unreachable: "+err.Error(), http.StatusBadGateway)
		return
	}
	defer ollamaResp.Body.Close()

	if ollamaResp.StatusCode != http.StatusOK {
		w.WriteHeader(ollamaResp.StatusCode)
		io.Copy(w, ollamaResp.Body)
		return
	}

	w.Header().Set("Content-Type", "application/x-ndjson")
	w.WriteHeader(http.StatusOK)

	flusher, ok := w.(http.Flusher)
	buf := bufio.NewReader(ollamaResp.Body)
	for {
		line, err := buf.ReadString('\n')
		if err != nil {
			break
		}
		w.Write([]byte(line))
		if ok {
			flusher.Flush()
		}
	}
}

func (d *Daemon) handleLocalChat(w http.ResponseWriter, r *http.Request) {
	if !d.verifyLocal(r) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	targetIP := r.URL.Query().Get("ip")
	if targetIP == "" {
		targetIP = "127.0.0.1"
	}

	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}

	if targetIP == "127.0.0.1" {
		ollamaReq, err := http.NewRequest(http.MethodPost, "http://127.0.0.1:11434/api/chat", bytes.NewReader(bodyBytes))
		if err != nil {
			http.Error(w, "Failed to prepare local Ollama request", http.StatusInternalServerError)
			return
		}
		ollamaReq.Header.Set("Content-Type", "application/json")

		client := &http.Client{Timeout: 5 * time.Minute}
		ollamaResp, err := client.Do(ollamaReq)
		if err != nil {
			http.Error(w, "Local Ollama server unreachable: "+err.Error(), http.StatusBadGateway)
			return
		}
		defer ollamaResp.Body.Close()

		w.Header().Set("Content-Type", "application/x-ndjson")
		w.WriteHeader(ollamaResp.StatusCode)

		flusher, ok := w.(http.Flusher)
		buf := bufio.NewReader(ollamaResp.Body)
		for {
			line, err := buf.ReadString('\n')
			if err != nil {
				break
			}
			w.Write([]byte(line))
			if ok {
				flusher.Flush()
			}
		}
		return
	}

	cert, err := tls.X509KeyPair([]byte(d.cfg.CertificatePEM), []byte(d.cfg.PrivateKeyPEM))
	if err != nil {
		http.Error(w, "Local cert loading failed", http.StatusInternalServerError)
		return
	}

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			Certificates:       []tls.Certificate{cert},
			InsecureSkipVerify: true,
			VerifyPeerCertificate: func(rawCerts [][]byte, verifiedChains [][]*x509.Certificate) error {
				if len(rawCerts) == 0 {
					return fmt.Errorf("no certs presented")
				}
				peerCert, _ := x509.ParseCertificate(rawCerts[0])
				hash := sha256.Sum256(peerCert.Raw)
				parts := make([]string, 4)
				for i := 0; i < 4; i++ {
					parts[i] = fmt.Sprintf("%02X", hash[i])
				}
				peerFingerprint := strings.Join(parts, ":")

				cfg, _ := LoadConfig()
				if peerFingerprint == cfg.PublicKey {
					return nil
				}
				for _, dev := range cfg.Devices {
					if dev.Fingerprint == peerFingerprint {
						return nil
					}
				}
				return fmt.Errorf("untrusted peer certificate fingerprint: %s", peerFingerprint)
			},
		},
	}

	client := &http.Client{Transport: tr, Timeout: 5 * time.Minute}
	remoteURL := fmt.Sprintf("https://%s:8911/remote/chat", targetIP)

	remoteReq, err := http.NewRequest(http.MethodPost, remoteURL, bytes.NewReader(bodyBytes))
	if err != nil {
		http.Error(w, "Failed to prepare remote P2P request: "+err.Error(), http.StatusInternalServerError)
		return
	}
	remoteReq.Header.Set("Content-Type", "application/json")

	remoteResp, err := client.Do(remoteReq)
	if err != nil {
		http.Error(w, "Remote peer connection failed: "+err.Error(), http.StatusBadGateway)
		return
	}
	defer remoteResp.Body.Close()

	w.Header().Set("Content-Type", "application/x-ndjson")
	w.WriteHeader(remoteResp.StatusCode)

	flusher, ok := w.(http.Flusher)
	buf := bufio.NewReader(remoteResp.Body)
	for {
		line, err := buf.ReadString('\n')
		if err != nil {
			break
		}
		w.Write([]byte(line))
		if ok {
			flusher.Flush()
		}
	}
}
