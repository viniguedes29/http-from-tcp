"""
TCP Listener — escuta conexões TCP e printa o request HTTP parseado.

Equivalente ao cmd/tcplistener/main.go do tutorial em Go.
Útil pra ver o protocolo HTTP em ação sem nenhuma camada de abstração.
"""

from __future__ import annotations

import os
import socket
import sys

sys.path.insert(0, os.path.dirname(os.path.dirname(os.path.dirname(os.path.abspath(__file__)))))

from internal.request import request_from_reader

PORT = 42069


def main() -> None:
    with socket.socket(socket.AF_INET, socket.SOCK_STREAM) as sock:
        sock.setsockopt(socket.SOL_SOCKET, socket.SO_REUSEADDR, 1)
        sock.bind(("", PORT))
        sock.listen(128)

        print(f"Listening on port {PORT}...")

        while True:
            conn, addr = sock.accept()
            try:
                reader = conn.makefile("rb")
                req = request_from_reader(reader)

                print("Request line:")
                print(f"- Method: {req.request_line.method}")
                print(f"- Target: {req.request_line.request_target}")
                print(f"- Version: {req.request_line.http_version}")
                print("Headers:")
                for k, v in req.headers.items():
                    print(f"- {k}: {v}")
                print("Body:")
                print(req.body.decode("utf-8", errors="replace"))
            except Exception as e:
                print(f"Error reading request: {e}")
            finally:
                conn.close()


if __name__ == "__main__":
    main()
