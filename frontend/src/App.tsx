import { ConfigProvider } from 'antd';
import { BrowserRouter, Routes, Route, Navigate, useLocation } from 'react-router-dom';
import { useEffect } from 'react';
import { AuthProvider, useAuth } from './contexts/AuthContext';
import MainLayout from './layouts/MainLayout';
import Login from './pages/Login';
import Home from './pages/Home';
import ProtectedRoute from './components/ProtectedRoute';
import Setting from './pages/Setting';
import Recording from './pages/Recording';
import Activity from './pages/Activity';
import RecordHistory from './pages/RecordHistory';

const theme = {
    token: {
        colorPrimary: '#2851e6',
        colorLink: '#2851e6',
        colorLinkHover: '#213fc9',
        borderRadius: 8,
        fontFamily: '-apple-system, BlinkMacSystemFont, SF Pro Display, Segoe UI, Roboto, sans-serif',
    },
};

function AuthSessionWatcher() {
    const location = useLocation();
    const { token, validateSession } = useAuth();

    useEffect(() => {
        if (!token) {
            return;
        }

        void validateSession();
    }, [location.pathname, token, validateSession]);

    return null;
}

function AppRoutes() {
    const { isAuthenticated } = useAuth();

    return (
        <>
            <AuthSessionWatcher />
            <Routes>
                <Route
                    path="/login"
                    element={isAuthenticated ? <Navigate to="/" replace /> : <Login />}
                />
                <Route
                    path="/"
                    element={
                        <ProtectedRoute>
                            <MainLayout />
                        </ProtectedRoute>
                    }
                >
                    <Route index element={<Home />} />
                    <Route path="settings" element={<Setting />} />
                    <Route path="recording" element={<Recording />} />
                    <Route path="record-history" element={<RecordHistory />} />
                    <Route path="activity" element={<Activity />} />
                </Route>
                <Route path="*" element={<Navigate to="/" replace />} />
            </Routes>
        </>
    );
}

function App() {
    return (
        <ConfigProvider theme={theme}>
            <BrowserRouter>
                <AuthProvider>
                    <AppRoutes />
                </AuthProvider>
            </BrowserRouter>
        </ConfigProvider>
    );
}

export default App;
