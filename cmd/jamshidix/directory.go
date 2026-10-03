package main

import (
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Node struct {
	ID               string `json:"id"`
	Name             string `json:"name"`
	Country          string `json:"country,omitempty"`
	City             string `json:"city,omitempty"`
	Region           string `json:"region,omitempty"`
	Priority         int    `json:"priority"`
	Source           string `json:"source"`
	Server           string `json:"server"`
	Port             int    `json:"port"`
	UUID             string `json:"uuid"`
	PublicKey        string `json:"public_key"`
	ShortID          string `json:"short_id"`
	SNI              string `json:"sni"`
	Flow             string `json:"flow"`
	Fingerprint      string `json:"fingerprint"`
	Protocol         string `json:"protocol"`
	Transport        string `json:"transport"`
	RemoteOK         bool   `json:"remote_ok"`
	RemoteLatencyMs  int    `json:"remote_latency_ms,omitempty"`
	FetchedAt        string `json:"fetched_at"`
	LocalOK          bool   `json:"-"`
	LocalLatencyMs   int    `json:"-"`
}

type Directory struct {
	Version     int       `json:"version"`
	GeneratedAt string    `json:"generated_at"`
	Nodes       []Node    `json:"nodes"`
	Sources     []string  `json:"sources,omitempty"`
}

var (
	directoryMirrors = []string{
		"https://cdn.jsdelivr.net/gh/jamshidimoh/jamshidix@main/directory/nodes.json",
		"https://raw.githubusercontent.com/jamshidimoh/jamshidix/main/directory/nodes.json",
		"https://raw.githack.com/jamshidimoh/jamshidix/main/directory/nodes.json",
		"https://cdn.statically.io/gh/jamshidimoh/jamshidix@main/directory/nodes.json",
	}
	directoryHTTPClient = &http.Client{Timeout: 12 * time.Second}
)

const directoryCacheName = "directory.json"

//go:embed assets/directory.seed.json
var embeddedDirectorySeed []byte

func directoryCachePath() string { return filepath.Join(dataDir, directoryCacheName) }

func decodeDirectory(b []byte) (Directory, error) {
	var d Directory
	if err := json.Unmarshal(b, &d); err != nil {
		return Directory{}, fmt.Errorf("invalid directory JSON: %w", err)
	}
	if d.Version < 1 {
		return Directory{}, errors.New("unsupported directory version")
	}
	valid := d.Nodes[:0]
	seen := make(map[string]bool)
	for _, n := range d.Nodes {
		if n.ID == "" || n.Server == "" || n.Port < 1 || n.Port > 65535 ||
			n.UUID == "" || n.PublicKey == "" || n.ShortID == "" || n.SNI == "" {
			continue
		}
		key := strings.Join([]string{n.Server, strconv.Itoa(n.Port), n.UUID, n.PublicKey, n.ShortID, n.SNI}, "|")
		if seen[key] {
			continue
		}
		seen[key] = true
		valid = append(valid, n)
	}
	d.Nodes = valid
	return d, nil
}

func loadDirectory() (Directory, error) {
	if b, err := os.ReadFile(directoryCachePath()); err == nil {
		if d, err := decodeDirectory(b); err == nil && len(d.Nodes) > 0 {
			return d, nil
		}
	}
	if len(embeddedDirectorySeed) > 0 {
		if d, err := decodeDirectory(embeddedDirectorySeed); err == nil {
			return d, nil
		}
	}
	return Directory{}, errors.New("no server directory is available")
}

func fetchDirectory(url string) (Directory, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return Directory{}, err
	}
	req.Header.Set("User-Agent", "Jamshidix/0.3")
	resp, err := directoryHTTPClient.Do(req)
	if err != nil {
		return Directory{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return Directory{}, fmt.Errorf("HTTP %s", resp.Status)
	}
	b, err := io.ReadAll(io.LimitReader(resp.Body, 6<<20))
	if err != nil {
		return Directory{}, err
	}
	return decodeDirectory(b)
}

func saveDirectory(d Directory) error {
	if err := os.MkdirAll(dataDir, 0o755); err != nil {
		return err
	}
	b, err := json.MarshalIndent(d, "", "  ")
	if err != nil {
		return err
	}
	tmp := directoryCachePath() + ".tmp"
	if err := os.WriteFile(tmp, b, 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, directoryCachePath())
}

func refreshDirectory() (Directory, string, error) {
	var errs []string
	for _, mirror := range directoryMirrors {
		d, err := fetchDirectory(mirror)
		if err != nil {
			errs = append(errs, mirror+": "+err.Error())
			continue
		}
		if len(d.Nodes) == 0 {
			errs = append(errs, mirror+": empty directory")
			continue
		}
		if err := saveDirectory(d); err != nil {
			return d, mirror, fmt.Errorf("directory fetched but cache write failed: %w", err)
		}
		return d, mirror, nil
	}
	return Directory{}, "", fmt.Errorf("all directory mirrors failed: %s", strings.Join(errs, " | "))
}

func probeNode(n Node) (bool, int) {
	start := time.Now()
	host, port := net.JoinHostPort(n.Server, strconv.Itoa(n.Port))
	conn, err := net.DialTimeout("tcp", host, 2200*time.Millisecond)
	if err != nil {
		return false, -1
	}
	_ = conn.Close()
	return true, int(time.Since(start).Milliseconds())
}

func applyNodeView(d Directory, sortMode, region string) Directory {
	filtered := make([]Node, 0, len(d.Nodes))
	for _, n := range d.Nodes {
		if region == "" || region == "همه مناطق" || strings.EqualFold(n.Region, region) || (n.Region == "" && region == "نامشخص") {
			filtered = append(filtered, n)
		}
	}
	sort.SliceStable(filtered, func(i, j int) bool {
		a, b := filtered[i], filtered[j]
		switch sortMode {
		case "سرعت":
			if a.LocalOK != b.LocalOK { return a.LocalOK }
			if a.LocalOK && b.LocalOK && a.LocalLatencyMs != b.LocalLatencyMs { return a.LocalLatencyMs < b.LocalLatencyMs }
			if a.RemoteOK != b.RemoteOK { return a.RemoteOK }
			return a.Priority > b.Priority
		case "منطقه":
			ar, br := a.Region, b.Region
			if ar == "" { ar = "نامشخص" }
			if br == "" { br = "نامشخص" }
			if ar != br { return ar < br }
			return a.Priority > b.Priority
		case "تازگی":
			return a.FetchedAt > b.FetchedAt
		default:
			if a.LocalOK != b.LocalOK { return a.LocalOK }
			if a.Priority != b.Priority { return a.Priority > b.Priority }
			if a.RemoteLatencyMs != b.RemoteLatencyMs { return a.RemoteLatencyMs < b.RemoteLatencyMs }
			return a.Name < b.Name
		}
	})
	d.Nodes = filtered
	return d
}

func decorateLocalReachability(d Directory) Directory {
	if len(d.Nodes) == 0 {
		return d
	}
	indices := make([]int, len(d.Nodes))
	for i := range indices {
		indices[i] = i
	}
	sort.SliceStable(indices, func(i, j int) bool {
		a, b := d.Nodes[indices[i]], d.Nodes[indices[j]]
		if a.RemoteOK != b.RemoteOK {
			return a.RemoteOK
		}
		if a.RemoteLatencyMs == 0 {
			return false
		}
		if b.RemoteLatencyMs == 0 {
			return true
		}
		return a.RemoteLatencyMs < b.RemoteLatencyMs
	})
	if len(indices) > 80 {
		indices = indices[:80]
	}

	var wg sync.WaitGroup
	sem := make(chan struct{}, 20)
	for _, idx := range indices {
		idx := idx
		wg.Add(1)
		go func() {
			defer wg.Done()
			sem <- struct{}{}
			ok, latency := probeNode(d.Nodes[idx])
			<-sem
			d.Nodes[idx].LocalOK = ok
			d.Nodes[idx].LocalLatencyMs = latency
		}()
	}
	wg.Wait()

	sort.SliceStable(d.Nodes, func(i, j int) bool {
		a, b := d.Nodes[i], d.Nodes[j]
		if a.LocalOK != b.LocalOK {
			return a.LocalOK
		}
		if a.LocalOK && b.LocalOK && a.LocalLatencyMs != b.LocalLatencyMs {
			return a.LocalLatencyMs < b.LocalLatencyMs
		}
		if a.RemoteOK != b.RemoteOK {
			return a.RemoteOK
		}
		return a.Name < b.Name
	})
	return d
}
