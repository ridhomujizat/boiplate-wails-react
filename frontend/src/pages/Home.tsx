import React, { useEffect, useState } from 'react';
import { Card, Row, Col, Typography, Space, Badge } from 'antd';
import {
    CheckCircleOutlined,
    CloseCircleOutlined,
    CloudOutlined,
    LoadingOutlined
} from '@ant-design/icons';
import { GetMQTTStatus } from '../../wailsjs/go/app/App';
import { EventsOn } from '../../wailsjs/runtime/runtime';
import { useAuth } from '../contexts/AuthContext';

const { Title, Text } = Typography;

interface MQTTStatus {
    connected: boolean;
    state: string;
    message: string;
}

type ConnectionState = 'connected' | 'disconnected' | 'connecting' | 'reconnecting';

const Home: React.FC = () => {
    const { user } = useAuth();
    const [mqttStatus, setMqttStatus] = useState<MQTTStatus>({
        connected: false,
        state: 'disconnected',
        message: 'Checking...'
    });
    const [connectionState, setConnectionState] = useState<ConnectionState>('connecting');

    const fetchMQTTStatus = async () => {
        try {
            const status = await GetMQTTStatus();
            setMqttStatus(status);
            setConnectionState(status.state as ConnectionState);
        } catch (error) {
            console.error('Failed to get MQTT status:', error);
            setMqttStatus({
                connected: false,
                state: 'disconnected',
                message: 'Connection failed'
            });
            setConnectionState('disconnected');
        }
    };

    useEffect(() => {
        fetchMQTTStatus();

        const unsubscribeStatus = EventsOn('mqtt-status', (data: { state: string }) => {
            const state = data.state as ConnectionState;
            setConnectionState(state);
            setMqttStatus(prev => ({
                ...prev,
                connected: state === 'connected',
                state: state,
            }));
        });

        return () => {
            unsubscribeStatus();
        };
    }, []);

    const getGreeting = () => {
        const hour = new Date().getHours();
        if (hour < 12) return 'Good Morning';
        if (hour < 18) return 'Good Afternoon';
        return 'Good Evening';
    };

    const getConnectionStyle = () => {
        switch (connectionState) {
            case 'connected':
                return {
                    background: '#f0fdf4',
                    border: '1px solid #bbf7d0',
                    iconBg: '#dcfce7',
                    iconColor: '#16a34a',
                    badgeStatus: 'success' as const,
                    badgeText: 'Connected',
                    message: 'Your device is connected with system.'
                };
            case 'connecting':
                return {
                    background: '#fefce8',
                    border: '1px solid #fef08a',
                    iconBg: '#fef9c3',
                    iconColor: '#ca8a04',
                    badgeStatus: 'processing' as const,
                    badgeText: 'Connecting...',
                    message: 'Establishing connection with system...'
                };
            case 'reconnecting':
                return {
                    background: '#fefce8',
                    border: '1px solid #fef08a',
                    iconBg: '#fef9c3',
                    iconColor: '#ca8a04',
                    badgeStatus: 'processing' as const,
                    badgeText: 'Reconnecting...',
                    message: 'Connection lost. Reconnecting to system...'
                };
            case 'disconnected':
            default:
                return {
                    background: '#fef2f2',
                    border: '1px solid #fecaca',
                    iconBg: '#fee2e2',
                    iconColor: '#dc2626',
                    badgeStatus: 'error' as const,
                    badgeText: 'Not Connected',
                    message: 'Your device is not connected with system.'
                };
        }
    };

    const style = getConnectionStyle();

    const getStatusIcon = () => {
        switch (connectionState) {
            case 'connected':
                return <CheckCircleOutlined style={{ fontSize: 28, color: '#16a34a' }} />;
            case 'connecting':
            case 'reconnecting':
                return <LoadingOutlined style={{ fontSize: 28, color: '#ca8a04' }} spin />;
            case 'disconnected':
            default:
                return <CloseCircleOutlined style={{ fontSize: 28, color: '#dc2626' }} />;
        }
    };

    return (
        <div className="p-6">
            <div style={{ marginBottom: 32 }}>
                <Title level={2} style={{ color: '#171717', marginBottom: 8, fontWeight: 600 }}>
                    {getGreeting()}{user?.email ? `, ${user.email.split('@')[0]}` : ''}!
                </Title>
                <Text type="secondary" style={{ fontSize: 15 }}>
                    Welcome to ONX Screen Record
                </Text>
            </div>

            <Row gutter={[24, 24]}>
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
                            <div style={{ display: 'flex', alignItems: 'center', gap: 16 }}>
                                <div style={{
                                    width: 48,
                                    height: 48,
                                    borderRadius: 10,
                                    background: style.iconBg,
                                    display: 'flex',
                                    alignItems: 'center',
                                    justifyContent: 'center'
                                }}>
                                    <CloudOutlined
                                        style={{
                                            fontSize: 22,
                                            color: style.iconColor
                                        }}
                                    />
                                </div>
                                <div style={{ flex: 1 }}>
                                    <Text strong style={{ fontSize: 16, display: 'block', color: '#171717' }}>
                                        Connection Status
                                    </Text>
                                    <Badge
                                        status={style.badgeStatus}
                                        text={style.badgeText}
                                        style={{ fontSize: 13 }}
                                    />
                                </div>
                                {getStatusIcon()}
                            </div>
                            <div style={{
                                padding: 14,
                                borderRadius: 8,
                                background: 'rgba(255, 255, 255, 0.6)'
                            }}>
                                <Text style={{ color: style.iconColor, fontSize: 14 }}>
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
