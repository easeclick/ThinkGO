# ThinkGo

**ThinkPHP-inspired Go web framework** — for building structured, production-ready APIs with a familiar MVC pattern.

```
go get github.com/easeclick/ThinkGO
```

---

## Quick Start

```go
package main

import (
    "log"
    "github.com/easeclick/ThinkGO/internal/framework"
)

func main() {
    app := thinkgo.NewApp()

    r := thinkgo.NewRouter()
    r.Use(thinkgo.Recovery(), thinkgo.LoggerMW())
    r.Get("/hello", func(c *thinkgo.Context) error {
        return c.JSON(thinkgo.Map{"msg": "hello"})
    })
    r.PrintRoutes()

    app.SetRouter(r)
    if err := app.Run(":8888"); err != nil {
        log.Fatal(err)
    }
}
```

---

## Architecture

### App (IoC Container)

`App` is the application kernel — config, logger, database, router, and service container.

```go
app := thinkgo.NewApp()
app.Config().Load("config.yaml")
app.SetLogger(slog.New(slog.NewTextHandler(os.Stdout, nil)))
app.SetDB(db)
app.SetRouter(router)

// IoC container
app.Bind("mailer", func() any { return NewMailer() })
mailer := app.Make("mailer").(*Mailer)

app.Singleton("cache", func() any { return thinkgo.NewMemoryCache() })
app.Run(":8888") // blocks until SIGINT/SIGTERM
```

`Run()` binds the listener synchronously (fails fast on port conflict) and manages graceful shutdown + DB teardown.

### Router

| Method | Description |
|---|---|
| `Get`, `Post`, `Put`, `Delete`, `Patch`, `Options` | Register route by method |
| `Any` | Matches all HTTP methods |
| `Group("/prefix", mws...)` | Route group with shared prefix + middleware |
| `Resource("/users", handler)` | RESTful resource (Index/Store/Show/Update/Delete) |

**Route groups nest to any depth:**

```go
api := r.Group("/api")
v1 := api.Group("/v1")
admin := v1.Group("/admin")
admin.Get("/users/:id", handler) // → GET /api/v1/admin/users/42
```

**Method Not Allowed (405):** if the path matches but the method doesn't, the router returns `405` with an `Allow` header listing valid methods.

**Not Found (404):** unmatched paths return `404` with the requested path.

**Path parameters** use `:name` syntax:

```go
r.Get("/users/:id/posts/:postId", func(c *thinkgo.Context) error {
    id := c.Param("id")
    postId := c.Param("postId")
    ...
})
```

### Context

`Context` wraps `http.ResponseWriter` and `*http.Request` for the duration of a request.

**Per-request key-value store** (replaces `context.WithValue` abuse):

```go
c.Set("user_id", 42)
uid := c.Get("user_id").(int)
c.Has("user_id") // → true
```

**HTTP method helpers:**

```go
c.IsGet()     c.IsPost()     c.IsPut()
c.IsDelete()  c.IsPatch()    c.IsOptions()
c.IsAjax()    // checks X-Requested-With: XMLHttpRequest
c.Abort()     c.IsAborted()  // middleware chain control
```

**Input reading:**

```go
c.Param("id")          // URL path parameter
c.Query("page")        // query string (?page=1)
c.DefaultQuery("sort", "desc")
c.QueryInt("limit")    // → int (0 on missing/invalid)
c.QuerySlice("ids")    // → []string (?ids=1&ids=2)
c.Form("email")        // form body
c.FormSlice("tags")    // multiple form values
c.Header("Content-Type")
c.BindJSON(&req)       // decode JSON body into struct
c.RemoteIP()           // respects X-Forwarded-For, X-Real-IP
c.UserAgent()
c.ContentType()
```

**Response helpers:**

```go
c.JSON(data)            // JSON with current status
c.Success("ok", data)   // {"code":1, "msg":"ok", "data":...}
c.Error("bad request")  // {"code":0, "msg":"bad request"} + 400
c.Text("plain")
c.HTML("<h1>hello</h1>")
c.Redirect("/login")
```

### Response

Direct `Response` object for when you need explicit control:

```go
resp := thinkgo.NewResponse(ctx)
resp.CodeJSON(http.StatusCreated, data)
resp.JSONP("callback", data)
resp.NoContent()       // 204
resp.Fail("denied")    // alias for Error
resp.XML(data)         // placeholder
```

---

## Middleware

**Built-in middleware:**

```go
r.Use(thinkgo.Recovery())          // panic → 500 JSON + stack trace
r.Use(thinkgo.LoggerMW())          // log method, path, remote, duration
r.Use(thinkgo.CORSMiddleware())    // allow all origins (dev)
r.Use(thinkgo.WithCORS(origin, methods, headers))  // custom CORS
r.Use(thinkgo.RequestID())         // X-Request-Id (pass-through or generated)
r.Use(thinkgo.Timeout(30*time.Second))  // context deadline → 504
r.Use(thinkgo.BasicAuth("admin", "secret"))  // HTTP Basic Auth
```

**Per-route middleware:**

```go
r.Get("/admin", handler, thinkgo.BasicAuth("admin", "secret"))
```

**Custom middleware:**

```go
func MyMiddleware() thinkgo.Middleware {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            // before
            next.ServeHTTP(w, r)
            // after
        })
    }
}
```

---

## Plugin System

Plugins are first-class citizens in ThinkGo — each plugin independently registers routes, boots, and shuts down. The global registry follows the `init()` auto-registration pattern (`database/sql` style).

### Plugin Interface

```go
type Plugin interface {
    ID() string             // unique identifier
    Version() string        // version
    Description() string    // description
    Routes() []RouteInfo    // route list for AI discovery
    RegisterRoutes(r *thinkgo.Router)
    Boot(app *thinkgo.App) error
    Shutdown() error
}
```

Embed `BasePlugin` for no-op defaults:

```go
type MyPlugin struct{ plugin.BasePlugin }
func (p *MyPlugin) ID() string          { return "myplugin" }
func (p *MyPlugin) RegisterRoutes(r *thinkgo.Router) {
    r.Get("/myplugin/hello", handler)
}
```

### Lifecycle

```
Register() → RegisterRoutes() → Boot() → ... → Shutdown()
```

### Global Registry (init auto-registration)

```go
func init() {
    plugin.Register(&MyPlugin{})
}
```

```go
import _ "path/to/myplugin" // blank import triggers init()
```

### PluginManager

```go
pm := plugin.NewManager(app, router)
// Load from global registry
for _, p := range plugin.Registered() {
    pm.Register(p)
}
pm.Boot() // calls RegisterRoutes → Boot for all plugins
// ...
pm.Shutdown() // graceful shutdown, reverse order
```

### AI Discovery Endpoints

The plugin manager registers these endpoints automatically for AI clients and API gateways:

| Endpoint | Description |
|----------|-------------|
| `GET /-/plugins` | Returns metadata and routes for all registered plugins |
| `GET /-/api.json` | Returns full API specification (framework + plugins + routes) |

### Built-in Plugins

| Plugin | ID | Description |
|--------|----|-------------|
| Shopee | `shopee` | Shopee OpenAPI v2 — items, orders, image upload |
| 1688 (Alibaba) | `alibaba` | 1688 OpenAPI — product search, drop shipping |
| ERP Core | `erpcore` | Core ERP — products, orders, purchases, profit reports |

All plugins fall back to Mock mode when API keys are not configured.

---

## Validation

Struct tag validation (ThinkPHP-style rules):

```go
type LoginRequest struct {
    Username string `validate:"required|minLen:3|maxLen:20"`
    Password string `validate:"required|minLen:6"`
    Email    string `validate:"required|email"`
}

v := thinkgo.NewValidator()
if !v.ValidateStruct(req) {
    fmt.Println(v.Errors())  // map[field]error_message
}
```

**Supported rules:**

| Rule | Example | Description |
|---|---|---|
| `required` | `required` | Non-empty |
| `minLen:N` | `minLen:3` | Minimum length |
| `maxLen:N` | `maxLen:20` | Maximum length |
| `len:N` | `len:11` | Exact length |
| `min:N` | `min:0` | Minimum numeric value |
| `max:N` | `max:100` | Maximum numeric value |
| `email` | `email` | Email format |
| `numeric` | `numeric` | Digits with optional decimal |
| `integer` | `integer` | Digits only |
| `alpha` | `alpha` | Letters only |
| `alphaNum` | `alphaNum` | Letters and digits |
| `phone` | `phone` | Chinese phone (1[3-9]XXXXXXXXX) |
| `url` | `url` | http/https URL |
| `in:a,b,c` | `in:get,post` | Must be one of |
| `regex:pattern` | `regex:^[A-Z]+$` | Custom regex |

**Explicit rule map:**

```go
rules := map[string]string{
    "username": "required|minLen:3",
}
v.ValidateRules(data, rules)
```

---

## Database & Models

GORM integration:

```go
import "gorm.io/gorm"

var db *gorm.DB // initialized during app bootstrap
thinkgo.DB = db

// Model embedding (with timestamps + soft delete)
type User struct {
    thinkgo.Model
    Name  string `gorm:"size:100"`
    Email string `gorm:"uniqueIndex"`
}

func (u *User) TableName() string { return "users" }

// CRUD
thinkgo.DB.Create(&user)
thinkgo.DB.First(&user, id)
thinkgo.DB.Model(&user).Update("name", "Alice")

// Chainable ModelOps (ThinkPHP-style)
var users []User
thinkgo.UseModel(&User{}).
    Where("age > ?", 18).
    Order("created_at DESC").
    Limit(10).
    Find(&users)

// Pagination
total, err := thinkgo.UseModel(&User{}).
    Where("status = ?", "active").
    Paginate(page, 20, &users)
```

**Multiple connections:**

```go
mgr := thinkgo.NewDBManager()
mgr.Register("default", db1)
mgr.Register("reporting", db2)
mgr.Get()          // default
mgr.Get("reporting")
```

---

## Configuration

YAML-based with dot-notation access:

```yaml
# config/app.yaml
server:
  host: "0.0.0.0"
  port: 8888

database:
  dsn: "erp.db"
  auto_migrate: true
```

```go
cfg := app.Config()
cfg.Load("config/app.yaml")
host := cfg.GetString("server.host")    // "0.0.0.0"
port := cfg.GetInt("server.port")        // 8888
mig  := cfg.GetBool("database.auto_migrate") // true
cfg.Set("app.debug", true)
```

---

## Caching

**In-memory cache** (default):

```go
cache := thinkgo.NewMemoryCache()
cache.Set("key", value, 5*time.Minute)
v, ok := cache.Get("key")
cache.Delete("key")
cache.Clear()
```

**File cache:**

```go
cache := thinkgo.NewFileCache("/tmp/cache")
cache.Set("key", data, time.Hour)
```

---

## Events

ThinkPHP-style pub/sub:

```go
// Define handler
type UserRegistered struct{}
func (UserRegistered) Handle(args ...any) error {
    fmt.Println("user registered:", args[0])
    return nil
}

// Register
events := thinkgo.NewEventSystem()
events.Listen("user.registered", UserRegistered{})

// Fire
events.Trigger("user.registered", user)

// Or use plain functions
events.Listen("order.created", func(orderID int) error {
    return sendEmail(orderID)
})
```

---

## Template Rendering

```go
// Init
view := thinkgo.NewViewEngine("view/")
view.SetExtension(".html")
view.AddFunc("uppercase", strings.ToUpper)

// In handler
html, err := view.Render("index", data)
ctx.HTML(html)
```

Dev mode (reload on every request):

```go
view.SetCached(false)
```

---

## Logging

```go
logger := thinkgo.NewLogger("debug")
logger.Info("server started", "port", 8888)
logger.With("request_id", "abc").Error("db failed", "err", err)
```

Or use `slog` directly (the framework uses `slog` internally):

```go
app.SetLogger(slog.New(slog.NewJSONHandler(os.Stdout, nil)))
```

---

## Helpers

```go
thinkgo.StructToMap(user)   // struct → map[string]any (supports json tags, nesting)
thinkgo.MD5("hello")        // hex MD5 hash
thinkgo.Token(16)           // crypto/rand hex token (32 chars)
thinkgo.Now("Y-m-d H:i:s")  // current time (ThinkPHP format)
thinkgo.InArray("x", []string{"x","y"})  // slice contains
thinkgo.ArrayColumn(items, "name")       // extract column from map slice
thinkgo.Default(val, fallback)            // ?? operator
thinkgo.DD(val)                          // dump + die
thinkgo.Dump(val)                        // dump
```

**Pagination:**

```go
total := int64(100)
p := thinkgo.NewPagination(total, 1, 20)
// p.Pages=5 p.HasPrev=false p.HasNext=true
```

---

## Resource Controller

```go
type UserController struct {
    thinkgo.Controller
}

func (c *UserController) Index(ctx *thinkgo.Context) error {
    return ctx.Success("users list")
}

// Register
r.Resource("/users", &UserController{})

// Generates:
// GET    /users       → Index
// POST   /users       → Store
// GET    /users/:id   → Show
// PUT    /users/:id   → Update
// DELETE /users/:id   → Delete
```

---

## Full Example

### API Server (Plugin Architecture)

See [`internal/api/server.go`](internal/api/server.go) — uses the PluginManager to auto-load all registered plugins:

```go
func Run() {
    app := thinkgo.NewApp()
    app.Config().Load("config/app.yaml")

    db, _ := gorm.Open(sqlite.Open(app.Config().GetString("database.dsn")))
    thinkgo.DB = db
    app.SetDB(db)

    // Auto migration
    model.MigrateDB(db)

    router := thinkgo.NewRouter()
    router.Use(thinkgo.Recovery(), thinkgo.LoggerMW(), thinkgo.CORSMiddleware())
    router.Get("/ping", func(c *thinkgo.Context) error {
        return c.JSON(thinkgo.Map{"message": "pong"})
    })

    // Plugin system — load from global registry
    pm := plugin.NewManager(app, router)
    for _, p := range plugin.Registered() {
        pm.Register(p)
    }
    pm.Boot() // auto-registers routes + Boot (includes AI discovery endpoints)

    router.PrintRoutes()
    app.SetRouter(router)
    app.Run(thinkgo.ListenAddr(
        app.Config().GetString("server.host"),
        app.Config().GetInt("server.port"),
    ))
}
```

### ERP Business Modules

The project includes a complete cross-border ERP demo (Mock mode, works out of the box):

| Module | Description |
|--------|-------------|
| [shopee](internal/shopee/) | Shopee API signing, items/orders interfaces |
| [alibaba](internal/alibaba/) | 1688 API signing, product search, purchase orders |
| [erpcore](internal/erpcore/) | Core ERP calculations: profit reports, auto-purchase |
| [model](internal/model/) | GORM data models + migration + seed data |
| [worker](internal/worker/) | Background tasks: order sync, low stock alerts |
| [api](internal/api/) | API server entry point with plugin integration |
| [plugins](plugins/) | Plugin implementations (shopee / alibaba / erpcore) |

**Business API Routes:**

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/products` | List all products |
| GET | `/api/v1/products/:id` | Get product by ID |
| GET | `/api/v1/orders` | List orders |
| GET | `/api/v1/orders/:id` | Get order by ID (supports order_id string) |
| GET | `/api/v1/purchases` | List 1688 purchase orders |
| GET | `/api/v1/search?keyword=&page=` | 1688 product search (Mock) |
| GET | `/api/v1/report/daily?date=2026-01-01` | Daily profit report |
| GET | `/api/v1/report/monthly?year=&month=` | Monthly profit report (best/worst seller analysis) |

**Running:**

```bash
go run main.go migrate   # Initialize tables
go run main.go seed      # Fill mock data (50 orders + 5 products + 20 purchases)
go run main.go api       # Start API server (default :8888)
go run main.go worker    # Start background worker (auto-sync/restock)
```

---

## License

MIT
