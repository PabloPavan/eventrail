# Basic Example

This example uses an in-memory broker, so no Redis is required.

## Run

```bash
go run ./examples/basic
```

Server starts at `http://localhost:3000`.

## Watch SSE

```bash
curl -N "http://localhost:3000/events?scope=1"
```

## Publish an Event

```bash
curl "http://localhost:3000/publish?scope=1&type=demo.ping&data={\"id\":123}"
```

You should see the event in the SSE client.

## Build

```bash
go build ./examples/basic
```
