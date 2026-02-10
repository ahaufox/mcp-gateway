---
trigger: always_on
description: 接口鉴权强制校验规范
---

# 鉴权强制校验规范 (Strict Auth Check)

## 1. 强制注入
- 所有的写操作接口 (`POST`, `PATCH`, `DELETE`) 以及敏感读接口必须使用鉴权依赖注入。
- 示例 (FastAPI):
  ```python
  @router.post("/items")
  async def create_item(
      item: ItemCreate,
      current_user: User = Depends(get_current_active_user)
  ):
      ...
  ```

## 2. 权限颗粒度
- 涉及用户私有数据时，必须校验 `data.user_id == current_user.id`。
- 管理后台接口必须包含 `require_superuser` 或 `require_role('admin')`。

## 3. 禁止项
- 禁止在接口中手动解析 `Authorization` Header，必须统一使用 `Depends`。
- 免鉴权接口必须由 `auth-white-list` 显式允许。
