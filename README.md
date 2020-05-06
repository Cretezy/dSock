<div align="center">
  <img src="https://i.imgur.com/2G9887I.png" width="150px"/>
  <br>
  <h1>dSock</h1>
  <p>dSock is a distributed WebSocket broker (in Go, using Redis).</p>
  <p>Clients can authenticate & connect, and you can send text/binary message as an API.</p>
</div>

## Features

**Multiple clients per user & authentication**

dSock can broadcast a message to all clients for a certain user (identified by user ID and optionally session ID) or a certain connection (by ID). Users can be authenticated using claims or JWTs (see below).

**Distributed**

dSock can be scaled up easily as it uses Redis as a central database & pub/sub, with clients connecting to worker. It's designed to run on the cloud using scalable platforms such as Kubernetes or Cloud Run.

**Text & binary messaging**

dSock is designed for text and binary messaging, enabling JSON (UTF-8), Protocol Buffers, or any custom protocol.

**Lightweight & fast**

dSock utilized Go's concurrency for great performance at scale, with easy distribution and safety. It is available as Docker images for convenience.

**Disconnects**

Disconnect clients from an external event (logout) from a session ID or for all user connections.

## Uses

The main use case for dSock is having stateful WebSocket connections act as a stateless API.

This enables you to not worry about your connection handling and simply send messages to all (or some) of a user's clients as any other HTTP API.

**Chat service**

Clients connect to dSock, and your back-end can broadcast messages to a specific user's clients

**More!**

## Architecture

![](https://i.imgur.com/pFo2zDU.png)

dSock is separated into 2 main services:

- **dSock Worker**
  This is the main server clients connect to. The worker distributed the messages to the clients ("last mile")

- **dSock API**
  The API receives messages and distributes it to the workers for target clients

This allows the worker (connections) and API (gateway) to scale independently and horizontally.

dSock uses Redis as a backend data store, to store connection locations and claims.

### Terminology

| Word                                             |                                                                     |
| ------------------------------------------------ | ------------------------------------------------------------------- |
| [WebSocket](https://tools.ietf.org/html/rfc6455) | Sockets "over" HTTP(S)                                              |
| [JWT](https://tools.ietf.org/html/rfc7519)       | JSON Web Token                                                      |
| Claim                                            | dSock authentication mention using a pre-registered claim ("token") |
| [Redis](https://redis.io/)                       | Open-source in-memory key-value database                            |

### Flow

- Authentication:
  - Client does request to your API, you either:
    - Hit the dSock API to create a claim for the user
    - Generate & sign a JWT for the user
  - You return the claim or JWT to client
- Connection:
  - User connects to a worker with claim or JWT
- Sending:
  - You hit the dSock API (`POST /send`) with the target (`user`, `session` optionally) and the message as body
  - Message sent to target(s)

## Setup

### Installation

dSock is published as binaries and as Docker images.

#### Binaries

Binaries are available on the [releases pages](https://github.com/Cretezy/dSock/releases).

You can simply run the binary for your architecture/OS.

You can configure dSock using environment variables or a config (see below).

#### Docker images

Docker images are published on Docker Hub:

- [`dsock/api`](https://hub.docker.com/r/dsock/api)
- [`dsock/worker`](https://hub.docker.com/r/dsock/worker)

The images are small (~15MB) and expose on port `80` by default (controllable by setting the `PORT` environment variable).

It is recommended to use the environment variables to configure dSock instead of a config when using the images.
Configs are still supported (can be mounted to `/config.toml` or `/config.$EXT`, see below).

### Options

dSock can be configured using a config file or using environment variables.

- `DSOCK_ADDRESS` (`address`, string): Address to listen to. Defaults to `:6241`. Uses `PORT` if empty.
- Redis:
  - `DSOCK_REDIS_HOST` (`redis_host`, string): Redis host. Defaults to `localhost:6379`
  - `DSOCK_REDIS_PASSWORD` (`redis_password`, string): Redis password. Defaults to no password
  - `DSOCK_REDIS_DB` (`redis_db`, integer): Redis database. Defaults to `0`
- `DSOCK_JWT_SECRET` (`jwt_secret`, string, optional): When set, enables JWT authentication
- `DSOCK_DEBUG` (`debug`, boolean): Enables debugging, useful for development. Defaults to `false`

You can write your config file in TOML (recommended), JSON, YAML, or any format supported by [viper](https://github.com/spf13/viper)

Configs are loaded from (in order):

- `$PWD/config.$EXT`
- `$HOME/.config/dsock/config.$EXT`
- `/etc/dsock/config.$EXT`

A default config will be created at `$PWD/config.toml` if no config is found.

## Usage

All API calls will return a `success` boolean.
If it is `false`, it will also add `error` (message) and `errorCode` (constant from `common/errors.go`).

All API calls (excluding `/connect` endpoint) requires authentication with a `token` query parameter, or set as a `Authorization` header in the format of: `Bearer $TOKEN`.

Having an invalid or missing token will result in the `INVALID_AUTHORIZATION` error code.

Most errors starting with `ERROR_` are downstream errors, usually from Redis. Check if your Redis connection is valid!

### Client authentication

#### Claims

Claims are the recommended way to authenticate with dSock. Before a client connects, they should hit your API (which you can use your usual authentication), and your API requests the dSock API to create a "claim", which you then return to the client.

Once a client has a claim, it can then connect to the worker using the `claim` query parameter.

You can create them by accessing the API as `POST /claim` with the following query options:

- `user` (required, string): The user ID
  - `session` (optional, string): The session ID (scoped per user)
- Time-related (not required, default expiration is 1 minute after the claim is created, only one used):
  - `expiration` (integer, seconds from epoch): Time the claim expires (takes precedence over `duration`)
  - `duration` (integer, seconds): Duration of the claim
- `token` (required, string): Authorization token for API set in config. Can also be a `Authorization` Bearer token
- `id` (optional, string): The claim ID to use. This should not be guessed, so long random string or UUIDv4 is recommended. If not set, it will generate a random string (recommended to let dSock generate the ID)

The returned body will contain the following keys:

- `claim`: The claim data
    - `id`: The claim ID
    - `expiration`: The expiration in seconds from epoch
    - `user`: The user for the claim
    - `session` (if session is provided): The user session for the claim

A claim is single-use, so once a client connects, it will instantly expire.

##### Examples

Create a claim for a user (`1`) expiring in 10 seconds:

```text
POST /claim?token=abcxyz&user=1&duration=10
```

Create a claim for a user (`1`) with a session (`a`) with a claim ID (`a1b2c3`) expiring at some time:

```text
POST /claim?user=1&session=a&expiration=1588473164&id=a1b2c3
Authorization: Bearer abcxyz
```

##### Errors

Creating a claim has the follow possible errors:

- `USER_ID_REQUIRED`: If the `user` parameter is not set
- `INVALID_EXPIRATION`: If the expiration is invalid (not parsable as integer)
- `NEGATIVE_EXPIRATION`: If the expiration is negative
- `INVALID_DURATION`: If the duration is invalid (not parsable as integer)
- `NEGATIVE_DURATION`: If the duration is negative
- `ERROR_CHECKING_CLAIM`: If an error occurred during checking if a claim exist (Redis error)
- `CLAIM_ID_ALREADY_USED`: If the claim ID is set and is already used

#### JWT

To authenticate a client, you can also create a JWT token and deliver it to the client before connecting. To enable this, set the `jwt_secret` to with your JWT secret (HMAC signature secret)

Payload options:

- `sub` (required, string): The user ID
- `sid` (optional, string): The session ID (scoped per user)
- Time-related (one is required):
  - [`iat`](https://tools.ietf.org/html/rfc7519#section-4.1.6) (integer, in seconds from epoch): Time the JWT is issued (expires 1 minute after this time)
  - [`exp`](https://tools.ietf.org/html/rfc7519#section-4.1.4) (integer, in seconds from epoch): Expiration time for the JWT, takes precedence over `iat`

### Client connections

Connect using a WebSocket to `ws://worker/connect` with the one of the following query parameter options:

- `claim`: The authentication claim created previously (takes precedence over `jwt`)
- `jwt`: JWT created previously

You can load-balance a cluster of workers, as long as the load-balancer supports WebSockets.

#### Errors

The following errors can happen during connection:

- `ERROR_GETTING_CLAIM`: If an error occurred during fetching the claim (Redis error)
- `MISSING_CLAIM`: If the claim ID doesn't exists. This can also happen if the claim has expired
- `INVALID_EXPIRATION`: If the claim has an invalid expiration (shouldn't happen unless Redis error)
- `EXPIRED_CLAIM`: If the claim has expired, but Redis hasn't expired the claim on it's own
- `INVALID_JWT`: If the JWT is malformed (bad JSON/JWT format) or is not signed with proper key
- `MISSING_AUTHENTICATION`: If no authentication is provided (no claim/JWT)

### Sending message

Sending a message is done through the `POST /send` API endpoint.

Query param options:

- Targeting (one is required):
  - `user` (string): The user ID to target
    - `session` (optional, string, when `user` is set): The specific session(s) to target from the user
  - `id` (string UUID): The specific internal connection ID
- `type` (required, string): Message (body) type. Can be `text` (UTF-8 text) or `binary`. This becomes the WebSocket message type.
- `token` (required, string): Authorization token for API set in config. Can also be a `Authorization` Bearer token

The body of the request is used as the message. This can be text/binary, and the `Content-Type` header is not used internally (only `type` is used).

#### Examples

Send a JSON message to a user (`1`)

```text
POST /send?token=abcxyz&user=1&type=text

{"message":"Hello world!","from":"Charles"}
```

Send a text value to a user (`1`) with a session (`a`)

```text
POST /send?user=1&session=a&type=text
Authorization: Bearer abcxyz

<Cretezy> Hey!
```

#### Errors

The following errors can happen during sending a message:

- `INVALID_AUTHORIZATION`: Invalid authentication (token). See errors section under usage
- `ERROR_GETTING_CONNECTION`: If could not fetch connection(s) (Redis error)
- `ERROR_GETTING_USER`: If `user` is set and could not fetch user (Redis error)
- `MISSING_CONNECTION_OR_USER`: If `id` or `user` is not provided
- `INVALID_MESSAGE_TYPE`: If the `type` is invalid
- `ERROR_READING_MESSAGE`: If an error occurred during reading the request body
- `ERROR_MARSHALLING_MESSAGE`: If an error occurred during preparing to send the message to the workers (shouldn't happen)

### Disconnecting

You can disconnect a client by user (and optionally session) ID.

This is useful when logging out a user, to make sure it also disconnects any connections.
Make sure to include a session in your claim/JWT to be able to disconnect only some of a user's connections.

The API endpoint is `POST /disconnect`, with the following query params:

- Targeting (one is required):
  - `user` (string): The user ID to target
    - `session` (optional, string, when `user` is set): The specific session(s) to target from the user
  - `id` (string UUID): The specific internal connection ID
- `token` (required, string): Authorization token for API set in config. Can also be a `Authorization` Bearer token
- `keepClaims` (optional, boolean): If to keep active claims for the target. By default, dSock will remove claims for the target to prevent race conditions

#### Examples

Disconnect a user (`1`) with a session (`a`):

```text
POST /send?token=abcxyz&user=1&session=a
```

#### Errors

The following errors can happen during disconnection:

- `INVALID_AUTHORIZATION`: Invalid authentication (token). See errors section under usage
- `ERROR_GETTING_CONNECTION`: If could not fetch connection(s) (Redis error)
- `ERROR_GETTING_USER`: If `user` is set and could not fetch user (Redis error)
- `MISSING_CONNECTION_OR_USER`: If `id` or `user` is not provided
- `ERROR_GETTING_CLAIM`: If an error occurred during fetching the claim(s) (Redis error)
- `ERROR_MARSHALLING_MESSAGE`: If an error occurred during preparing to send the message to the workers (shouldn't happen)

### Info

You can access info about connections and claims using the `GET /info` API endpoint. The following query params are supported:

- Targeting (one is required):
  - `user` (string): The user ID to target
    - `session` (optional, string, when `user` is set): The specific session(s) to target from the user
  - `id` (string UUID): The specific internal connection ID
- `token` (required, string): Authorization token for API set in config. Can also be a `Authorization` Bearer token

The API will return all opened connections and non-expired claims for the target.

The returned object contains:

- `connections` (array of objects): List of open connections for the target
  - `id`: Internal connection ID
  - `worker`: Internal worker holding the connection
  - `lastPing`: Last ping from client in seconds from epoch
  - `user`: The connection's user
  - `session` (optional): The connection's session
- `claims` (array of objects): List of non-expired claims for the target:
  - `id`: Claim ID (what a client would connect with)
  - `expiration`: Claim expiration in seconds from epoch
  - `user`: The claim's user
  - `session` (optional): The claim's session

#### Examples

Get info for a user (`1`) with a session (`a`):

```text
GET /info?token=abcxyz&user=1&session=a
```

#### Errors

The following errors can happen during getting info:

- `INVALID_AUTHORIZATION`: Invalid authentication (token). See errors section under usage
- `ERROR_GETTING_CLAIM`: If an error occurred during fetching the claim(s) (Redis error)
- `ERROR_GETTING_CONNECTION`: If could not fetch connection(s) (Redis error)
- `ERROR_GETTING_USER`: If `user` is set and could not fetch user (Redis error)
- `MISSING_CONNECTION_OR_USER`: If `id` or `user` is not provided

## Internals

dSock uses Redis as it's database (for claims and connection information) and for it's publish/subscribe capabilities.
Redis was chosen because it is widely used, is performant, and supports all requried features.

### Claims

When creating a claim, dSock does the following operations:

- Set `claim:$id` to the claim information (user, session, expiration)
- Add the claim ID to `claim-user:$user` (to be able to lookup all of a user's claims)
- Add the claim ID to `claim-user-session:$user-$session` if session is passed (to be able to lookup all of a user session's claims)

When a user connects, dSock retrieves the claim by ID and validates it's expiration. It then removes the claim from the user and user session storages.

When getting information or disconnecting, it retrieves or deletes the claim(s).

### Connections

When a user connects and authenticates, dSock does the following operations:

- Set `conn:$id` to the connection's information (using a random UUID, with user, session, worker ID, and last ping)
- Add the connection ID to `user:$user` (to be able to lookup all of a user's connections)
- Add the connection ID to `user-sesion:$user-$session` (to be able to lookup all of a user session's connections)

When receiving a ping or pong from the client, it updates the last ping time. A ping is sent from the server every minute.

Connections are kept alive until a client disconnects, or is forcibly disconnected using `POST /disconnect`

### Sending

When sending a message, the API resolves of all of the workers that hold connections for the target user/session/connection, and sends the message through Redis to that worker's channel (`worker:$id`).

API to worker messages are encoded using [Protocol Buffer](https://developers.google.com/protocol-buffers) for efficiency;
they are fast to encode/decode, and binary messages to not need to be encoded as strings during communication.

## FAQ

**Why is built-in HTTPS not supported?**

To remove complexity inside dSock, TLS is not implemented.
It is expected that the API and worker nodes are behind load-balancers, which would be able to do TLS termination.

If you need TLS, you can either add a TLS-terminating load-balancer, or a reverse proxy (such as nginx or Caddy).

## Development

### Setup

- Install [Go](https://golang.org)
- Install [Docker](https://www.docker.com) and [Docker Compose](https://docs.docker.com/compose/)
- Pull the [dSock repository](https://github.com/Cretzy/dSock)
- Run `docker-compose up`
- Develop! API is available at `:3000`, and worker at `:3001`. Configs are in their respective folders

### Protocol Buffers

If making changes to the Protocol Buffer definitions (under `protos`), make sure you have the [`protoc`](https://github.com/protocolbuffers/protobuf) compiler and [`protoc-gen-go`](https://developers.google.com/protocol-buffers/docs/reference/go-generated).

Once changes are done to the definitions, run the `scripts/build-protos` file to generate the associated Go code.

### Docker

You can build the Docker images using `scripts/build-docker`. This will create the `dsock-worker` and `dsock-api` images.

### Tests

dSock has multiple types of tests to ensure stability and maximum coverage.

You can run all tests by running `scripts/run-tests`. You can also run individual test suites (see below)

#### End-to-end (E2E)

You can run the E2E tests by running `scripts/run-e2e`. The E2E tests are located inside the `e2e` directory. 

#### Unit

You can run the unit tests by running `scripts/run-unit`. The units tests are located inside the `common`/`api`/`worker` directories. 

### Contributing

[Pull requests](https://github.com/Cretezy/dSock/pulls) are encouraged!

## License & support

dSock is MIT licensed ([see license](./LICENSE)).

Community support is available through [GitHub issues](https://github.com/Cretezy/dSock/issues).
For professional support, please contact [charles@cretezy.com](mailto:charles@cretezy.com?subject=dSock%20Support).

## Credit

Icon made by Freepik from flaticon.com.

Project was created & currently maintained by [Charles Crete](https://github.com/Cretezy).
