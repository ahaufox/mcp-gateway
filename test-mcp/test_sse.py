#!/usr/bin/env python3
"""
MCP Streamable HTTP 传输层测试工具

测试 MCP 协议的 initialize / initialized / tools/list / ping 等流程，
支持 Streamable HTTP（默认）和 SSE 两种传输模式。

用法:
    # 默认使用 Streamable HTTP 测试 context7 服务
    python test-mcp/test_sse.py

    # 测试其他服务器
    python test-mcp/test_sse.py http://localhost:9090/jules/mcp MyToken

    # SSE 模式（需服务端启用 SSE 传输）
    python test-mcp/test_sse.py --transport sse http://localhost:1122/sse
"""
import requests
import json
import sys
import time


def test_streamable_http(url, token):
    """Streamable HTTP 模式: 直接 POST JSON-RPC 请求，同步获取响应。"""
    headers = {
        "Content-Type": "application/json",
        "Authorization": f"Bearer {token}",
    }

    session_id = None

    def send_request(payload):
        nonlocal session_id
        h = dict(headers)
        if session_id:
            h["mcp-session-id"] = session_id
        resp = requests.post(url, json=payload, headers=h, timeout=10)
        print(f"[{resp.status_code}] {payload.get('method', '(notification)')}")
        if resp.headers.get("mcp-session-id"):
            session_id = resp.headers["mcp-session-id"]
        try:
            body = resp.json()
            print(f"    -> {json.dumps(body, ensure_ascii=False, indent=2)}")
        except Exception:
            print(f"    -> {resp.text}")
        return resp

    print(f"[*] Streamable HTTP 模式 - 端点: {url}")
    print(f"[*] Token: {token}")
    print()

    # 1. initialize
    r = send_request({
        "jsonrpc": "2.0", "id": 1, "method": "initialize",
        "params": {
            "protocolVersion": "2024-11-05",
            "capabilities": {},
            "clientInfo": {"name": "mcp-test-client", "version": "1.0.0"},
        },
    })
    if r.status_code != 200:
        print("[!] initialize 失败，退出。")
        return False

    # 2. initialized notification (无需响应)
    send_request({
        "jsonrpc": "2.0", "method": "notifications/initialized",
    })

    # 3. tools/list
    send_request({
        "jsonrpc": "2.0", "id": 2, "method": "tools/list",
    })

    # 4. ping
    send_request({
        "jsonrpc": "2.0", "id": 3, "method": "ping",
    })

    print("[+] 测试完成。")
    return True


def test_sse(url, token):
    """SSE 模式: 建立 SSE 长连接，通过 endpoint 事件获取投递 URL 后进行交互。"""
    sse_headers = {
        "Accept": "text/event-stream",
        "Authorization": f"Bearer {token}",
        "Cache-Control": "no-cache",
    }

    print(f"[*] SSE 模式 - 端点: {url}")
    try:
        response = requests.get(url, headers=sse_headers, stream=True, timeout=10)
        response.raise_for_status()
    except Exception as e:
        print(f"[!] 无法建立 SSE 连接: {e}")
        return False

    print(f"[+] SSE 连接建立成功，状态码: {response.status_code}")

    post_url = None
    for line in response.iter_lines():
        if not line:
            continue
        decoded = line.decode("utf-8").strip()
        if decoded.startswith("data:"):
            data = decoded[5:].strip()
            # 首次收到的 data 即为 endpoint URL
            if data.startswith("/"):
                base = url.rstrip("/").rsplit("/", 1)[0]
                post_url = base + data
            elif data.startswith("http"):
                post_url = data
            print(f"[<-] endpoint: {post_url}")
            break

    if not post_url:
        print("[!] 未收到 endpoint 事件")
        return False

    # 发送 initialize
    init_payload = {
        "jsonrpc": "2.0", "id": 1, "method": "initialize",
        "params": {
            "protocolVersion": "2024-11-05",
            "capabilities": {},
            "clientInfo": {"name": "mcp-test-client", "version": "1.0.0"},
        },
    }
    post_headers = {"Content-Type": "application/json", "Authorization": f"Bearer {token}"}
    r = requests.post(post_url, json=init_payload, headers=post_headers, timeout=5)
    print(f"[{r.status_code}] initialize -> {r.text}")

    print("[+] SSE 测试完成。")
    return True


if __name__ == "__main__":
    transport = "streamable-http"
    args = [a for a in sys.argv[1:] if not a.startswith("--")]
    flags = [a for a in sys.argv[1:] if a.startswith("--")]

    for f in flags:
        if f in ("--sse", "--transport sse"):
            transport = "sse"

    url = "http://localhost:9090/context7/mcp"
    token = "DefaultTokens"
    if len(args) > 0:
        url = args[0]
    if len(args) > 1:
        token = args[1]

    if transport == "sse":
        test_sse(url, token)
    else:
        test_streamable_http(url, token)
