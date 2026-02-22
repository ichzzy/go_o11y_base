# go-o11y-base

## Setup

### Run server
```
$ cp ./.env.example ./.env  # 用於配置本地環境變數
$ go run cmd/server/main.go
```
### 常用指令
```
$ make install-bin   # 安裝本專案指定版本工具
$ make uninstall-bin # 移除本專案指定版本工具
$ make mock          # 生成 Mock
$ make test          # 執行 Unit testing
$ make lint          # 執行 Linter
$ make mysql-init    # 建立 MySQL DB (本地要裝mysql)
$ make mysql-up      # 按順序遷移 DB schema changes
$ make mysql-down    # 按順序回滾 DB schema changes
$ make pprof-heap    # 執行 pprof heap 分析
$ make pprof-profile # 執行 pprof profile 分析
$ make pprof-goleak  # 執行 pprof goroutine leak 分析
```
### 生成 Mock
- 指令:
  ```
  $ make mock
  ```
- 使用時機:
  - 新增 interface 後, 須更新 `./Makefile` 中的 `mock` 命令, 以添加 genmock 對象, 再執行指令
  - 變更 interface 定義後, 再執行指令
  - CI 執行 unit testing 前, 須執行指令

### 生成 JWT Key Pair (Ed25519)
```
$ openssl genpkey -algorithm ed25519 -out ed25519_private.pem          # 生成私鑰
$ openssl pkey -in ed25519_private.pem -pubout -out ed25519_public.pem # 生成公鑰
$ cat ed25519_private.pem | base64 | tr -d '\n'                        # 輸出成私鑰 Base64
$ cat ed25519_public.pem | base64 | tr -d '\n'                         # 輸出成公鑰 Base64
```