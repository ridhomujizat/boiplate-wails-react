import React, { useCallback, useEffect, useMemo, useState } from 'react';
import { Alert, Button, Card, Col, Empty, Row, Space, Table, Tag, Tooltip, Typography } from 'antd';
import {
    CheckCircleOutlined,
    ClockCircleOutlined,
    ExclamationCircleOutlined,
    HistoryOutlined,
    ReloadOutlined,
    SyncOutlined
} from '@ant-design/icons';
import { EventsOn } from '../../wailsjs/runtime/runtime';
import { GetUploadHistory, type UploadHistoryItem } from '../lib/uploadHistory';

const { Title, Text } = Typography;

const formatFileSize = (bytes: number): string => {
    if (bytes <= 0) {
        return '0 B';
    }

    const units = ['B', 'KB', 'MB', 'GB'];
    const index = Math.min(Math.floor(Math.log(bytes) / Math.log(1024)), units.length - 1);
    const value = bytes / Math.pow(1024, index);
    return `${value.toFixed(index === 0 ? 0 : 1)} ${units[index]}`;
};

const formatDateTime = (value: string): string => {
    if (!value) {
        return '-';
    }

    const date = new Date(value);
    if (Number.isNaN(date.getTime())) {
        return value;
    }

    return new Intl.DateTimeFormat('en-US', {
        year: 'numeric',
        month: 'short',
        day: '2-digit',
        hour: '2-digit',
        minute: '2-digit',
        second: '2-digit',
    }).format(date);
};

const formatCompactDateTime = (value: string): string => {
    if (!value) {
        return '-';
    }

    const date = new Date(value);
    if (Number.isNaN(date.getTime())) {
        return value;
    }

    return new Intl.DateTimeFormat('en-US', {
        month: 'short',
        day: '2-digit',
        hour: '2-digit',
        minute: '2-digit',
    }).format(date);
};

const getStatusTag = (status: string) => {
    switch (status) {
        case 'done':
            return <Tag color="success">Done</Tag>;
        case 'retry_wait':
            return <Tag color="warning">Retry Waiting</Tag>;
        case 'blocked_auth':
            return <Tag color="gold">Blocked Auth</Tag>;
        case 'uploaded_unconfirmed':
            return <Tag color="processing">Uploaded</Tag>;
        case 'missing_file':
            return <Tag color="error">Missing File</Tag>;
        case 'dead':
            return <Tag color="error">Dead</Tag>;
        case 'putting':
            return <Tag color="processing">Uploading</Tag>;
        case 'initing':
            return <Tag color="blue">Initializing</Tag>;
        default:
            return <Tag>{status}</Tag>;
    }
};

const StatCard: React.FC<{
    title: string;
    value: number;
    color: string;
    icon: React.ReactNode;
}> = ({ title, value, color, icon }) => (
    <Card style={{ borderRadius: 12, border: '1px solid #e5e5e5', boxShadow: 'none' }}>
        <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
            <div>
                <Text style={{ color: '#737373', fontSize: 12, textTransform: 'uppercase', fontWeight: 600 }}>
                    {title}
                </Text>
                <div style={{ color: '#171717', fontSize: 32, fontWeight: 700, marginTop: 6 }}>
                    {value}
                </div>
            </div>
            <div
                style={{
                    width: 48,
                    height: 48,
                    borderRadius: 10,
                    background: `${color}15`,
                    display: 'flex',
                    alignItems: 'center',
                    justifyContent: 'center',
                    color,
                }}
            >
                {icon}
            </div>
        </div>
    </Card>
);

const RecordHistory: React.FC = () => {
    const [items, setItems] = useState<UploadHistoryItem[]>([]);
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState<string>('');

    const fetchHistory = useCallback(async () => {
        setLoading(true);
        try {
            const data = await GetUploadHistory();
            setItems(data);
            setError('');
        } catch (err) {
            console.error('Failed to load upload history:', err);
            setError('Failed to load upload history');
        } finally {
            setLoading(false);
        }
    }, []);

    useEffect(() => {
        void fetchHistory();

        const unsubscribe = EventsOn('recording-uploaded', () => {
            void fetchHistory();
        });

        return () => {
            unsubscribe();
        };
    }, [fetchHistory]);

    const stats = useMemo(() => {
        return items.reduce(
            (acc, item) => {
                acc.total += 1;
                if (item.status === 'done') {
                    acc.done += 1;
                }
                if (item.status === 'retry_wait' || item.status === 'blocked_auth') {
                    acc.pending += 1;
                }
                if (item.status === 'dead' || item.status === 'missing_file') {
                    acc.failed += 1;
                }
                return acc;
            },
            { total: 0, done: 0, pending: 0, failed: 0 }
        );
    }, [items]);

    const columns = [
        {
            title: 'Session',
            dataIndex: 'sessionId',
            key: 'sessionId',
            width: 180,
            ellipsis: true,
            render: (value: string) => (
                <Tooltip title={value}>
                    <Text style={{ fontFamily: 'SF Mono, Monaco, Consolas, monospace' }}>{value || '-'}</Text>
                </Tooltip>
            ),
        },
        {
            title: 'File',
            dataIndex: 'filename',
            key: 'filename',
            width: 340,
            render: (_: string, record: UploadHistoryItem) => (
                <Space direction="vertical" size={0}>
                    <Tooltip title={record.filename}>
                        <Text strong style={{ color: '#171717', maxWidth: 280 }} ellipsis>
                            {record.filename || '-'}
                        </Text>
                    </Tooltip>
                    <Tooltip title={record.filePath}>
                        <Text type="secondary" style={{ maxWidth: 300, display: 'inline-block' }} ellipsis>
                            {record.filePath}
                        </Text>
                    </Tooltip>
                    <Text type="secondary" style={{ fontSize: 12 }}>
                        {formatFileSize(record.fileSize)} • {record.contentType || '-'}
                    </Text>
                </Space>
            ),
        },
        {
            title: 'Status',
            dataIndex: 'status',
            key: 'status',
            width: 220,
            render: (value: string, record: UploadHistoryItem) => (
                <Space direction="vertical" size={4}>
                    {getStatusTag(value)}
                    {record.lastError ? (
                        <Tooltip title={record.lastError}>
                            <Text style={{ color: '#b91c1c', maxWidth: 180, display: 'inline-block' }} ellipsis>
                                {record.lastHttpStatus > 0 ? `[${record.lastHttpStatus}] ${record.lastError}` : record.lastError}
                            </Text>
                        </Tooltip>
                    ) : (
                        <Text type="secondary">No error</Text>
                    )}
                </Space>
            ),
        },
        {
            title: 'Attempts',
            dataIndex: 'attemptCount',
            key: 'attemptCount',
            width: 100,
            align: 'center' as const,
        },
        {
            title: 'Schedule',
            key: 'schedule',
            width: 190,
            render: (_: unknown, record: UploadHistoryItem) => {
                const label = record.nextAttemptAt || record.confirmedAt || record.gcsUploadedAt || record.updatedAt;
                const caption = record.nextAttemptAt
                    ? 'Next attempt'
                    : record.confirmedAt
                        ? 'Confirmed'
                        : record.gcsUploadedAt
                            ? 'Uploaded'
                            : 'Updated';

                return (
                    <Space direction="vertical" size={0}>
                        <Text style={{ color: '#171717' }}>{formatCompactDateTime(label)}</Text>
                        <Text type="secondary" style={{ fontSize: 12 }}>{caption}</Text>
                    </Space>
                );
            },
        },
        {
            title: 'Updated',
            dataIndex: 'updatedAt',
            key: 'updatedAt',
            width: 160,
            render: (value: string) => (
                <Space direction="vertical" size={0}>
                    <Text style={{ color: '#171717' }}>{formatCompactDateTime(value)}</Text>
                    <Text type="secondary" style={{ fontSize: 12 }}>{formatDateTime(value)}</Text>
                </Space>
            ),
        },
    ];

    return (
        <div className="p-6" style={{ maxWidth: 1400, margin: '0 auto' }}>
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', gap: 16, marginBottom: 24, flexWrap: 'wrap' }}>
                <div>
                    <Title level={2} style={{ color: '#171717', margin: 0, fontWeight: 700 }}>
                        Record History
                    </Title>
                    <Text type="secondary">
                        Local upload queue history from the `upload_jobs` database.
                    </Text>
                </div>
                <Button
                    icon={<ReloadOutlined />}
                    onClick={() => void fetchHistory()}
                    loading={loading}
                    size="large"
                    style={{ borderRadius: 8 }}
                >
                    Refresh
                </Button>
            </div>

            {error ? (
                <Alert
                    type="error"
                    message={error}
                    showIcon
                    style={{ marginBottom: 24, borderRadius: 10 }}
                />
            ) : null}

            <Row gutter={[20, 20]} style={{ marginBottom: 24 }}>
                <Col xs={24} sm={12} lg={6}>
                    <StatCard title="Total Jobs" value={stats.total} color="#2851e6" icon={<HistoryOutlined style={{ fontSize: 22 }} />} />
                </Col>
                <Col xs={24} sm={12} lg={6}>
                    <StatCard title="Completed" value={stats.done} color="#16a34a" icon={<CheckCircleOutlined style={{ fontSize: 22 }} />} />
                </Col>
                <Col xs={24} sm={12} lg={6}>
                    <StatCard title="Pending" value={stats.pending} color="#ca8a04" icon={<SyncOutlined spin style={{ fontSize: 22 }} />} />
                </Col>
                <Col xs={24} sm={12} lg={6}>
                    <StatCard title="Failed" value={stats.failed} color="#dc2626" icon={<ExclamationCircleOutlined style={{ fontSize: 22 }} />} />
                </Col>
            </Row>

            <Card style={{ borderRadius: 12, border: '1px solid #e5e5e5', boxShadow: 'none' }}>
                <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 16, gap: 12, flexWrap: 'wrap' }}>
                    <Space size="middle">
                        <Text strong style={{ color: '#171717', fontSize: 16 }}>Upload Queue History</Text>
                        <Tag icon={<ClockCircleOutlined />} color="default">
                            Auto-refresh on upload events
                        </Tag>
                    </Space>
                </div>

                <Table<UploadHistoryItem>
                    rowKey="id"
                    columns={columns}
                    dataSource={items}
                    loading={loading}
                    tableLayout="fixed"
                    scroll={{ x: 1100 }}
                    locale={{
                        emptyText: (
                            <Empty
                                image={Empty.PRESENTED_IMAGE_SIMPLE}
                                description="No upload history yet"
                            />
                        ),
                    }}
                    pagination={{ pageSize: 8, showSizeChanger: false }}
                    expandable={{
                        expandedRowRender: (record) => (
                            <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(220px, 1fr))', gap: 12 }}>
                                <div>
                                    <Text type="secondary">Upload ID</Text>
                                    <div><Text code>{record.uploadId || '-'}</Text></div>
                                </div>
                                <div>
                                    <Text type="secondary">Full Error</Text>
                                    <div><Text>{record.lastError || '-'}</Text></div>
                                </div>
                                <div>
                                    <Text type="secondary">HTTP Status</Text>
                                    <div><Text>{record.lastHttpStatus || '-'}</Text></div>
                                </div>
                                <div>
                                    <Text type="secondary">Content Type</Text>
                                    <div><Text>{record.contentType || '-'}</Text></div>
                                </div>
                                <div>
                                    <Text type="secondary">File Size</Text>
                                    <div><Text>{formatFileSize(record.fileSize)}</Text></div>
                                </div>
                                <div>
                                    <Text type="secondary">Created</Text>
                                    <div><Text>{formatDateTime(record.createdAt)}</Text></div>
                                </div>
                                <div>
                                    <Text type="secondary">Last Attempt</Text>
                                    <div><Text>{formatDateTime(record.lastAttemptAt)}</Text></div>
                                </div>
                                <div>
                                    <Text type="secondary">Signed URL Expires</Text>
                                    <div><Text>{formatDateTime(record.signedUrlExpiresAt)}</Text></div>
                                </div>
                                <div>
                                    <Text type="secondary">Uploaded To GCS</Text>
                                    <div><Text>{formatDateTime(record.gcsUploadedAt)}</Text></div>
                                </div>
                                <div>
                                    <Text type="secondary">Confirmed</Text>
                                    <div><Text>{formatDateTime(record.confirmedAt)}</Text></div>
                                </div>
                                <div style={{ gridColumn: '1 / -1' }}>
                                    <Text type="secondary">Local File Path</Text>
                                    <div><Text code>{record.filePath}</Text></div>
                                </div>
                            </div>
                        ),
                    }}
                />
            </Card>
        </div>
    );
};

export default RecordHistory;
