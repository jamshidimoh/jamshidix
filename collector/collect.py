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
    ("ebrasha", "https://raw.githubusercontent.com/ebrasha/free-v2ray-public-list/refs/heads/main/vless_configs.txt"),
    ("baarcuda", "https://raw.githubusercontent.com/Baarcuda/vpn-configs/master/top100-vless.txt"),
    ("radikal-verified", "https://raw.githubusercontent.com/0xRadikal/Free-v2ray-Configs/main/verified/configs.txt"),
    ("radikal", "https://raw.githubusercontent.com/0xRadikal/Free-v2ray-Configs/main/protocols/vless.txt"),
    ("morpheus", "https://raw.githubusercontent.com/morpheusadam/v2ray-config/main/subs/bundles/vless.txt"),
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


SOURCE_PRIORITY = {"radikal-verified": 95, "ebrasha": 90, "baarcuda": 86, "morpheus": 84, "radikal": 82, "freevlessnode": 78, "vlessnode": 80}

def infer_region(name, host, sni):
    text = f"{name} {host} {sni}".lower()
    flags = {"🇩🇪":"DE","🇳🇱":"NL","🇫🇷":"FR","🇬🇧":"GB","🇺🇸":"US","🇨🇦":"CA","🇯🇵":"JP","🇭🇰":"HK","🇸🇬":"SG","🇰🇷":"KR","🇷🇺":"RU","🇫🇮":"FI","🇵🇱":"PL","🇹🇷":"TR","🇦🇿":"AZ","🇸🇪":"SE","🇨🇭":"CH"}
    for flag, code in flags.items():
        if flag in text:
            return code
    patterns = [
        (r"\b(germany|de|fra|france)\b", "DE/FR"),
        (r"\b(netherlands|nl|amsterdam)\b", "NL"),
        (r"\b(france|paris)\b", "FR"),
        (r"\b(uk|united.?kingdom|london|gb)\b", "GB"),
        (r"\b(us|usa|america|new.?york|los.?angeles)\b", "US"),
        (r"\b(canada|toronto|montreal)\b", "CA"),
        (r"\b(japan|tokyo|jp)\b", "JP"),
        (r"\b(hong.?kong|hk)\b", "HK"),
        (r"\b(singapore|sg)\b", "SG"),
        (r"\b(korea|seoul|kr)\b", "KR"),
        (r"\b(russia|moscow|ru)\b", "RU"),
        (r"\b(finland|helsinki|fi)\b", "FI"),
        (r"\b(poland|warsaw|pl)\b", "PL"),
        (r"\b(turkey|istanbul|tr)\b", "TR"),
    ]
    for pattern, code in patterns:
        if re.search(pattern, text):
            return code
    return ""

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
        "region": infer_region(name, host, sni),
        "priority": SOURCE_PRIORITY.get(source, 70),
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
        node["priority"] = min(100, SOURCE_PRIORITY.get(node["source"], 70) + max(0, 18 - node["remote_latency_ms"] // 50) + (4 if node.get("region") else 0))
        return node
    except Exception:
        return None


def main():
    raw_blobs = []
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
        raise SystemExit("no validated VLESS/REALITY candidates found")


if __name__ == "__main__":
    main()
