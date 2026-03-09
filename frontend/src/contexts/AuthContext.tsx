import React, { createContext, useContext, useState, ReactNode, useEffect } from 'react';
import { Login, Logout, ConnectMQTT } from '../../wailsjs/go/app/App';
import { EventsOn } from '../../wailsjs/runtime/runtime';
import { useNavigate } from 'react-router-dom';

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
    logout: () => Promise<void>;
    isAuthenticated: boolean;
    deepLinkLoading: boolean;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

const TOKEN_KEY = 'auth_token';
const USER_KEY = 'auth_user';

export const AuthProvider: React.FC<{ children: ReactNode }> = ({ children }) => {
    const [user, setUser] = useState<User | null>(null);
    const [token, setToken] = useState<string | null>(null);
    const [deepLinkLoading, setDeepLinkLoading] = useState(false);
    const navigate = useNavigate();

    const clearAuthState = () => {
        setUser(null);
        setToken(null);
        localStorage.removeItem(TOKEN_KEY);
        localStorage.removeItem(USER_KEY);
    };

    // Load token and user from localStorage on mount
    useEffect(() => {
        const storedToken = localStorage.getItem(TOKEN_KEY);
        const storedUser = localStorage.getItem(USER_KEY);
        if (storedToken && storedUser) {
            setToken(storedToken);
            setUser(JSON.parse(storedUser));

            // Reconnect MQTT if user session exists
            console.log('Existing session detected, connecting MQTT...');
            ConnectMQTT()
                .then((response) => {
                    console.log('MQTT connection result:', response);
                })
                .catch((error) => {
                    console.error('Failed to connect MQTT on app reload:', error);
                });
        }

        const unsubLoading = EventsOn('deep-link-auth-loading', () => {
            setDeepLinkLoading(true);
        });

        const unsubSuccess = EventsOn('deep-link-auth-success', (data: { message: string; user: User; token: string }) => {
            console.log('Deep link auth success:', data);
            setDeepLinkLoading(false);
            if (data.user && data.token) {
                setUser(data.user);
                setToken(data.token);
                localStorage.setItem(TOKEN_KEY, data.token);
                localStorage.setItem(USER_KEY, JSON.stringify(data.user));

                // Connect MQTT after deep link auth
                ConnectMQTT()
                    .then((response) => {
                        console.log('MQTT connected after deep link auth:', response);
                    })
                    .catch((error) => {
                        console.error('Failed to connect MQTT after deep link:', error);
                    });
            }
        });

        const unsubError = EventsOn('deep-link-auth-error', (data: { message: string }) => {
            console.error('Deep link auth error:', data);
            setDeepLinkLoading(false);
        });

        const unsubLogout = EventsOn('auth-logout', (data: { success: boolean; message: string }) => {
            console.log('Auth logout event received:', data);
            setDeepLinkLoading(false);
            clearAuthState();
            navigate('/login', { replace: true });
        });

        return () => {
            unsubLoading();
            unsubSuccess();
            unsubError();
            unsubLogout();
        };
    }, [navigate]);

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

    const logout = async () => {
        try {
            // Call backend logout with token
            if (token) {
                await Logout(token);
                console.log('Backend logout successful');
            }
        } catch (error) {
            console.error('Backend logout error:', error);
        } finally {
            // Clear frontend state regardless of backend result
            clearAuthState();
        }
    };

    return (
        <AuthContext.Provider value={{ user, token, login, logout, isAuthenticated: user !== null, deepLinkLoading }}>
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
