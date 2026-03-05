import React, { useState, useEffect, useRef } from 'react';
import { Card, Button, Typography, Space, Tag, Progress, message, Divider } from 'antd';
import {
    VideoCameraOutlined,
    PauseCircleOutlined,
    PlayCircleOutlined,
    SaveOutlined,
    LoadingOutlined,
    CheckCircleOutlined,
    ClockCircleOutlined,
    FolderOpenOutlined
} from '@ant-design/icons';
import {
    StartRecording,
    StopRecording,
    GetRecordingStatus
} from '../../wailsjs/go/app/App';

const { Title, Text } = Typography;

type RecordingState = 'idle' | 'recording' | 'processing' | 'playing' | 'error';

interface RecordingStatusData {
    state: RecordingState;
    duration: number;
    filePath: string;
    error: string;
}

const Recording: React.FC = () => {
    const [status, setStatus] = useState<RecordingStatusData>({
        state: 'idle',
        duration: 0,
        filePath: '',
        error: ''
    });
    const [isLoading, setIsLoading] = useState(false);
    const [lastSavedPath, setLastSavedPath] = useState<string>('');
    const timerRef = useRef<NodeJS.Timeout | null>(null);

    useEffect(() => {
        const initializeSession = async () => {
            try {
                const currentStatus = await GetRecordingStatus();
                if (currentStatus.state === 'recording') {
                    setStatus({
                        state: 'recording',
                        duration: currentStatus.duration,
                        filePath: currentStatus.filePath,
                        error: currentStatus.error
                    });
                } else {
                    setStatus({
                        state: 'idle',
                        duration: 0,
                        filePath: '',
                        error: ''
                    });
                }
            } catch (error) {
                console.error('Failed to initialize recording session:', error);
                setStatus({
                    state: 'idle',
                    duration: 0,
                    filePath: '',
                    error: ''
                });
            }
        };

        initializeSession();

        return () => {
            if (timerRef.current) {
                clearInterval(timerRef.current);
                timerRef.current = null;
            }
        };
    }, []);

    useEffect(() => {
        if (status.state === 'recording') {
            timerRef.current = setInterval(async () => {
                try {
                    const newStatus = await GetRecordingStatus();
                    setStatus({
                        state: newStatus.state as RecordingState,
                        duration: newStatus.duration,
                        filePath: newStatus.filePath,
                        error: newStatus.error
                    });
                } catch (error) {
                    console.error('Failed to get status:', error);
                }
            }, 1000);
        }

        return () => {
            if (timerRef.current) {
                clearInterval(timerRef.current);
            }
        };
    }, [status.state]);

    const formatDuration = (seconds: number): string => {
        const hrs = Math.floor(seconds / 3600);
        const mins = Math.floor((seconds % 3600) / 60);
        const secs = seconds % 60;

        if (hrs > 0) {
            return `${hrs.toString().padStart(2, '0')}:${mins.toString().padStart(2, '0')}:${secs.toString().padStart(2, '0')}`;
        }
        return `${mins.toString().padStart(2, '0')}:${secs.toString().padStart(2, '0')}`;
    };

    const handleStartRecording = async () => {
        setIsLoading(true);
        try {
            const result = await StartRecording();
            if (result.success) {
                setStatus({
                    state: 'recording',
                    duration: 0,
                    filePath: '',
                    error: ''
                });
                message.success('Recording started');
            } else {
                message.error(result.message || 'Failed to start recording');
            }
        } catch (error) {
            console.error('Failed to start recording:', error);
            message.error('Failed to start recording');
        } finally {
            setIsLoading(false);
        }
    };

    const handleStopRecording = async () => {
        setIsLoading(true);
        setStatus(prev => ({ ...prev, state: 'processing' }));

        try {
            const result = await StopRecording();
            if (result.success) {
                setStatus({
                    state: 'idle',
                    duration: 0,
                    filePath: result.filePath,
                    error: ''
                });
                setLastSavedPath(result.filePath);
                message.success('Recording saved successfully!');
            } else {
                message.error(result.message || 'Failed to stop recording');
                setStatus(prev => ({ ...prev, state: 'error', error: result.message }));
            }
        } catch (error) {
            console.error('Failed to stop recording:', error);
            message.error('Failed to stop recording');
            setStatus(prev => ({ ...prev, state: 'error', error: 'Failed to stop recording' }));
        } finally {
            setIsLoading(false);
        }
    };

    const getStateTag = () => {
        switch (status.state) {
            case 'recording':
                return <Tag color="red" icon={<VideoCameraOutlined />}>Recording</Tag>;
            case 'processing':
                return <Tag color="blue" icon={<LoadingOutlined spin />}>Processing</Tag>;
            case 'error':
                return <Tag color="error">Error</Tag>;
            default:
                return <Tag color="default" icon={<PauseCircleOutlined />}>Ready</Tag>;
        }
    };

    return (
        <div className="p-6" style={{ maxWidth: 800 }}>
            <Title level={2} style={{ color: '#171717', marginBottom: 24, fontWeight: 600 }}>
                <VideoCameraOutlined style={{ marginRight: 12 }} />
                Screen Recording
            </Title>

            <Card
                style={{ marginBottom: 24 }}
            >
                <div style={{ textAlign: 'center', padding: '32px 0' }}>
                    <div style={{ marginBottom: 24 }}>
                        {getStateTag()}
                    </div>

                    <div style={{
                        fontSize: 56,
                        fontWeight: 600,
                        color: status.state === 'recording' ? '#dc2626' : '#171717',
                        fontFamily: 'SF Mono, Monaco, Consolas, monospace',
                        marginBottom: 32,
                        letterSpacing: '-1px'
                    }}>
                        {formatDuration(status.duration)}
                    </div>

                    {status.state === 'recording' && (
                        <div style={{ marginBottom: 32 }}>
                            <div style={{
                                width: 12,
                                height: 12,
                                borderRadius: 2,
                                background: '#dc2626',
                                display: 'inline-block',
                                marginRight: 8
                            }} />
                            <Text type="secondary">Recording in progress</Text>
                        </div>
                    )}

                    <Space size="large">
                        {status.state === 'idle' ? (
                            <Button
                                type="primary"
                                size="large"
                                icon={<PlayCircleOutlined />}
                                loading={isLoading}
                                onClick={handleStartRecording}
                                className="h-14 px-8 text-base font-medium"
                            >
                                Start Recording
                            </Button>
                        ) : status.state === 'recording' ? (
                            <Button
                                danger
                                type="primary"
                                size="large"
                                icon={<PauseCircleOutlined />}
                                loading={isLoading}
                                onClick={handleStopRecording}
                                className="h-14 px-8 text-base font-medium"
                            >
                                Stop Recording
                            </Button>
                        ) : status.state === 'processing' ? (
                            <Button
                                size="large"
                                disabled
                                icon={<LoadingOutlined spin />}
                                className="h-14 px-8 text-base"
                            >
                                Processing...
                            </Button>
                        ) : null}
                    </Space>
                </div>
            </Card>

            {lastSavedPath && (
                <Card>
                    <Space direction="vertical" style={{ width: '100%' }}>
                        <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
                            <CheckCircleOutlined style={{ color: '#16a34a' }} />
                            <Text strong>Saved to:</Text>
                        </div>
                        <Text
                            code
                            style={{
                                display: 'block',
                                padding: 12,
                                backgroundColor: '#f5f5f5',
                                borderRadius: 6,
                                wordBreak: 'break-all',
                                fontSize: 13
                            }}
                        >
                            {lastSavedPath}
                        </Text>
                    </Space>
                </Card>
            )}

            {status.error && (
                <Card style={{ marginTop: 16, borderColor: '#dc2626' }}>
                    <Text type="danger">{status.error}</Text>
                </Card>
            )}
        </div>
    );
};

export default Recording;
