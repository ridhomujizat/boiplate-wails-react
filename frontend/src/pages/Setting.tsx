import React, { useState, useEffect } from 'react';
import { Card, Form, Input, Button, Typography, message, Space, Spin, Select, Switch, Tag, Divider, InputNumber } from 'antd';
import {
    SettingOutlined,
    SaveOutlined,
    GlobalOutlined,
    CloudServerOutlined,
    BankOutlined,
    LoadingOutlined,
    SafetyCertificateOutlined,
    DesktopOutlined,
    AudioOutlined,
    SoundOutlined,
    CheckCircleOutlined,
    ExclamationCircleOutlined,
    FieldTimeOutlined,
    VideoCameraOutlined
} from '@ant-design/icons';
import {
    GetSettings,
    SaveSettings,
    CheckScreenPermission,
    RequestScreenPermission,
    CheckAccessibilityPermission,
    RequestAccessibilityPermission,
    GetCaptureDevices,
    GetAudioSettings,
    SaveAudioSettings,
    GetActivitySettings,
    SaveActivitySettings,
    GetRecordingSettings,
    SaveRecordingSettings
} from '../../wailsjs/go/app/App';
import { app } from '../../wailsjs/go/models';

const { Title, Text } = Typography;
const { Option } = Select;

interface SettingFormValues {
    tenantCode: string;
    baseUrl: string;
    mqttBroker: string;
}

interface AudioFormValues {
    microphoneId: string;
    systemAudioEnabled: boolean;
}

interface ActivityFormValues {
    pollingInterval: number;
    afkThreshold: number;
}

interface RecordingFormValues {
    maxRecordingTimeEnabled: boolean;
    maxRecordingTimeSeconds: number;
}

interface PermissionState {
    screen: { granted: boolean; message: string };
    accessibility: { granted: boolean; message: string };
}

const Setting: React.FC = () => {
    const [loading, setLoading] = useState(false);
    const [audioLoading, setAudioLoading] = useState(false);
    const [activityLoading, setActivityLoading] = useState(false);
    const [recordingLoading, setRecordingLoading] = useState(false);
    const [initialLoading, setInitialLoading] = useState(true);
    const [form] = Form.useForm<SettingFormValues>();
    const [audioForm] = Form.useForm<AudioFormValues>();
    const [activityForm] = Form.useForm<ActivityFormValues>();
    const [recordingForm] = Form.useForm<RecordingFormValues>();

    // Permission states
    const [permissions, setPermissions] = useState<PermissionState>({
        screen: { granted: false, message: 'Checking...' },
        accessibility: { granted: false, message: 'Checking...' }
    });

    // Audio devices
    const [microphones, setMicrophones] = useState<app.AudioDevice[]>([]);

    // Load settings on component mount
    useEffect(() => {
        loadSettings();
        checkPermissions();
        loadAudioDevices();
        loadActivitySettings();
        loadRecordingSettings();
    }, []);

    const loadSettings = async () => {
        try {
            const settings = await GetSettings();
            form.setFieldsValue({
                tenantCode: settings.tenantCode || '',
                baseUrl: settings.baseUrl || '',
                mqttBroker: settings.mqttBroker || ''
            });

            const audioSettings = await GetAudioSettings();
            audioForm.setFieldsValue({
                microphoneId: audioSettings.microphoneId || '',
                systemAudioEnabled: audioSettings.systemAudioEnabled || false
            });
        } catch (error) {
            console.error('Failed to load settings:', error);
            message.error('Failed to load settings');
        } finally {
            setInitialLoading(false);
        }
    };

    const checkPermissions = async () => {
        try {
            const screenStatus = await CheckScreenPermission();
            const accessibilityStatus = await CheckAccessibilityPermission();

            setPermissions({
                screen: { granted: screenStatus.granted, message: screenStatus.message },
                accessibility: { granted: accessibilityStatus.granted, message: accessibilityStatus.message }
            });
        } catch (error) {
            console.error('Failed to check permissions:', error);
        }
    };

    const loadAudioDevices = async () => {
        try {
            const devices = await GetCaptureDevices();
            setMicrophones(devices || []);
        } catch (error) {
            console.error('Failed to load audio devices:', error);
        }
    };

    const loadActivitySettings = async () => {
        try {
            const settings = await GetActivitySettings();
            activityForm.setFieldsValue({
                pollingInterval: settings.pollingInterval || 5,
                afkThreshold: settings.afkThreshold || 180
            });
        } catch (error) {
            console.error('Failed to load activity settings:', error);
        }
    };

    const loadRecordingSettings = async () => {
        try {
            const settings = await GetRecordingSettings();
            recordingForm.setFieldsValue({
                maxRecordingTimeEnabled: settings.maxRecordingTimeEnabled || false,
                maxRecordingTimeSeconds: settings.maxRecordingTimeSeconds || 3600
            });
        } catch (error) {
            console.error('Failed to load recording settings:', error);
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

    const handleAudioSave = async (values: AudioFormValues) => {
        setAudioLoading(true);
        try {
            const result = await SaveAudioSettings({
                microphoneId: values.microphoneId,
                systemAudioEnabled: values.systemAudioEnabled
            });

            if (result.success) {
                message.success(result.message || 'Audio settings saved successfully!');
            } else {
                message.error(result.message || 'Failed to save audio settings');
            }
        } catch (error) {
            console.error('Failed to save audio settings:', error);
            message.error('Failed to save audio settings');
        } finally {
            setAudioLoading(false);
        }
    };

    const handleActivitySave = async (values: ActivityFormValues) => {
        setActivityLoading(true);
        try {
            const result = await SaveActivitySettings({
                pollingInterval: values.pollingInterval,
                afkThreshold: values.afkThreshold
            });

            if (result.success) {
                message.success(result.message || 'Activity settings saved successfully!');
            } else {
                message.error(result.message || 'Failed to save activity settings');
            }
        } catch (error) {
            console.error('Failed to save activity settings:', error);
            message.error('Failed to save activity settings');
        } finally {
            setActivityLoading(false);
        }
    };

    const handleRecordingSave = async (values: RecordingFormValues) => {
        setRecordingLoading(true);
        try {
            // Use default value when disabled to avoid validation issues
            const seconds = values.maxRecordingTimeEnabled ? values.maxRecordingTimeSeconds : 3600;

            const result = await SaveRecordingSettings({
                maxRecordingTimeEnabled: values.maxRecordingTimeEnabled,
                maxRecordingTimeSeconds: seconds
            });

            if (result.success) {
                message.success(result.message || 'Recording settings saved successfully!');
            } else {
                message.error(result.message || 'Failed to save recording settings');
            }
        } catch (error) {
            console.error('Failed to save recording settings:', error);
            message.error('Failed to save recording settings');
        } finally {
            setRecordingLoading(false);
        }
    };

    const handleRequestScreenPermission = async () => {
        try {
            await RequestScreenPermission();
            message.info('Opening System Preferences...');
            // Re-check permissions after a short delay
            setTimeout(checkPermissions, 2000);
        } catch (error) {
            console.error('Failed to request screen permission:', error);
        }
    };

    const handleRequestAccessibilityPermission = async () => {
        try {
            await RequestAccessibilityPermission();
            message.info('Opening System Preferences...');
            // Re-check permissions after a short delay
            setTimeout(checkPermissions, 2000);
        } catch (error) {
            console.error('Failed to request accessibility permission:', error);
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
        <div className="p-6" style={{ maxWidth: 800 }}>
            <Title level={2} style={{ color: '#4c1d95', marginBottom: 24 }}>
                <SettingOutlined style={{ marginRight: 12 }} />
                Settings
            </Title>

            {/* Permissions Section */}
            <Card
                title={
                    <Space>
                        <SafetyCertificateOutlined style={{ color: '#7c3aed' }} />
                        <Text strong>Permissions</Text>
                    </Space>
                }
                style={{ marginBottom: 24 }}
            >
                <div style={{ display: 'flex', flexDirection: 'column', gap: 16 }}>
                    {/* Screen Recording Permission */}
                    <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                        <Space>
                            <DesktopOutlined style={{ fontSize: 18, color: '#7c3aed' }} />
                            <div>
                                <Text strong>Screen Recording</Text>
                                <br />
                                <Text type="secondary" style={{ fontSize: 12 }}>
                                    Required to capture screen content
                                </Text>
                            </div>
                        </Space>
                        <Space>
                            <Tag
                                icon={permissions.screen.granted ? <CheckCircleOutlined /> : <ExclamationCircleOutlined />}
                                color={permissions.screen.granted ? 'success' : 'warning'}
                            >
                                {permissions.screen.granted ? 'Granted' : 'Not Granted'}
                            </Tag>
                            {!permissions.screen.granted && (
                                <Button
                                    type="primary"
                                    size="small"
                                    onClick={handleRequestScreenPermission}
                                    style={{ backgroundColor: '#7c3aed', borderColor: '#7c3aed' }}
                                >
                                    Request
                                </Button>
                            )}
                        </Space>
                    </div>

                    <Divider style={{ margin: '8px 0' }} />

                    {/* Accessibility Permission */}
                    <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                        <Space>
                            <SafetyCertificateOutlined style={{ fontSize: 18, color: '#7c3aed' }} />
                            <div>
                                <Text strong>Accessibility</Text>
                                <br />
                                <Text type="secondary" style={{ fontSize: 12 }}>
                                    Required for advanced features
                                </Text>
                            </div>
                        </Space>
                        <Space>
                            <Tag
                                icon={permissions.accessibility.granted ? <CheckCircleOutlined /> : <ExclamationCircleOutlined />}
                                color={permissions.accessibility.granted ? 'success' : 'warning'}
                            >
                                {permissions.accessibility.granted ? 'Granted' : 'Not Granted'}
                            </Tag>
                            {!permissions.accessibility.granted && (
                                <Button
                                    type="primary"
                                    size="small"
                                    onClick={handleRequestAccessibilityPermission}
                                    style={{ backgroundColor: '#7c3aed', borderColor: '#7c3aed' }}
                                >
                                    Request
                                </Button>
                            )}
                        </Space>
                    </div>
                </div>
            </Card>

            {/* Audio Settings Section */}
            <Card
                title={
                    <Space>
                        <AudioOutlined style={{ color: '#7c3aed' }} />
                        <Text strong>Audio Settings</Text>
                    </Space>
                }
                style={{ marginBottom: 24 }}
            >
                <Form
                    form={audioForm}
                    layout="vertical"
                    onFinish={handleAudioSave}
                    initialValues={{
                        microphoneId: '',
                        systemAudioEnabled: false
                    }}
                >
                    <Form.Item
                        label={
                            <Space>
                                <AudioOutlined style={{ color: '#7c3aed' }} />
                                <span>Microphone</span>
                            </Space>
                        }
                        name="microphoneId"
                    >
                        <Select
                            placeholder="Select a microphone"
                            size="large"
                            allowClear
                        >
                            {microphones.map((device) => (
                                <Option key={device.id} value={device.id}>
                                    {device.name}
                                </Option>
                            ))}
                        </Select>
                    </Form.Item>

                    <Form.Item
                        label={
                            <Space>
                                <SoundOutlined style={{ color: '#7c3aed' }} />
                                <span>System Audio</span>
                            </Space>
                        }
                        name="systemAudioEnabled"
                        valuePropName="checked"
                    >
                        <Switch
                            checkedChildren="ON"
                            unCheckedChildren="OFF"
                        />
                    </Form.Item>

                    <Form.Item style={{ marginTop: 16, marginBottom: 0 }}>
                        <Button
                            type="primary"
                            htmlType="submit"
                            icon={<SaveOutlined />}
                            loading={audioLoading}
                            size="large"
                            style={{
                                backgroundColor: '#7c3aed',
                                borderColor: '#7c3aed'
                            }}
                        >
                            Save Audio Settings
                        </Button>
                    </Form.Item>
                </Form>
            </Card>

            {/* Recording Settings Section */}
            <Card
                title={
                    <Space>
                        <VideoCameraOutlined style={{ color: '#7c3aed' }} />
                        <Text strong>Recording Settings</Text>
                    </Space>
                }
                style={{ marginBottom: 24 }}
            >
                <Form
                    form={recordingForm}
                    layout="vertical"
                    onFinish={handleRecordingSave}
                    initialValues={{
                        maxRecordingTimeEnabled: false,
                        maxRecordingTimeSeconds: 3600
                    }}
                >
                    <Form.Item
                        label={
                            <Space>
                                <FieldTimeOutlined style={{ color: '#7c3aed' }} />
                                <span>Maximum Recording Time</span>
                            </Space>
                        }
                        name="maxRecordingTimeEnabled"
                        valuePropName="checked"
                        extra="Automatically stop recording after the specified duration"
                    >
                        <Switch
                            checkedChildren="ON"
                            unCheckedChildren="OFF"
                        />
                    </Form.Item>

                    <Form.Item noStyle shouldUpdate={(prevValues, currentValues) => prevValues.maxRecordingTimeEnabled !== currentValues.maxRecordingTimeEnabled}>
                        {({ getFieldValue }) =>
                            getFieldValue('maxRecordingTimeEnabled') ? (
                                <Form.Item
                                    label={
                                        <Space>
                                            <FieldTimeOutlined style={{ color: '#7c3aed' }} />
                                            <span>Duration (seconds)</span>
                                        </Space>
                                    }
                                    name="maxRecordingTimeSeconds"
                                    extra="Minimum: 30 seconds, Maximum: 36000 seconds (10 hours)"
                                    rules={[
                                        { required: true, message: 'Please enter duration' },
                                        { type: 'number', min: 30, message: 'Minimum is 30 seconds' },
                                        { type: 'number', max: 36000, message: 'Maximum is 36000 seconds (10 hours)' }
                                    ]}
                                >
                                    <InputNumber
                                        min={30}
                                        max={36000}
                                        size="large"
                                        style={{ width: '100%' }}
                                        placeholder="3600"
                                        addonAfter="seconds"
                                    />
                                </Form.Item>
                            ) : null
                        }
                    </Form.Item>

                    <Form.Item style={{ marginTop: 16, marginBottom: 0 }}>
                        <Button
                            type="primary"
                            htmlType="submit"
                            icon={<SaveOutlined />}
                            loading={recordingLoading}
                            size="large"
                            style={{
                                backgroundColor: '#7c3aed',
                                borderColor: '#7c3aed'
                            }}
                        >
                            Save Recording Settings
                        </Button>
                    </Form.Item>
                </Form>
            </Card>

            {/* Activity Tracking Section */}
            <Card
                title={
                    <Space>
                        <FieldTimeOutlined style={{ color: '#7c3aed' }} />
                        <Text strong>Activity Tracking</Text>
                    </Space>
                }
                style={{ marginBottom: 24 }}
            >
                <Form
                    form={activityForm}
                    layout="vertical"
                    onFinish={handleActivitySave}
                    initialValues={{
                        pollingInterval: 5,
                        afkThreshold: 180
                    }}
                >
                    <Form.Item
                        label={
                            <Space>
                                <FieldTimeOutlined style={{ color: '#7c3aed' }} />
                                <span>Polling Interval</span>
                            </Space>
                        }
                        name="pollingInterval"
                        extra="How often to check active window (seconds)"
                        rules={[{ required: true, message: 'Please enter polling interval' }]}
                    >
                        <InputNumber
                            min={1}
                            max={60}
                            size="large"
                            style={{ width: '100%' }}
                            placeholder="5"
                        />
                    </Form.Item>

                    <Form.Item
                        label={
                            <Space>
                                <DesktopOutlined style={{ color: '#7c3aed' }} />
                                <span>AFK Threshold</span>
                            </Space>
                        }
                        name="afkThreshold"
                        extra="Idle time before marking as away (seconds)"
                        rules={[{ required: true, message: 'Please enter AFK threshold' }]}
                    >
                        <InputNumber
                            min={10}
                            max={3600}
                            size="large"
                            style={{ width: '100%' }}
                            placeholder="180"
                        />
                    </Form.Item>

                    <Form.Item style={{ marginTop: 16, marginBottom: 0 }}>
                        <Button
                            type="primary"
                            htmlType="submit"
                            icon={<SaveOutlined />}
                            loading={activityLoading}
                            size="large"
                            style={{
                                backgroundColor: '#7c3aed',
                                borderColor: '#7c3aed'
                            }}
                        >
                            Save Activity Settings
                        </Button>
                    </Form.Item>
                </Form>
            </Card>

            {/* Configuration Section */}
            <Card
                title={
                    <Space>
                        <CloudServerOutlined style={{ color: '#7c3aed' }} />
                        <Text strong>Configuration</Text>
                    </Space>
                }
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
