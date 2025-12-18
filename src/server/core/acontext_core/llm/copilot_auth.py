import asyncio
import json
import os
from dataclasses import dataclass
from time import time
from typing import Any, Optional

from httpx import AsyncClient, Response

from ..env import LOG, DEFAULT_CORE_CONFIG


COPILOT_CLIENT_ID = "Iv1.b507a08c87ecfe98"

# Must match https://github.com/sst/opencode-copilot-auth
COPILOT_DEFAULT_HEADERS: dict[str, str] = {
    "User-Agent": "GitHubCopilotChat/0.32.4",
    "Editor-Version": "vscode/1.105.1",
    "Editor-Plugin-Version": "copilot-chat/0.32.4",
    "Copilot-Integration-Id": "vscode-chat",
}

# Must match https://github.com/sst/opencode-copilot-auth
DEVICE_FLOW_USER_AGENT = "GitHubCopilotChat/0.35.0"


@dataclass
class CopilotOAuthInfo:
    # Matches plugin naming: refresh is the GitHub OAuth device-flow token
    refresh: str
    access: str
    expires: int
    enterprise_url: Optional[str] = None


def _normalize_domain(url_or_domain: str) -> str:
    s = url_or_domain.strip()
    s = s.removeprefix("https://").removeprefix("http://")
    return s.removesuffix("/")


def _get_urls(domain: str) -> dict[str, str]:
    return {
        "DEVICE_CODE_URL": f"https://{domain}/login/device/code",
        "ACCESS_TOKEN_URL": f"https://{domain}/login/oauth/access_token",
        "COPILOT_API_KEY_URL": f"https://api.{domain}/copilot_internal/v2/token",
    }


def _copilot_base_url(domain: str, enterprise_url: Optional[str]) -> str:
    # Plugin uses https://api.githubcopilot.com for github.com
    # and https://copilot-api.<enterpriseDomain> for enterprise
    if enterprise_url:
        return f"https://copilot-api.{domain}"
    return "https://api.githubcopilot.com"


def _now_ms() -> int:
    return int(time() * 1000)


def _load_oauth_info_from_disk(path: str) -> Optional[CopilotOAuthInfo]:
    try:
        if not os.path.isfile(path):
            return None
        with open(path, "r", encoding="utf-8") as f:
            raw = json.load(f)
        if not isinstance(raw, dict):
            return None
        refresh = raw.get("refresh")
        if not refresh:
            return None
        return CopilotOAuthInfo(
            refresh=str(refresh),
            access=str(raw.get("access") or ""),
            expires=int(raw.get("expires") or 0),
            enterprise_url=raw.get("enterpriseUrl") or raw.get("enterprise_url"),
        )
    except Exception as e:
        LOG.warning(f"Failed to load Copilot token store: {e}")
        return None


def _save_oauth_info_to_disk(path: str, info: CopilotOAuthInfo) -> None:
    parent = os.path.dirname(path)
    if parent:
        os.makedirs(parent, exist_ok=True)
    payload: dict[str, Any] = {
        "type": "oauth",
        "refresh": info.refresh,
        "access": info.access,
        "expires": info.expires,
    }
    if info.enterprise_url:
        # Keep the plugin casing too for compatibility
        payload["enterpriseUrl"] = info.enterprise_url
        payload["enterprise_url"] = info.enterprise_url

    with open(path, "w", encoding="utf-8") as f:
        json.dump(payload, f)


async def _ensure_ok(resp: Response, err: str) -> None:
    if resp.is_success:
        return
    body = ""
    try:
        body = resp.text
    except Exception:
        body = "<unreadable>"
    raise RuntimeError(f"{err} (status={resp.status_code}) body={body}")


async def _start_device_flow(domain: str) -> dict[str, Any]:
    urls = _get_urls(domain)
    async with AsyncClient(timeout=30.0) as http:
        resp = await http.post(
            urls["DEVICE_CODE_URL"],
            headers={
                "Accept": "application/json",
                "Content-Type": "application/json",
                "User-Agent": DEVICE_FLOW_USER_AGENT,
            },
            json={"client_id": COPILOT_CLIENT_ID, "scope": "read:user"},
        )
        await _ensure_ok(resp, "Failed to initiate device authorization")
        return resp.json()


async def _poll_device_flow(domain: str, device_code: str, interval_s: int) -> str:
    urls = _get_urls(domain)
    async with AsyncClient(timeout=30.0) as http:
        while True:
            resp = await http.post(
                urls["ACCESS_TOKEN_URL"],
                headers={
                    "Accept": "application/json",
                    "Content-Type": "application/json",
                    "User-Agent": DEVICE_FLOW_USER_AGENT,
                },
                json={
                    "client_id": COPILOT_CLIENT_ID,
                    "device_code": device_code,
                    "grant_type": "urn:ietf:params:oauth:grant-type:device_code",
                },
            )
            if not resp.is_success:
                raise RuntimeError(
                    f"Device flow polling failed (status={resp.status_code})"
                )
            data = resp.json()

            access_token = data.get("access_token")
            if access_token:
                return str(access_token)

            if data.get("error") == "authorization_pending":
                await asyncio.sleep(interval_s)
                continue

            if data.get("error"):
                raise RuntimeError(f"Device flow failed: {data.get('error')}")

            await asyncio.sleep(interval_s)


async def login_via_device_flow_if_needed() -> CopilotOAuthInfo:
    """Ensure we have a stored GitHub OAuth token (refresh token in plugin terms).

    If token store exists, returns it. Otherwise, performs device flow and stores it.
    """

    token_path = DEFAULT_CORE_CONFIG.copilot_token_store_path
    existing = _load_oauth_info_from_disk(token_path)
    if existing and existing.refresh:
        return existing

    enterprise_url = DEFAULT_CORE_CONFIG.copilot_enterprise_url
    domain = _normalize_domain(enterprise_url) if enterprise_url else "github.com"

    device_data = await _start_device_flow(domain)

    verification_uri = device_data.get("verification_uri")
    user_code = device_data.get("user_code")
    interval_s = int(device_data.get("interval") or 5)

    # IMPORTANT: Surface this in container logs.
    LOG.warning("GitHub Copilot login required (device flow)")
    LOG.warning(f"Open: {verification_uri}")
    LOG.warning(f"Enter code: {user_code}")

    refresh_token = await _poll_device_flow(
        domain, device_code=str(device_data.get("device_code")), interval_s=interval_s
    )

    info = CopilotOAuthInfo(
        refresh=refresh_token,
        access="",
        expires=0,
        enterprise_url=domain if enterprise_url else None,
    )
    _save_oauth_info_to_disk(token_path, info)
    return info


async def get_copilot_access_token() -> str:
    """Return valid Copilot access token. Refreshes if expired."""

    token_path = DEFAULT_CORE_CONFIG.copilot_token_store_path
    info = await login_via_device_flow_if_needed()

    if info.access and info.expires and info.expires > _now_ms():
        return info.access

    enterprise_url = DEFAULT_CORE_CONFIG.copilot_enterprise_url
    domain = _normalize_domain(enterprise_url) if enterprise_url else "github.com"
    urls = _get_urls(domain)

    async with AsyncClient(timeout=30.0) as http:
        resp = await http.get(
            urls["COPILOT_API_KEY_URL"],
            headers={
                "Accept": "application/json",
                "Authorization": f"Bearer {info.refresh}",
                **COPILOT_DEFAULT_HEADERS,
            },
        )
        await _ensure_ok(resp, "Token refresh failed")
        token_data = resp.json()

    info.access = str(token_data.get("token") or "")
    info.expires = int(token_data.get("expires_at") or 0) * 1000
    if enterprise_url:
        info.enterprise_url = domain

    _save_oauth_info_to_disk(token_path, info)

    if not info.access:
        raise RuntimeError("Copilot token exchange returned empty token")

    return info.access


def copilot_openai_base_url() -> str:
    enterprise_url = DEFAULT_CORE_CONFIG.copilot_enterprise_url
    domain = _normalize_domain(enterprise_url) if enterprise_url else "github.com"
    return _copilot_base_url(domain=domain, enterprise_url=enterprise_url)
