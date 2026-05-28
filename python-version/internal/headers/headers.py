"""
HTTP Headers — parsing e gerenciamento de headers HTTP.

Headers são case-insensitive (armazenados em lowercase).
Validação do nome segue a spec RFC 9110 (tchar).
"""

from __future__ import annotations

CRLF = b"\r\n"

_TCHAR_EXTRAS = frozenset(b"!#$%&'*+-.^_`|~")


class InvalidHeaderError(Exception):
    pass


def _is_valid_header_name(name: str) -> bool:
    if not name:
        return False
    for ch in name.encode("ascii"):
        if ch in _TCHAR_EXTRAS:
            continue
        if ord("A") <= ch <= ord("Z"):
            continue
        if ord("a") <= ch <= ord("z"):
            continue
        if ord("0") <= ch <= ord("9"):
            continue
        return False
    return True


class Headers(dict[str, str]):
    """Dict-like para headers HTTP, com chaves normalizadas em lowercase."""

    def get_header(self, key: str) -> tuple[str, bool]:
        k = key.lower()
        if k in self:
            return self[k], True
        return "", False

    def set(self, key: str, value: str) -> None:
        if not _is_valid_header_name(key):
            raise InvalidHeaderError(f"invalid header name: {key!r}")
        k = key.lower()
        if k in self:
            self[k] = f"{self[k]}, {value}"
        else:
            self[k] = value

    def set_override(self, key: str, value: str) -> None:
        if not _is_valid_header_name(key):
            raise InvalidHeaderError(f"invalid header name: {key!r}")
        self[key.lower()] = value

    def delete(self, key: str) -> None:
        k = key.lower()
        self.pop(k, None)

    def parse(self, data: bytes) -> tuple[int, bool]:
        """
        Parseia UM header a partir de `data`.

        Retorna (bytes_consumidos, done).
        - done=True quando encontra o CRLF final (fim dos headers).
        - bytes_consumidos=0 se não tem dados suficientes ainda.
        - Levanta InvalidHeaderError se o formato for inválido.
        """
        rn_idx = data.find(CRLF)

        if rn_idx == -1:
            return 0, False

        # CRLF no início → fim dos headers
        if rn_idx == 0:
            return 2, True

        header_line = data[:rn_idx]

        # permite leading spaces na linha
        header_line = header_line.lstrip(b" ")

        colon_idx = header_line.find(b":")
        if colon_idx == -1:
            raise InvalidHeaderError("missing colon separator")

        key_bytes = header_line[:colon_idx]
        value_bytes = header_line[colon_idx + 1:]

        # chave NÃO pode ter espaços antes/depois
        if key_bytes.startswith(b" ") or key_bytes.endswith(b" "):
            raise InvalidHeaderError("header name has leading/trailing spaces")

        key = key_bytes.decode("ascii")
        value = value_bytes.decode("ascii").strip()

        self.set(key, value)

        return rn_idx + 2, False
