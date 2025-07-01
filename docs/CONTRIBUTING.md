# 贡献指南

我们欢迎任何形式的贡献！为了保持项目的高质量和协作的顺畅，请在提交代码或 Issue 前仔细阅读本指南。

## 1. 任务与 Bug 追踪

我们使用 **GitHub Issues** 来追踪所有的开发任务、功能需求和 Bug。

- **创建 Issue**: 在开始开发新功能或修复 Bug 之前，请先创建一个 Issue 来描述它。这有助于我们讨论和确认需求。
- **关联 PR**: 当你提交一个用于解决某个 Issue 的 Pull Request 时，请在 PR 的描述中添加 `Closes #<issue-number>` 或 `Fixes #<issue-number>`。这样，当 PR 被合并后，对应的 Issue 会被自动关闭。
- **使用标签 (Labels)**: 我们会使用标签来对 Issue 进行分类（如 `bug`, `feature`, `p1-mvp` 等），方便管理和筛选。

## 2. Git 工作流

我们采用一个简化的 **GitHub Flow** 作为我们的协作模型。

### 2.1. 分支管理

- **`main` 分支**: 这是项目的主分支，永远处于**稳定且可部署**的状态。
- **禁止直接提交**: **严禁**直接向 `main` 分支 `push` 代码。所有代码变更都必须通过 Pull Request (PR) 进行。
- **功能与修复分支**:
  - 从 `main` 分支创建你的工作分支。
  - 分支命名应清晰地反映其目的，并遵循以下约定：
    - **新功能**: `feature/<short-feature-name>` (例如: `feature/user-login-api`)
    - **Bug修复**: `fix/<short-fix-name>` (例如: `fix/post-slug-generation`)
    - **文档**: `docs/<topic>` (例如: `docs/update-contributing-guide`)
    - **重构**: `refactor/<area>` (例如: `refactor/database-layer`)

### 2.2. 提交信息规范 (Commit Message)

我们严格遵循 **[Conventional Commits](https://www.conventionalcommits.org/zh-hans/v1.0.0/)** 规范。一个格式良好的 Commit Message 能极大地帮助我们追踪变更、生成发布日志。

- **格式**: `<type>(<scope>): <subject>`
- **常用 `type`**:
  - `feat`: 引入新功能。
  - `fix`: 修复 Bug。
  - `docs`: 只修改了文档。
  - `style`: 代码格式调整（不影响代码逻辑，如空格、分号等）。
  - `refactor`: 代码重构（既不是新增功能，也不是修复 Bug）。
  - `perf`: 提升性能的改动。
  - `test`: 增加或修改测试用例。
  - `build`: 影响构建系统或外部依赖的变更（如修改 `go.mod`）。
  - `ci`: CI/CD 流程相关的变更。
  - `chore`: 其他不修改源码或测试的杂项变动。
- **`scope` (可选)**: 括号内用于说明本次提交影响的范围（如 `auth`, `posts`, `dao` 等）。
- **`subject`**: 简明扼要地描述本次提交的目的。

- **示例**:
  - `feat(auth): add JWT generation for user login`
  - `fix(posts): correct pagination logic for draft posts`
  - `test(dao): add unit tests for user dao layer`
  - `docs: update go-zero development guidelines`

## 3. Pull Request (PR) 流程

1.  **保持原子性**: 一个 PR 应该只做一件事。避免将多个不相关的功能或修复混合在同一个 PR 中。
2.  **代码自检**: 在提交 PR 前，请确保：
    - 代码能够成功编译 (`go build ./...`)。
    - 所有测试都已通过 (`go test ./...`)。
    - 代码遵循了我们约定的 `GO-ZERO-GUIDELINES.md`。
3.  **清晰的描述**: 你的 PR 标题和描述应该清晰地说明：
    - **What**: 你做了什么？
    - **Why**: 为什么要做这个改动？
    - **How**: （可选）你是如何实现的？
4.  **关联 Issue**: 如上文所述，使用关键词关联对应的 Issue。
5.  **Code Review**: 提交 PR 后，至少需要一名团队成员 Review 并批准后，方可合并到 `main` 分支。 