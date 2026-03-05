import React, { useState, useEffect } from 'react';
import { Card, Form, Input, InputNumber, Button, Typography, message, Space, Spin, Select, Switch, Tag, Divider } from 'antd';
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
    VideoCameraOutlined,
    CloudUploadOutlined,
    DeleteOutlined
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
    SaveRecordingSettings,
    GetUploadSettings,
    SaveUploadSettings
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

interface UploadFormValues {
    deleteAfterUpload: boolean;
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
    const [uploadLoading, setUploadLoading] = useState(false);
    const [initialLoading, setInitialLoading] = useState(true);
    const [form] = Form.useForm<SettingFormValues>();
    const [audioForm] = Form.useForm<AudioFormValues>();
    const [activityForm] = Form.useForm<ActivityFormValues>();
    const [recordingForm] = Form.useForm<RecordingFormValues>();
    const [uploadForm] = Form.useForm<UploadFormValues>();

    const [permissions, setPermissions] = useState<PermissionState>({
        screen: { granted: false, message: 'Checking...' },
        accessibility: { granted: false, message: 'Checking...' }
    });

    const [microphones, setMicrophones] = useState<app.AudioDevice[]>([]);

    useEffect(() => {
        loadSettings();
        checkPermissions();
        loadAudioDevices();
        loadActivitySettings();
        loadRecordingSettings();
        loadUploadSettings();
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

    const loadUploadSettings = async () => {
        try {
            const settings = await GetUploadSettings();
            uploadForm.setFieldsValue({
                deleteAfterUpload: settings.deleteAfterUpload || false
            });
        } catch (error) {
            console.error('Failed to load upload settings:', error);
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

    const handleUploadSave = async (values: UploadFormValues) => {
        setUploadLoading(true);
        try {
            const result = await SaveUploadSettings({
                deleteAfterUpload: values.deleteAfterUpload
            });

            if (result.success) {
                message.success(result.message || 'Upload settings saved successfully!');
            } else {
                message.error(result.message || 'Failed to save upload settings');
            }
        } catch (error) {
            console.error('Failed to save upload settings:', error);
            message.error('Failed to save upload settings');
        } finally {
            setUploadLoading(false);
        }
    };

    const handleRequestScreenPermission = async () => {
        try {
            await RequestScreenPermission();
            message.info('Opening System Preferences...');
            setTimeout(checkPermissions, 2000);
        } catch (error) {
            console.error('Failed to request screen permission:', error);
        }
    };

    const handleRequestAccessibilityPermission = async () => {
        try {
            await RequestAccessibilityPermission();
            message.info('Opening System Preferences...');
            setTimeout(checkPermissions, 2000);
        } catch (error) {
            console.error('Failed to request accessibility permission:', error);
        }
    };

    if (initialLoading) {
        return (
            <div className="p-6 flex justify-center items-center" style={{ minHeight: 400 }}>
                <Spin indicator={<LoadingOutlined style={{ fontSize: 48 }} spin />} />
            </div>
        );
    }

    return (
        <div className="p-6" style={{ maxWidth: 800 }}>
            <Title level={2} style={{ color: '#171717', marginBottom: 24, fontWeight: 600 }}>
                <SettingOutlined style={{ marginRight: 12 }} />
                Settings
            </Title>

            <Card
                title={
                    <Space>
                        <SafetyCertificateOutlined />
                        <Text strong>Permissions</Text>
                    </Space>
                }
                style={{ marginBottom: 24 }}
            >
                <div style={{ display: 'flex', flexDirection: 'column', gap: 16 }}>
                    <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                        <Space>
                            <DesktopOutlined style={{ fontSize: 18 }} />
                            <div>
                                <Text strong style={{ display: 'block', color: '#171717' }}>Screen Recording</Text>
                                <Text type="secondary" style={{ fontSize: 13 }}>
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
                                    className="font-medium"
                                >
                                    Request
                                </Button>
                            )}
                        </Space>
                    </div>

                    <Divider style={{ margin: '8px 0' }} />

                    <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                        <Space>
                            <SafetyCertificateOutlined style={{ fontSize: 18 }} />
                            <div>
                                <Text strong style={{ display: 'block', color: '#171717' }}>Accessibility</Text>
                                <Text type="secondary" style={{ fontSize: 13 }}>
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
                                    className="font-medium"
                                >
                                    Request
                                </Button>
                            )}
                        </Space>
                    </div>
                </div>
            </Card>

            <Card
                title={
                    <Space>
                        <AudioOutlined />
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
                                <AudioOutlined />
                                <span className="font-medium">Microphone</span>
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
                                <SoundOutlined />
                                <span className="font-medium">System Audio</span>
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
                            className="font-medium"
                        >
                            Save Audio Settings
                        </Button>
                    </Form.Item>
                </Form>
            </Card>

            <Card
                title={
                    <Space>
                        <VideoCameraOutlined />
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
                                <FieldTimeOutlined />
                                <span className="font-medium">Maximum Recording Time</span>
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
                                            <FieldTimeOutlined />
                                            <span className="font-medium">Duration (seconds)</span>
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
                            className="font-medium"
                        >
                            Save Recording Settings
                        </Button>
                    </Form.Item>
                </Form>
            </Card>

            <Card
                title={
                    <Space>
                        <CloudUploadOutlined />
                        <Text strong>Upload Settings</Text>
                    </Space>
                }
                style={{ marginBottom: 24 }}
            >
                <Form
                    form={uploadForm}
                    layout="vertical"
                    onFinish={handleUploadSave}
                    initialValues={{
                        deleteAfterUpload: false
                    }}
                >
                    <Form.Item
                        label={
                            <Space>
                                <DeleteOutlined />
                                <span className="font-medium">Delete file after upload</span>
                            </Space>
                        }
                        name="deleteAfterUpload"
                        valuePropName="checked"
                        extra="Automatically delete the local recording file after successful upload"
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
                            loading={uploadLoading}
                            size="large"
                            className="font-medium"
                        >
                            Save Upload Settings
                        </Button>
                    </Form.Item>
                </Form>
            </Card>

            <Card
                title={
                    <Space>
                        <FieldTimeOutlined />
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
                                <FieldTimeOutlined />
                                <span className="font-medium">Polling Interval</span>
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
                                <DesktopOutlined />
                                <span className="font-medium">AFK Threshold</span>
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
                            className="font-medium"
                        >
                            Save Activity Settings
                        </Button>
                    </Form.Item>
                </Form>
            </Card>

            <Card
                title={
                    <Space>
                        <CloudServerOutlined />
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
                                <BankOutlined />
                                <span className="font-medium">Tenant Code</span>
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
                        <Input
                            placeholder="https://api.example.com"
                            size="large"
                        />
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
                            className="font-medium"
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
