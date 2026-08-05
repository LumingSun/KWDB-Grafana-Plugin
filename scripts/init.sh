#!/bin/sh
set -eu

KWDB_BIN=${KWDB_BIN:-./kwbase}
KWDB_STORE=${KWDB_STORE:-/kaiwudb/deploy/kaiwudb-container}
KWDB_HOST=${KWDB_HOST:-127.0.0.1:26257}
INIT_DIR=${INIT_DIR:-/docker-entrypoint-initdb.d}
MARKER=${KWDB_STORE}/.demo-init-complete

echo "[kwdb-init] starting KWDB single node"
"${KWDB_BIN}" start-single-node \
  --insecure \
  --listen-addr=0.0.0.0:26257 \
  --http-addr=0.0.0.0:8080 \
  --store="${KWDB_STORE}" &
KWDB_PID=$!

trap 'kill "${KWDB_PID}" 2>/dev/null || true' EXIT INT TERM

i=0
while [ "${i}" -lt 60 ]; do
  if "${KWDB_BIN}" sql --insecure --host="${KWDB_HOST}" -e "SELECT 1" >/dev/null 2>&1; then
    break
  fi
  if ! kill -0 "${KWDB_PID}" 2>/dev/null; then
    echo "[kwdb-init] KWDB exited before becoming ready" >&2
    exit 1
  fi
  i=$((i + 1))
  sleep 1
done

if ! "${KWDB_BIN}" sql --insecure --host="${KWDB_HOST}" -e "SELECT 1" >/dev/null 2>&1; then
  echo "[kwdb-init] KWDB did not become ready within 60s" >&2
  exit 1
fi

if [ -f "${INIT_DIR}/init.sql" ] && [ ! -f "${MARKER}" ]; then
  echo "[kwdb-init] applying ${INIT_DIR}/init.sql"
  "${KWDB_BIN}" sql --insecure --host="${KWDB_HOST}" < "${INIT_DIR}/init.sql"
  touch "${MARKER}"
  echo "[kwdb-init] demo database and sample data initialized"
else
  echo "[kwdb-init] init.sql already applied or not found; skipping"
fi

echo "[kwdb-init] KWDB ready on ${KWDB_HOST}"
wait "${KWDB_PID}"
