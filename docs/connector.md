# Connector（连接器）

连接器接口负责管理导入任务所需的原始文件：上传、预览、导入以及下载/删除。本文档基于 `RawClient`，涵盖 `connector.go` 中的全部方法。

## 功能总览

| 方法 | 说明 |
| ---- | ---- |
| `UploadLocalFiles` / `UploadLocalFile` / `UploadLocalFileFromPath` | 将本地文件上传到临时存储，获得 `conn_file_id` |
| `FilePreview` | 预览已上传文件的结构，用于构建 `TableConfig` |
| `UploadConnectorFile` | 引用上传的文件并发起导入任务 |
| `DownloadConnectorFile` | 为指定 `conn_file_id` 生成一次性下载链接 |
| `DeleteConnectorFile` | 删除指定的 `conn_file_id`，释放临时存储 |

> 所有示例默认已通过 `sdk.NewRawClient(baseURL, apiKey)` 创建 `rawClient`，并准备好 `ctx := context.Background()`。

## 上传本地文件

```go
filePath := "/tmp/data.csv"
meta := []sdk.FileMeta{
	{Filename: "data.csv", Path: "/uploads"},
}

resp, err := rawClient.UploadLocalFileFromPath(ctx, filePath, meta)
if err != nil {
	log.Fatal(err)
}

connFileID := resp.ConnFileIds[0]
fmt.Printf("Uploaded connector file: %s\n", connFileID)
```

`UploadLocalFiles` 用于多文件上传，`UploadLocalFile`/`UploadLocalFileFromPath` 是便捷包装。

## 预览文件

```go
previewResp, err := rawClient.FilePreview(ctx, &sdk.FilePreviewRequest{
	ConnFileId:    connFileID,
	IsColumnName:  true,
	ColumnNameRow: 1,
	RowStart:      2,
})
if err != nil {
	log.Fatal(err)
}

for _, row := range previewResp.Rows {
	fmt.Printf("Row %d: %v\n", row.Number, row.ColumnValues)
}
```

## 发起导入任务

```go
taskResp, err := rawClient.UploadConnectorFile(ctx, &sdk.UploadFileRequest{
	VolumeID: "your-volume-id",
	Meta:     meta,
	TableConfig: &sdk.TableConfig{
		NewTable:    true,
		DatabaseID:  123,
		ConnFileIDs: []string{connFileID},
		// ...
	},
})
if err != nil {
	log.Fatal(err)
}
fmt.Printf("Import task id: %d\n", taskResp.TaskId)
```

## 下载已上传文件

`DownloadConnectorFile` 会返回一次性下载 URL，可直接通过 HTTP 客户端获取文件内容。下面示例在 SDK 测试用例中也采用相同逻辑：上传 -> 获取下载链接 -> 校验内容。

```go
downloadResp, err := rawClient.DownloadConnectorFile(ctx, &sdk.ConnectorFileDownloadRequest{
	ConnFileId: connFileID,
})
if err != nil {
	log.Fatal(err)
}

httpResp, err := http.Get(downloadResp.URL)
if err != nil {
	log.Fatal(err)
}
defer httpResp.Body.Close()

body, err := io.ReadAll(httpResp.Body)
if err != nil {
	log.Fatal(err)
}
fmt.Printf("Downloaded %d bytes\n", len(body))
```

> 该 URL 通常带有过期时间，请在短时间内使用并避免重复传播。

## 删除已上传文件

连接器临时文件不会自动清理，完成导入或下载后，请调用 `DeleteConnectorFile` 释放空间。

```go
_, err := rawClient.DeleteConnectorFile(ctx, &sdk.ConnectorFileDeleteRequest{
	ConnFileId: connFileID,
})
if err != nil {
	log.Fatal(err)
}
fmt.Printf("Deleted connector file %s\n", connFileID)
```

## 常见问题

1. **下载返回 403/404？** 下载 URL 已过期，请重新调用 `DownloadConnectorFile`。
2. **删除失败？** 确认 `conn_file_id` 来自同一账号/API Key，避免重复删除。
3. **能否批量删除？** 当前接口仅支持单个删除，可在客户端遍历实现。

