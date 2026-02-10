import React, { useEffect, useState } from 'react';
import { Card, Row, Col, Typography, Space, Badge, Spin } from 'antd';
import {
    CheckCircleOutlined,
    CloseCircleOutlined,
    CloudOutlined,
    SyncOutlined,
    LoadingOutlined
} from '@ant-design/icons';
import { GetMQTTStatus } from '../../wailsjs/go/app/App';
import { EventsOn } from '../../wailsjs/runtime/runtime';
import { useAuth } from '../contexts/AuthContext';

const { Title, Text } = Typography;

interface MQTTStatus {
    connected: boolean;
    message: string;
}

type ConnectionState = 'connected' | 'disconnected' | 'connecting';

const Home: React.FC = () => {
    const { user } = useAuth();
    const [mqttStatus, setMqttStatus] = useState<MQTTStatus>({
        connected: false,
        message: 'Checking...'
    });
    const [connectionState, setConnectionState] = useState<ConnectionState>('connecting');

    const fetchMQTTStatus = async () => {
        try {
            const status = await GetMQTTStatus();
            setMqttStatus(status);

            // Determine connection state based on message
            if (status.connected) {
                setConnectionState('connected');
            } else if (
                status.message.includes('Checking') ||
                status.message.includes('initiated') ||
                status.message.includes('Attempting')
            ) {
                setConnectionState('connecting');
            } else {
                setConnectionState('disconnected');
            }
        } catch (error) {
            console.error('Failed to get MQTT status:', error);
            setMqttStatus({
                connected: false,
                message: 'Connection failed'
            });
            setConnectionState('disconnected');
        }
    };

    useEffect(() => {
        fetchMQTTStatus();
        const interval = setInterval(fetchMQTTStatus, 3000); // Check every 3 seconds
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

    // Get styling based on connection state
    const getConnectionStyle = () => {
        switch (connectionState) {
            case 'connected':
                return {
                    background: '#f6ffed',
                    border: '2px solid #b7eb8f',
                    iconBg: '#f6ffed',
                    iconColor: '#52c41a',
                    badgeStatus: 'success' as const,
                    badgeText: 'Connected',
                    message: 'Your device is connected with system.'
                };
            case 'connecting':
                return {
                    background: '#fffbe6',
                    border: '2px solid #ffe58f',
                    iconBg: '#fffbe6',
                    iconColor: '#faad14',
                    badgeStatus: 'processing' as const,
                    badgeText: 'Connecting...',
                    message: 'Establishing connection with system...'
                };
            case 'disconnected':
            default:
                return {
                    background: '#fff1f0',
                    border: '2px solid #ffccc7',
                    iconBg: '#fff1f0',
                    iconColor: '#ff4d4f',
                    badgeStatus: 'error' as const,
                    badgeText: 'Not Connected',
                    message: 'Your device is not connected with system.'
                };
        }
    };

    const style = getConnectionStyle();

    // Get icon based on connection state
    const getStatusIcon = () => {
        switch (connectionState) {
            case 'connected':
                return <CheckCircleOutlined style={{ fontSize: 32, color: '#52c41a' }} />;
            case 'connecting':
                return <LoadingOutlined style={{ fontSize: 32, color: '#faad14' }} spin />;
            case 'disconnected':
            default:
                return <CloseCircleOutlined style={{ fontSize: 32, color: '#ff4d4f' }} />;
        }
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
                <Col xs={24} md={24}>
                    <Card
                        style={{
                            borderRadius: 12,
                            height: '100%',
                            background: style.background,
                            border: style.border
                        }}
                    >
                        <Space direction="vertical" size="large" style={{ width: '100%' }}>
                            <div style={{ display: 'flex', alignItems: 'center', gap: 12 }}>
                                <div style={{
                                    width: 48,
                                    height: 48,
                                    borderRadius: 12,
                                    background: style.iconBg,
                                    display: 'flex',
                                    alignItems: 'center',
                                    justifyContent: 'center'
                                }}>
                                    <CloudOutlined
                                        style={{
                                            fontSize: 24,
                                            color: style.iconColor
                                        }}
                                    />
                                </div>
                                <div style={{ flex: 1 }}>
                                    <Text strong style={{ fontSize: 18, display: 'block' }}>
                                        Connection
                                    </Text>
                                    <Badge
                                        status={style.badgeStatus}
                                        text={style.badgeText}
                                        style={{ fontSize: 14 }}
                                    />
                                </div>
                                {getStatusIcon()}
                            </div>
                            <div style={{
                                padding: 16,
                                borderRadius: 8,
                                background: 'rgba(255, 255, 255, 0.8)'
                            }}>
                                <Text style={{ color: style.iconColor }}>
                                    {style.message}
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
