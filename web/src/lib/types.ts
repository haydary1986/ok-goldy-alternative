export interface User {
  id: string;
  primary_email: string;
  given_name?: string;
  family_name?: string;
  org_unit_path?: string;
  suspended: boolean;
  is_admin: boolean;
}

export interface UsersListResponse {
  users: User[];
  next_page_token?: string;
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
