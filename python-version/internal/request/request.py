"""
HTTP Request Parser — parseia requests HTTP/1.1 a partir de um stream TCP raw.

Usa uma máquina de estados:
  INIT -> PARSING_HEADERS -> PARSING_BODY -> DONE
"""

from __future__ import annotations

from dataclasses import dataclass, field
from enum import IntEnum, auto
from typing import BinaryIO

from internal.headers import Headers, InvalidHeaderError

CRLF = b"\r\n"
BUFFER_SIZE = 1024


class ParseState(IntEnum):
    INIT = auto()
    PARSING_HEADERS = auto()
    PARSING_BODY = auto()
    DONE = auto()


class MalformedRequestError(Exception):
    pass


class UnsupportedHTTPVersionError(Exception):
    pass


@dataclass
class RequestLine:
    method: str = ""
    request_target: str = ""
    http_version: str = ""


@dataclass
class Request:
    request_line: RequestLine = field(default_factory=RequestLine)
    headers: Headers = field(default_factory=Headers)
    body: bytes = b""
    state: ParseState = ParseState.INIT


def request_from_reader(reader: BinaryIO) -> Request:
    """
    Lê um request HTTP completo de um reader (socket, BytesIO, etc).
    Parseia incrementalmente à medida que dados chegam.

    Usa read1() quando disponível (BufferedReader de sockets) para não
    bloquear esperando o buffer inteiro — equivale ao behavior de Read() em Go.
    """
    buf = bytearray(BUFFER_SIZE)
    read_to_index = 0
    req = Request()

    # read1 retorna assim que QUALQUER dado está disponível (não bloqueia pro buffer cheio)
    _read = getattr(reader, "read1", None) or reader.read

    while req.state != ParseState.DONE:
        if read_to_index >= len(buf):
            buf.extend(bytearray(len(buf)))

        chunk = _read(BUFFER_SIZE)

        if not chunk:
            if req.state == ParseState.PARSING_BODY:
                cl_str, has_cl = req.headers.get_header("content-length")
                if has_cl and cl_str:
                    cl = int(cl_str)
                    if len(req.body) < cl:
                        raise MalformedRequestError(
                            "body shorter than content-length"
                        )
            req.state = ParseState.DONE
            break

        end = read_to_index + len(chunk)
        if end > len(buf):
            buf.extend(bytearray(end - len(buf)))
        buf[read_to_index:end] = chunk
        read_to_index = end

        bytes_consumed = _parse(req, bytes(buf[:read_to_index]))

        if bytes_consumed > 0:
            buf[:read_to_index - bytes_consumed] = buf[bytes_consumed:read_to_index]
            read_to_index -= bytes_consumed

    return req


def _parse(req: Request, data: bytes) -> int:
    total = 0
    while req.state != ParseState.DONE:
        n = _parse_single(req, data[total:])
        if n == 0:
            break
        total += n
    return total


def _parse_single(req: Request, data: bytes) -> int:
    if req.state == ParseState.INIT:
        return _parse_request_line(req, data)
    elif req.state == ParseState.PARSING_HEADERS:
        return _parse_headers(req, data)
    elif req.state == ParseState.PARSING_BODY:
        return _parse_body(req, data)
    elif req.state == ParseState.DONE:
        raise MalformedRequestError("trying to parse when already done")
    else:
        raise MalformedRequestError("invalid state")


def _parse_request_line(req: Request, data: bytes) -> int:
    rn_idx = data.find(CRLF)
    if rn_idx == -1:
        return 0

    line = data[:rn_idx]
    parts = line.split(b" ")
    if len(parts) != 3:
        raise MalformedRequestError("request line must have 3 parts")

    method, target, version = parts

    if version != b"HTTP/1.1":
        raise UnsupportedHTTPVersionError(f"unsupported: {version.decode()}")

    req.request_line = RequestLine(
        method=method.decode("ascii"),
        request_target=target.decode("ascii"),
        http_version="1.1",
    )
    req.state = ParseState.PARSING_HEADERS
    return rn_idx + 2


def _parse_headers(req: Request, data: bytes) -> int:
    n, done = req.headers.parse(data)
    if n == 0:
        return 0
    if done:
        req.state = ParseState.PARSING_BODY
    return n


def _parse_body(req: Request, data: bytes) -> int:
    cl_str, has_cl = req.headers.get_header("content-length")

    if not has_cl or not cl_str or cl_str == "0":
        req.state = ParseState.DONE
        return 0

    try:
        content_length = int(cl_str)
    except ValueError:
        raise MalformedRequestError("invalid content-length")

    if content_length < 0:
        raise MalformedRequestError("negative content-length")

    req.body += data

    if len(req.body) > content_length:
        raise MalformedRequestError("body exceeds content-length")

    if len(req.body) == content_length:
        req.state = ParseState.DONE

    return len(data)
