#!/usr/bin/env sh
set -eu

APP_DIR=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
cd "$APP_DIR"

if docker compose version >/dev/null 2>&1; then
    COMPOSE="docker compose"
elif docker-compose version >/dev/null 2>&1; then
    COMPOSE="docker-compose"
else
    echo "Docker Compose is required." >&2
    exit 1
fi

if [ -n "${IMAGE_REPOSITORY:-}" ] && [ -n "${IMAGE_TAG:-}" ]; then
    $COMPOSE pull app
    $COMPOSE up -d --no-build
else
    $COMPOSE up -d --build
fi

APP_CONTAINER=$($COMPOSE ps -q app)
if [ -z "$APP_CONTAINER" ]; then
    echo "API container was not created." >&2
    exit 1
fi

ATTEMPT=0
while [ "$ATTEMPT" -lt 30 ]; do
    STATUS=$(docker inspect --format '{{if .State.Health}}{{.State.Health.Status}}{{else}}{{.State.Status}}{{end}}' "$APP_CONTAINER")
    if [ "$STATUS" = "healthy" ]; then
        $COMPOSE ps
        exit 0
    fi
    if [ "$STATUS" = "exited" ] || [ "$STATUS" = "dead" ]; then
        $COMPOSE logs --tail=100 app
        exit 1
    fi

    ATTEMPT=$((ATTEMPT + 1))
    sleep 2
done

echo "API did not become healthy in time." >&2
$COMPOSE logs --tail=100 app
exit 1
