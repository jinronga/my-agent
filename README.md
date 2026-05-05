# AI Agent 企业级学习与落地手册

这个仓库整理的是一套面向工程落地的 AI Agent 文档和示例工程。

它不是只解释概念，而是按下面这条主线组织：

> 从 Agent 基础概念，到评估、安全、工程化，再到企业级 RAG 和 Agent 项目交付标准。

## 当前内容结构

```text
docs/
├── 00-阅读地图与术语约定.md
├── 01-什么是-ai-agent.md
├── 02-ai-agent-能解决什么问题.md
├── 03-ai-agent-是如何构建的.md
├── 04-ai-agent-的常见工作流模式.md
├── 05-单-agent-和多-agent.md
├── 06-agent-背后的大模型基础.md
├── 07-agent-的记忆设计.md
├── 08-agent-的工具调用与-rag.md
├── 09-agent-的评估与测试.md
├── 10-agent-的安全、权限与人机协作.md
├── 11-agent-的工程化与生产落地.md
├── 12-rag-企业级架构与落地/
├── 13-agent-企业级项目标准/
└── 14-autoresearch-自动研究型-agent.md

examples/
└── rag-service/
```

## 建议阅读顺序

1. 先读 [阅读地图与术语约定](./docs/00-阅读地图与术语约定.md)
2. 再按 `01` 到 `11` 理解 Agent 的核心机制
3. 如果重点是知识库、检索、企业问答，读 [企业级 RAG 架构与落地](./docs/12-rag-企业级架构与落地/README.md)
4. 如果重点是项目交付、上线、安全合规和验收，读 [Agent 企业级项目标准](./docs/13-agent-企业级项目标准/README.md)
5. 如果想理解自动科研和持续实验型 Agent，读 [AutoResearch 与自动研究型 Agent](./docs/14-autoresearch-自动研究型-agent.md)
6. 如果想看最小可运行骨架，进入 [examples/rag-service](./examples/rag-service/README.md)

## 企业级项目标准

新增的企业级项目标准章节重点回答这些问题：

- 一个 Agent 项目到什么程度才算可以进企业生产环境？
- 平台层应该有哪些模块边界？
- 安全、权限、提示注入、越权动作和审计怎么治理？
- 评估集、上线门槛、SLO 和 Runbook 怎么定？
- 交付时应该准备哪些文档、接口、配置和验收资产？

入口见：[docs/13-agent-企业级项目标准/README.md](./docs/13-agent-企业级项目标准/README.md)

## AutoResearch 专题

[AutoResearch 与自动研究型 Agent](./docs/14-autoresearch-自动研究型-agent.md) 基于 `karpathy/autoresearch` 项目，整理了自动研究型 Agent 的核心模式：

- 固定评估基座
- 限定可修改面
- 固定实验预算
- 统一指标判断
- 实验日志和回滚
- 企业落地时的沙箱、权限、审计和人工准入

## 示例工程

[examples/rag-service](./examples/rag-service/README.md) 是一个最小 Go 服务骨架，包含：

- `POST /v1/query`
- 租户、用户、角色上下文
- 输入校验和简单策略控制
- 内存检索和引用返回
- trace id
- 单元测试
- OpenAPI 草案

它不是完整 RAG 引擎，目标是给企业级项目提供可继续扩展的工程起点。

## 企业级落地判断

如果一个项目准备上线，至少要能回答下面这些问题：

1. 任务状态、工具调用、检索证据和最终输出是否可追踪？
2. 读权限、写权限和高风险动作是否分层？
3. 是否有评估集、回归流程和上线准入阈值？
4. 是否有 SLO、告警、Runbook 和回滚演练？
5. 是否能证明数据来源、索引版本、prompt 版本和模型版本？
6. 是否有审计日志、敏感信息处理和事故响应流程？

这些问题如果答不上来，项目通常还处于原型阶段。
