import React from 'react';
import { Button } from './ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from './ui/card';
import { LogIn, AlertCircle } from 'lucide-react';
import { useGoogleLogin, useGoogleCallback } from '@/hooks';

interface LoginProps {
  onLogin: (user: any) => void;
}

export const Login: React.FC<LoginProps> = ({ onLogin }) => {
  const [error, setError] = React.useState<string | null>(null);
  const googleLoginMutation = useGoogleLogin();
  const googleCallbackMutation = useGoogleCallback();

  const isLoading = googleLoginMutation.isPending || googleCallbackMutation.isPending;

  React.useEffect(() => {
    // Handle OAuth callbacks
    const urlParams = new URLSearchParams(window.location.search);
    const code = urlParams.get('code');
    const state = urlParams.get('state');
    
    if (code && state) {
      handleOAuthCallback(code, state);
    }
  }, []);

  const handleGoogleLogin = async () => {
    setError(null);
    
    try {
      const data = await googleLoginMutation.mutateAsync();
      
      if (data.auth_url) {
        // Store state for verification
        localStorage.setItem('oauth_state', data.state);
        window.location.href = data.auth_url;
      } else {
        throw new Error('Failed to get authentication URL');
      }
    } catch (err) {
      console.error('Login error:', err);
      setError('Failed to initiate Google login');
    }
  };

  const handleOAuthCallback = async (code: string, state: string) => {
    const storedState = localStorage.getItem('oauth_state');
    if (state !== storedState) {
      setError('Invalid authentication state');
      return;
    }

    try {
      const data = await googleCallbackMutation.mutateAsync({ code, state });
      const userData = data.user;
      
      localStorage.setItem('user', JSON.stringify(userData));
      localStorage.setItem('access_token', data.access_token);
      localStorage.removeItem('oauth_state');
      
      onLogin(userData);
      
      // Clean up URL
      window.history.replaceState({}, document.title, window.location.pathname);
    } catch (err) {
      console.error('Authentication callback error:', err);
      setError('Authentication failed');
    }
  };

  return (
    <div className="min-h-screen bg-gradient-to-br from-blue-50 to-indigo-100 flex items-center justify-center p-4">
      <Card className="w-full max-w-md">
        <CardHeader className="text-center">
          <div className="w-16 h-16 rounded-full bg-gradient-to-r from-blue-500 to-purple-600 mx-auto mb-4 flex items-center justify-center">
            <LogIn className="w-8 h-8 text-white" />
          </div>
          <CardTitle className="text-2xl">AI Code Agent</CardTitle>
          <CardDescription>
            Sign in to start building amazing projects with AI assistance
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          {error && (
            <div className="bg-red-50 border border-red-200 rounded-lg p-3 flex items-center gap-2">
              <AlertCircle className="w-4 h-4 text-red-500" />
              <span className="text-red-700 text-sm">{error}</span>
            </div>
          )}
          
          <Button 
            className="w-full" 
            onClick={handleGoogleLogin}
            disabled={isLoading}
          >
            {isLoading ? (
              <div className="flex items-center gap-2">
                <div className="w-4 h-4 border-2 border-white border-t-transparent rounded-full animate-spin" />
                Connecting...
              </div>
            ) : (
              <div className="flex items-center gap-2">
                <svg className="w-4 h-4" viewBox="0 0 24 24">
                  <path fill="currentColor" d="M22.56 12.25c0-.78-.07-1.53-.2-2.25H12v4.26h5.92c-.26 1.37-1.04 2.53-2.21 3.31v2.77h3.57c2.08-1.92 3.28-4.74 3.28-8.09z"/>
                  <path fill="currentColor" d="M12 23c2.97 0 5.46-.98 7.28-2.66l-3.57-2.77c-.98.66-2.23 1.06-3.71 1.06-2.86 0-5.29-1.93-6.16-4.53H2.18v2.84C3.99 20.53 7.7 23 12 23z"/>
                  <path fill="currentColor" d="M5.84 14.09c-.22-.66-.35-1.36-.35-2.09s.13-1.43.35-2.09V7.07H2.18C1.43 8.55 1 10.22 1 12s.43 3.45 1.18 4.93l2.85-2.22.81-.62z"/>
                  <path fill="currentColor" d="M12 5.38c1.62 0 3.06.56 4.21 1.64l3.15-3.15C17.45 2.09 14.97 1 12 1 7.7 1 3.99 3.47 2.18 7.07l3.66 2.84c.87-2.6 3.3-4.53 6.16-4.53z"/>
                </svg>
                Continue with Google
              </div>
            )}
          </Button>
          
          <div className="text-center text-sm text-gray-500">
            By continuing, you agree to our terms of service and privacy policy.
          </div>
        </CardContent>
      </Card>
    </div>
  );
};
