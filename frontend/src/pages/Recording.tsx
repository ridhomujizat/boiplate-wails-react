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

type RecordingState = 'idle' | 'recording' | 'processing' | 'error';

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
        // Poll status while recording
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
            <Title level={2} style={{ color: '#4c1d95', marginBottom: 24 }}>
                <VideoCameraOutlined style={{ marginRight: 12 }} />
                Screen Recording
            </Title>

            {/* Recording Controls */}
            <Card
                title={
                    <Space>
                        <VideoCameraOutlined style={{ color: '#7c3aed' }} />
                        <Text strong>Recording Controls</Text>
                    </Space>
                }
                style={{ marginBottom: 24 }}
            >
                <div style={{ textAlign: 'center', padding: '24px 0' }}>
                    {/* Status */}
                    <div style={{ marginBottom: 24 }}>
                        {getStateTag()}
                    </div>

                    {/* Timer Display */}
                    <div style={{
                        fontSize: 64,
                        fontWeight: 'bold',
                        color: status.state === 'recording' ? '#ef4444' : '#4c1d95',
                        fontFamily: 'monospace',
                        marginBottom: 24
                    }}>
                        <ClockCircleOutlined style={{ marginRight: 12, fontSize: 48 }} />
                        {formatDuration(status.duration)}
                    </div>

                    {/* Recording Animation */}
                    {status.state === 'recording' && (
                        <div style={{ marginBottom: 24 }}>
                            <Progress
                                type="circle"
                                percent={100}
                                status="active"
                                strokeColor="#ef4444"
                                format={() => (
                                    <div style={{ color: '#ef4444' }}>
                                        <VideoCameraOutlined style={{ fontSize: 32 }} />
                                    </div>
                                )}
                            />
                        </div>
                    )}

                    {/* Control Buttons */}
                    <Space size="large">
                        {status.state === 'idle' ? (
                            <Button
                                type="primary"
                                size="large"
                                icon={<PlayCircleOutlined />}
                                loading={isLoading}
                                onClick={handleStartRecording}
                                style={{
                                    backgroundColor: '#7c3aed',
                                    borderColor: '#7c3aed',
                                    height: 56,
                                    paddingLeft: 32,
                                    paddingRight: 32,
                                    fontSize: 18
                                }}
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
                                style={{
                                    height: 56,
                                    paddingLeft: 32,
                                    paddingRight: 32,
                                    fontSize: 18
                                }}
                            >
                                Stop Recording
                            </Button>
                        ) : status.state === 'processing' ? (
                            <Button
                                size="large"
                                disabled
                                icon={<LoadingOutlined spin />}
                                style={{
                                    height: 56,
                                    paddingLeft: 32,
                                    paddingRight: 32,
                                    fontSize: 18
                                }}
                            >
                                Processing...
                            </Button>
                        ) : null}
                    </Space>
                </div>
            </Card>

            {/* Last Recording Info */}
            {lastSavedPath && (
                <Card
                    title={
                        <Space>
                            <CheckCircleOutlined style={{ color: '#22c55e' }} />
                            <Text strong>Last Recording</Text>
                        </Space>
                    }
                >
                    <Space direction="vertical" style={{ width: '100%' }}>
                        <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
                            <FolderOpenOutlined style={{ color: '#7c3aed' }} />
                            <Text strong>Saved to:</Text>
                        </div>
                        <Text
                            code
                            style={{
                                display: 'block',
                                padding: 12,
                                backgroundColor: '#f5f5f5',
                                borderRadius: 6,
                                wordBreak: 'break-all'
                            }}
                        >
                            {lastSavedPath}
                        </Text>
                    </Space>
                </Card>
            )}

            {/* Error Display */}
            {status.error && (
                <Card style={{ marginTop: 16, borderColor: '#ef4444' }}>
                    <Text type="danger">{status.error}</Text>
                </Card>
            )}
        </div>
    );
};

export default Recording;
