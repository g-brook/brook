# Brook Docker Image

This image contains the Brook server and client applications, pre-configured for easy deployment in containerized environments.

## What is Brook?

Brook is a cross-platform, high-performance network tunneling and proxy toolkit implemented in Go. It supports TCP, UDP, HTTP(S), and WebSocket protocols, making it compatible with SSH, HTTP, Redis, MySQL, and other application protocols. The built-in web UI simplifies configuration and management.

Project GitHub Repository: [https://github.com/g-brook/brook](https://github.com/g-brook/brook)

## Tags

- `vX.X.X-arm64` - Specific version tags (e.g., `v1.0.0-arm64`)
- `vX.X.X-amd64` - Specific version tags (e.g., `v1.0.0-amd64`)
- `edge` - Development builds from main branch

## Supported Architectures

- `amd64`
- `arm64`

## How to Use This Image

### Start a Brook Server Instance

docker run -d \
  --name brook-server \
  -p 8000:8000 \
  -p 8909:8909 \
  -p 8919:8919 \
  -v /path/to/config:/config \
  gbrook/brook:latest \
  brook-sev -c /config/server.json

### Start a Brook Client Instance

docker run -d \
  --name brook-client \
  -v /path/to/config:/config \
  gbrook/brook:latest \
  brook-cli -c /config/client.json

### Using Docker Compose

Create a `docker-compose.yml` file:

version: '3.8'

services:
  brook-server:
    image: gbrook/brook:latest
    container_name: brook-server
    ports:
      - "8000:8000"   # Web UI
      - "8909:8909"   # Control port
      - "8919:8919"   # Tunnel port
    volumes:
      - ./config:/config
    command: brook-sev -c /config/server.json
    restart: unless-stopped

  brook-client:
    image: gbrook/brook:latest
    container_name: brook-client
    volumes:
      - ./config:/config
    command: brook-cli -c /config/client.json
    restart: unless-stopped
    depends_on:
      - brook-server

Then run:

docker-compose up -d

## Configuration

### Server Configuration (`server.json`)

Place your server configuration in a `server.json` file:

{
  "enableWeb": true,
  "webPort": 8000,
  "serverPort": 8909,
  "tunnelPort": 8919,
  "token": "",
  "logger": {
    "logLevel": "info",
    "logPath": "/config/logs",
    "outs": "file"
  }
}

### Client Configuration (`client.json`)

Place your client configuration in a `client.json` file:

{
  "serverPort": 8909,
  "serverHost": "brook-server",
  "token": "<Token generated in server UI>",
  "pingTime": 2000,
  "tunnels": [
    {
      "type": "udp",
      "destination": "127.0.0.1:9000",
      "proxyId": "333223"
    },
    {
      "type": "http",
      "destination": "127.0.0.1:8081",
      "proxyId": "HttpLocal-2",
      "httpId": "local"
    }
  ]
}

## Environment Variables

The Brook images do not currently support environment variable configuration. Please use JSON configuration files instead.

## Volumes

- `/config` - Configuration files and logs

## Ports

- `8000/tcp` - Web UI (when enabled)
- `8909/tcp` - Server control port
- `8919/tcp` - Data tunnel port (default, can be changed in config)

## Logging

Logs are written to the `/config/logs` directory by default. You can change this path in the `logger.logPath` setting in your configuration file.

## Docker Tips

### Running in Background

docker run -d --restart unless-stopped gbrook/brook:latest

### Viewing Logs

docker logs brook-server

### Executing Commands Inside Container

docker exec -it brook-server sh

## Building From Source

To build the Docker image from source:

git clone https://github.com/g-brook/brook.git
cd brook
docker build -t brook-local .

## License

View [license information](https://github.com/g-brook/brook/blob/main/LICENSE) for the software contained in this image.

As with all Docker images, these likely also contain other software which may be under other licenses (such as Bash, etc from the base distribution, along with any direct or indirect dependencies of the primary software being contained).

Some additional license information which was able to be auto-detected might be found in [the repo-info repository](https://github.com/docker-library/repo-info) for this image.

As for any pre-built image usage, it is the image user's responsibility to ensure that any use of this image complies with any relevant licenses for all software contained within.
