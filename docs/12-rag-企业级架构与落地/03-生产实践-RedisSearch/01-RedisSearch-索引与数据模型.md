# 01. RedisSearch 索引与数据模型

## 1. 先说结论：RedisSearch 很适合做企业级 RAG 的在线热数据检索层，但前提是你接受它的内存模型和治理要求

RedisSearch 在企业 RAG 里常被选中，  
主要因为它同时具备：

- 低延迟
- 较强的标签过滤能力
- 全文和向量混合检索能力
- 与 Go 服务集成简单

但它不是“默认万能解”。  
如果不提前理解它的边界，后面会遇到：

- 内存成本超预期
- 大规模索引重建窗口太长
- 多租户隔离混乱
- 文档和索引版本难以切换

所以本篇先明确一个判断：

**RedisSearch 最适合承担在线服务的热路径检索，不一定适合作为企业全部知识资产的唯一长期底座。**

## 2. RedisSearch 更适合什么场景

### 2.1 低延迟在线问答

如果你希望：

- 检索 p95 尽量压低
- 线上服务并发高
- 查询时需要标签、租户、文档类型过滤

RedisSearch 通常会比较顺手。

### 2.2 热数据知识库

适合：

- 客服 FAQ
- 高频访问知识
- 最近 3 到 6 个月的热门制度和流程文档

### 2.3 Agent 的实时检索子系统

Agent 编排里常常要求：

- 先快速取到相关上下文
- 再决定是否调用工具

这种场景下，RedisSearch 作为低延迟召回层很合适。

## 3. 不太适合直接用 RedisSearch 承担的情况

- 全量超大规模冷数据长期沉淀
- 需要复杂聚合分析和搜索报表
- 内存预算非常敏感
- 索引重建量极大且窗口非常紧

这时更常见的做法是：

- RedisSearch 承担热层
- Elasticsearch / 对象存储 / 数据库承担冷层或源数据层

## 4. 推荐的数据模型

建议把索引对象按 `chunk` 而不是整篇文档组织。  
一个比较实用的字段集合如下：

| 字段 | 类型 | 作用 |
| --- | --- | --- |
| `tenant_id` | `TAG` | 租户隔离 |
| `doc_id` | `TAG` | 文档主键 |
| `chunk_id` | `TAG` | chunk 主键 |
| `source_system` | `TAG` | 来源系统 |
| `doc_type` | `TAG` | 文档类型 |
| `title` | `TEXT` | 标题召回 |
| `body` | `TEXT` | 正文全文检索 |
| `path` | `TEXT` | 面包屑 / 路径 |
| `tags` | `TAG` | 业务标签 |
| `acl_roles` | `TAG` | 角色权限 |
| `acl_users` | `TAG` | 用户级权限 |
| `updated_at` | `NUMERIC` | 新鲜度排序 |
| `dataset_version` | `TAG` | 数据版本 |
| `index_version` | `TAG` | 索引版本 |
| `embedding` | `VECTOR` | 向量召回 |

## 5. 推荐的 key 设计

建议用可读且稳定的 key：

```text
rag:{tenant}:{dataset_version}:{doc_id}:{chunk_id}
```

这样做的好处是：

- 容易定位数据来源
- 便于按版本回收
- 便于批量扫描和统计

## 6. 一条推荐的索引定义

下面是一条常见的 `FT.CREATE` 示例：

```text
FT.CREATE idx:rag:prod:202604 ON HASH PREFIX 1 "rag:tenant-a:202604:" SCHEMA
tenant_id TAG
doc_id TAG
chunk_id TAG
source_system TAG
doc_type TAG
title TEXT WEIGHT 3.0
body TEXT WEIGHT 1.0
path TEXT WEIGHT 1.5
tags TAG SEPARATOR ","
acl_roles TAG SEPARATOR ","
acl_users TAG SEPARATOR ","
updated_at NUMERIC SORTABLE
dataset_version TAG
index_version TAG
embedding VECTOR HNSW 10 TYPE FLOAT32 DIM 1024 DISTANCE_METRIC COSINE
M 16 EF_CONSTRUCTION 200 EF_RUNTIME 64
```

这里面最关键的几个参数是：

- `DIM`：必须和 embedding 维度一致
- `DISTANCE_METRIC`：通常 `COSINE` 最常见
- `M`：图连接度，影响召回和内存
- `EF_CONSTRUCTION`：建图质量和建索引成本
- `EF_RUNTIME`：查询时的召回质量和延迟权衡

## 7. HNSW 参数调优建议

### 7.1 `M`

经验上：

- `12 - 16`：常见默认起点
- 更高：召回可能更好，但内存更大

### 7.2 `EF_CONSTRUCTION`

经验上：

- `100 - 200`：比较稳的起点
- 更高：索引构建更慢，但图质量更好

### 7.3 `EF_RUNTIME`

经验上建议按业务分档：

- 在线低延迟问答：`32 - 64`
- 精度优先查询：`64 - 128`

不要把这个值写死，  
更适合做成可动态配置。

## 8. FLAT 还是 HNSW

### 8.1 什么时候用 FLAT

- 数据量小
- 追求精确最近邻
- 可接受更高查询成本

### 8.2 什么时候用 HNSW

- 数据量中到大
- 更关注低延迟
- 可以接受 ANN 近似召回

企业在线服务里，  
绝大多数情况下会优先考虑 `HNSW`。

## 9. 多租户和多业务线隔离建议

不要只靠一个字段过滤来想象“已经隔离”。  
建议至少分三层考虑：

1. **逻辑租户字段**：`tenant_id`
2. **索引版本隔离**：不同业务线或环境使用不同 index alias
3. **物理资源隔离**：高敏感业务可单独 Redis 集群

## 10. RedisSearch 的索引设计建议

### 10.1 一份文档不要只保留正文

建议至少保留：

- 标题
- 面包屑路径
- 标签
- 更新时间
- 来源系统
- ACL

### 10.2 不要把 ACL 做成回答后过滤

更稳的方式是：

- 在检索阶段通过 `TAG` 先过滤
- 回答阶段再做最后校验

### 10.3 版本字段必须内置

如果没有 `dataset_version` 和 `index_version`，  
后面做重建、回滚和对账会很痛苦。

## 11. 推荐的容量规划思路

至少提前估算：

1. chunk 总量
2. 平均 chunk 长度
3. embedding 维度
4. 各类 TAG 字段数量
5. 副本数和高可用策略

RedisSearch 的成本估算，  
不能只按原始文本大小算。  
向量、索引结构和元数据都会占内存。

## 12. 代码参考

RedisSearch 的 Go 索引创建和查询示例见：  
[examples/redisearch_repository.go](./examples/redisearch_repository.go)

## 13. 小结

- RedisSearch 适合低延迟在线热数据检索
- 设计时必须先想好 key、字段、ACL、版本和索引参数
- 它适合作为企业级 RAG 的一层，但不一定是全部知识底座
