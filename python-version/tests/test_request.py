"""Testes para o módulo de request HTTP."""

import io
import sys
import os

sys.path.insert(0, os.path.dirname(os.path.dirname(os.path.abspath(__file__))))

import pytest
from internal.request import request_from_reader, MalformedRequestError


class ChunkReader:
    """Simula leitura em chunks de uma conexão de rede."""

    def __init__(self, data: bytes, num_bytes_per_read: int):
        self._data = data
        self._chunk_size = num_bytes_per_read
        self._pos = 0

    def read(self, n: int = -1) -> bytes:
        if self._pos >= len(self._data):
            return b""
        end = min(self._pos + self._chunk_size, len(self._data))
        chunk = self._data[self._pos:end]
        self._pos = end
        return chunk


class TestRequestLineParse:
    def test_good_get_request(self):
        reader = ChunkReader(
            b"GET / HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n",
            num_bytes_per_read=3,
        )
        req = request_from_reader(reader)

        assert req.request_line.method == "GET"
        assert req.request_line.request_target == "/"
        assert req.request_line.http_version == "1.1"

    def test_good_get_with_path(self):
        reader = ChunkReader(
            b"GET /coffee HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n",
            num_bytes_per_read=1,
        )
        req = request_from_reader(reader)

        assert req.request_line.method == "GET"
        assert req.request_line.request_target == "/coffee"
        assert req.request_line.http_version == "1.1"

    def test_standard_headers(self):
        reader = ChunkReader(
            b"GET / HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n",
            num_bytes_per_read=3,
        )
        req = request_from_reader(reader)

        assert req.headers["host"] == "localhost:42069"
        assert req.headers["user-agent"] == "curl/7.81.0"
        assert req.headers["accept"] == "*/*"

    def test_malformed_header(self):
        reader = ChunkReader(
            b"GET / HTTP/1.1\r\nHost localhost:42069\r\n\r\n",
            num_bytes_per_read=3,
        )
        with pytest.raises(Exception):
            request_from_reader(reader)

    def test_standard_body(self):
        reader = ChunkReader(
            b"POST /submit HTTP/1.1\r\n"
            b"Host: localhost:42069\r\n"
            b"Content-Length: 13\r\n"
            b"\r\n"
            b"hello world!\n",
            num_bytes_per_read=3,
        )
        req = request_from_reader(reader)

        assert req.body == b"hello world!\n"

    def test_body_shorter_than_content_length(self):
        reader = ChunkReader(
            b"POST /submit HTTP/1.1\r\n"
            b"Host: localhost:42069\r\n"
            b"Content-Length: 20\r\n"
            b"\r\n"
            b"partial content",
            num_bytes_per_read=3,
        )
        with pytest.raises(MalformedRequestError):
            request_from_reader(reader)
