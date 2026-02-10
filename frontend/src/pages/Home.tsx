import React, { useEffect, useState } from 'react';
import { Card, Row, Col, Typography, Space, Badge } from 'antd';
import {
    CheckCircleOutlined,
    CloseCircleOutlined,
    VideoCameraOutlined,
    CloudOutlined
} from '@ant-design/icons';
import { GetMQTTStatus } from '../../wailsjs/go/app/App';
import { EventsOn } from '../../wailsjs/runtime/runtime';
import { useAuth } from '../contexts/AuthContext';

const { Title, Text, Paragraph } = Typography;

interface MQTTStatus {
    connected: boolean;
    message: string;
}

const Home: React.FC = () => {
    const { user } = useAuth();
    const [mqttStatus, setMqttStatus] = useState<MQTTStatus>({
        connected: false,
        message: 'Checking...'
    });

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
        }
    };

    useEffect(() => {
        fetchMQTTStatus();
        const interval = setInterval(fetchMQTTStatus, 5000);
        const unsubscribe = EventsOn('mqtt-message', fetchMQTTStatus);

        return () => {
            clearInterval(interval);
            unsubscribe();
        };
    }, []);

    const getGreeting = () => {
        const hour = new Date().getHours();
        if (hour < 12) return 'Good Morning';
        if (hour < 18) return 'Good Afternoon';
        return 'Good Evening';
    };

    return (
        <div className="p-6">
            {/* Welcome Section */}
            <div style={{ marginBottom: 32 }}>
                <Title level={2} style={{ color: '#4c1d95', marginBottom: 8 }}>
                    {getGreeting()}{user?.email ? `, ${user.email.split('@')[0]}` : ''}!
                </Title>
                <Text type="secondary" style={{ fontSize: 16 }}>
                    Welcome to ONX Screen Record
                </Text>
            </div>

            <Row gutter={[24, 24]}>
                {/* Connection Status Card */}
                <Col xs={24} md={12}>
                    <Card
                        style={{
                            borderRadius: 12,
                            height: '100%',
                            background: mqttStatus.connected
                                ? 'linear-gradient(135deg, #f6ffed 0%, #ffffff 100%)'
                                : 'linear-gradient(135deg, #fff1f0 0%, #ffffff 100%)',
                            border: mqttStatus.connected
                                ? '2px solid #b7eb8f'
                                : '2px solid #ffccc7'
                        }}
                    >
                        <Space direction="vertical" size="large" style={{ width: '100%' }}>
                            <div style={{ display: 'flex', alignItems: 'center', gap: 12 }}>
                                <div style={{
                                    width: 48,
                                    height: 48,
                                    borderRadius: 12,
                                    background: mqttStatus.connected ? '#f6ffed' : '#fff1f0',
                                    display: 'flex',
                                    alignItems: 'center',
                                    justifyContent: 'center'
                                }}>
                                    <CloudOutlined
                                        style={{
                                            fontSize: 24,
                                            color: mqttStatus.connected ? '#52c41a' : '#ff4d4f'
                                        }}
                                    />
                                </div>
                                <div style={{ flex: 1 }}>
                                    <Text strong style={{ fontSize: 18, display: 'block' }}>
                                        Connection
                                    </Text>
                                    <Badge
                                        status={mqttStatus.connected ? 'success' : 'error'}
                                        text={mqttStatus.connected ? 'Connected' : 'Not Connected'}
                                        style={{ fontSize: 14 }}
                                    />
                                </div>
                                {mqttStatus.connected ? (
                                    <CheckCircleOutlined
                                        style={{ fontSize: 32, color: '#52c41a' }}
                                    />
                                ) : (
                                    <CloseCircleOutlined
                                        style={{ fontSize: 32, color: '#ff4d4f' }}
                                    />
                                )}
                            </div>
                        </Space>
                    </Card>
                </Col>
            </Row>

        </div>
    );
};

export default Home;
