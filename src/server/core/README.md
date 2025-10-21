

## Local Development

- Start infras

```bash
# current path: ./src/server
sh scripts/local_dev_deps.sh
```

- Setup python deps

```bash
# current path: ./src/server/core
uv sync
```

- Launch Core in dev mode (with hot reload)

```bash
# current path: ./src/server/core
uv run -m fastapi dev
```

- Launch Core in prod mode

```bash
# current path: ./src/server/core
uv run -m gunicorn api:app \
    --workers 4 \
    --worker-class uvicorn.workers.UvicornWorker \
    --bind 0.0.0.0:8000 \
    --timeout 120
```

- Test the core
```bash
# current path: ./src/server/core
cp config.yaml.example config.yaml
uv run -m pytest
```