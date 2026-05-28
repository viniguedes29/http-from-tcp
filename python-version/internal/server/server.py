"""
TCP Server — aceita conexões TCP e despacha pra um handler HTTP.

Equivalente ao internal/server/server.go do tutorial em Go.
Cada conexão é tratada numa thread separada (similar a goroutines).
"""

from __future__ import annotations

import logging
import socket
import threading
from typing import Callable

from internal.headers import Headers
from internal.request import Request, request_from_reader
from internal.response import Writer, StatusCode, get_default_headers

logger = logging.getLogger(__name__)

Handler = Callable[[Writer, Request], None]


class Server:
    def __init__(self, port: int, handler: Handler):
        self._port = port
        self._handler = handler
        self._sock: socket.socket | None = None
        self._closed = threading.Event()

    def serve(self) -> None:
        self._sock = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        self._sock.setsockopt(socket.SOL_SOCKET, socket.SO_REUSEADDR, 1)
        self._sock.bind(("", self._port))
        self._sock.listen(128)
        self._sock.settimeout(1.0)

        logger.info("Server listening on port %d", self._port)

        while not self._closed.is_set():
            try:
                conn, addr = self._sock.accept()
            except socket.timeout:
                continue
            except OSError:
                break

            t = threading.Thread(target=self._handle, args=(conn,), daemon=True)
            t.start()

    def serve_background(self) -> None:
        t = threading.Thread(target=self.serve, daemon=True)
        t.start()

    def close(self) -> None:
        self._closed.set()
        if self._sock:
            self._sock.close()

    def _handle(self, conn: socket.socket) -> None:
        try:
            reader = conn.makefile("rb")
            try:
                req = request_from_reader(reader)
            except Exception as e:
                logger.debug("Bad request: %s", e)
                writer = Writer(conn.makefile("wb"))
                writer.write_status_line(StatusCode.BAD_REQUEST)

                error_body = b"""<html>
  <head>
    <title>400 Bad Request</title>
  </head>
  <body>
    <h1>Bad Request</h1>
    <p>Your request honestly kinda sucked.</p>
  </body>
</html>"""
                headers = get_default_headers(0)
                headers.set_override("Content-Type", "text/html")
                headers.set_override("Content-Length", str(len(error_body)))
                writer.write_headers(headers)
                writer.write_body(error_body)
                return

            if self._handler:
                writer = Writer(conn.makefile("wb"))
                self._handler(writer, req)
        finally:
            conn.close()
