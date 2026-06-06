---
trigger: model_decision
description: 后端工作流动态参数自适应解析与单元测试异步死锁、全局覆写残留污染的防范规则。
---

# 后端工作流与测试健壮性防范规则 (Backend Workflow & Testing Safeguards)

## 1. 异步流式接口单元测试死锁防范 (StreamingResponse/SSE Test Safeguard)
### 1.1 问题场景与痛点
在使用 FastAPI `StreamingResponse` 响应流（如 Server-Sent Events 流式生成内容）的场景中，底层通常使用 `asyncio.Queue` 搭配异步生成器来消费和推送事件。
如果在单元测试中，我们 Mock 了后台负责入队的生产者方法（如 `enqueue_events`），但仅简单 Patch 成了无动作的空白 `Mock`，会导致底层的 SSE 消费生成器一直阻塞在 `await queue.get()`，令客户端的异步请求永远无法结束，引发整个 `pytest` 测试套件在执行该测试时发生进程死锁卡死。

### 1.2 强制规范与最佳实践
- **强制使用 Side Effect 模拟退出**：当 Mock 流式事件流的队列写入端时，**必须**配置一个有副作用 of Mock 动作（使用 `side_effect`）。该动作被调用时，需要向事件分发器发送结束符，或显式触发关联流的关闭操作（如调用 `WorkflowRunManager.complete()`，往队列中写入 `None` 作为结束符等）。
- **必须配合超时退出机制**：为相关的流式响应消费测试用例配置显式的超时断言或超时管理器，防止意外死锁发生时阻塞 CI。

---

## 2. FastAPI `dependency_overrides` 环境隔离与污染防范 (Dependency Override Safeguard)
### 2.1 问题场景与痛点
为了在单元测试中 Mock 鉴权用户、特定领域服务或系统配置，我们需要修改全局 FastAPI 实例的 `app.dependency_overrides`。
如果在**模块导入级别（即文件最外层）**直接改写全局 `app.dependency_overrides`，或者在测试用例的类周期中进行修改但未做妥善清理，该覆盖会常驻内存，直接污染之后执行的其他单元测试。
当运行全量 pytest 测试套件时，这会导致：
- 跨事件循环（Cross-loop）数据库连接池抢占冲突与报错。
- 其他测试类因覆盖残留直接报 401 越权或 403 权限拒绝。
- 业务依赖残留导致其他用例获取到脏的 Mock 对象。

### 2.2 强制规范与最佳实践
- **严禁模块导入级覆盖**：严禁在测试文件模块导入（Import）级别直接覆写全局 `app.dependency_overrides`。
- **生命周期动态覆盖与彻底清理**：
  1. 必须在测试用例类的 `setUp` / `asyncSetUp` 钩子或局部 Context 中配置 `app.dependency_overrides`。
  2. **最核心的要求**：必须在用例执行完毕后的 `tearDown` / `asyncTearDown` 钩子中显式调用 `app.dependency_overrides.clear()` 清空所有覆盖。
- **Pytest Fixture 最佳实践**：如果使用 pytest 的 fixture 进行依赖覆盖，必须将其作用域限制为函数级（`scope="function"`），并使用 `yield` 关键字在 fixture 退出时执行清除动作：
  ```python
  @pytest.fixture(autouse=True)
  def mock_dependencies():
      app.dependency_overrides[get_current_user] = override_user
      yield
      app.dependency_overrides.clear()
  ```

---

## 3. 工作流去硬编码与自适应反射健壮性规范 (Workflow Robustness & Decoupling Standard)
### 3.1 问题场景与痛点
在工作流节点的输入/输出去硬编码重构中，当把硬编码的字面量提取逻辑重构为基于 Schema 反射（探测 context 中的数据结构特征）来动态组装参数时：
- **数据冲突与覆盖风险**：在合并 Context 并展平（如 `flatten_ctx_data`）时，粗暴地遍历 `ctx_data.values()` 并执行 `flat.update(val)` 会导致在输入数据同构（例如传入了多个 dict，或多个 string 作为主内容）时，后遍历到的数据会完全覆写前面的关键信息，导致数据流失。
- **空配置与越界 Crash**：若节点的 input/output 没有定义，或者在读取输出键时裸用了越界索引（如直接取 `self.node_def.outputs[0]` 而未判空），在执行空流程或配置变更的流程时会导致系统直接抛出 `IndexError` 或 `KeyError` 从而瘫痪。

### 3.2 强制规范与最佳实践
- **参数分配优先级规则**：动态参数分配必须严格遵守以下优先级路径：
  1. **配置绑定优先**：优先基于 `self.node_def.inputs` 中声明的指定键名去 `ctx_data` 提取数据。
  2. **特征匹配降级**：当配置键找不到时，作为降级 fallback，根据数据的 Schema 结构特征去 `ctx_data.values()` 中探测提取。
  3. **全局默认兜底**：若上述均未提取到，使用 `state["global_ctx"]` 内的默认字段（如 `raw_input`）进行底线兜底。
- **展平与合并时的防覆盖隔离**：在扁平化 Context 数据时，对于核心公共字段（例如主案情文本 `chat_content`），应当在反射处理前或后做显式分配和逻辑锁定，禁止用 `flat.update()` 任意覆盖已经填充的主参数。
- **输出索引安全防御**：严禁在不确定 outputs 数量时直接使用 `self.node_def.outputs[0]`。在映射输出前，必须确保做长度校验或提供默认值备用：
  ```python
  output_key = self.node_def.outputs[0] if self.node_def.outputs else "default_node_output"
  ```
