# go-o11y-base

基於 Go 的 codebase，旨在提供一個結構清晰、開發體驗優良且具備強大可觀測性（Observability）的基底。

## Features

- **分層架構 (Layered Architecture)**: 採用類似 Clean Architecture 的設計模式，將業務邏輯（Usecase）、領域模型（Domain）與基礎設施（Infrastructure）及交付層（Delivery）解耦，提高測試覆蓋率與後續維護性。
- **全方位可觀測性整合 (Integrated Observability)**:
  - **Logger & Error 深度整合**: 專案封裝了 Logger 與 Error 處理機制。透過 Middleware 的串聯，確保所有的日誌輸出與錯誤捕捉都能自動關聯到當前的 Trace 鏈路中。
  - **Trace 鏈路追踪**: 整合 OpenTelemetry (OTEL)，實現跨組件的請求追踪鏈路。
  - **Metrics**: 內建 Prometheus 指標搜集，支援即時監控服務健康狀態與性能數據。
  - **Continuous Profiling**: 整合 Pyroscope，實現開發環境與生產環境的持續效能剖析。
- **依賴注入 (Dependency Injection)**: 使用 `uber-go/dig` 進行依賴管理，簡化組件間的初始化與組裝流程。

---

## Setup

### Run server
```bash
$ cp ./.env.example ./.env  # 用於配置本地環境變數
$ go run cmd/server/main.go
```

### 常用指令
```bash
$ make install-bin   # 安裝專案指定版本工具
$ make uninstall-bin # 移除專案指定版本工具
$ make mock          # Generate Go mock files
$ make test          # 執行 Unit testing
$ make lint          # 執行 Linter
$ make mysql-up      # 按順序遷移 DB schema changes
$ make mysql-down    # 按順序回滾 DB schema changes
$ make pprof-heap    # 執行 pprof heap 分析
$ make pprof-profile # 執行 pprof profile 分析
$ make pprof-goleak  # 執行 pprof goroutine leak 分析
```

### 生成 Mock
- **指令**:
  ```bash
  $ make mock
  ```
- **使用時機**:
  - 新增 interface 後, 須更新 `./Makefile` 中的 `mock` 命令, 以添加 genmock 對象, 再執行指令
  - 變更 interface 定義後, 再執行指令
  - CI 執行 unit testing 前, 須執行指令

### 生成 JWT Key Pair (Ed25519)
```bash
$ openssl genpkey -algorithm ed25519 -out ed25519_private.pem          # 生成私鑰
$ openssl pkey -in ed25519_private.pem -pubout -out ed25519_public.pem # 生成公鑰
$ cat ed25519_private.pem | base64 | tr -d '\n'                        # 輸出成私鑰 Base64
$ cat ed25519_public.pem | base64 | tr -d '\n'                         # 輸出成公鑰 Base64
```