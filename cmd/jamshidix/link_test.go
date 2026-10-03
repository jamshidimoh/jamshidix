package main

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

const (
	testUUID = "11111111-2222-3333-4444-555555555555"
	testPBK  = "eRq5LAVjEDHmQzmUqRIskIm5pQ36svx4YkKSaiLmmR8"
)

func goodProfile() Profile {
	return Profile{Server: "192.0.2.10", Port: 443, UUID: testUUID, PublicKey: testPBK, ShortID: "0123abcd", SNI: "www.cloudflare.com"}
}

func TestParseVLESSLink(t *testing.T) {
	link := "vless://" + testUUID + "@192.0.2.10:8443?encryption=none&flow=xtls-rprx-vision&security=reality&sni=www.cloudflare.com&fp=Chrome&pbk=" + testPBK + "&sid=0123abcd&type=tcp#My%20Server"
	p, err := parseVLESSLink(link)
	if err != nil {
		t.Fatal(err)
	}
	want := Profile{Server: "192.0.2.10", Port: 8443, UUID: testUUID, PublicKey: testPBK, ShortID: "0123abcd", SNI: "www.cloudflare.com", Flow: defaultFlow, Fingerprint: "chrome"}
	if p != want {
		t.Fatalf("got %+v want %+v", p, want)
	}
}

func TestParseDefaults(t *testing.T) {
	p, err := parseVLESSLink("vless://" + testUUID + "@example.org?security=reality&sni=www.cloudflare.com&pbk=" + testPBK)
	if err != nil {
		t.Fatal(err)
	}
	if p.Port != 443 || p.Flow != defaultFlow || p.Fingerprint != defaultFingerprint || p.ShortID != "" {
		t.Fatalf("unexpected defaults: %+v", p)
	}
}

func TestParseRejects(t *testing.T) {
	base := "vless://" + testUUID + "@192.0.2.10:443?security=reality&sni=www.cloudflare.com&pbk=" + testPBK + "&sid=0123abcd"
	bad := map[string]string{
		"scheme":       strings.Replace(base, "vless://", "vmess://", 1),
		"no reality":   strings.Replace(base, "security=reality", "security=tls", 1),
		"ws transport": base + "&type=ws",
		"grpc":         base + "&type=grpc",
		"bad uuid":     strings.Replace(base, testUUID, "not-a-uuid", 1),
		"bad pbk":      strings.Replace(base, testPBK, "short", 1),
		"odd sid":      strings.Replace(base, "sid=0123abcd", "sid=abc", 1),
		"non-hex sid":  strings.Replace(base, "sid=0123abcd", "sid=zzzz", 1),
		"port 0":       strings.Replace(base, ":443?", ":0?", 1),
		"port 70000":   strings.Replace(base, ":443?", ":70000?", 1),
		"quote in sni": strings.Replace(base, "sni=www.cloudflare.com", "sni=a%22b.com", 1),
		"space in sni": strings.Replace(base, "sni=www.cloudflare.com", "sni=a%20b.com", 1),
		"bad flow":     base + "&flow=xtls-rprx-direct",
		"bad fp":       base + "&fp=evil",
		"encryption":   base + "&encryption=aes",
	}
	for name, link := range bad {
		if _, err := parseVLESSLink(link); err == nil {
			t.Errorf("%s: expected an error", name)
		}
	}
}

func TestFindLink(t *testing.T) {
	text := "my server:\r\n  vless://" + testUUID + "@1.2.3.4:443?security=reality#x, thanks\r\n"
	got := findLink(text)
	if !strings.HasPrefix(got, "vless://"+testUUID) || strings.ContainsAny(got, " \r\n") {
		t.Fatalf("bad extraction: %q", got)
	}
	if findLink("nothing here") != "" {
		t.Fatal("expected empty result")
	}
}

func TestLinkRoundTrip(t *testing.T) {
	in := goodProfile()
	link, err := buildLink(in, "Jamshidix")
	if err != nil {
		t.Fatal(err)
	}
	out, err := parseVLESSLink(link)
	if err != nil {
		t.Fatal(err)
	}
	in.Flow, in.Fingerprint = defaultFlow, defaultFingerprint
	if in != out {
		t.Fatalf("round trip mismatch:\n in=%+v\nout=%+v", in, out)
	}
}

func TestRenderClientConfig(t *testing.T) {
	raw, err := renderClientConfig(goodProfile())
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(raw), "REPLACE_WITH") {
		t.Fatal("placeholder left in rendered config")
	}
	var cfg struct {
		Inbounds  []map[string]any `json:"inbounds"`
		Outbounds []map[string]any `json:"outbounds"`
	}
	if err := json.Unmarshal(raw, &cfg); err != nil {
		t.Fatal(err)
	}
	var proxy map[string]any
	for _, o := range cfg.Outbounds {
		if o["tag"] == "proxy" {
			proxy = o
		}
	}
	if proxy["server"] != "192.0.2.10" || proxy["uuid"] != testUUID || proxy["server_port"].(float64) != 443 {
		t.Fatalf("proxy outbound not filled: %v", proxy)
	}
	reality := proxy["tls"].(map[string]any)["reality"].(map[string]any)
	if reality["public_key"] != testPBK || reality["short_id"] != "0123abcd" {
		t.Fatalf("reality not filled: %v", reality)
	}
	if len(cfg.Inbounds) != 2 {
		t.Fatalf("tun + mixed inbounds expected, got %d", len(cfg.Inbounds))
	}
}

func TestRenderRejectsInvalidProfile(t *testing.T) {
	p := goodProfile()
	p.SNI = `x","evil":"1`
	if _, err := renderClientConfig(p); err == nil {
		t.Fatal("expected validation error")
	}
}

// The Windows templates exist twice (go:embed cannot reach outside the package).
func TestTemplateCopiesInSync(t *testing.T) {
	other, err := os.ReadFile(filepath.Join("..", "..", "config", "client.config.template.json"))
	if err != nil {
		t.Fatal(err)
	}
	var a, b any
	if err := json.Unmarshal(clientTemplate, &a); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(other, &b); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(a, b) {
		t.Fatal("config/client.config.template.json and cmd/jamshidix/assets/client.config.template.json differ")
	}
}

// Optional end-to-end check with the real sing-box parser: SINGBOX_BIN=/path/to/sing-box
func TestRenderedConfigPassesSingBoxCheck(t *testing.T) {
	bin := os.Getenv("SINGBOX_BIN")
	if bin == "" {
		t.Skip("SINGBOX_BIN not set")
	}
	raw, err := renderClientConfig(goodProfile())
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "client.json")
	if err := os.WriteFile(path, raw, 0o600); err != nil {
		t.Fatal(err)
	}
	if out, err := exec.Command(bin, "check", "-c", path).CombinedOutput(); err != nil {
		t.Fatalf("sing-box check failed: %v\n%s", err, out)
	}
}
