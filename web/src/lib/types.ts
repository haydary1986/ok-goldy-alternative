export interface User {
  id: string;
  primary_email: string;
  given_name?: string;
  family_name?: string;
  org_unit_path?: string;
  suspended: boolean;
  is_admin: boolean;
  creation_time?: string;
  last_login_time?: string;
}

export interface UsersListResponse {
  users: User[];
  next_page_token?: string;
}

export interface CreateUserRequest {
  primary_email: string;
  given_name: string;
  family_name: string;
  password: string;
  org_unit_path?: string;
}

export interface UpdateUserRequest {
  given_name?: string;
  family_name?: string;
  org_unit_path?: string;
  suspended?: boolean;
}

export interface Alias {
  alias: string;
}

export interface Group {
  id: string;
  email: string;
  name?: string;
  description?: string;
  direct_members_count?: number;
}

export interface GroupsListResponse {
  groups: Group[];
  next_page_token?: string;
}

export interface CreateGroupRequest {
  email: string;
  name?: string;
  description?: string;
}

export interface AddMemberRequest {
  email: string;
  role?: 'OWNER' | 'MANAGER' | 'MEMBER';
}

export interface MembersListResponse {
  members: Member[];
  next_page_token?: string;
}

export interface OrgUnit {
  org_unit_id?: string;
  name: string;
  org_unit_path: string;
  parent_org_unit_path?: string;
  description?: string;
}

export interface OrgUnitsListResponse {
  org_units: OrgUnit[];
}

export interface CreateOrgUnitRequest {
  name: string;
  parent_org_unit_path: string;
  description?: string;
}

export interface OUUserCount {
  org_unit_path: string;
  total: number;
  active: number;
  suspended: number;
}

export interface StatsOverview {
  generated_at: string;
  duration_ms: number;

  total_users: number;
  active_users: number;
  suspended_users: number;
  admin_users: number;
  never_logged_in: number;

  inactive_30d: number;
  inactive_90d: number;
  inactive_180d: number;
  inactive_365d: number;

  created_last_7d: number;
  created_last_30d: number;
  created_last_90d: number;

  users_by_ou: OUUserCount[];

  total_groups: number;
  empty_groups: number;
  total_group_members: number;
}

export interface InactiveListResponse {
  users: User[];
  total: number;
  days: number;
  cutoff: string;
}

export interface BulkSuspendRequest {
  user_ids: string[];
  suspended: boolean;
}

export interface BulkSuspendResult {
  user_id: string;
  ok: boolean;
  error?: string;
}

export interface BulkSuspendResponse {
  total: number;
  successful: number;
  failed: number;
  results: BulkSuspendResult[];
}

export interface AuditEntry {
  id: number;
  occurred_at: string;
  actor: string;
  action: string;
  resource_type: string;
  resource_id: string;
  request_id?: string;
  before?: unknown;
  after?: unknown;
  ok: boolean;
  error_message?: string;
}

export interface AuditListResponse {
  entries: AuditEntry[];
  total: number;
  limit: number;
  offset: number;
}

export interface Member {
  id?: string;
  email: string;
  role?: 'OWNER' | 'MANAGER' | 'MEMBER';
  type?: 'USER' | 'GROUP' | 'EXTERNAL';
}

export interface WorkspaceStatus {
  configured: boolean;
  source?: 'db' | 'env' | '';
  delegated_admin?: string;
  customer_id?: string;
  sa_email?: string;
  sa_client_id?: string;
  project_id?: string;
  updated_at?: string;
  required_scopes: string[];
}

export interface WorkspaceCredentials {
  delegated_admin: string;
  customer_id: string;
  sa_email?: string;
  sa_client_id?: string;
  project_id?: string;
  updated_at: string;
}

export interface ScopeProbe {
  scope: string;
  ok: boolean;
  error?: string;
}

export interface WorkspaceDiagnostic {
  sa_client_id?: string;
  delegated_admin?: string;
  probes: ScopeProbe[];
  summary: string;
}

export interface ScopeProbe {
  scope: string;
  ok: boolean;
  error?: string;
}

export interface DiagnosticResponse {
  sa_client_id?: string;
  delegated_admin?: string;
  probes: ScopeProbe[];
  summary: string;
}
