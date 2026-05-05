# rag-service 示例工程

这是一个最小可运行的企业级 RAG / Agent 查询服务骨架。

它不直接接真实大模型或 RedisSearch，目标是先把企业级项目最容易漏掉的工程边界放进去：

- 租户、用户、角色上下文
- 输入校验
- 简单策略控制
- 检索候选和引用返回
- 无证据时拒答
- trace id
- 单元测试
- OpenAPI 草案

## 快速开始

```bash
go test ./...
go run ./cmd/query-api
```

启动后调用：

```bash
curl -s http://localhost:8080/v1/query \
  -H 'Content-Type: application/json' \
  -H 'X-Tenant-ID: tenant_a' \
  -H 'X-User-ID: u_001' \
  -H 'X-Roles: support_agent' \
  -d '{"query":"怎么处理客户投诉？","scene":"customer_support","max_results":2}'
```

## 目录

```text
.
├── api/openapi.yaml
├── cmd/query-api/main.go
├── configs/config.example.yaml
├── internal/
│   ├── app/
│   ├── httpapi/
│   ├── observability/
│   ├── policy/
│   ├── retrieval/
│   └── types/
├── Makefile
└── go.mod
```

## 后续扩展点

- 将 `internal/retrieval.MemoryRetriever` 替换为 RedisSearch / Elasticsearch / pgvector。
- 将 `app.Service` 中的生成逻辑替换为 Model Gateway。
- 将 `policy.ValidateQuery` 扩展成独立 Policy Engine。
- 增加 Tool Gateway、任务状态机、审计日志和评估 runner。

