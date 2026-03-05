import React, { useState, useEffect } from 'react';
import { Form, Input, Button, Card, Typography, message, Modal, Space, Tooltip, Spin } from 'antd';
import { UserOutlined, LockOutlined, SettingOutlined, SaveOutlined, BankOutlined, GlobalOutlined, CloudServerOutlined, LoadingOutlined } from '@ant-design/icons';
import { useAuth } from '../contexts/AuthContext';
import { useNavigate } from 'react-router-dom';
import { GetSettings, SaveSettings } from '../../wailsjs/go/app/App';

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
    const { login } = useAuth();
    const navigate = useNavigate();
    const [settingsForm] = Form.useForm<SettingFormValues>();

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
        <div className="min-h-screen flex items-center justify-center bg-gray-50">
            <div className="w-full max-w-md">
                <Card
                    className="w-full shadow-lg"
                    styles={{ body: { padding: '40px' } }}
                >

                    <div className="text-center mb-8">
                        {/* <Title level={2} className="!mb-2" style={{ color: '#4c1d95' }}>
                        Welcome Back
                    </Title> */}
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
                                prefix={<UserOutlined className="text-gray-400" />}
                                placeholder="Email"
                            />
                        </Form.Item>

                        <Form.Item
                            name="password"
                            rules={[{ required: true, message: 'Please input your password!' }]}
                        >
                            <Input.Password
                                prefix={<LockOutlined className="text-gray-400" />}
                                placeholder="Password"
                            />
                        </Form.Item>

                        <Form.Item className="mb-0">
                            <Button
                                type="primary"
                                htmlType="submit"
                                loading={loading}
                                block
                                style={{
                                    backgroundColor: '#7c3aed',
                                    borderColor: '#7c3aed',
                                    height: '48px'
                                }}
                            >
                                Sign In
                            </Button>
                        </Form.Item>
                    </Form>
                </Card>
                {/* Settings gear icon - below card */}
                <div style={{ display: 'flex', justifyContent: 'flex-end', marginTop: 8 }}>
                    <Tooltip title="Settings">
                        <Button
                            type="text"
                            icon={<SettingOutlined style={{ fontSize: 20, color: '#7c3aed' }} />}
                            onClick={openSettings}
                            id="login-settings-btn"
                        />
                    </Tooltip>
                </div>
            </div>

            {/* Settings Modal */}
            <Modal
                title={
                    <Space>
                        <SettingOutlined style={{ color: '#7c3aed' }} />
                        <span>Server Configuration</span>
                    </Space>
                }
                open={settingsOpen}
                onCancel={() => setSettingsOpen(false)}
                footer={null}
                destroyOnClose
            >
                {settingsLoading ? (
                    <div style={{ textAlign: 'center', padding: '40px 0' }}>
                        <Spin indicator={<LoadingOutlined style={{ fontSize: 36, color: '#7c3aed' }} spin />} />
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
                                    <BankOutlined style={{ color: '#7c3aed' }} />
                                    <span>Tenant Code</span>
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
                                    <GlobalOutlined style={{ color: '#7c3aed' }} />
                                    <span>Base URL</span>
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
                                    <CloudServerOutlined style={{ color: '#7c3aed' }} />
                                    <span>MQTT Broker</span>
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
                                style={{
                                    backgroundColor: '#7c3aed',
                                    borderColor: '#7c3aed',
                                    height: '48px'
                                }}
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
