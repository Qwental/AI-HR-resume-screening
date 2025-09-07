# Из корня проекта
docker compose \
  -f database/docker-compose.db.yml \
  -f auth/docker-compose.yml \
  down -v

docker compose \
  -f database/docker-compose.db.yml \
  -f auth/docker-compose.yml \
  up -d --build
