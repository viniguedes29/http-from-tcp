"""
HTTP Server — servidor HTTP completo com rotas, proxy e streaming.

Equivalente ao cmd/httpserver/main.go do tutorial em Go.

Endpoints:
  GET /            → 200 OK com HTML
  GET /yourproblem → 400 Bad Request
  GET /myproblem   → 500 Internal Server Error
  GET /video       → Serve arquivo de vídeo (binary)
  GET /httpbin/*   → Proxy reverso pro httpbin.org com chunked encoding
"""

from __future__ import annotations

import hashlib
import logging
import os
import signal
import sys

import httpx

sys.path.insert(0, os.path.dirname(os.path.dirname(os.path.dirname(os.path.abspath(__file__)))))

from internal.headers import Headers
from internal.request import Request
from internal.response import Writer, StatusCode
from internal.server import Server

PORT = 42069

logging.basicConfig(
    level=logging.INFO,
    format="%(asctime)s [%(levelname)s] %(message)s",
)
logger = logging.getLogger(__name__)


def handler(w: Writer, req: Request) -> None:
    target = req.request_line.request_target

    if target.startswith("/httpbin"):
        handle_httpbin_proxy(w, req)
        return

    if target == "/yourproblem":
        w.write_status_line(StatusCode.BAD_REQUEST)
        body = b"""<html>
  <head>
    <title>400 Bad Request</title>
  </head>
  <body>
    <h1>Bad Request</h1>
    <p>Your request honestly kinda sucked.</p>
  </body>
</html>"""
        h = Headers()
        h.set_override("Content-Type", "text/html")
        h.set_override("Content-Length", str(len(body)))
        h.set_override("Connection", "close")
        w.write_headers(h)
        w.write_body(body)

    elif target == "/myproblem":
        w.write_status_line(StatusCode.INTERNAL_SERVER_ERROR)
        body = b"""<html>
  <head>
    <title>500 Internal Server Error</title>
  </head>
  <body>
    <h1>Internal Server Error</h1>
    <p>Okay, you know what? This one is on me.</p>
  </body>
</html>"""
        h = Headers()
        h.set_override("Content-Type", "text/html")
        h.set_override("Content-Length", str(len(body)))
        h.set_override("Connection", "close")
        w.write_headers(h)
        w.write_body(body)

    elif target == "/video":
        serve_video(w)

    else:
        w.write_status_line(StatusCode.OK)
        body = b"""<html>
  <head>
    <title>200 OK</title>
  </head>
  <body>
    <h1>Success!</h1>
    <p>Your request was an absolute banger.</p>
  </body>
</html>"""
        h = Headers()
        h.set_override("Content-Type", "text/html")
        h.set_override("Content-Length", str(len(body)))
        h.set_override("Connection", "close")
        w.write_headers(h)
        w.write_body(body)


def serve_video(w: Writer) -> None:
    project_root = os.path.dirname(os.path.dirname(os.path.dirname(os.path.abspath(__file__))))
    video_path = os.path.join(project_root, "..", "assets", "vim.mp4")
    try:
        with open(video_path, "rb") as f:
            video = f.read()
    except OSError:
        w.write_status_line(StatusCode.INTERNAL_SERVER_ERROR)
        body = b"""<html>
  <head>
    <title>500 Internal Server Error</title>
  </head>
  <body>
    <h1>Internal Server Error</h1>
    <p>Failed to read video file</p>
  </body>
</html>"""
        h = Headers()
        h.set_override("Content-Type", "text/html")
        h.set_override("Content-Length", str(len(body)))
        h.set_override("Connection", "close")
        w.write_headers(h)
        w.write_body(body)
        return

    h = Headers()
    h.set_override("Content-Type", "video/mp4")
    h.set_override("Content-Length", str(len(video)))
    h.set_override("Connection", "close")

    w.write_status_line(StatusCode.OK)
    w.write_headers(h)
    w.write_body(video)


def handle_httpbin_proxy(w: Writer, req: Request) -> None:
    path = req.request_line.request_target.removeprefix("/httpbin")
    if not path:
        path = "/"

    httpbin_url = f"https://httpbin.org{path}"

    try:
        with httpx.stream("GET", httpbin_url) as resp:
            w.write_status_line(StatusCode.OK)

            h = Headers()
            content_type = resp.headers.get("content-type", "application/octet-stream")
            h.set_override("Content-Type", content_type)
            h.set_override("Transfer-Encoding", "chunked")
            h.set_override("Connection", "close")
            w.write_headers(h)

            full_body = bytearray()

            for chunk in resp.iter_bytes(chunk_size=32):
                full_body.extend(chunk)
                w.write_chunked_body(chunk)

            # trailers com SHA256
            hash_hex = hashlib.sha256(bytes(full_body)).hexdigest()
            trailers = Headers()
            trailers.set_override("X-Content-SHA256", hash_hex)
            trailers.set_override("X-Content-Length", str(len(full_body)))
            w.write_trailers(trailers)

    except Exception as e:
        logger.error("Proxy error: %s", e)
        w.write_status_line(StatusCode.INTERNAL_SERVER_ERROR)
        body = b"""<html>
  <head>
    <title>500 Internal Server Error</title>
  </head>
  <body>
    <h1>Internal Server Error</h1>
    <p>Failed to proxy request to httpbin.org</p>
  </body>
</html>"""
        h = Headers()
        h.set_override("Content-Type", "text/html")
        h.set_override("Content-Length", str(len(body)))
        h.set_override("Connection", "close")
        w.write_headers(h)
        w.write_body(body)


def main() -> None:
    server = Server(PORT, handler)

    def shutdown(signum, frame):
        logger.info("Shutting down server...")
        server.close()
        sys.exit(0)

    signal.signal(signal.SIGINT, shutdown)
    signal.signal(signal.SIGTERM, shutdown)

    server.serve()


if __name__ == "__main__":
    main()
