GOOS=linux go build .
docker build -t jetbench .
docker run --network=nats-network jetbench