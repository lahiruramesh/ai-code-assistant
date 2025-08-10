import { useMutation } from '@tanstack/react-query';
import { authApi, projectsApi } from '@/services/api';
import type { User } from '@/types';

export const useIntegrations = () => {
  // Mutation for connecting GitHub
  const connectGitHubMutation = useMutation({
    mutationFn: () => authApi.connectGitHub(),
    onSuccess: (data) => {
      if (data.auth_url) {
        localStorage.setItem('oauth_state', data.state);
        window.location.href = data.auth_url;
      } else {
        throw new Error('Failed to get GitHub authentication URL');
      }
    },
  });

  // Mutation for connecting Vercel
  const connectVercelMutation = useMutation({
    mutationFn: () => authApi.connectVercel(),
    onSuccess: (data) => {
      if (data.auth_url) {
        localStorage.setItem('oauth_state', data.state);
        window.location.href = data.auth_url;
      } else {
        throw new Error('Failed to get Vercel authentication URL');
      }
    },
  });

  // Mutation for disconnecting integrations
  const disconnectIntegrationMutation = useMutation({
    mutationFn: (provider: 'github' | 'vercel') => authApi.disconnectIntegration(provider),
    onSuccess: (_, provider) => {
      // Update user data in localStorage
      const userData = localStorage.getItem('user');
      if (userData) {
        const user: User = JSON.parse(userData);
        const updatedUser = {
          ...user,
          [`${provider}_connected`]: false,
          [`${provider}_username`]: undefined,
        };
        localStorage.setItem('user', JSON.stringify(updatedUser));
      }
    },
  });

  // Mutation for deploying to Vercel
  const deployToVercelMutation = useMutation({
    mutationFn: (projectId: string) => projectsApi.deployToVercel(projectId),
  });

  // Mutation for pushing to GitHub
  const pushToGitHubMutation = useMutation({
    mutationFn: ({ projectId, repoName }: { projectId: string; repoName?: string }) => 
      projectsApi.pushToGitHub(projectId, repoName),
  });

  const connectGitHub = async (): Promise<void> => {
    await connectGitHubMutation.mutateAsync();
  };

  const connectVercel = async (): Promise<void> => {
    await connectVercelMutation.mutateAsync();
  };

  const disconnectIntegration = async (provider: 'github' | 'vercel'): Promise<void> => {
    await disconnectIntegrationMutation.mutateAsync(provider);
  };

  const deployToVercel = async (projectId: string): Promise<string> => {
    const data = await deployToVercelMutation.mutateAsync(projectId);
    return data.deployment_url || data.url;
  };

  const pushToGitHub = async (projectId: string, repoName?: string): Promise<string> => {
    const data = await pushToGitHubMutation.mutateAsync({ projectId, repoName });
    return data.repo_url || data.url;
  };

  const isLoading = 
    connectGitHubMutation.isPending ||
    connectVercelMutation.isPending ||
    disconnectIntegrationMutation.isPending ||
    deployToVercelMutation.isPending ||
    pushToGitHubMutation.isPending;

  const error = 
    connectGitHubMutation.error?.message ||
    connectVercelMutation.error?.message ||
    disconnectIntegrationMutation.error?.message ||
    deployToVercelMutation.error?.message ||
    pushToGitHubMutation.error?.message ||
    null;

  return {
    isLoading,
    error,
    connectGitHub,
    connectVercel,
    disconnectIntegration,
    deployToVercel,
    pushToGitHub,
  };
};
