import React, { createContext, useContext, useEffect, useState } from 'react';
import { Login } from '@/components/Login';
import { useLogout } from '@/hooks';
import type { User } from '@/types';

interface AuthContextType {
    user: User | null;
    isLoading: boolean;
    login: (user: User) => void;
    logout: () => Promise<void>;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

export const useAuth = () => {
    const context = useContext(AuthContext);
    if (context === undefined) {
        throw new Error('useAuth must be used within an AuthProvider');
    }
    return context;
};

interface AuthProviderProps {
    children: React.ReactNode;
}

export const AuthProvider: React.FC<AuthProviderProps> = ({ children }) => {
    const [user, setUser] = useState<User | null>(null);
    const [isLoading, setIsLoading] = useState(true);
    const logoutMutation = useLogout();

    useEffect(() => {
        const initAuth = async () => {
            try {
                // Check localStorage for existing user session
                const storedUser = localStorage.getItem('user');
                if (storedUser) {
                    const parsedUser = JSON.parse(storedUser);
                    setUser(parsedUser);
                }
            } catch (error) {
                console.error('Auth initialization error:', error);
                localStorage.removeItem('user');
                setUser(null);
            } finally {
                setIsLoading(false);
            }
        };

        initAuth();
    }, []);

    const login = (userData: User) => {
        setUser(userData);
        localStorage.setItem('user', JSON.stringify(userData));
    };

    const logout = async () => {
        try {
            // Call backend logout if user exists
            if (user?.id) {
                await logoutMutation.mutateAsync(user.id);
            }

            setUser(null);
            localStorage.removeItem('user');
            localStorage.removeItem('access_token');
        } catch (error) {
            console.error('Logout failed:', error);
            // Clear local state even if backend call fails
            setUser(null);
            localStorage.removeItem('user');
            localStorage.removeItem('access_token');
        }
    };

    const value = {
        user,
        isLoading,
        login,
        logout,
    };

    if (isLoading) {
        return (
            <div className="min-h-screen bg-gradient-to-br from-blue-50 to-indigo-100 flex items-center justify-center">
                <div className="text-center space-y-4">
                    <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary mx-auto"></div>
                    <p className="text-muted-foreground">Loading...</p>
                </div>
            </div>
        );
    }

    if (!user) {
        return <Login onLogin={login} />;
    }

    return (
        <AuthContext.Provider value={value}>
            {children}
        </AuthContext.Provider>
    );
};
