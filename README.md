# Auction Service

A Go auction API backed by MongoDB. Auctions are closed automatically by a background goroutine after the duration configured in `AUCTION_DURATION`.

This project is based on [devfullcycle/labs-auction-goexpert](https://github.com/devfullcycle/labs-auction-goexpert).

## Run with Docker Compose

Start the API and MongoDB:

```bash
docker compose up --build
```

The API will be available at http://localhost:8080.

The default auction duration is `20s` in Docker Compose. Override it when starting the environment:

```bash
AUCTION_DURATION=30s docker compose up --build
```

The value must use Go duration syntax, such as `500ms`, `20s`, `5m`, or `1h`. Invalid, missing, or non-positive values use the default duration of five minutes.

## Create an auction

```bash
curl --request POST http://localhost:8080/auction \
  --header "Content-Type: application/json" \
  --data '{
    "product_name": "Mechanical Keyboard",
    "category": "Electronics",
    "description": "A mechanical keyboard in excellent condition",
    "condition": 1
  }'
```

The endpoint returns `201 Created`. The auction starts with status `0` (`Active`) and changes automatically to status `1` (`Completed`) after `AUCTION_DURATION`.

List auctions:

```bash
curl "http://localhost:8080/auction?status=0"
```

## Run locally without Docker

Start MongoDB and create the environment file:

```bash
cp cmd/auction/.env.example cmd/auction/.env
go run ./cmd/auction
```

## Tests

Run the unit tests:

```bash
go test ./...
```

Run the automatic closing integration test against the MongoDB container:

```bash
docker compose up -d mongodb
MONGODB_TEST_URL="mongodb://admin:admin@localhost:27017/?authSource=admin" \
  go test ./internal/infra/database/auction \
  -run TestAuctionAutomaticallyCloses \
  -count=1
```

Stop the environment:

```bash
docker compose down
```
