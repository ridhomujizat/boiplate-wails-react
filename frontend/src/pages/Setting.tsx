import React, { useState, useEffect } from 'react';
import { Card, Form, Input, Button, Typography, message, Space, Spin } from 'antd';
import {
    SettingOutlined,
    SaveOutlined,
    GlobalOutlined,
    CloudServerOutlined,
    BankOutlined,
    LoadingOutlined
} from '@ant-design/icons';
import { GetSettings, SaveSettings } from '../../wailsjs/go/app/App';

const { Title, Text } = Typography;

interface SettingFormValues {
    tenantCode: string;
    baseUrl: string;
    mqttBroker: string;
}

const Setting: React.FC = () => {
    const [loading, setLoading] = useState(false);
    const [initialLoading, setInitialLoading] = useState(true);
    const [form] = Form.useForm<SettingFormValues>();

    // Load settings on component mount
    useEffect(() => {
        loadSettings();
    }, []);

    const loadSettings = async () => {
        try {
            const settings = await GetSettings();
            form.setFieldsValue({
                tenantCode: settings.tenantCode || '',
                baseUrl: settings.baseUrl || '',
                mqttBroker: settings.mqttBroker || ''
            });
        } catch (error) {
            console.error('Failed to load settings:', error);
            message.error('Failed to load settings');
        } finally {
            setInitialLoading(false);
        }
    };

    const handleSave = async (values: SettingFormValues) => {
        setLoading(true);
        try {
            const result = await SaveSettings({
                tenantCode: values.tenantCode,
                baseUrl: values.baseUrl,
                mqttBroker: values.mqttBroker
            });

            if (result.success) {
                message.success(result.message || 'Settings saved successfully!');
            } else {
                message.error(result.message || 'Failed to save settings');
            }
        } catch (error) {
            console.error('Failed to save settings:', error);
            message.error('Failed to save settings');
        } finally {
            setLoading(false);
        }
    };

    if (initialLoading) {
        return (
            <div className="p-6 flex justify-center items-center" style={{ minHeight: 400 }}>
                <Spin indicator={<LoadingOutlined style={{ fontSize: 48, color: '#7c3aed' }} spin />} />
            </div>
        );
    }

    return (
        <div className="p-6">
            <Title level={2} style={{ color: '#4c1d95', marginBottom: 24 }}>
                <SettingOutlined style={{ marginRight: 12 }} />
                Settings
            </Title>

            <Card
                title={
                    <Space>
                        <CloudServerOutlined style={{ color: '#7c3aed' }} />
                        <Text strong>Configuration</Text>
                    </Space>
                }
                style={{ maxWidth: 600 }}
            >
                <Form
                    form={form}
                    layout="vertical"
                    onFinish={handleSave}
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
                        <Input
                            placeholder="Enter tenant code"
                            size="large"
                        />
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
                        <Input
                            placeholder="https://api.example.com"
                            size="large"
                        />
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
                        <Input
                            placeholder="mqtt://broker.example.com:1883"
                            size="large"
                        />
                    </Form.Item>

                    <Form.Item style={{ marginTop: 24, marginBottom: 0 }}>
                        <Button
                            type="primary"
                            htmlType="submit"
                            icon={<SaveOutlined />}
                            loading={loading}
                            size="large"
                            style={{
                                backgroundColor: '#7c3aed',
                                borderColor: '#7c3aed'
                            }}
                        >
                            Save Settings
                        </Button>
                    </Form.Item>
                </Form>
            </Card>
        </div>
    );
};

export default Setting;
