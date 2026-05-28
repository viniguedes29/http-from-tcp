# Python Version — HTTP-from-TCP

Reescrita em Python do servidor HTTP from scratch. Veja o [README principal](../README.md) para a comparação completa Go vs Python.

## Rodar

```bash
pip install -r requirements.txt
python app/httpserver/main.py
```

## Testes

```bash
python -m pytest tests/ -v
```

## Estrutura

```
python-version/
├── internal/
│   ├── headers/headers.py      # Parsing e validação de headers
│   ├── request/request.py      # Parser incremental (state machine)
│   ├── response/response.py    # Writer HTTP (state machine)
│   └── server/server.py        # Servidor TCP multithreaded
├── app/
│   ├── httpserver/main.py      # Servidor com rotas e proxy httpbin
│   └── tcplistener/main.py     # Listener TCP básico (debug)
├── tests/
│   ├── test_headers.py
│   └── test_request.py
└── requirements.txt
```
