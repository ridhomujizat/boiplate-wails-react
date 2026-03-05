import React, { useState, useEffect, useCallback } from 'react';
import { Card, Row, Col, Statistic, Typography, Table, Progress, DatePicker, Button, Space, Tooltip } from 'antd';
import {
    ClockCircleOutlined,
    AppstoreOutlined,
    TrophyOutlined,
    PauseCircleOutlined,
    ReloadOutlined,
    PlayCircleOutlined,
    StopOutlined
} from '@ant-design/icons';
import dayjs, { Dayjs } from 'dayjs';
import {
    GetActivityTimeline,
    GetTopApplications,
    GetActivitySummary,
    GetDashboardStats,
    StartActivityTracking,
    StopActivityTracking,
    IsActivityTrackingActive
} from '../../wailsjs/go/app/App';
import { app } from '../../wailsjs/go/models';

const { Title } = Typography;
const { RangePicker } = DatePicker;

type DateRange = app.DateRange;
type TimelineEvent = app.TimelineEvent;
type TopApplication = app.TopApplication;
type ActivitySummary = app.ActivitySummary;
type DashboardStats = app.DashboardStats;

const formatDuration = (seconds: number): string => {
    if (seconds < 60) return `${seconds}s`;
    const hours = Math.floor(seconds / 3600);
    const minutes = Math.floor((seconds % 3600) / 60);
    if (hours > 0) return `${hours}h ${minutes}m`;
    return `${minutes}m`;
};

const Activity: React.FC = () => {
    const [dateRange, setDateRange] = useState<[Dayjs, Dayjs]>([dayjs(), dayjs()]);
    const [stats, setStats] = useState<DashboardStats>({ totalActiveTime: 0, totalAfkTime: 0, totalApps: 0, topApp: '' });
    const [timeline, setTimeline] = useState<TimelineEvent[]>([]);
    const [topApps, setTopApps] = useState<TopApplication[]>([]);
    const [summary, setSummary] = useState<ActivitySummary[]>([]);
    const [isTracking, setIsTracking] = useState(false);
    const [loading, setLoading] = useState(false);

    const getDateRangeParam = useCallback((): DateRange => ({
        startDate: dateRange[0].format('YYYY-MM-DD'),
        endDate: dateRange[1].format('YYYY-MM-DD')
    }), [dateRange]);

    const fetchData = useCallback(async () => {
        setLoading(true);
        try {
            const range = getDateRangeParam();
            const [statsData, timelineData, topAppsData, summaryData, trackingStatus] = await Promise.all([
                GetDashboardStats(range),
                GetActivityTimeline(range),
                GetTopApplications(range, 10),
                GetActivitySummary(range),
                IsActivityTrackingActive()
            ]);
            setStats(statsData);
            setTimeline(timelineData);
            setTopApps(topAppsData);
            setSummary(summaryData);
            setIsTracking(trackingStatus);
        } catch {
        } finally {
            setLoading(false);
        }
    }, [getDateRangeParam]);

    useEffect(() => {
        fetchData();
    }, [fetchData]);

    const handleToggleTracking = async () => {
        if (isTracking) {
            await StopActivityTracking();
        } else {
            await StartActivityTracking();
        }
        setIsTracking(!isTracking);
    };

    const handleDateRangeChange = (dates: [Dayjs | null, Dayjs | null] | null) => {
        if (dates && dates[0] && dates[1]) {
            setDateRange([dates[0], dates[1]]);
        }
    };

    const summaryColumns = [
        { title: 'Application', dataIndex: 'appName', key: 'appName', ellipsis: true },
        { title: 'Active Time', dataIndex: 'activeTime', key: 'activeTime', render: (v: number) => formatDuration(v), width: 100 },
        { title: 'AFK Time', dataIndex: 'afkTime', key: 'afkTime', render: (v: number) => formatDuration(v), width: 100 },
        { title: 'Sessions', dataIndex: 'sessionCount', key: 'sessionCount', width: 80 },
        {
            title: 'Usage',
            dataIndex: 'percentage',
            key: 'percentage',
            width: 150,
            render: (v: number) => <Progress percent={Math.round(v)} size="small" strokeColor="#2851e6" showInfo={false} />
        }
    ];

    const getTimelineBlocks = () => {
        if (timeline.length === 0) return null;

        const hourWidth = 100 / 24;
        return (
            <div style={{ position: 'relative', height: 40, background: '#f5f5f5', borderRadius: 6, overflow: 'hidden' }}>
                {timeline.map((event, idx) => {
                    const start = dayjs(event.startTime);
                    const end = event.endTime ? dayjs(event.endTime) : dayjs();
                    const startHour = start.hour() + start.minute() / 60;
                    const duration = end.diff(start, 'minute') / 60;
                    const left = startHour * hourWidth;
                    const width = Math.max(duration * hourWidth, 0.5);

                    return (
                        <Tooltip key={idx} title={`${event.appName}: ${formatDuration(event.duration)}`}>
                            <div
                                style={{
                                    position: 'absolute',
                                    left: `${left}%`,
                                    width: `${width}%`,
                                    height: '100%',
                                    background: event.status === 'active' ? '#22c55e' : '#d4d4d4',
                                    borderRight: '1px solid #ffffff'
                                }}
                            />
                        </Tooltip>
                    );
                })}
            </div>
        );
    };

    const getHourLabels = () => {
        const hours = [0, 6, 12, 18, 24];
        return (
            <div style={{ display: 'flex', justifyContent: 'space-between', fontSize: 12, color: '#a3a3a3', marginTop: 8 }}>
                {hours.map(h => <span key={h}>{h === 24 ? '24:00' : `${h}:00`}</span>)}
            </div>
        );
    };

    return (
        <div className="p-6">
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 24 }}>
                <Title level={2} style={{ color: '#171717', margin: 0, fontWeight: 600 }}>Activity Monitor</Title>
                <Space>
                    <RangePicker
                        value={dateRange}
                        onChange={handleDateRangeChange}
                        allowClear={false}
                    />
                    <Button icon={<ReloadOutlined />} onClick={fetchData} loading={loading}>Refresh</Button>
                    <Button
                        type={isTracking ? 'default' : 'primary'}
                        icon={isTracking ? <StopOutlined /> : <PlayCircleOutlined />}
                        onClick={handleToggleTracking}
                        danger={isTracking}
                        className="font-medium"
                    >
                        {isTracking ? 'Stop Tracking' : 'Start Tracking'}
                    </Button>
                </Space>
            </div>

            <Row gutter={[16, 16]} className="mb-6">
                <Col xs={24} sm={12} md={6}>
                    <Card>
                        <Statistic
                            title="Active Time"
                            value={formatDuration(stats.totalActiveTime)}
                            prefix={<ClockCircleOutlined style={{ color: '#22c55e' }} />}
                            valueStyle={{ color: '#171717', fontWeight: 600 }}
                        />
                    </Card>
                </Col>
                <Col xs={24} sm={12} md={6}>
                    <Card>
                        <Statistic
                            title="AFK Time"
                            value={formatDuration(stats.totalAfkTime)}
                            prefix={<PauseCircleOutlined style={{ color: '#f59e0b' }} />}
                            valueStyle={{ color: '#171717', fontWeight: 600 }}
                        />
                    </Card>
                </Col>
                <Col xs={24} sm={12} md={6}>
                    <Card>
                        <Statistic
                            title="Apps Used"
                            value={stats.totalApps}
                            prefix={<AppstoreOutlined style={{ color: '#2851e6' }} />}
                            valueStyle={{ color: '#171717', fontWeight: 600 }}
                        />
                    </Card>
                </Col>
                <Col xs={24} sm={12} md={6}>
                    <Card>
                        <Statistic
                            title="Most Used"
                            value={stats.topApp || '-'}
                            prefix={<TrophyOutlined style={{ color: '#7c3aed' }} />}
                            valueStyle={{ color: '#171717', fontSize: 16, fontWeight: 600 }}
                        />
                    </Card>
                </Col>
            </Row>

            <Card title="Timeline" className="mb-4" style={{ marginBottom: 16 }}>
                {getTimelineBlocks()}
                {getHourLabels()}
                <div style={{ display: 'flex', gap: 20, marginTop: 16, fontSize: 13 }}>
                    <span style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
                        <span style={{ display: 'inline-block', width: 10, height: 10, background: '#22c55e', borderRadius: 2 }} />
                        <span className="text-neutral-600">Active</span>
                    </span>
                    <span style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
                        <span style={{ display: 'inline-block', width: 10, height: 10, background: '#d4d4d4', borderRadius: 2 }} />
                        <span className="text-neutral-600">AFK</span>
                    </span>
                </div>
            </Card>

            <Row gutter={[16, 16]}>
                <Col xs={24} md={8}>
                    <Card title="Top Applications">
                        {topApps.map((app, idx) => (
                            <div key={idx} style={{ marginBottom: 16 }}>
                                <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: 8 }}>
                                    <span style={{ fontWeight: 500, color: '#171717' }}>{app.appName}</span>
                                    <span style={{ color: '#737373', fontSize: 13 }}>{formatDuration(app.totalDuration)}</span>
                                </div>
                                <Progress percent={Math.round(app.percentage)} strokeColor="#2851e6" showInfo={false} trailColor="#f5f5f5" />
                            </div>
                        ))}
                        {topApps.length === 0 && <div style={{ color: '#a3a3a3', textAlign: 'center', padding: '20px 0' }}>No data</div>}
                    </Card>
                </Col>
                <Col xs={24} md={16}>
                    <Card title="Activity Summary">
                        <Table
                            dataSource={summary}
                            columns={summaryColumns}
                            rowKey="appName"
                            pagination={false}
                            size="small"
                            scroll={{ y: 300 }}
                        />
                    </Card>
                </Col>
            </Row>
        </div>
    );
};

export default Activity;
