# CLAUDE - AI 助手指南

这是为 AI 助手（如 Claude）准备的项目文档入口。

## 📚 完整项目文档

**请访问：[docs/INDEX.md](docs/INDEX.md)**

SingBox Config 是一个基于 Go 语言的sing-box 代理配置管理系统

## 🚨 文档管理严格约束

### 必须遵守的规则

1. **必须先阅读文档**
   - 在回答任何项目相关问题前，必须先阅读 [docs/INDEX.md](docs/INDEX.md)
   - 根据问题类型查阅对应的具体文档，不得凭空猜测或编造信息

2. **禁止随意创建文档目录**
   - 文档目录结构已在 [docs/INDEX.md](docs/INDEX.md) 中明确定义
   - 不得在 `docs/` 下创建任何未在 INDEX.md 中声明的新目录
   - 不得在项目根目录创建除 `CLAUDE.md`、`CODEX.md`、`README.md` 之外的其他 `.md` 文件

3. **严格遵循文档结构约定**
   - 项目根目录只允许存在以下文档文件：
     - `README.md` - 项目简介（指向 docs/INDEX.md）
     - `CLAUDE.md` - AI 助手指南（本文件）
     - `CODEX.md` - 开发者指南（指向 docs/INDEX.md）
   - 所有详细文档必须放在 `docs/` 目录下
   - `docs/` 目录结构必须严格遵循 [docs/INDEX.md](docs/INDEX.md) 中定义的结构

4. **文档修改流程**
   - 修改现有文档：直接编辑对应文件
   - 新增文档：必须先在 [docs/INDEX.md](docs/INDEX.md) 中添加目录条目，然后再创建文件
   - 调整目录结构：必须先更新 [docs/INDEX.md](docs/INDEX.md)，确保结构图和实际目录一致

5. **禁止的操作**
   - ❌ 在项目根目录创建 `ARCHITECTURE.md`、`DEVELOPMENT.md` 等文档
   - ❌ 在 `docs/` 下创建 `docs/new-feature/` 等未定义的目录
   - ❌ 创建与现有文档重复的内容
   - ❌ 不经过 INDEX.md 直接创建新文档

## AI 助手使用建议

当用户询问项目相关问题时，请先阅读 [docs/INDEX.md](docs/INDEX.md)，然后根据问题类型查阅相应文档：

- **项目概况** → [docs/architecture/overview.md](docs/architecture/overview.md)
- **快速开始** → [docs/guides/quickstart.md](docs/guides/quickstart.md)
- **开发指南** → [docs/guides/development.md](docs/guides/development.md)
- **API 使用** → [docs/guides/api-guide.md](docs/guides/api-guide.md)
- **模块功能** → [docs/modules/](docs/modules/) 目录下对应文档
- **配置说明** → [docs/reference/configuration.md](docs/reference/configuration.md)
- **常见问题** → [docs/reference/faq.md](docs/reference/faq.md)

---

所有项目信息都在 [docs/INDEX.md](docs/INDEX.md)，请从那里开始阅读。