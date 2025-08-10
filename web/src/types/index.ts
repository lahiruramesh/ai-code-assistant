// User types
export interface User {
  id: string;
  name: string;
  email: string;
  avatar_url?: string;
  github_connected: boolean;
  vercel_connected: boolean;
  github_username?: string;
  vercel_username?: string;
}

// Project types
export interface Project {
  id: string;
  name: string;
  template: string;
  status: 'created' | 'building' | 'running' | 'stopped' | 'error';
  created_at: string;
  updated_at: string;
  docker_container?: string;
  port?: number;
  github_repo?: string;
  vercel_project?: string;
  description?: string;
}

// API Response types
export interface ApiResponse<T = any> {
  success: boolean;
  data?: T;
  error?: string;
  message?: string;
}

export interface ProjectsResponse {
  projects: Project[];
  total: number;
}

// Auth types
export interface AuthState {
  user: User | null;
  isAuthenticated: boolean;
  isLoading: boolean;
}

export interface OAuthResponse {
  auth_url: string;
  state: string;
}

export interface AuthCallbackResponse {
  user: User;
  access_token: string;
}

// Navigation types
export interface NavItem {
  title: string;
  href: string;
  icon?: React.ComponentType<{ className?: string }>;
  disabled?: boolean;
}

// Modal types
export interface ModalState {
  isOpen: boolean;
  type?: 'create-project' | 'connect-github' | 'connect-vercel';
  data?: any;
}

// Layout types
export interface LayoutProps {
  children: React.ReactNode;
  showSidebar?: boolean;
  showTopNav?: boolean;
}

// Hook types
export interface UseProjectsOptions {
  enabled?: boolean;
  refetchInterval?: number;
}

export interface UseProjectsReturn {
  projects: Project[];
  isLoading: boolean;
  error: string | null;
  refetch: () => Promise<void>;
  createProject: (projectData: CreateProjectData) => Promise<Project>;
  deleteProject: (projectId: string) => Promise<void>;
}

// Form types
export interface CreateProjectData {
  name: string;
  template: string;
  message?: string;
}

export interface ConnectIntegrationData {
  provider: 'github' | 'vercel';
  redirect_uri?: string;
}

// Template types
export interface ProjectTemplate {
  id: string;
  name: string;
  description: string;
  tags: string[];
  icon?: string;
  popular?: boolean;
}

// Chat session types
export interface CreateSessionData {
  message: string;
}

export interface CreateSessionResponse {
  project_id: string;
  session_id: string;
}

// Hook types for TanStack Query
export interface UseCreateSessionOptions {
  onSuccess?: (data: CreateSessionResponse) => void;
  onError?: (error: Error) => void;
}
