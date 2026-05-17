# pkg/storage

`pkg/storage` 是业务无关的文件处理封装，统一 afero 文件系统、文件复制、MIME 检测和 OS 文件监听。它不承载业务流程、传输 DTO、领域模型或持久化模型。

## 设计边界

- `pkg/storage` 不导入 `internal`，只提供可复用文件能力。
- 所有传入路径都是 storage-relative path；绝对路径和 `..` 越界会返回 `ErrInvalidPath`。
- OS 存储使用调用方提供的 root 作为边界；内存存储用于快速单元测试和 mock。
- 文件监听只支持 OS-rooted storage；内存和 custom storage 会返回 `ErrUnsupported`。
- Excel 和图片处理不在当前版本实现，后续可在明确需求下追加 `excelize` 和 `imaging` 能力。

## 使用说明

### OS 文件系统

```go
store, err := storage.NewOS("data")
if err != nil {
    return err
}

if err := store.WriteFile("uploads/readme.txt", []byte("hello")); err != nil {
    return err
}

content, err := store.ReadFile("uploads/readme.txt")
if err != nil {
    return err
}
_ = content
```

### 内存文件系统

```go
store, err := storage.NewMemory()
if err != nil {
    return err
}

if err := store.WriteFile("fixtures/a.txt", []byte("mock")); err != nil {
    return err
}
```

### 只读包装

```go
base, _ := storage.NewMemory()
_ = base.WriteFile("a.txt", []byte("content"))

readonly, err := storage.NewReadOnly(base)
if err != nil {
    return err
}

_, err = readonly.ReadFile("a.txt")
```

### 复制与 MIME 检测

```go
if err := store.Copy("uploads/readme.txt", "backup/readme.txt"); err != nil {
    return err
}

info, err := store.DetectMIME("backup/readme.txt")
if err != nil {
    return err
}
_ = info.MIME
```

复制默认不覆盖目标；需要覆盖时显式传入：

```go
err := store.Copy("source.txt", "target.txt", storage.WithOverwrite(true))
```

### 文件监听

```go
watcher, err := store.Watch("uploads", storage.WithRecursiveWatch(true))
if err != nil {
    return err
}
defer watcher.Close()

for {
    select {
    case event := <-watcher.Events():
        _ = event.Path
    case err := <-watcher.Errors():
        return err
    }
}
```

## 错误处理

调用方可以用 `errors.Is` 判断稳定错误：

- `ErrInvalidConfig`：配置不完整或不合法。
- `ErrInvalidPath`：路径为空、绝对路径或逃逸 storage root。
- `ErrUnsupported`：当前后端不支持该操作。
- `ErrAlreadyExists`：复制目标已存在且未允许覆盖。
