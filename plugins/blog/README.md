# Blog Plugin

`plugins/blog` is an independently deployed remote HTTP plugin service. It owns Blog data and exposes Article use cases through the main service plugin gateway.

## Runtime

Required environment:

```powershell
$env:BLOG_MYSQL_DSN="blog:blog@tcp(127.0.0.1:3306)/blog?charset=utf8mb4&parseTime=True&loc=UTC"
$env:KEIYAKU_HOST="http://127.0.0.1:8080"
$env:BLOG_REGISTRATION_SECRET="change-me-blog-registration-secret-32b"
$env:BLOG_GATEWAY_SECRET="change-me-blog-gateway-secret-32bytes"
go run ./cmd/blog
```

Optional environment:

- `BLOG_ADDR`, default `:9091`
- `BLOG_BASE_URL`, default `http://127.0.0.1${BLOG_ADDR}`
- `BLOG_INSTANCE_ID`, default `blog-local`
- `BLOG_HEARTBEAT_INTERVAL`, default `10s`
- `BLOG_REGISTRATION_SECRET`, required
- `BLOG_GATEWAY_SECRET`, required
- `BLOG_SNOWFLAKE_NODE`, default `2`

## Gateway Paths

- `POST /api/v1/extensions/blog/articles`
- `GET /api/v1/extensions/blog/articles`
- `GET /api/v1/extensions/blog/articles/{id}`

The plugin verifies the gateway HMAC signature before serving article routes. It reads authenticated user context from `X-Keiyaku-User-ID` and does not read the main service `users` or `articles` tables.
