"""
Disk and artifact endpoints.
"""

import json
from collections.abc import Mapping, MutableMapping
from typing import Any, BinaryIO, cast

from ..client_types import RequesterProtocol
from ..uploads import FileUpload, normalize_file_upload


def _bool_to_str(value: bool) -> str:
    return "true" if value else "false"


class DisksAPI:
    def __init__(self, requester: RequesterProtocol) -> None:
        self._requester = requester
        self.artifacts = DiskArtifactsAPI(requester)

    def list(self) -> Any:
        return self._requester.request("GET", "/disk")

    def create(self) -> Any:
        return self._requester.request("POST", "/disk")

    def delete(self, disk_id: str) -> None:
        self._requester.request("DELETE", f"/disk/{disk_id}")


class DiskArtifactsAPI:
    def __init__(self, requester: RequesterProtocol) -> None:
        self._requester = requester

    def upsert(
        self,
        disk_id: str,
        *,
        file: FileUpload
        | tuple[str, BinaryIO | bytes]
        | tuple[str, BinaryIO | bytes, str | None],
        file_path: str | None = None,
        meta: Mapping[str, Any] | MutableMapping[str, Any] | None = None,
    ) -> Any:
        upload = normalize_file_upload(file)
        files = {"file": upload.as_httpx()}
        form: dict[str, Any] = {}
        if file_path:
            form["file_path"] = file_path
        if meta is not None:
            form["meta"] = json.dumps(cast(Mapping[str, Any], meta))
        return self._requester.request(
            "POST",
            f"/disk/{disk_id}/artifact",
            data=form or None,
            files=files,
        )

    def get(
        self,
        disk_id: str,
        *,
        file_path: str,
        with_public_url: bool | None = None,
        with_content: bool | None = None,
        expire: int | None = None,
    ) -> Any:
        params: dict[str, Any] = {"file_path": file_path}
        if with_public_url is not None:
            params["with_public_url"] = _bool_to_str(with_public_url)
        if with_content is not None:
            params["with_content"] = _bool_to_str(with_content)
        if expire is not None:
            params["expire"] = expire
        return self._requester.request("GET", f"/disk/{disk_id}/artifact", params=params)

    def update(
        self,
        disk_id: str,
        *,
        file_path: str,
        meta: Mapping[str, Any] | MutableMapping[str, Any],
    ) -> Any:
        payload = {
            "file_path": file_path,
            "meta": json.dumps(cast(Mapping[str, Any], meta)),
        }
        return self._requester.request("PUT", f"/disk/{disk_id}/artifact", json_data=payload)

    def delete(
        self,
        disk_id: str,
        *,
        file_path: str,
    ) -> None:
        params = {"file_path": file_path}
        self._requester.request("DELETE", f"/disk/{disk_id}/artifact", params=params)

    def list(
        self,
        disk_id: str,
        *,
        path: str | None = None,
    ) -> Any:
        params: dict[str, Any] = {}
        if path is not None:
            params["path"] = path
        return self._requester.request("GET", f"/disk/{disk_id}/artifact/ls", params=params or None)
