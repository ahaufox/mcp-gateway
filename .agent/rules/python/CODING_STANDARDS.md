# Coding Standards (编码规范)

## Python (Python 语言)

### Attribute Access vs Getters (属性访问与 Getter 方法)
- **Direct Access** (直接访问): Prefer accessing attributes directly (e.g., `obj.base_date`) over using getter methods (e.g., `obj.get_base_date()`). (优先直接访问属性（例如 `obj.base_date`），而不是使用 getter 方法（例如 `obj.get_base_date()`）。)
- **@property** (@property 装饰器): If validation or computation is needed in the future, use the `@property` decorator to maintain backward compatibility without changing the call site. (如果未来需要校验或计算，请使用 `@property` 装饰器，以便在不更改调用端的情况下保持向后兼容性。)
- **Rationale** (理由): This is the "Pythonic" way. It avoids premature optimization and properly utilizes Python's language features (Descriptors/Properties) to handle complexity only when necessary. (这是“Pythonic”的方式。它避免了过早优化，并在必要时利用 Python 的语言特性（描述符/属性）来处理复杂性。)
