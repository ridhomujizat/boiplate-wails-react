import React, { useState } from 'react';
import { Layout, Menu } from 'antd';
import {
    HomeOutlined,
    SettingOutlined,
    UserOutlined,
    LogoutOutlined,
    MenuFoldOutlined,
    MenuUnfoldOutlined
} from '@ant-design/icons';
import { useNavigate, useLocation, Outlet } from 'react-router-dom';
import { useAuth } from '../contexts/AuthContext';

const { Header, Sider, Content } = Layout;

const MainLayout: React.FC = () => {
    const [collapsed, setCollapsed] = useState(false);
    const navigate = useNavigate();
    const location = useLocation();
    const { user, logout } = useAuth();

    const menuItems = [
        {
            key: '/',
            icon: <HomeOutlined />,
            label: 'Home',
        },
        {
            key: '/profile',
            icon: <UserOutlined />,
            label: 'Profile',
        },
        {
            key: '/settings',
            icon: <SettingOutlined />,
            label: 'Settings',
        },
    ];

    const handleMenuClick = ({ key }: { key: string }) => {
        navigate(key);
    };

    const handleLogout = () => {
        logout();
        navigate('/login');
    };

    return (
        <Layout className="min-h-screen">
            <Sider
                trigger={null}
                collapsible
                collapsed={collapsed}
                style={{ background: '#4c1d95' }}
            >
                <div className="h-16 flex items-center justify-center">
                    <span className="text-white text-lg font-bold">
                        {collapsed ? 'ONX' : 'ONX Screen'}
                    </span>
                </div>
                <Menu
                    theme="dark"
                    mode="inline"
                    selectedKeys={[location.pathname]}
                    items={menuItems}
                    onClick={handleMenuClick}
                    style={{ background: '#4c1d95' }}
                />
                <div className="absolute bottom-4 left-0 right-0 px-4">
                    <Menu
                        theme="dark"
                        mode="inline"
                        selectable={false}
                        items={[
                            {
                                key: 'logout',
                                icon: <LogoutOutlined />,
                                label: 'Logout',
                                onClick: handleLogout,
                            }
                        ]}
                        style={{ background: '#4c1d95' }}
                    />
                </div>
            </Sider>
            <Layout>
                <Header
                    className="flex items-center justify-between px-4"
                    style={{ background: '#fff', padding: '0 24px' }}
                >
                    <div
                        className="cursor-pointer text-lg"
                        onClick={() => setCollapsed(!collapsed)}
                    >
                        {collapsed ? <MenuUnfoldOutlined /> : <MenuFoldOutlined />}
                    </div>
                    <div className="flex items-center gap-2">
                        <UserOutlined style={{ color: '#7c3aed' }} />
                        <span className="text-gray-600">{user?.email}</span>
                    </div>
                </Header>
                <Content style={{ background: '#f5f5f5', minHeight: 'calc(100vh - 64px)' }}>
                    <Outlet />
                </Content>
            </Layout>
        </Layout>
    );
};

export default MainLayout;
