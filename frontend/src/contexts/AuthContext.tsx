import React, { createContext, useContext, useState, ReactNode, useEffect } from 'react';
import { Login } from '../../wailsjs/go/app/App';

interface Role {
    id: number;
    name: string;
    description: string;
}

interface User {
    id: number;
    email: string;
    device_id: string;
    role: Role;
}

interface LoginData {
    token: string;
    expires_at: string;
    user: User;
}

interface LoginResponse {
    data: LoginData;
    message: string;
}

interface AuthContextType {
    user: User | null;
    token: string | null;
    login: (email: string, password: string) => Promise<{ success: boolean; message: string }>;
    logout: () => void;
    isAuthenticated: boolean;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

const TOKEN_KEY = 'auth_token';
const USER_KEY = 'auth_user';

export const AuthProvider: React.FC<{ children: ReactNode }> = ({ children }) => {
    const [user, setUser] = useState<User | null>(null);
    const [token, setToken] = useState<string | null>(null);

    // Load token and user from localStorage on mount
    useEffect(() => {
        const storedToken = localStorage.getItem(TOKEN_KEY);
        const storedUser = localStorage.getItem(USER_KEY);
        if (storedToken && storedUser) {
            setToken(storedToken);
            setUser(JSON.parse(storedUser));
        }
    }, []);

    const login = async (email: string, password: string): Promise<{ success: boolean; message: string }> => {
        try {
            const response: LoginResponse = await Login(email, password);

            if (response.data && response.data.token) {
                setToken(response.data.token);
                setUser(response.data.user);

                // Store in localStorage
                localStorage.setItem(TOKEN_KEY, response.data.token);
                localStorage.setItem(USER_KEY, JSON.stringify(response.data.user));

                return { success: true, message: response.message };
            }

            return { success: false, message: response.message || 'Login failed' };
        } catch (error) {
            console.error('Login error:', error);
            return { success: false, message: 'Login failed. Please try again.' };
        }
    };

    const logout = () => {
        setUser(null);
        setToken(null);
        localStorage.removeItem(TOKEN_KEY);
        localStorage.removeItem(USER_KEY);
    };

    return (
        <AuthContext.Provider value={{ user, token, login, logout, isAuthenticated: user !== null }}>
            {children}
        </AuthContext.Provider>
    );
};

export const useAuth = (): AuthContextType => {
    const context = useContext(AuthContext);
    if (!context) {
        throw new Error('useAuth must be used within an AuthProvider');
    }
    return context;
};
