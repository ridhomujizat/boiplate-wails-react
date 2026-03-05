import React, { useState } from 'react';
import { Layout, Menu } from 'antd';
import {
    HomeOutlined,
    SettingOutlined,
    UserOutlined,
    LogoutOutlined,
    MenuFoldOutlined,
    MenuUnfoldOutlined,
    VideoCameraOutlined,
    BarChartOutlined
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
            key: '/activity',
            icon: <BarChartOutlined />,
            label: 'Activity',
        },
        {
            key: '/recording',
            icon: <VideoCameraOutlined />,
            label: 'Recording',
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

    const handleLogout = async () => {
        await logout();
        navigate('/login');
    };

    return (
        <Layout style={{ minHeight: '100vh' }}>
            <div
                style={{
                    position: 'fixed',
                    left: 0,
                    top: 0,
                    bottom: 0,
                    width: collapsed ? 80 : 200,
                    background: '#171717',
                    display: 'flex',
                    flexDirection: 'column',
                    zIndex: 1000,
                    transition: 'width 0.2s'
                }}
            >
                <div className="h-16 flex items-center justify-center" style={{ flexShrink: 0 }}>
                    <span className="text-white text-lg font-semibold tracking-tight">
                        {collapsed ? 'ONX' : 'ONX Screen'}
                    </span>
                </div>
                <div style={{ flex: 1, overflowY: 'auto', overflowX: 'hidden' }}>
                    <Menu
                        theme="dark"
                        mode="inline"
                        selectedKeys={[location.pathname]}
                        items={menuItems}
                        onClick={handleMenuClick}
                        style={{ background: '#171717', border: 'none' }}
                        inlineCollapsed={collapsed}
                    />
                </div>
                <div style={{
                    flexShrink: 0,
                    paddingLeft: collapsed ? 0 : 16,
                    paddingRight: collapsed ? 0 : 16,
                    paddingBottom: 12,
                    paddingTop: 8
                }}>
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
                        style={{ background: '#171717', border: 'none' }}
                        inlineCollapsed={collapsed}
                    />
                </div>
            </div>
            <Layout style={{ marginLeft: collapsed ? 80 : 200, transition: 'margin-left 0.2s' }}>
                <Header
                    className="flex items-center justify-between px-4"
                    style={{ background: '#ffffff', padding: '0 24px', borderBottom: '1px solid #e5e5e5' }}
                >
                    <div
                        className="cursor-pointer text-lg text-neutral-600"
                        onClick={() => setCollapsed(!collapsed)}
                    >
                        {collapsed ? <MenuUnfoldOutlined /> : <MenuFoldOutlined />}
                    </div>
                    <div className="flex items-center gap-3">
                        <div className="w-8 h-8 rounded-full bg-primary-100 flex items-center justify-center">
                            <UserOutlined style={{ color: '#2851e6' }} />
                        </div>
                        <span className="text-neutral-600 text-sm font-medium">{user?.email}</span>
                    </div>
                </Header>
                <Content style={{ background: '#fafafa', minHeight: 'calc(100vh - 64px)' }}>
                    <Outlet />
                </Content>
            </Layout>
        </Layout>
    );
};

export default MainLayout;
