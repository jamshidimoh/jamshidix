package main

import (
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/url"
	"regexp"
	"strconv"
	"strings"
)

//go:embed assets/client.config.template.json
var clientTemplate []byte

// Profile is everything a client needs to reach a VLESS + REALITY gateway.
type Profile struct {
	Server      string
	Port        int
	UUID        string
	PublicKey   string
	ShortID     string
	SNI         string
	Flow        string
	Fingerprint string
}

const (
	defaultFlow        = "xtls-rprx-vision"
	defaultFingerprint = "chrome"
)

var (
	uuidRe    = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)
	pbkRe     = regexp.MustCompile(`^[A-Za-z0-9_-]{43}$`)
	shortIDRe = regexp.MustCompile(`^([0-9a-fA-F]{2}){0,8}$`)
	hostRe    = regexp.MustCompile(`^[A-Za-z0-9]([A-Za-z0-9-]{0,61}[A-Za-z0-9])?(\.[A-Za-z0-9]([A-Za-z0-9-]{0,61}[A-Za-z0-9])?)*$`)
	linkRe    = regexp.MustCompile(`vless://[^\s"'<>]+`)

	allowedFingerprints = map[string]bool{
		"chrome": true, "firefox": true, "safari": true, "edge": true,
		"ios": true, "android": true, "random": true, "randomized": true,
		"360": true, "qq": true,
	}
)

func validHost(h string) bool {
	if h == "" || len(h) > 253 {
		return false
	}
	return net.ParseIP(h) != nil || hostRe.MatchString(h)
}

// Validate rejects anything that is not safe to place in a sing-box config.
func (p *Profile) Validate() error {
	if p.Flow == "" {
		p.Flow = defaultFlow
	}
	if p.Fingerprint == "" {
		p.Fingerprint = defaultFingerprint
	}
	switch {
	case !validHost(p.Server):
		return fmt.Errorf("invalid server address %q", p.Server)
	case p.Port < 1 || p.Port > 65535:
		return fmt.Errorf("invalid port %d", p.Port)
	case !uuidRe.MatchString(p.UUID):
		return errors.New("invalid UUID")
	case !pbkRe.MatchString(p.PublicKey):
		return errors.New("invalid REALITY public key (expected 43 base64url characters)")
	case !shortIDRe.MatchString(p.ShortID):
		return errors.New("invalid short id (expected 0-16 hex characters, even length)")
	case !validHost(p.SNI):
		return fmt.Errorf("invalid handshake host (SNI) %q", p.SNI)
	case p.Flow != defaultFlow:
		return fmt.Errorf("unsupported flow %q", p.Flow)
	case !allowedFingerprints[p.Fingerprint]:
		return fmt.Errorf("unsupported fingerprint %q", p.Fingerprint)
	}
	return nil
}

// findLink extracts the first vless:// link from arbitrary text (clipboard, file).
func findLink(text string) string {
	return strings.TrimRight(linkRe.FindString(text), ".,;)")
}

// parseVLESSLink understands the common share-link form:
// vless://UUID@HOST:PORT?security=reality&pbk=KEY&sid=ID&sni=HOST&flow=...&fp=...#name
func parseVLESSLink(raw string) (Profile, error) {
	raw = strings.TrimSpace(raw)
	if !strings.HasPrefix(strings.ToLower(raw), "vless://") {
		return Profile{}, errors.New("not a vless:// link")
	}
	u, err := url.Parse(raw)
	if err != nil {
		return Profile{}, fmt.Errorf("malformed link: %w", err)
	}
	q := u.Query()
	if sec := strings.ToLower(q.Get("security")); sec != "reality" {
		return Profile{}, fmt.Errorf("unsupported security %q: only REALITY links are supported", sec)
	}
	if t := strings.ToLower(q.Get("type")); t != "" && t != "tcp" {
		return Profile{}, fmt.Errorf("unsupported transport %q: only tcp is supported", t)
	}
	if e := strings.ToLower(q.Get("encryption")); e != "" && e != "none" {
		return Profile{}, fmt.Errorf("unsupported encryption %q", e)
	}
	p := Profile{
		Server:      u.Hostname(),
		UUID:        u.User.Username(),
		PublicKey:   q.Get("pbk"),
		ShortID:     q.Get("sid"),
		SNI:         q.Get("sni"),
		Flow:        q.Get("flow"),
		Fingerprint: strings.ToLower(q.Get("fp")),
	}
	p.Port = 443
	if ps := u.Port(); ps != "" {
		if p.Port, err = strconv.Atoi(ps); err != nil {
			return Profile{}, fmt.Errorf("invalid port %q", ps)
		}
	}
	if err := p.Validate(); err != nil {
		return Profile{}, err
	}
	return p, nil
}

// buildLink is the inverse of parseVLESSLink.
func buildLink(p Profile, name string) (string, error) {
	if err := p.Validate(); err != nil {
		return "", err
	}
	q := url.Values{}
	q.Set("encryption", "none")
	q.Set("flow", p.Flow)
	q.Set("security", "reality")
	q.Set("sni", p.SNI)
	q.Set("fp", p.Fingerprint)
	q.Set("pbk", p.PublicKey)
	q.Set("sid", p.ShortID)
	q.Set("type", "tcp")
	hostport := net.JoinHostPort(p.Server, strconv.Itoa(p.Port))
	return fmt.Sprintf("vless://%s@%s?%s#%s", p.UUID, hostport, q.Encode(), url.PathEscape(name)), nil
}

func child(m map[string]any, key string) (map[string]any, error) {
	v, ok := m[key].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("client template is missing %q", key)
	}
	return v, nil
}

// renderClientConfig fills the embedded template through structured JSON edits
// (never string substitution), so no input can alter the config structure.
func renderClientConfig(p Profile) ([]byte, error) {
	if err := p.Validate(); err != nil {
		return nil, err
	}
	var cfg map[string]any
	if err := json.Unmarshal(clientTemplate, &cfg); err != nil {
		return nil, fmt.Errorf("client template is not valid JSON: %w", err)
	}
	outs, ok := cfg["outbounds"].([]any)
	if !ok {
		return nil, errors.New("client template has no outbounds")
	}
	var proxy map[string]any
	for _, o := range outs {
		if m, ok := o.(map[string]any); ok && m["tag"] == "proxy" {
			proxy = m
		}
	}
	if proxy == nil {
		return nil, errors.New("client template has no \"proxy\" outbound")
	}
	tls, err := child(proxy, "tls")
	if err != nil {
		return nil, err
	}
	utls, err := child(tls, "utls")
	if err != nil {
		return nil, err
	}
	reality, err := child(tls, "reality")
	if err != nil {
		return nil, err
	}
	proxy["server"] = p.Server
	proxy["server_port"] = p.Port
	proxy["uuid"] = p.UUID
	proxy["flow"] = p.Flow
	tls["server_name"] = p.SNI
	utls["fingerprint"] = p.Fingerprint
	reality["public_key"] = p.PublicKey
	reality["short_id"] = p.ShortID
	return json.MarshalIndent(cfg, "", "  ")
}
