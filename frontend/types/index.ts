export interface ApiResponse<T = unknown> {
  code: number
  message: string
  data: T
  request_id: string
}

export interface DependencyStatus {
  status: string
  error?: string
}

export interface HealthData {
  status: string
  db: DependencyStatus
  redis: DependencyStatus
  time: string
}

export interface TokenPair {
  access_token: string
  refresh_token: string
  expires_in: number
  token_type: string
}

export interface UserInfo {
  id: number
  username: string
  email: string
  nickname: string
  avatar_url: string
  status: number
  roles: string[]
  role_names: string[]
  permissions: string[]
}

export interface UserRow {
  id: number
  username: string
  email: string
  nickname: string
  avatar_url: string
  status: number
  roles: string[]
  created_at: string
}

export interface RoleItem {
  id: number
  code: string
  name: string
  description: string
  is_system: boolean
  scope: string
  permissions: string[]
}

export interface PermissionItem {
  id: number
  code: string
  name: string
  module: string
  description: string
}

export interface PageResult<T = unknown> {
  total: number
  items: T[]
}

// ---------- ECS ----------

export interface PortMapping {
  container_port: number
  host_port: number
  protocol: string
}

export interface EcsInstance {
  id: number
  instance_no: string
  owner_id: number
  name: string
  description: string
  image: string
  cpu: number
  memory_mb: number
  disk_gb: number
  ports: PortMapping[]
  env: string[]
  command: string[]
  restart_policy: string
  readonly_rootfs: boolean
  network_id: string
  fixed_ip: string
  mounts: MountInfo[]
  desired_state: string
  observed_state: string
  container_id: string
  container_name: string
  last_error: string
  created_at: string
}

export interface MountInfo {
  volume_name: string
  target: string
  read_only: boolean
}

// ---------- 镜像 / 网络 / 存储 / Registry ----------

export interface DockerImage {
  id: number
  repo: string
  tag: string
  image_id: string
  size_bytes: number
  docker_created_at: string | null
  source: string
  status: string
  pull_error: string
  created_at: string
}

export interface CloudNetwork {
  id: number
  name: string
  docker_name: string
  docker_network_id: string
  driver: string
  subnet: string
  gateway: string
  ip_range: string
  internal: boolean
  owner_id: number
  created_at: string
}

export interface CloudVolume {
  id: number
  name: string
  docker_name: string
  driver: string
  mountpoint: string
  capacity_gb: number
  used_mb: number
  owner_id: number
  created_at: string
}

export interface RegistryItem {
  id: number
  name: string
  url: string
  type: string
  status: number
}

export interface RegistryRepo {
  name: string
  tags: string[]
}

// ---------- 项目 / 应用 / 部署 / 域名 ----------

export interface Project {
  id: number
  org_id: number
  name: string
  code: string
  description: string
  status: number
  created_at: string
}

export interface ProjectEnv {
  id: number
  project_id: number
  name: string
  seq: number
}

export interface Application {
  id: number
  project_id: number | null
  owner_id: number
  name: string
  type: string
  image: string
  git_url: string
  git_branch: string
  port: number
  health_check_path: string
  env: string
  domain: string
  active_deployment_id: number | null
  status: number
  created_at: string
}

export interface AppVersion {
  id: number
  application_id: number
  version: string
  image_ref: string
  commit_sha: string
  status: string
  created_at: string
}

export interface Deployment {
  id: number
  application_id: number
  version: string
  image_ref: string
  strategy: string
  status: string
  health_status: string
  trigger: string
  container_id: string
  container_name: string
  note: string
  started_at: string | null
  finished_at: string | null
  created_at: string
}

export interface DomainItem {
  id: number
  domain: string
  application_id: number | null
  target_port: number
  tls: boolean
  status: number
  created_at: string
}

export const DeployStatusNames: Record<string, string> = {
  pending: '等待中',
  deploying: '部署中',
  success: '成功',
  failed: '失败',
  'rolled-back': '已回滚',
}

export const DeployStatusType: Record<string, 'success' | 'default' | 'warning' | 'error'> = {
  pending: 'warning',
  deploying: 'warning',
  success: 'success',
  failed: 'error',
  'rolled-back': 'default',
}

// ---------- Pipeline ----------

export interface Pipeline {
  id: number
  name: string
  description: string
  definition: string
  status: number
  created_at: string
}

export interface PipelineRun {
  id: number
  pipeline_id: number
  run_no: number
  trigger: string
  ref: string
  commit_sha: string
  status: string
  started_at: string | null
  finished_at: string | null
  duration_ms: number
  triggered_by: number | null
  created_at: string
}

export interface PipelineJob {
  id: number
  pipeline_run_id: number
  step_id: number
  name: string
  type: string
  status: string
  exit_code: number
  container_id: string
  log_path: string
  started_at: string | null
  finished_at: string | null
  created_at: string
}

export const PipeStatusNames: Record<string, string> = {
  pending: '等待中',
  running: '运行中',
  success: '成功',
  failed: '失败',
  canceled: '已取消',
}

export const PipeStatusType: Record<string, 'success' | 'default' | 'warning' | 'error' | 'info'> = {
  pending: 'warning',
  running: 'info',
  success: 'success',
  failed: 'error',
  canceled: 'default',
}

export const JobStatusNames: Record<string, string> = {
  pending: '等待中',
  running: '运行中',
  success: '成功',
  failed: '失败',
  skipped: '已跳过',
}

export const JobStatusType: Record<string, 'success' | 'default' | 'warning' | 'error' | 'info'> = {
  pending: 'default',
  running: 'info',
  success: 'success',
  failed: 'error',
  skipped: 'warning',
}

// ---------- Webhook ----------

export interface WebhookItem {
  id: number
  pipeline_id: number
  provider: string
  branch_filter: string
  events: string
  status: number
  hook_code: string
  created_at: string
}

export interface WebhookCreated {
  id: number
  hook_code: string
  provider: string
  url: string
  secret: string
}

export interface EcsEvent {
  id: number
  instance_id: number
  event_type: string
  level: string
  message: string
  actor_id: number | null
  request_id: string
  created_at: string
}

export interface EcsStats {
  cpu_percent: number
  mem_used: number
  mem_limit: number
  mem_percent: number
  net_rx: number
  net_tx: number
  disk_read: number
  disk_write: number
  pids: number
}

export interface EcsLogs {
  logs: string
  tail: number
}

export const EcsStateNames: Record<string, string> = {
  creating: '创建中',
  starting: '启动中',
  running: '运行中',
  stopping: '停止中',
  stopped: '已停止',
  restarting: '重启中',
  deleting: '删除中',
  failed: '失败',
  unknown: '未知',
}

export const EcsStateType: Record<string, 'success' | 'default' | 'warning' | 'error'> = {
  running: 'success',
  stopped: 'default',
  creating: 'warning',
  starting: 'warning',
  stopping: 'warning',
  restarting: 'warning',
  deleting: 'warning',
  failed: 'error',
  unknown: 'error',
}

// 与后端 pkg/errcode 对齐
export const ErrCode = {
  Success: 0,
  BadRequest: 40000,
  Forbidden: 40001,
  Conflict: 40009,
  TooManyRequests: 42900,
  Unauthorized: 40100,
  NotFound: 40400,
  NotImplemented: 40401,
  Internal: 50000,
} as const

// ---------- 组织 / 配额 / 计费（Phase 10 Multi-Tenant） ----------

export interface Organization {
  id: number
  name: string
  code: string
  plan: string
  credit: number
  status: number
  created_by: number | null
  created_at: string
}

export interface OrgMember {
  id: number
  org_id: number
  user_id: number
  org_role: string
  status: number
  created_at: string
}

export interface ResourceQuota {
  id: number
  org_id: number
  project_id: number | null
  resource_type: string
  limit_value: number
  updated_at: string
}

export interface ResourceUsage {
  id: number
  org_id: number
  project_id: number | null
  resource_type: string
  used_value: number
  period: string
  created_at: string
}

export interface BillingSummary {
  credit: number
  usage_month: Record<string, number>
  price: Record<string, number>
}

export const QuotaTypeNames: Record<string, string> = {
  ecs_count: '云主机实例数（台）',
  cpu: 'CPU（核）',
  memory: '内存（MB）',
  storage: '存储（GB）',
  network: '网络（个）',
  pipeline: '流水线（条）',
}

export const UsageTypeNames: Record<string, string> = {
  cpu_hour: 'CPU 核时',
  mem_gb_hour: '内存 GB·时',
  disk_gb_hour: '磁盘 GB·时',
}

export const OrgRoleNames: Record<string, string> = {
  owner: '所有者',
  admin: '管理员',
  member: '成员',
}

export const PlanNames: Record<string, string> = {
  free: '免费版',
  pro: '专业版',
  enterprise: '企业版',
}

// ---------- 安全中心 / 密钥托管（Phase 11） ----------

export interface SecurityFinding {
  severity: 'high' | 'medium' | 'low' | 'info'
  kind: 'baseline' | 'image'
  target: string
  message: string
}

export interface SecurityReportItem {
  id: number
  kind: string
  score: number
  finding_count: number
  created_at: string
}

export interface SecurityReportBlock {
  kind: string
  score: number
  finding_count: number
  findings: SecurityFinding[]
  scanned_at: string
}

export interface SecurityDashboard {
  score: number
  finding_count: number
  reports: SecurityReportBlock[]
  baseline_rules: string[]
  image_rules: string[]
}

export interface SecretItem {
  id: number
  org_id: number
  name: string
  created_by: number | null
  created_at: string
  updated_at: string
}

export const SeverityType: Record<string, 'error' | 'warning' | 'default' | 'info'> = {
  high: 'error',
  medium: 'warning',
  low: 'default',
  info: 'info',
}

export const SeverityNames: Record<string, string> = {
  high: '高危',
  medium: '中危',
  low: '低危',
  info: '提示',
}
