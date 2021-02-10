# Binance Ingest

Ingestion pipeline for Binance market data via this websocket stream

## How to run
1. Update `symbols` in `Dockerfile` and add maximum of **1024** comma separated symbols.
2. Build docker image:
```shell
docker build -t binance-ingest .
```
3. Run the container:
```shell
docker run binance-ingest
```
