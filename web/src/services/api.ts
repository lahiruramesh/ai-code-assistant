import type { Project, CreateProjectData, User} from '@/types';

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL;

// Base API client
class ApiClient {
  private baseUrl: string;

  constructor(baseUrl: string) {
    this.baseUrl = baseUrl;
  }

  private getAuthHeaders(): HeadersInit {
    const token = localStorage.getItem('access_token');
    return {
      'Content-Type': 'application/json',
      ...(token && { 'Authorization': `Bearer ${token}` }),
    };
  }

  async request<T>(
    endpoint: string,
    options: RequestInit = {}
  ): Promise<T> {
    const url = `${this.baseUrl}${endpoint}`;
    
    const config: RequestInit = {
      ...options,
      headers: {
        ...this.getAuthHeaders(),
        ...options.headers,
      },
    };

    const response = await fetch(url, config);

    if (!response.ok) {
      let errorMessage = `HTTP ${response.status}: Request failed`;
      try {
        const errorData = await response.json();
        errorMessage = errorData.error || errorData.message || errorMessage;
      } catch {
        // If we can't parse error response, use the default message
      }
      throw new Error(errorMessage);
    }

    return response.json();
  }

  async get<T>(endpoint: string): Promise<T> {
    return this.request<T>(endpoint, { method: 'GET' });
  }

  async post<T>(endpoint: string, data?: any): Promise<T> {
    return this.request<T>(endpoint, {
      method: 'POST',
      body: data ? JSON.stringify(data) : undefined,
    });
  }

  async put<T>(endpoint: string, data?: any): Promise<T> {
    return this.request<T>(endpoint, {
      method: 'PUT',
      body: data ? JSON.stringify(data) : undefined,
    });
  }

  async delete<T>(endpoint: string): Promise<T> {
    return this.request<T>(endpoint, { method: 'DELETE' });
  }
}

// Create the API client instance
export const apiClient = new ApiClient(API_BASE_URL);

// API endpoints
export const chatApi = {
  createSession: (message: string) =>
    apiClient.post<{ project_id: string; session_id: string }>('/chat/create-session', { message }),
  getMessages: (projectId: string) =>
    apiClient.get<{ messages: any[] }>(`/projects/${projectId}/messages`),
  getChatById: (chatId: string) =>
    apiClient.get<any>(`/chat/${chatId}`),
  cancelSession: (sessionId: string) =>
    apiClient.post<void>(`/chat/${sessionId}/cancel`),
};

export const projectsApi = {
  getAll: () => apiClient.get<{ projects: Project[] }>('/projects'),
  getById: (id: string) => apiClient.get<Project>(`/projects/${id}`),
  create: (data: CreateProjectData) => apiClient.post<Project>('/projects/', data),
  delete: (id: string) => apiClient.delete<{ message: string; cleanup_result: any }>(`/projects/${id}`),
  deployToVercel: (id: string) => 
    apiClient.post<{ deployment_url: string; url: string }>(`/projects/${id}/deploy/vercel`),
  pushToGitHub: (id: string, repoName?: string) =>
    apiClient.post<{ repo_url: string; url: string }>(`/projects/${id}/github/push`, { repo_name: repoName }),
  getPreview: (projectName: string) =>
    apiClient.get<{ preview_url: string; host_path: string }>(`/projects/${projectName}/preview`),
  getFiles: (projectName: string, source = 'aiagent') =>
    apiClient.get<{ files: any[] }>(`/projects/${projectName}/files?source=${source}`),
  getFileContent: (projectName: string, filePath: string, source = 'aiagent') =>
    apiClient.get<{ file_path: string; content: string }>(`/projects/${projectName}/files/${filePath}?source=${source}`),
};

export const authApi = {
  googleLogin: () => apiClient.get<{ auth_url: string; state: string }>('/auth/google/login'),
  googleCallback: (code: string, state: string) =>
    apiClient.post<{ user: User; access_token: string }>('/auth/google/callback', { code, state }),
  logout: (userId: string) => apiClient.post<void>('/auth/logout', { user_id: userId }),
  connectGitHub: () => apiClient.get<{ auth_url: string; state: string }>('/auth/github/connect'),
  connectVercel: () => apiClient.get<{ auth_url: string; state: string }>('/auth/vercel/connect'),
  disconnectIntegration: (provider: 'github' | 'vercel') =>
    apiClient.post<void>(`/auth/${provider}/disconnect`),
};

export const modelsApi = {
  getAll: () => apiClient.get<{ models: any }>('/models/all'),
  getProviderInfo: () => apiClient.get<{ current_provider: string }>('/models'),
};

export const tokensApi = {
  getSessionUsage: (sessionId: string) =>
    apiClient.get<any>(`/tokens/usage/${sessionId}`),
  getGlobalStats: () => apiClient.get<any>('/tokens/stats'),
};

// Re-export types for convenience
export type { Project, CreateProjectData, User, OAuthResponse } from '@/types';
