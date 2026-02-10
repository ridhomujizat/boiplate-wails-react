import React, { useEffect, useState } from 'react';
import { Card, Row, Col, Typography, Tag, Space, Spin } from 'antd';
import {
    CheckCircleOutlined,
    CloseCircleOutlined,
    ApiOutlined,
    ReloadOutlined
} from '@ant-design/icons';
import { GetMQTTStatus } from '../../wailsjs/go/app/App';
import { EventsOn } from '../../wailsjs/runtime/runtime';

const { Title, Text } = Typography;

interface MQTTStatus {
    connected: boolean;
    message: string;
}

const Home: React.FC = () => {
    const [mqttStatus, setMqttStatus] = useState<MQTTStatus>({
        connected: false,
        message: 'Checking...'
    });
    const [loading, setLoading] = useState(true);

    const fetchMQTTStatus = async () => {
        try {
            const status = await GetMQTTStatus();
            setMqttStatus(status);
        } catch (error) {
            console.error('Failed to get MQTT status:', error);
            setMqttStatus({
                connected: false,
                message: 'Failed to check status'
            });
        } finally {
            setLoading(false);
        }
    };

    useEffect(() => {
        // Initial fetch
        fetchMQTTStatus();

        // Poll status every 5 seconds
        const interval = setInterval(fetchMQTTStatus, 5000);

        // Listen for MQTT messages (indicates connection is active)
        const unsubscribe = EventsOn('mqtt-message', () => {
            // When we receive MQTT messages, update status immediately
            fetchMQTTStatus();
        });

        return () => {
            clearInterval(interval);
            unsubscribe();
        };
    }, []);

    const handleRefresh = () => {
        setLoading(true);
        fetchMQTTStatus();
    };

    return (
        <div className="p-6">
            <Title level={2} style={{ color: '#4c1d95', marginBottom: 24 }}>
                System Status
            </Title>

            <Row gutter={[16, 16]}>
                <Col xs={24} lg={12}>
                    <Card
                        title={
                            <Space>
                                <ApiOutlined style={{ color: '#7c3aed' }} />
                                <span>MQTT Connection</span>
                            </Space>
                        }
                        extra={
                            <ReloadOutlined
                                onClick={handleRefresh}
                                style={{ cursor: 'pointer', fontSize: 16 }}
                                spin={loading}
                            />
                        }
                    >
                        {loading ? (
                            <div style={{ textAlign: 'center', padding: '20px 0' }}>
                                <Spin />
                            </div>
                        ) : (
                            <Space direction="vertical" size="large" style={{ width: '100%' }}>
                                <div style={{
                                    display: 'flex',
                                    alignItems: 'center',
                                    justifyContent: 'space-between'
                                }}>
                                    <Text strong style={{ fontSize: 16 }}>Status:</Text>
                                    <Tag
                                        icon={mqttStatus.connected ? <CheckCircleOutlined /> : <CloseCircleOutlined />}
                                        color={mqttStatus.connected ? 'success' : 'error'}
                                        style={{ fontSize: 14, padding: '4px 12px' }}
                                    >
                                        {mqttStatus.connected ? 'Connected' : 'Disconnected'}
                                    </Tag>
                                </div>

                                <div>
                                    <Text type="secondary">{mqttStatus.message}</Text>
                                </div>

                                {mqttStatus.connected && (
                                    <div style={{
                                        background: '#f6ffed',
                                        border: '1px solid #b7eb8f',
                                        borderRadius: 4,
                                        padding: 12,
                                        marginTop: 8
                                    }}>
                                        <Text style={{ color: '#52c41a' }}>
                                            ✓ Ready to receive real-time commands
                                        </Text>
                                    </div>
                                )}

                                {!mqttStatus.connected && (
                                    <div style={{
                                        background: '#fff1f0',
                                        border: '1px solid #ffccc7',
                                        borderRadius: 4,
                                        padding: 12,
                                        marginTop: 8
                                    }}>
                                        <Text style={{ color: '#ff4d4f' }}>
                                            ⓘ Please check your MQTT broker settings
                                        </Text>
                                    </div>
                                )}
                            </Space>
                        )}
                    </Card>
                </Col>

                <Col xs={24} lg={12}>
                    <Card
                        title={
                            <Space>
                                <CheckCircleOutlined style={{ color: '#52c41a' }} />
                                <span>Quick Info</span>
                            </Space>
                        }
                    >
                        <Space direction="vertical" size="middle" style={{ width: '100%' }}>
                            <div>
                                <Text strong>Welcome to ONX Screen Record</Text>
                            </div>
                            <div style={{ paddingTop: 8 }}>
                                <Text type="secondary">
                                    Your screen recording application is ready to use.
                                    The MQTT connection enables remote control and
                                    real-time synchronization with the backend system.
                                </Text>
                            </div>
                            <div style={{
                                background: '#f0f5ff',
                                border: '1px solid #adc6ff',
                                borderRadius: 4,
                                padding: 12,
                                marginTop: 8
                            }}>
                                <Text style={{ color: '#1890ff' }}>
                                    💡 Navigate to Recording to start capturing your screen
                                </Text>
                            </div>
                        </Space>
                    </Card>
                </Col>
            </Row>
        </div>
    );
};

export default Home;
