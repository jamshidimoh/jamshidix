package main

import "testing"

func TestDecodeDirectoryDeduplicates(t *testing.T) {
	b := []byte(`{"version":1,"nodes":[{"id":"a","server":"1.2.3.4","port":443,"uuid":"11111111-1111-1111-1111-111111111111","public_key":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","short_id":"a1","sni":"example.com","flow":"xtls-rprx-vision","fingerprint":"chrome"},{"id":"b","server":"1.2.3.4","port":443,"uuid":"11111111-1111-1111-1111-111111111111","public_key":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","short_id":"a1","sni":"example.com","flow":"xtls-rprx-vision","fingerprint":"chrome"}]}`)
	d, err := decodeDirectory(b)
	if err != nil {
		t.Fatal(err)
	}
	if len(d.Nodes) != 1 {
		t.Fatalf("want 1 node, got %d", len(d.Nodes))
	}
}

func TestApplyNodeViewPriority(t *testing.T) {
	d := Directory{Version: 1, Nodes: []Node{
		{ID: "slow", Name: "slow", Server: "1.1.1.1", Priority: 60, RemoteOK: true},
		{ID: "fast", Name: "fast", Server: "2.2.2.2", Priority: 90, RemoteOK: true},
	}}
	got := applyNodeView(d, "اولویت", "همه مناطق")
	if len(got.Nodes) != 2 || got.Nodes[0].ID != "fast" {
		t.Fatalf("priority sort failed: %+v", got.Nodes)
	}
}

func TestApplyNodeViewRegionFilter(t *testing.T) {
	d := Directory{Version: 1, Nodes: []Node{
		{ID: "de", Server: "1.1.1.1", Region: "DE"},
		{ID: "jp", Server: "2.2.2.2", Region: "JP"},
	}}
	got := applyNodeView(d, "منطقه", "JP")
	if len(got.Nodes) != 1 || got.Nodes[0].ID != "jp" {
		t.Fatalf("region filter failed: %+v", got.Nodes)
	}
}
