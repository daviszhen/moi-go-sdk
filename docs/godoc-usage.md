# Go Docs 使用指南

本文档说明如何使用 Go 的 `godoc` 工具查看 SDK 的文档。

## 查看包文档

```bash
# 查看整个包的文档
go doc .

# 查看包的简要信息
go doc -short .
```

## 查看类型和函数文档

```bash
# 查看类型文档
go doc RawClient
go doc SDKClient
go doc APIError

# 查看函数文档
go doc NewRawClient
go doc RawClient.CreateCatalog
go doc SDKClient.CreateTableRole

# 查看方法文档
go doc RawClient.CreateCatalog
go doc SDKClient.ImportLocalFileToTable
```

## 查看所有文档

```bash
# 查看包中所有导出符号的文档
go doc -all .
```

## 启动本地文档服务器

```bash
# 启动 HTTP 服务器，在浏览器中查看文档
go doc -http=:6060

# 然后在浏览器中访问 http://localhost:6060/pkg/github.com/matrixorigin/moi-go-sdk/
```

## 文档注释规范

所有导出的函数、类型和变量都已添加符合 Go 文档规范的注释：

1. **包注释**: 在 `package` 声明之前，描述包的用途
2. **函数注释**: 以函数名开头，描述功能、参数和返回值
3. **类型注释**: 以类型名开头，描述类型的用途和字段
4. **示例代码**: 使用 `Example:` 标记提供使用示例

## 已添加文档的文件

以下文件已添加完整的 Go 文档注释：

- ✅ `client.go` - 包注释、RawClient 类型、NewRawClient 函数
- ✅ `catalog.go` - 所有 Catalog 相关方法的文档
- ✅ `database.go` - 所有 Database 相关方法的文档
- ✅ `errors.go` - 错误类型和变量的文档
- ✅ `options.go` - 所有选项函数的文档
- ✅ `sdk_client.go` - SDKClient 类型和高级方法的文档

## 示例

### 查看包文档

```bash
$ go doc .
package sdk // import "github.com/matrixorigin/moi-go-sdk"

Package sdk provides a Go client library for interacting with the MOI Catalog
Service.

The package provides two types of clients:
  - RawClient: Low-level client that provides direct access to API endpoints
  - SDKClient: High-level client that provides convenient business-oriented APIs
...
```

### 查看函数文档

```bash
$ go doc RawClient.CreateCatalog
func (c *RawClient) CreateCatalog(ctx context.Context, req *CatalogCreateRequest, opts ...CallOption) (*CatalogCreateResponse, error)
    CreateCatalog creates a new catalog.

    The catalog is the top-level organizational structure for managing
    databases, tables, and volumes.

    Example:

        resp, err := client.CreateCatalog(ctx, &sdk.CatalogCreateRequest{
        	CatalogName: "my-catalog",
        	Comment:     "My catalog description",
        })
        if err != nil {
        	return err
        }
        fmt.Printf("Created catalog ID: %d\n", resp.CatalogID)
```

### 查看类型文档

```bash
$ go doc APIError
type APIError struct {
	Code       string // Error code returned by the server (e.g., "ErrInternal")
	Message    string // Human-readable error message
	RequestID  string // Unique request identifier for tracking purposes
	HTTPStatus int    // HTTP status code of the response
}
    APIError captures an application-level error returned by the catalog service
    envelope.

    APIError represents business logic errors returned by the server, such as
    validation errors, resource not found, permission denied, etc.
```

## 在 IDE 中使用

大多数 Go IDE（如 VS Code、GoLand）会自动显示这些文档注释：

- **悬停提示**: 将鼠标悬停在函数或类型上时显示文档
- **自动补全**: 在输入代码时显示函数签名和文档
- **快速查看**: 使用快捷键（如 VS Code 的 `Ctrl+Q`）查看完整文档

## 注意事项

1. 文档注释必须紧接在被注释的声明之前，中间不能有空行
2. 注释应该以被注释的声明名称开头
3. 使用 `Example:` 标记提供代码示例
4. 使用空行分隔不同的段落
5. 使用代码块格式（缩进）显示示例代码

## 更多信息

- [Go 文档注释规范](https://go.dev/doc/comment)
- [godoc 工具文档](https://pkg.go.dev/cmd/go#hdr-Show_documentation_for_package_or_symbol)

