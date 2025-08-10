import { useQuery, useMutation } from '@tanstack/react-query';
import { authApi } from '@/services/api';
import type { User } from '@/types';

export const useGoogleLogin = () => {
  return useMutation({
    mutationFn: () => authApi.googleLogin(),
  });
};

export const useGoogleCallback = () => {
  return useMutation({
    mutationFn: ({ code, state }: { code: string; state: string }) => 
      authApi.googleCallback(code, state),
  });
};

export const useLogout = () => {
  return useMutation({
    mutationFn: (userId: string) => authApi.logout(userId),
  });
};

export const useAuth = () => {
  const googleLoginMutation = useGoogleLogin();
  const googleCallbackMutation = useGoogleCallback();
  const logoutMutation = useLogout();

  const initiateGoogleLogin = async () => {
    const data = await googleLoginMutation.mutateAsync();
    if (data.auth_url) {
      localStorage.setItem('oauth_state', data.state);
      window.location.href = data.auth_url;
    } else {
      throw new Error('Failed to get authentication URL');
    }
  };

  const handleGoogleCallback = async (code: string, state: string) => {
    const storedState = localStorage.getItem('oauth_state');
    if (state !== storedState) {
      throw new Error('Invalid authentication state');
    }

    const data = await googleCallbackMutation.mutateAsync({ code, state });
    localStorage.setItem('user', JSON.stringify(data.user));
    localStorage.setItem('access_token', data.access_token);
    localStorage.removeItem('oauth_state');
    
    return data.user;
  };

  const logout = async (userId: string) => {
    await logoutMutation.mutateAsync(userId);
    localStorage.removeItem('user');
    localStorage.removeItem('access_token');
  };

  return {
    initiateGoogleLogin,
    handleGoogleCallback,
    logout,
    isLoading: googleLoginMutation.isPending || googleCallbackMutation.isPending || logoutMutation.isPending,
    error: googleLoginMutation.error || googleCallbackMutation.error || logoutMutation.error,
  };
};
