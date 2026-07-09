with open("../docker-compose.full.yml", "r") as f:
    config = f.read()

config = config.replace("- sovereign_network", "- sovereign-network")

with open("../docker-compose.full.yml", "w") as f:
    f.write(config)
