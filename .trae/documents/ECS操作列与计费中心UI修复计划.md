# ECS 操作列与计费中心 UI 修复计划

## 摘要

用户反馈两个问题并要求全局 UI 审视：
1. **ECS 云主机页**：表格「操作」列太宽（`width: 300`），8 列固定宽度合计约 1305px + 弹性镜像列，1280px 屏幕下表格拥挤/横向滚动。
2. **计费中心页**：① 页面底部存在两个与中部完全重复的卡片（「本月用量明细」「账单流水」，上次编辑遗留的旧模板块）；② 费用构成环形图中心金额显示 `¥0.6` 而非 `¥0.64`（`toFixed(1)` 精度错误）；③ 图例数值显示 `0.3` 而非 `0.30`（未格式化小数位）。

全局扫描结论：重复块仅计费页存在（已用脚本确认所有页面卡片 title 重复情况）；其他页面操作列宽度：containers 220 / iam-users 260 / apps 200 / images 170 / networks 280，其中 networks 280 与 iam-users 260 偏宽，一并微调。

## 现状分析（基于实际代码）

### frontend/pages/ecs/index.vue
- 操作列（L104-129）：`width: 300`，渲染最多 6 个 tiny 按钮（详情/启动|停止/重启/强制停止/删除）平铺在 `NSpace` 里。
- 其余列：实例ID 170 / 名称 150 / 镜像(弹性) / 规格 140 / 状态 90 / IP 130 / 端口 160 / 创建时间 165。
- 右侧面板（状态分布环形图 + 资源用量进度条）280px，布局本身合理，不动。

### frontend/pages/billing/index.vue
- L154-187：grid 内的「本月用量明细」+（费用构成环形图 + 账单流水表格）——正常内容。
- **L190-198：底部重复的「本月用量明细」卡片（待删除）**
- **L200-202：底部重复的「账单流水（用量记录）」卡片（待删除）**
- L173：`:center-text="'¥' + totalCost.toFixed(1)"` → 0.64 显示成 0.6（待改 toFixed(2)）。

### frontend/components/DonutChart.vue
- L83：图例 `<span class="val">{{ a.value }}</span>` 直接输出原始数值，无小数位格式化。计费页传入 `Number((u.value * u.price).toFixed(2))`，如 0.3 会显示成 `0.3`。
- 组件被 9+ 个页面使用，**不能改变默认行为**，需加可选 prop。

## 修改内容

### 1. frontend/pages/ecs/index.vue — 操作列重构（核心）
**做什么**：操作列从 6 按钮平铺改为「详情 + 主操作 + 更多下拉」三段式（对齐阿里云/腾讯云控制台模式，也与本项目 containers 页 `fixed: 'right'` 的既有做法一致）。

- 引入 `NDropdown`（naive-ui）。
- 列定义改为 `width: 190, fixed: 'right'`。
- 直接渲染的按钮：
  - `详情`（primary ghost，不变）
  - 状态主操作：stopped/failed → `启动`；running → `停止`
- `更多 ▾` 下拉（`NDropdown` + options 按行渲染）：
  - running 时：`重启`、`强制停止`
  - 有 `ecs:delete` 权限时：`删除`（红色文字，点击仍走现有 `dialog.warning` 二次确认）
- `act()` 函数逻辑完全不动，仅改列渲染。
- 同步微调：创建时间列 165 → 150（够显示 `YYYY-MM-DD HH:mm`），端口列 160 → 140（配合 `ellipsis` 已有 tooltip 的镜像列）。

**为什么**：300px 操作列是"字段栏右边太长"的直接原因；收纳次要操作后表格主信息列获得约 175px 空间，1280px 屏不再挤压。

### 2. frontend/pages/billing/index.vue — 删重复块 + 图表精度
- 删除 L190-202 两个重复卡片（底部「本月用量明细」+「账单流水（用量记录）」）。
- L173 `toFixed(1)` → `toFixed(2)`，中心金额显示 `¥0.64`。
- 环形图组件调用处新增 `:value-decimals="2"`（见 3）。

### 3. frontend/components/DonutChart.vue — 图例数值格式化（可选 prop，向后兼容）
- 新增 prop `valueDecimals?: number`（默认 `-1` = 保持原样，直接 `{{ a.value }}`）。
- 图例 val 渲染：`valueDecimals >= 0` 时用 `a.value.toFixed(valueDecimals)`。
- 其他 9 个使用方不传该 prop，行为不变。

### 4. 轻量微调（同问题预防性收窄，非必须但顺手）
- frontend/pages/networks/index.vue：操作列 `width: 280` → `200`（详情/连接容器/删除 3 按钮）。
- frontend/pages/iam/users/index.vue：操作列 `width: 260` → `190`（分配角色/禁用/删除 3 按钮）。

## 假设与决策
- **不跑 `npm run build` 做验证**（上次教训：宿主机构建会覆盖容器 dev 模式的 `.nuxt/`，导致全部静态资源 404）。验证只用 SSR HTTP 请求 + 用户浏览器强刷确认。
- 环形图中心字号 `Math.max(13, size/7)`（size=130 → 18.5px）容纳 `¥0.64` 五个字符无压力，无需调字号。
- 截图中"孤立蓝色加载图标"判断为账单流水表格的 loading spinner 摄于加载瞬间 + 重复块造成的空旷感，删除重复块后消失；不为它单独加逻辑。
- hero 统计胶囊与 StatTile 信息有部分重复（虚拟余额/本月费用出现两次），属常见摘要模式，保留不改。

## 验证步骤
1. `Invoke-WebRequest http://localhost/ecs`、`/billing`、`/networks`、`/iam/users` → 全部 200（dev 容器热更新自动生效）。
2. 用户浏览器 Ctrl+F5 强刷：
   - ECS 列表：操作列收窄为「详情/停止/更多」，点「更多」可见重启/强制停止/删除，删除仍弹二次确认；
   - 计费中心：底部不再有重复的两个卡片；环形图中心显示 `¥0.64`，图例显示 `0.30 / 0.04 / 0.30`；
   - 其余使用 DonutChart 的页面（容器/镜像/网络/域名/仓库/用户/应用等）图例显示不受影响。
