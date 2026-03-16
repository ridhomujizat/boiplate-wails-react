import React, { createContext, useContext, useState, ReactNode, useEffect, useRef, useCallback } from 'react';
import { Login, Logout, ConnectMQTT, AuthMe } from '../../wailsjs/go/app/App';
import { EventsOn } from '../../wailsjs/runtime/runtime';

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

interface AuthMeResponse {
    statusCode: number;
    success: boolean;
    message: string;
    user?: User | null;
}

interface AuthContextType {
    user: User | null;
    token: string | null;
    login: (email: string, password: string) => Promise<{ success: boolean; message: string }>;
    logout: () => Promise<void>;
    isAuthenticated: boolean;
    isReady: boolean;
    deepLinkLoading: boolean;
    validateSession: () => Promise<boolean>;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

const TOKEN_KEY = 'auth_token';
const USER_KEY = 'auth_user';

export const AuthProvider: React.FC<{ children: ReactNode }> = ({ children }) => {
    const [user, setUser] = useState<User | null>(null);
    const [token, setToken] = useState<string | null>(null);
    const [isReady, setIsReady] = useState(false);
    const [deepLinkLoading, setDeepLinkLoading] = useState(false);
    const tokenRef = useRef<string | null>(null);
    const validationPromiseRef = useRef<Promise<boolean> | null>(null);

    const clearSession = useCallback(() => {
        tokenRef.current = null;
        setUser(null);
        setToken(null);
        localStorage.removeItem(TOKEN_KEY);
        localStorage.removeItem(USER_KEY);
        setIsReady(true);
    }, []);

    const persistSession = useCallback((nextToken: string, nextUser: User) => {
        tokenRef.current = nextToken;
        setToken(nextToken);
        setUser(nextUser);
        localStorage.setItem(TOKEN_KEY, nextToken);
        localStorage.setItem(USER_KEY, JSON.stringify(nextUser));
        setIsReady(true);
    }, []);

    const logout = useCallback(async () => {
        const currentToken = tokenRef.current;

        try {
            if (currentToken) {
                await Logout(currentToken);
                console.log('Backend logout successful');
            }
        } catch (error) {
            console.error('Backend logout error:', error);
        } finally {
            clearSession();
        }
    }, [clearSession]);

    const validateSession = useCallback(async (): Promise<boolean> => {
        const currentToken = tokenRef.current ?? localStorage.getItem(TOKEN_KEY);

        if (!currentToken) {
            setIsReady(true);
            return false;
        }

        if (validationPromiseRef.current) {
            return validationPromiseRef.current;
        }

        const validationPromise = (async () => {
            try {
                const response: AuthMeResponse = await AuthMe(currentToken);

                if (response.statusCode === 401) {
                    clearSession();
                    return false;
                }

                if (!response.success) {
                    console.error('Session validation failed with status:', response.statusCode, response.message);
                    setIsReady(true);
                    return true;
                }

                const nextUser = response.user ?? null;

                if (nextUser) {
                    setUser(nextUser);
                    localStorage.setItem(USER_KEY, JSON.stringify(nextUser));
                }

                setIsReady(true);
                return true;
            } catch (error) {
                console.error('Session validation error:', error);
                setIsReady(true);
                return true;
            } finally {
                validationPromiseRef.current = null;
            }
        })();

        validationPromiseRef.current = validationPromise;
        return validationPromise;
    }, [clearSession]);

    // Load token and user from localStorage on mount
    useEffect(() => {
        const storedToken = localStorage.getItem(TOKEN_KEY);
        const storedUser = localStorage.getItem(USER_KEY);

        if (storedToken) {
            tokenRef.current = storedToken;
            setToken(storedToken);

            if (storedUser) {
                try {
                    setUser(JSON.parse(storedUser));
                } catch (error) {
                    console.error('Failed to parse stored user:', error);
                    localStorage.removeItem(USER_KEY);
                }
            }

            // Reconnect MQTT if user session exists
            console.log('Existing session detected, connecting MQTT...');
            ConnectMQTT()
                .then((response) => {
                    console.log('MQTT connection result:', response);
                })
                .catch((error) => {
                    console.error('Failed to connect MQTT on app reload:', error);
                });
        } else {
            setIsReady(true);
        }

        const unsubLoading = EventsOn('deep-link-auth-loading', () => {
            setDeepLinkLoading(true);
        });

        const unsubSuccess = EventsOn('deep-link-auth-success', (data: { message: string; user: User; token: string }) => {
            console.log('Deep link auth success:', data);
            setDeepLinkLoading(false);
            if (data.user && data.token) {
                persistSession(data.token, data.user);

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

        return () => {
            unsubLoading();
            unsubSuccess();
            unsubError();
        };
    }, [persistSession]);

    const login = async (email: string, password: string): Promise<{ success: boolean; message: string }> => {
        try {
            const response: LoginResponse = await Login(email, password);

            if (response.data && response.data.token) {
                persistSession(response.data.token, response.data.user);

                return { success: true, message: response.message };
            }

            return { success: false, message: response.message || 'Login failed' };
        } catch (error) {
            console.error('Login error:', error);
            return { success: false, message: 'Login failed. Please try again.' };
        }
    };

    return (
        <AuthContext.Provider
            value={{
                user,
                token,
                login,
                logout,
                isAuthenticated: user !== null,
                isReady,
                deepLinkLoading,
                validateSession,
            }}
        >
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
