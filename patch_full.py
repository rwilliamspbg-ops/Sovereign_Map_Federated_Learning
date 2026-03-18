with open("docker-compose.full.yml", "r") as f:
    config = f.read()

# find "services:" and put frontend after it
frontend_service = """
  frontend:
    build:
      context: .
      dockerfile: Dockerfile.frontend
    image: frontend:latest
    container_name: sovereign-frontend
    ports:
      - "${FRONTEND_HOST_PORT:-3000}:80"
    environment:
      - NODE_ENV=production
    depends_on:
      - backend
    networks:
      - sovereign_network
    restart: always
"""
config = config.replace("services:\n", "services:\n" + frontend_service)

with open("docker-compose.full.yml", "w") as f:
    f.write(config)
