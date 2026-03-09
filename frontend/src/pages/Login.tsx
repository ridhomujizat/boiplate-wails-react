import React, { useState, useEffect } from 'react';
import { Form, Input, Button, Card, Typography, message, Modal, Space, Tooltip, Spin } from 'antd';
import { UserOutlined, LockOutlined, SettingOutlined, SaveOutlined, BankOutlined, GlobalOutlined, CloudServerOutlined, LoadingOutlined } from '@ant-design/icons';
import { useAuth } from '../contexts/AuthContext';
import { useNavigate } from 'react-router-dom';
import { GetSettings, SaveSettings } from '../../wailsjs/go/app/App';
import { EventsOn } from '../../wailsjs/runtime/runtime';

const { Text } = Typography;

interface LoginForm {
    email: string;
    password: string;
}

interface SettingFormValues {
    tenantCode: string;
    baseUrl: string;
    mqttBroker: string;
}

const Login: React.FC = () => {
    const [loading, setLoading] = useState(false);
    const [settingsOpen, setSettingsOpen] = useState(false);
    const [settingsLoading, setSettingsLoading] = useState(false);
    const [settingsSaving, setSettingsSaving] = useState(false);
    const { login, deepLinkLoading, isAuthenticated } = useAuth();
    const navigate = useNavigate();
    const [settingsForm] = Form.useForm<SettingFormValues>();

    // Listen for deep link auth success/error notifications
    useEffect(() => {
        const unsubSuccess = EventsOn('deep-link-auth-success', (data: { message: string }) => {
            message.success(data.message || 'Login successful!');
        });
        const unsubError = EventsOn('deep-link-auth-error', (data: { message: string }) => {
            message.error(data.message || 'Deep link login failed');
        });
        return () => {
            unsubSuccess();
            unsubError();
        };
    }, []);

    // Navigate to home when authenticated (handles deep link auth redirect)
    useEffect(() => {
        if (isAuthenticated) {
            navigate('/');
        }
    }, [isAuthenticated, navigate]);

    const onFinish = async (values: LoginForm) => {
        setLoading(true);
        try {
            const result = await login(values.email, values.password);
            if (result.success) {
                message.success(result.message || 'Login successful!');
                navigate('/');
            } else {
                message.error(result.message || 'Invalid credentials');
            }
        } catch {
            message.error('Login failed');
        } finally {
            setLoading(false);
        }
    };

    const openSettings = async () => {
        setSettingsOpen(true);
        setSettingsLoading(true);
        try {
            const settings = await GetSettings();
            settingsForm.setFieldsValue({
                tenantCode: settings.tenantCode || '',
                baseUrl: settings.baseUrl || '',
                mqttBroker: settings.mqttBroker || ''
            });
        } catch (error) {
            console.error('Failed to load settings:', error);
            message.error('Failed to load settings');
        } finally {
            setSettingsLoading(false);
        }
    };

    const handleSettingsSave = async (values: SettingFormValues) => {
        setSettingsSaving(true);
        try {
            const result = await SaveSettings({
                tenantCode: values.tenantCode,
                baseUrl: values.baseUrl,
                mqttBroker: values.mqttBroker
            });

            if (result.success) {
                message.success(result.message || 'Settings saved successfully!');
                setSettingsOpen(false);
            } else {
                message.error(result.message || 'Failed to save settings');
            }
        } catch (error) {
            console.error('Failed to save settings:', error);
            message.error('Failed to save settings');
        } finally {
            setSettingsSaving(false);
        }
    };

    return (
        <div className="min-h-screen flex items-center justify-center bg-neutral-50">
            {deepLinkLoading && (
                <div style={{
                    position: 'fixed',
                    inset: 0,
                    zIndex: 1000,
                    display: 'flex',
                    flexDirection: 'column',
                    alignItems: 'center',
                    justifyContent: 'center',
                    background: 'rgba(255,255,255,0.85)',
                    backdropFilter: 'blur(4px)',
                }}>
                    <Spin indicator={<LoadingOutlined style={{ fontSize: 48 }} spin />} />
                    <p style={{ marginTop: 16, fontSize: 16, color: '#555' }}>Signing in...</p>
                </div>
            )}
            <div className="w-full max-w-md">
                <Card
                    className="w-full shadow-sm"
                    styles={{ body: { padding: '48px 40px' } }}
                >
                    <div className="mb-10">
                        <div className="w-12 h-12 rounded-xl bg-primary-600 flex items-center justify-center mb-6">
                            <UserOutlined className="text-white text-xl" />
                        </div>
                        <h1 className="text-2xl font-semibold text-neutral-900 mb-2">Welcome back</h1>
                        <Text type="secondary">Sign in to your account</Text>
                    </div>

                    <Form
                        name="login"
                        onFinish={onFinish}
                        layout="vertical"
                        size="large"
                    >
                        <Form.Item
                            name="email"
                            rules={[
                                { required: true, message: 'Please input your email!' },
                                { type: 'email', message: 'Please enter a valid email!' }
                            ]}
                        >
                            <Input
                                prefix={<UserOutlined className="text-neutral-400" />}
                                placeholder="Email"
                                className="h-12"
                            />
                        </Form.Item>

                        <Form.Item
                            name="password"
                            rules={[{ required: true, message: 'Please input your password!' }]}
                        >
                            <Input.Password
                                prefix={<LockOutlined className="text-neutral-400" />}
                                placeholder="Password"
                                className="h-12"
                            />
                        </Form.Item>

                        <Form.Item className="mb-0">
                            <Button
                                type="primary"
                                htmlType="submit"
                                loading={loading}
                                block
                                className="h-12 text-base font-medium"
                            >
                                Sign In
                            </Button>
                        </Form.Item>
                    </Form>
                </Card>
                <div style={{ display: 'flex', justifyContent: 'flex-end', marginTop: 12 }}>
                    <Tooltip title="Settings">
                        <Button
                            type="text"
                            icon={<SettingOutlined style={{ fontSize: 20 }} />}
                            onClick={openSettings}
                            id="login-settings-btn"
                        />
                    </Tooltip>
                </div>
            </div>

            <Modal
                title={
                    <Space>
                        <SettingOutlined />
                        <span className="font-medium">Server Configuration</span>
                    </Space>
                }
                open={settingsOpen}
                onCancel={() => setSettingsOpen(false)}
                footer={null}
                destroyOnClose
            >
                {settingsLoading ? (
                    <div style={{ textAlign: 'center', padding: '40px 0' }}>
                        <Spin indicator={<LoadingOutlined style={{ fontSize: 36 }} spin />} />
                    </div>
                ) : (
                    <Form
                        form={settingsForm}
                        layout="vertical"
                        onFinish={handleSettingsSave}
                        initialValues={{
                            tenantCode: '',
                            baseUrl: '',
                            mqttBroker: ''
                        }}
                    >
                        <Form.Item
                            label={
                                <Space>
                                    <BankOutlined />
                                    <span className="font-medium">Tenant Code</span>
                                </Space>
                            }
                            name="tenantCode"
                            rules={[{ required: true, message: 'Please enter tenant code' }]}
                        >
                            <Input placeholder="Enter tenant code" size="large" />
                        </Form.Item>

                        <Form.Item
                            label={
                                <Space>
                                    <GlobalOutlined />
                                    <span className="font-medium">Base URL</span>
                                </Space>
                            }
                            name="baseUrl"
                            rules={[
                                { required: true, message: 'Please enter base URL' },
                                { type: 'url', message: 'Please enter a valid URL' }
                            ]}
                        >
                            <Input placeholder="https://api.example.com" size="large" />
                        </Form.Item>

                        <Form.Item
                            label={
                                <Space>
                                    <CloudServerOutlined />
                                    <span className="font-medium">MQTT Broker</span>
                                </Space>
                            }
                            name="mqttBroker"
                            rules={[{ required: true, message: 'Please enter MQTT broker address' }]}
                        >
                            <Input placeholder="mqtt://broker.example.com:1883" size="large" />
                        </Form.Item>

                        <Form.Item style={{ marginTop: 24, marginBottom: 0 }}>
                            <Button
                                type="primary"
                                htmlType="submit"
                                icon={<SaveOutlined />}
                                loading={settingsSaving}
                                block
                                size="large"
                                className="h-12 font-medium"
                            >
                                Save Settings
                            </Button>
                        </Form.Item>
                    </Form>
                )}
            </Modal>
        </div>
    );
};

export default Login;
