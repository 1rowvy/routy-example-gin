# Routy example: a Gin shop API

A small [Gin](https://github.com/gin-gonic/gin) service and its API described in
[Routy](https://github.com/1rowvy/routy): requests, checks and scenarios live in `api/` next to the
Go code, run from the terminal, the desktop app, VS Code or CI, and fail the build when the code and
the requests drift apart.

## Try it

You need Go and the `routy` CLI:

```sh
curl -fsSL https://raw.githubusercontent.com/1rowvy/routy/master/install.sh | sh   # Linux
```

On macOS and Windows take the archive from the
[latest release](https://github.com/1rowvy/routy/releases/latest).

```sh
git clone https://github.com/1rowvy/routy-example-gin && cd routy-example-gin
go run . &                         # the shop on :8080, admin password "admin"
export ROUTY_ADMIN_PASSWORD=admin  # or keep it in the keychain: routy secret set admin_password
routy run api                      # every request and flow
```

```console
✓ api/auth/login/post.routy  Login  POST http://localhost:8080/auth/login  200 OK  1ms 62B
    ✓ status == 200
    ✓ body matches Token
    ✓ body.expires_in == 3600
✓ api/flows/checkout.routy  Checkout  flow
    ↳ CreateUser  201  0ms
      ↳ Login  cached
    ↳ CreateOrder  201  0ms
    ↳ PayOrder  202  0ms
    ↳ WaitUntilPaid  200  1ms
    ✓ paid.body.total == 45.5
    ✓ paid.body.user_id == user.body.id
    → saved last_order
…
22 passed, 0 failed
```

`DeleteUser` has `confirm: true`, so `routy run` asks before sending it; `--yes` skips the question.

## Change the code, watch the requests notice

The requests know which Go handler they belong to (`handler: s.CreateUser`), and Routy reads the
handler: the struct the body is bound into, query parameters, headers, middleware and the type it
responds with. Rename a field in `internal/model/model.go`:

```diff
 type CreateUser struct {
-	Name  string `json:"name" binding:"required"`  // full name
+	FullName string `json:"full_name" binding:"required"` // full name
 	Email string `json:"email" binding:"required"` // must be unique
+	Phone string `json:"phone" binding:"required"`
```

Nothing has to compile or run for this — `routy check` reads the source:

```console
$ routy check
api/v1/users/post.routy:17:8: error: required field `full_name` (string) is missing from body  ← internal/api/users.go:13
api/v1/users/post.routy:17:8: error: required field `phone` (string) is missing from body  ← internal/api/users.go:13
api/v1/users/post.routy:17:10: error: body field `name` is not in model.CreateUser  ← internal/api/users.go:13
14 files, 2 environments, 12 Go routes: 3 errors, 0 warnings; 3 fixable with `routy import go --fix`

$ routy import go . --fix
```

In a pull request the same errors appear as annotations on the changed lines — see
[`.github/workflows/api.yml`](.github/workflows/api.yml). Add a route and `routy import go .` creates
its request, with the body, query, headers and a response check already filled in.

## What is where

Every file in `api/` shows something; open them in order.

| Feature | File |
|---------|------|
| Environments, shared variables, a secret, middleware → headers | [`api/env.toml`](api/env.toml) |
| Checks: status, body fields, headers, timing, a JSON Schema | [`api/health.routy`](api/health.routy), [`api/schemas/`](api/schemas/health.schema.json) |
| Parameters with defaults, a secret as a parameter, `cache: 50m` between runs | [`api/auth/login/post.routy`](api/auth/login/post.routy) |
| A negative test next to the main request | [`api/auth/login/post.routy`](api/auth/login/post.routy), [`api/v1/users/post.routy`](api/v1/users/post.routy) |
| `form` body, the cookie jar, `cookies.session` | [`api/auth/session.routy`](api/auth/session.routy) |
| Requests calling requests: `Login().body.token`, `CreateUser().body.id`, sent once per run | [`api/v1/users/get-by-id.routy`](api/v1/users/get-by-id.routy) |
| Calls with arguments: log in as another user | [`api/v1/users/delete-by-id.routy`](api/v1/users/delete-by-id.routy) |
| `only: [dev]` and `confirm: true` for destructive requests | [`api/v1/users/delete-by-id.routy`](api/v1/users/delete-by-id.routy) |
| `multipart` upload of a file | [`api/v1/users/avatar/post.routy`](api/v1/users/avatar/post.routy) |
| File-level `let`, `uuid()`, `nowIso()`, arithmetic, `all` / `map` over arrays, `save` | [`api/v1/orders/post.routy`](api/v1/orders/post.routy) |
| Query parameters that are left out when `null` | [`api/v1/orders/get.routy`](api/v1/orders/get.routy) |
| Polling until the order is paid | [`api/v1/orders/pay/post.routy`](api/v1/orders/pay/post.routy) |
| Flows: a checkout scenario, an idempotent retry with `fresh` | [`api/flows/checkout.routy`](api/flows/checkout.routy) |
| Shapes generated from Go structs, made stricter by hand | [`api/shapes.routy`](api/shapes.routy) |

Run a single scenario by name — `routy run Checkout` — or a single request — `routy run CreateOrder`.
`routy check --env staging` checks the staging environment without sending anything.

## How `api/` was made

```sh
routy init            # api/env.toml
routy import go .     # a request for every Gin route and a shape for every response type
```

Then the generated requests got checks, parameters and calls, files were moved around (the session
requests live together in `api/auth/session.routy`) and the flows were written by hand. `routy import`
still recognizes every request by its `handler:`.

## Editors

- **VS Code**: the [Routy extension](https://marketplace.visualstudio.com/items?itemName=routy.routy-vscode)
  (recommended in `.vscode/extensions.json`) — completion, errors as you type, the Go drift as quick
  fixes, a **Send** lens over every request.
- **Desktop app**: open the folder in [Routy](https://github.com/1rowvy/routy/releases/latest); the
  **Routes** tab shows every Gin route and its state.
- **Neovim, Helix, Zed**: `routy lsp` and the tree-sitter grammar — see the
  [editors guide](https://1rowvy.github.io/routy/guides/editors/).

## Layout

```txt
main.go                    starts the server
internal/api/router.go     routes, groups, middleware
internal/api/*.go          handlers and the in-memory store
internal/model/model.go    request and response types
api/                       the Routy project
.github/workflows/api.yml  routy check + routy run on every push and PR
```

The service keeps everything in memory: restart it to start over.
