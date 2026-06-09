# ThinkGo

**ThinkPHP 风格的 Go Web 框架** — 用于构建结构清晰、生产就绪的 API 服务，采用熟悉的 MVC 模式。

---

## 快速开始

```go
package main

import (
    "log"
    "github.com/user/thinkgo/internal/framework"
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

## 架构

### App（IoC 容器）

`App` 是应用内核 — 统一管理配置、日志、数据库、路由和服务容器。

```go
app := thinkgo.NewApp()
app.Config().Load("config.yaml")
app.SetLogger(slog.New(slog.NewTextHandler(os.Stdout, nil)))
app.SetDB(db)
app.SetRouter(router)

// IoC 容器
app.Bind("mailer", func() any { return NewMailer() })
mailer := app.Make("mailer").(*Mailer)

app.Singleton("cache", func() any { return thinkgo.NewMemoryCache() })
app.Run(":8888") // 阻塞等待 SIGINT/SIGTERM
```

`Run()` 使用 `net.Listen` 同步绑定端口（端口冲突立即报错），内置优雅关闭 + 数据库清理。

### Router（路由器）

| 方法 | 说明 |
|---|---|
| `Get` / `Post` / `Put` / `Delete` / `Patch` / `Options` | 按方法注册路由 |
| `Any` | 匹配所有 HTTP 方法 |
| `Group("/prefix", mws...)` | 路由分组，共享前缀 + 中间件 |
| `Resource("/users", handler)` | RESTful 资源路由（Index/Store/Show/Update/Delete） |

**路由组支持任意深度嵌套：**

```go
api := r.Group("/api")
v1 := api.Group("/v1")
admin := v1.Group("/admin")
admin.Get("/users/:id", handler) // → GET /api/v1/admin/users/42
```

**405 Method Not Allowed：** 路径匹配但方法不对时，返回 `405`，附带 `Allow` 头列出允许的方法。

**404 Not Found：** 未匹配的路径返回 `404`，包含请求路径。

**路径参数** 使用 `:name` 语法：

```go
r.Get("/users/:id/posts/:postId", func(c *thinkgo.Context) error {
    id := c.Param("id")
    postId := c.Param("postId")
    // ...
})
```

### Context（上下文）

`Context` 包装 `http.ResponseWriter` 和 `*http.Request`，贯穿请求生命周期。

**请求级键值存储**（替代 `context.WithValue` 滥用）：

```go
c.Set("user_id", 42)
uid := c.Get("user_id").(int)
c.Has("user_id") // → true
```

**HTTP 方法判断：**

```go
c.IsGet()     c.IsPost()     c.IsPut()
c.IsDelete()  c.IsPatch()    c.IsOptions()
c.IsAjax()    // 检查 X-Requested-With: XMLHttpRequest
c.Abort()     c.IsAborted()  // 中间件链控制
```

**输入读取：**

```go
c.Param("id")          // URL 路径参数
c.Query("page")        // 查询字符串 (?page=1)
c.DefaultQuery("sort", "desc")
c.QueryInt("limit")    // → int（没有或无效返回 0）
c.QuerySlice("ids")    // → []string (?ids=1&ids=2)
c.Form("email")        // 表单字段
c.FormSlice("tags")    // 多个表单值
c.Header("Content-Type")
c.BindJSON(&req)       // JSON 解码到结构体
c.RemoteIP()           // 支持 X-Forwarded-For, X-Real-IP
c.UserAgent()
c.ContentType()
```

**输出响应：**

```go
c.JSON(data)            // JSON，使用当前状态码
c.Success("ok", data)   // {"code":1, "msg":"ok", "data":...}
c.Error("bad request")  // {"code":0, "msg":"bad request"} + 400
c.Text("plain")
c.HTML("<h1>hello</h1>")
c.Redirect("/login")
```

### Response（响应对象）

需要精细控制时直接使用 `Response`：

```go
resp := thinkgo.NewResponse(ctx)
resp.CodeJSON(http.StatusCreated, data)
resp.JSONP("callback", data)
resp.NoContent()       // 204
resp.Fail("denied")    // Error 的别名
resp.XML(data)         // 占位
```

---

## 中间件

**内置中间件：**

```go
r.Use(thinkgo.Recovery())          // 捕获 panic → 500 JSON + 堆栈
r.Use(thinkgo.LoggerMW())          // 记录方法、路径、客户端、耗时
r.Use(thinkgo.CORSMiddleware())    // 允许所有来源（开发用）
r.Use(thinkgo.WithCORS(origin, methods, headers))  // 自定义 CORS
r.Use(thinkgo.RequestID())         // X-Request-Id（透传或自动生成）
r.Use(thinkgo.Timeout(30*time.Second))  // context 超时 → 504
r.Use(thinkgo.BasicAuth("admin", "secret"))  // HTTP Basic 认证
```

**路由级中间件：**

```go
r.Get("/admin", handler, thinkgo.BasicAuth("admin", "secret"))
```

**自定义中间件：**

```go
func MyMiddleware() thinkgo.Middleware {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            // 前置处理
            next.ServeHTTP(w, r)
            // 后置处理
        })
    }
}
```

---

## 参数验证

结构体标签验证（ThinkPHP 风格规则）：

```go
type LoginRequest struct {
    Username string `validate:"required|minLen:3|maxLen:20"`
    Password string `validate:"required|minLen:6"`
    Email    string `validate:"required|email"`
}

v := thinkgo.NewValidator()
if !v.ValidateStruct(req) {
    fmt.Println(v.Errors())  // map[字段名]错误信息
}
```

**支持的规则：**

| 规则 | 示例 | 说明 |
|---|---|---|
| `required` | `required` | 非空 |
| `minLen:N` | `minLen:3` | 最小长度 |
| `maxLen:N` | `maxLen:20` | 最大长度 |
| `len:N` | `len:11` | 固定长度 |
| `min:N` | `min:0` | 最小值 |
| `max:N` | `max:100` | 最大值 |
| `email` | `email` | 邮箱格式 |
| `numeric` | `numeric` | 数字（含小数） |
| `integer` | `integer` | 整数 |
| `alpha` | `alpha` | 纯字母 |
| `alphaNum` | `alphaNum` | 字母和数字 |
| `phone` | `phone` | 手机号（1[3-9]XXXXXXXXX） |
| `url` | `url` | http/https URL |
| `in:a,b,c` | `in:get,post` | 枚举值 |
| `regex:pattern` | `regex:^[A-Z]+$` | 自定义正则 |

**显式规则映射：**

```go
rules := map[string]string{
    "username": "required|minLen:3",
}
v.ValidateRules(data, rules)
```

---

## 数据库与模型

GORM 集成：

```go
import "gorm.io/gorm"

var db *gorm.DB // 应用启动时初始化
thinkgo.DB = db

// 嵌入模型（自带时间戳 + 软删除）
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

// 链式查询（ThinkPHP 风格）
var users []User
thinkgo.UseModel(&User{}).
    Where("age > ?", 18).
    Order("created_at DESC").
    Limit(10).
    Find(&users)

// 分页查询
total, err := thinkgo.UseModel(&User{}).
    Where("status = ?", "active").
    Paginate(page, 20, &users)
```

**多数据库连接：**

```go
mgr := thinkgo.NewDBManager()
mgr.Register("default", db1)
mgr.Register("reporting", db2)
mgr.Get()          // 默认连接
mgr.Get("reporting")
```

---

## 配置管理

YAML 配置文件，点号语法访问：

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

## 缓存

**内存缓存（默认）：**

```go
cache := thinkgo.NewMemoryCache()
cache.Set("key", value, 5*time.Minute)
v, ok := cache.Get("key")
cache.Delete("key")
cache.Clear()
```

**文件缓存：**

```go
cache := thinkgo.NewFileCache("/tmp/cache")
cache.Set("key", data, time.Hour)
```

---

## 事件系统

ThinkPHP 风格发布订阅：

```go
// 定义监听器
type UserRegistered struct{}
func (UserRegistered) Handle(args ...any) error {
    fmt.Println("用户注册:", args[0])
    return nil
}

// 注册
events := thinkgo.NewEventSystem()
events.Listen("user.registered", UserRegistered{})

// 触发
events.Trigger("user.registered", user)

// 也支持普通函数
events.Listen("order.created", func(orderID int) error {
    return sendEmail(orderID)
})
```

---

## 模板渲染

```go
// 初始化
view := thinkgo.NewViewEngine("view/")
view.SetExtension(".html")
view.AddFunc("uppercase", strings.ToUpper)

// 控制器中使用
html, err := view.Render("index", data)
ctx.HTML(html)
```

开发模式（每次请求重新加载模板）：

```go
view.SetCached(false)
```

---

## 日志

```go
logger := thinkgo.NewLogger("debug")
logger.Info("服务已启动", "port", 8888)
logger.With("request_id", "abc").Error("数据库错误", "err", err)
```

或直接使用 `slog`（框架内部使用 `slog`）：

```go
app.SetLogger(slog.New(slog.NewJSONHandler(os.Stdout, nil)))
```

---

## 辅助函数

```go
thinkgo.StructToMap(user)   // 结构体 → map[string]any（支持 json 标签、嵌套）
thinkgo.MD5("hello")        // MD5 哈希
thinkgo.Token(16)           // 随机 Token（32 位十六进制）
thinkgo.Now("Y-m-d H:i:s")  // 当前时间（ThinkPHP 日期格式）
thinkgo.InArray("x", []string{"x","y"})  // 判断元素是否在切片中
thinkgo.ArrayColumn(items, "name")       // 提取 map 切片中的字段
thinkgo.Default(val, fallback)            // 空值默认值（?? 运算符）
thinkgo.DD(val)                          // 打印并终止
thinkgo.Dump(val)                        // 打印
```

**分页辅助：**

```go
total := int64(100)
p := thinkgo.NewPagination(total, 1, 20)
// p.Pages=5  p.HasPrev=false  p.HasNext=true
```

---

## RESTful 资源控制器

```go
type UserController struct {
    thinkgo.Controller
}

func (c *UserController) Index(ctx *thinkgo.Context) error {
    return ctx.Success("用户列表")
}

// 注册
r.Resource("/users", &UserController{})

// 自动生成：
// GET    /users       → Index
// POST   /users       → Store
// GET    /users/:id   → Show
// PUT    /users/:id   → Update
// DELETE /users/:id   → Delete
```

---

## 完整示例

参见 [`cmd/api/main.go`](cmd/api/main.go) 和 [`internal/api/server.go`](internal/api/server.go) 的完整 API 服务搭建，包含数据库、迁移、路由注册和优雅关闭。

```go
func Run() {
    app := thinkgo.NewApp()
    app.Config().Load("config/app.yaml")
    app.SetLogger(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug})))

    db, _ := gorm.Open(sqlite.Open(app.Config().GetString("database.dsn")))
    thinkgo.DB = db
    app.SetDB(db)

    router := thinkgo.NewRouter()
    router.Use(thinkgo.Recovery(), thinkgo.LoggerMW(), thinkgo.CORSMiddleware())
    router.Get("/ping", func(c *thinkgo.Context) error {
        return c.JSON(thinkgo.Map{"message": "pong"})
    })
    router.PrintRoutes()

    app.SetRouter(router)
    app.Run(thinkgo.ListenAddr(
        app.Config().GetString("server.host"),
        app.Config().GetInt("server.port"),
    ))
}
```

---

## License

MIT
