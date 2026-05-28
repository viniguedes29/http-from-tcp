"""Testes para o módulo de headers HTTP."""

import sys
import os

sys.path.insert(0, os.path.dirname(os.path.dirname(os.path.abspath(__file__))))

import pytest
from internal.headers import Headers, InvalidHeaderError


class TestHeadersParse:
    def test_valid_single_header(self):
        headers = Headers()
        data = b"Host: localhost:42069\r\n\r\n"
        n, done = headers.parse(data)

        assert n == 23
        assert done is False
        assert headers["host"] == "localhost:42069"

    def test_valid_single_header_with_extra_whitespace(self):
        headers = Headers()
        data = b"       Host: localhost:42069                           \r\n\r\n"
        n, done = headers.parse(data)

        assert done is False
        assert n == 57
        assert headers["host"] == "localhost:42069"

    def test_valid_two_headers_with_existing(self):
        headers = Headers({"host": "localhost:42069"})
        data = b"User-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n"
        n, done = headers.parse(data)

        assert n == 25
        assert done is False
        assert headers["host"] == "localhost:42069"
        assert headers["user-agent"] == "curl/7.81.0"

    def test_done_on_empty_line(self):
        headers = Headers()
        data = b"\r\n a bunch of other stuff"
        n, done = headers.parse(data)

        assert n == 2
        assert done is True
        assert len(headers) == 0

    def test_invalid_spacing_in_key(self):
        headers = Headers()
        data = b"       Host : localhost:42069       \r\n\r\n"

        with pytest.raises(InvalidHeaderError):
            headers.parse(data)

    def test_invalid_character_in_key(self):
        headers = Headers()
        data = "H©st: localhost:42069\r\n\r\n".encode("utf-8")

        with pytest.raises((InvalidHeaderError, UnicodeDecodeError)):
            headers.parse(data)

    def test_duplicate_header_combines_values(self):
        headers = Headers({"host": "localhost:8000"})
        data = b"Host: localhost:42069\r\n\r\n"
        n, done = headers.parse(data)

        assert n == 23
        assert done is False
        assert headers["host"] == "localhost:8000, localhost:42069"
