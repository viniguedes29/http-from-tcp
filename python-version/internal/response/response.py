"""
HTTP Response Writer — escreve respostas HTTP/1.1 formatadas no socket TCP.

Usa uma máquina de estados para garantir a ordem correta:
  INITIAL -> STATUS_LINE -> HEADERS -> BODY (ou CHUNKED_BODY -> BODY)
"""

from __future__ import annotations

from enum import IntEnum, auto
from typing import BinaryIO

from internal.headers import Headers


class StatusCode(IntEnum):
    OK = 200
    BAD_REQUEST = 400
    INTERNAL_SERVER_ERROR = 500


_REASON_PHRASES = {
    StatusCode.OK: "OK",
    StatusCode.BAD_REQUEST: "Bad Request",
    StatusCode.INTERNAL_SERVER_ERROR: "Internal Server Error",
}


class WriterState(IntEnum):
    INITIAL = auto()
    STATUS_LINE = auto()
    HEADERS = auto()
    BODY = auto()
    CHUNKED_BODY = auto()


class WriterStateError(Exception):
    pass


class Writer:
    """Escreve uma resposta HTTP/1.1 incrementalmente num socket/writer."""

    def __init__(self, conn: BinaryIO):
        self._conn = conn
        self._state = WriterState.INITIAL

    def write_status_line(self, status: StatusCode) -> None:
        if self._state != WriterState.INITIAL:
            raise WriterStateError(
                f"expected INITIAL, got {self._state.name}"
            )

        reason = _REASON_PHRASES.get(status, "")
        line = f"HTTP/1.1 {status.value} {reason}\r\n"
        self._conn.write(line.encode("ascii"))
        self._state = WriterState.STATUS_LINE

    def write_headers(self, headers: Headers) -> None:
        if self._state != WriterState.STATUS_LINE:
            raise WriterStateError(
                f"expected STATUS_LINE, got {self._state.name}"
            )

        buf = bytearray()
        for key, value in headers.items():
            buf.extend(f"{key}: {value}\r\n".encode("ascii"))
        buf.extend(b"\r\n")

        self._conn.write(bytes(buf))
        self._state = WriterState.HEADERS

    def write_body(self, body: bytes) -> None:
        if self._state != WriterState.HEADERS:
            raise WriterStateError(
                f"expected HEADERS, got {self._state.name}"
            )

        self._conn.write(body)
        self._state = WriterState.BODY

    def write_chunked_body(self, chunk: bytes) -> int:
        if self._state not in (WriterState.HEADERS, WriterState.CHUNKED_BODY):
            raise WriterStateError(
                f"expected HEADERS or CHUNKED_BODY, got {self._state.name}"
            )

        if self._state == WriterState.HEADERS:
            self._state = WriterState.CHUNKED_BODY

        size_line = f"{len(chunk):x}\r\n".encode("ascii")
        self._conn.write(size_line)
        self._conn.write(chunk)
        self._conn.write(b"\r\n")
        return len(chunk)

    def write_chunked_body_end(self) -> None:
        if self._state != WriterState.CHUNKED_BODY:
            raise WriterStateError(
                f"expected CHUNKED_BODY, got {self._state.name}"
            )

        self._conn.write(b"0\r\n\r\n")
        self._state = WriterState.BODY

    def write_trailers(self, trailers: Headers) -> None:
        if self._state != WriterState.CHUNKED_BODY:
            raise WriterStateError(
                f"expected CHUNKED_BODY, got {self._state.name}"
            )

        # chunk final (tamanho 0)
        self._conn.write(b"0\r\n")

        buf = bytearray()
        for key, value in trailers.items():
            buf.extend(f"{key}: {value}\r\n".encode("ascii"))
        buf.extend(b"\r\n")

        self._conn.write(bytes(buf))
        self._state = WriterState.BODY


def get_default_headers(content_length: int) -> Headers:
    h = Headers()
    h.set("Content-Length", str(content_length))
    h.set("Content-Type", "text/plain")
    h.set("Connection", "close")
    return h
