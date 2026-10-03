#!/usr/bin/env python3
import concurrent.futures
import json
import os
import re
import socket
import time
import urllib.parse
import urllib.request
from datetime import datetime, timezone
from pathlib import Path

SOURCES = [
    ("ebrasha", "https://github.com/ebrasha/free-v2ray-public-list/raw/refs/heads/main/vless_configs.txt"),
    ("gfpcom", "https://raw.githubusercontent.com/wiki/gfpcom/free-proxy-list/lists/vless.txt"),
    ("kort0881", "https://github.com/kort0881/vpn-vless-configs-russia/raw/refs/heads/main/data/githubmirror/ru-sni/vless.txt"),
    ("radikal-fast", "https://github.com/0xRadikal/Free-v2ray-Configs/raw/refs/heads/main/fast/configs.txt"),
]
DISCOVERY_PAGES = [
    ("vlessnode", "https://vlessnode.github.io/"),
    ("freevlessnode", "https://freevlessnode.github.io/"),
]
OUT = Path("directory/nodes.json")
SEED = Path("cmd/jamshidix/assets/directory.seed.json")
MAX_CANDIDATES = 1500
MAX_OUTPUT = 250
TIMEOUT = 10
FALLBACK_LINKS = [
    "vless://3b002dc2-fd76-4eca-b6cd-7459ecca765e@xray2-direct-mci1-fs-ce.freesocks.work:443?security=reality&encryption=none&pbk=7j3hmlHrEDk3Kjx-dFG3E3jf0c6PZF45trIPFsuKWmI&headerType=none&fp=chrome&type=tcp&flow=xtls-rprx-vision&sni=dl.google.com&sid=1",
    "vless://3c092bd5-f723-4a64-b576-9fb0fd9d3543@2.27.54.106:50098?type=tcp&security=reality&flow=xtls-rprx-vision&fp=firefox&pbk=0_FWbcTe6tvBuSZ7_7PrMaNmg44a8VxYRXPKjZuERwQ&sid=a150dff4&sni=eh.vk.com",
    "vless://d56716f5-a22a-463b-89ad-4f65a54d994a@65.109.217.192:48349?security=reality&encryption=none&pbk=bdtDRp63jl4nMZRPN2wNlKJ6Yt77SF1uSrfkQ8NBKks&headerType=none&fp=chrome&type=tcp&flow=xtls-rprx-vision&sni=amp-api-edge.apps.apple.com&sid=c679641e0190177b",
    "vless://d5e5f5a1-0f23-4c49-af9f-4be445bf9d8e@82.118.16.189:443?encryption=none&flow=xtls-rprx-vision&fp=chrome&pbk=SjGkzJQ9I6ZZjC3V73atcPMczhMm0oiTrB8mUQXz1HY&security=reality&sid=37fbf155c7a9f3ff&sni=www.cloudflare.com&type=tcp",
    "vless://ff6e1028-b337-4062-a4db-86c784a9cdb4@nl1.nevcore.ru:8443?security=reality&encryption=none&headerType=none&fp=firefox&type=tcp&flow=xtls-rprx-vision&sni=www.amazon.com&sid=83a1c3d1295348be&pbk=8d--Q-ukEleKlt5yMPF50BK_76VhZ9jiBvZzGONX0yw",
    "vless://d00d9d18-20cf-4191-be76-6f73d43a27ec@178.79.149.80:22117?security=reality&type=tcp&sni=www.cloudflare.com&fp=chrome&flow=xtls-rprx-vision&sid=c7994be0&pbk=Vqwhml0SG253PsBtx9xi3GJY-hhyA6sgGUJ_wVI7Q1g&encryption=none",
    "vless://4bdeee92-97e8-414d-bef6-ec1d5e2ab73b@164.37.100.195:443?flow=xtls-rprx-vision&encryption=none&security=reality&sni=tomaz1337shellbuy.vvzzkontaktezzvv.garden&pbk=s0ypPx0AL3k2KOk74tbm-aE-i7eIR_MYVzfLHZeLaUk&sid=2e124af254a65e7c&type=tcp&headerType=none",
]

UUID_RE = re.compile(r"^[0-9a-fA-F-]{36}$")
PBK_RE = re.compile(r"^[A-Za-z0-9_-]{43}$")
SID_RE = re.compile(r"^([0-9a-fA-F]{2}){0,8}$")


def get_text(url):
    req = urllib.request.Request(url, headers={"User-Agent": "JamshidixCollector/0.3"})
    with urllib.request.urlopen(req, timeout=TIMEOUT) as r:
        return r.read(8_000_000).decode("utf-8", "replace")


def discover_latest_text_page(base_url):
    text = get_text(base_url)
    links = re.findall(r'https://[A-Za-z0-9.-]+/uploads/[^\s"\']+\.(?:txt|yaml|yml|json)', text)
    if not links:
        return None
    dated = []
    for link in links:
        m = re.search(r"(20\d{6})", link)
        dated.append((m.group(1) if m else "00000000", link))
    return sorted(dated, reverse=True)[0][1]


def parse_vless(line, source):
    m = re.search(r"(vless://[^\s'\"<>]+)", line.strip())
    if not m:
        return None
    raw = m.group(1).rstrip(".,;)")
    try:
        u = urllib.parse.urlsplit(raw)
        q = urllib.parse.parse_qs(u.query)
        security = q.get("security", [""])[0].lower()
        transport = q.get("type", ["tcp"])[0].lower()
        encryption = q.get("encryption", ["none"])[0].lower()
        flow = q.get("flow", [""])[0]
        sni = q.get("sni", [""])[0]
        fp = q.get("fp", ["chrome"])[0].lower()
        pbk = q.get("pbk", [""])[0]
        sid = q.get("sid", [""])[0]
        uuid = urllib.parse.unquote(u.username or "")
        host = u.hostname or ""
        port = u.port or 443
    except Exception:
        return None
    if security != "reality" or transport != "tcp" or encryption != "none":
        return None
    if flow != "xtls-rprx-vision":
        return None
    if not UUID_RE.match(uuid) or not PBK_RE.match(pbk) or not SID_RE.match(sid):
        return None
    if not sni or not host or len(host) > 253:
        return None
    if not re.fullmatch(r"[A-Za-z0-9](?:[A-Za-z0-9-]{0,61}[A-Za-z0-9])?(?:\.[A-Za-z0-9](?:[A-Za-z0-9-]{0,61}[A-Za-z0-9])?)*", sni):
        return None
    if fp not in {"chrome", "firefox", "safari", "edge", "ios", "android", "random", "randomized", "360", "qq"}:
        return None
    try:
        socket.inet_aton(host)
        if host.startswith(("10.", "127.", "192.168.", "172.16.")):
            return None
    except OSError:
        if "." not in host:
            return None
    name = urllib.parse.unquote(u.fragment or "") or host
    key = "|".join([host.lower(), str(port), uuid.lower(), pbk, sid.lower(), sni.lower()])
    return {
        "id": __import__("hashlib").sha256(key.encode()).hexdigest()[:16],
        "name": name[:100],
        "source": source,
        "server": host,
        "port": port,
        "uuid": uuid,
        "public_key": pbk,
        "short_id": sid,
        "sni": sni,
        "flow": flow,
        "fingerprint": fp,
        "protocol": "VLESS/REALITY",
        "transport": "tcp",
        "remote_ok": False,
        "remote_latency_ms": 0,
        "fetched_at": datetime.now(timezone.utc).isoformat(),
    }


def tcp_probe(node):
    start = time.monotonic()
    try:
        with socket.create_connection((node["server"], node["port"]), timeout=2.0):
            pass
        node["remote_ok"] = True
        node["remote_latency_ms"] = max(1, int((time.monotonic() - start) * 1000))
        return node
    except Exception:
        return None


def main():
    raw_blobs = []
    for link in FALLBACK_LINKS:
        node = parse_vless(link, "seed-fallback")
        if node:
            raw_blobs.append(("seed-fallback", link))
    for name, url in SOURCES:
        try:
            raw_blobs.append((name, get_text(url)))
        except Exception as exc:
            print(f"source failed: {name}: {exc}")

    for name, page in DISCOVERY_PAGES:
        try:
            latest = discover_latest_text_page(page)
            if latest:
                raw_blobs.append((name, get_text(latest)))
        except Exception as exc:
            print(f"discovery failed: {name}: {exc}")

    candidates = {}
    for source, blob in raw_blobs:
        lines = blob.splitlines() if "\n" in blob else [blob]
        for line in lines:
            node = parse_vless(line, source)
            if node:
                candidates[node["id"]] = node
                if len(candidates) >= MAX_CANDIDATES:
                    break

    print(f"parsed candidates: {len(candidates)}")
    nodes = list(candidates.values())
    with concurrent.futures.ThreadPoolExecutor(max_workers=50) as pool:
        checked = [x for x in pool.map(tcp_probe, nodes) if x]
    checked.sort(key=lambda x: (x["remote_latency_ms"], x["server"]))
    checked = checked[:MAX_OUTPUT]
    if not checked:
        print("remote TCP probing returned no reachable nodes; publishing validated candidates for local preflight")
        checked = nodes[:MAX_OUTPUT]

    now = datetime.now(timezone.utc).isoformat()
    directory = {
        "version": 1,
        "generated_at": now,
        "nodes": checked,
        "sources": [x[0] for x in raw_blobs],
    }
    OUT.parent.mkdir(parents=True, exist_ok=True)
    data = json.dumps(directory, ensure_ascii=False, indent=2) + "\n"
    OUT.write_text(data, encoding="utf-8")
    SEED.parent.mkdir(parents=True, exist_ok=True)
    SEED.write_text(data, encoding="utf-8")
    print(f"published nodes: {len(checked)}")
    if len(checked) == 0:
        # Keep a valid directory even if the runner cannot reach any public node.
        # Local client-side preflight will decide actual usability.
        checked = nodes[:MAX_OUTPUT]
    if len(checked) == 0:
        raise SystemExit("no validated VLESS/REALITY candidates found")


if __name__ == "__main__":
    main()
