import React, { useState, useEffect, useCallback } from 'react';
import { Card, Row, Col, Typography, Table, Progress, DatePicker, Button, Space, Tooltip, Empty } from 'antd';
import {
    ClockCircleOutlined,
    AppstoreOutlined,
    TrophyOutlined,
    PauseCircleOutlined,
    ReloadOutlined,
    PlayCircleOutlined,
    StopOutlined,
    ThunderboltOutlined,
    DashboardOutlined
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

const { Title, Text } = Typography;
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
        { 
            title: 'Application', 
            dataIndex: 'appName', 
            key: 'appName',
            ellipsis: true,
            render: (text: string) => <span style={{ fontWeight: 500, color: '#171717' }}>{text}</span>
        },
        { 
            title: 'Active Time', 
            dataIndex: 'activeTime', 
            key: 'activeTime', 
            render: (v: number) => <span style={{ color: '#22c55e', fontWeight: 600 }}>{formatDuration(v)}</span>,
            width: 100 
        },
        { 
            title: 'AFK Time', 
            dataIndex: 'afkTime', 
            key: 'afkTime', 
            render: (v: number) => <span style={{ color: '#f59e0b' }}>{formatDuration(v)}</span>,
            width: 100 
        },
        { 
            title: 'Sessions', 
            dataIndex: 'sessionCount', 
            key: 'sessionCount',
            render: (v: number) => <span style={{ color: '#737373' }}>{v}</span>,
            width: 90 
        },
        {
            title: 'Usage',
            dataIndex: 'percentage',
            key: 'percentage',
            width: 140,
            render: (v: number) => (
                <Progress 
                    percent={Math.round(v)} 
                    size="small" 
                    strokeColor={{
                        '0%': '#2851e6',
                        '100%': '#1d307a',
                    }}
                    trailColor="#f0f0f0"
                    showInfo={false}
                    strokeWidth={6}
                />
            )
        }
    ];

    const getTimelineBlocks = () => {
        if (timeline.length === 0) {
            return (
                <div style={{ 
                    padding: '40px 20px', 
                    textAlign: 'center',
                    background: '#fafafa',
                    borderRadius: 8
                }}>
                    <Empty 
                        image={Empty.PRESENTED_IMAGE_SIMPLE} 
                        description="No activity data for selected period"
                    />
                </div>
            );
        }

        const hourWidth = 100 / 24;
        return (
            <div style={{ position: 'relative', height: 48, background: '#f5f5f5', borderRadius: 8, overflow: 'hidden' }}>
                {timeline.map((event, idx) => {
                    const start = dayjs(event.startTime);
                    const end = event.endTime ? dayjs(event.endTime) : dayjs();
                    const startHour = start.hour() + start.minute() / 60;
                    const duration = end.diff(start, 'minute') / 60;
                    const left = startHour * hourWidth;
                    const width = Math.max(duration * hourWidth, 0.5);

                    return (
                        <Tooltip 
                            key={idx} 
                            title={
                                <div>
                                    <div style={{ fontWeight: 600, marginBottom: 4 }}>{event.appName}</div>
                                    <div style={{ fontSize: 12 }}>{formatDuration(event.duration)}</div>
                                </div>
                            }
                        >
                            <div
                                style={{
                                    position: 'absolute',
                                    left: `${left}%`,
                                    width: `${width}%`,
                                    height: '100%',
                                    background: event.status === 'active' ? '#22c55e' : '#d4d4d4',
                                    borderRight: '1px solid #ffffff',
                                    cursor: 'pointer',
                                    transition: 'opacity 0.2s',
                                    opacity: 0.85
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
            <div style={{ display: 'flex', justifyContent: 'space-between', fontSize: 11, color: '#a3a3a3', marginTop: 8, fontWeight: 500 }}>
                {hours.map(h => (
                    <span key={h}>{h === 24 ? '24:00' : `${h.toString().padStart(2, '0')}:00`}</span>
                ))}
            </div>
        );
    };

    const StatCard = ({ title, value, icon, color, subtext }: any) => (
        <Card style={{ borderRadius: 12, border: '1px solid #e5e5e5', boxShadow: 'none' }}>
            <div style={{ display: 'flex', alignItems: 'flex-start', justifyContent: 'space-between' }}>
                <div>
                    <Text style={{ fontSize: 13, color: '#737373', textTransform: 'uppercase', letterSpacing: '0.5px', fontWeight: 600 }}>
                        {title}
                    </Text>
                    <div style={{ fontSize: 32, fontWeight: 700, color: '#171717', marginTop: 8, letterSpacing: '-0.5px' }}>
                        {value}
                    </div>
                    {subtext && (
                        <Text style={{ fontSize: 12, color: '#a3a3a3', marginTop: 4, display: 'block' }}>
                            {subtext}
                        </Text>
                    )}
                </div>
                <div style={{ 
                    width: 48, 
                    height: 48, 
                    borderRadius: 10, 
                    background: color + '15',
                    display: 'flex', 
                    alignItems: 'center', 
                    justifyContent: 'center',
                    color: color
                }}>
                    {icon}
                </div>
            </div>
        </Card>
    );

    return (
        <div className="p-6" style={{ maxWidth: 1400, margin: '0 auto' }}>
            {/* Header */}
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 32, flexWrap: 'wrap', gap: 16 }}>
                <div>
                    <Title level={2} style={{ color: '#171717', margin: 0, fontWeight: 700, fontSize: 28 }}>
                        Activity Monitor
                    </Title>
                    <Text type="secondary" style={{ fontSize: 14 }}>
                        Track and analyze your application usage
                    </Text>
                </div>
                <Space size="middle">
                    <RangePicker
                        value={dateRange}
                        onChange={handleDateRangeChange}
                        allowClear={false}
                        style={{ borderRadius: 8 }}
                    />
                    <Button 
                        icon={<ReloadOutlined />} 
                        onClick={fetchData} 
                        loading={loading}
                        size="large"
                        style={{ borderRadius: 8 }}
                    >
                        Refresh
                    </Button>
                    <Button
                        type={isTracking ? 'default' : 'primary'}
                        icon={isTracking ? <StopOutlined /> : <PlayCircleOutlined />}
                        onClick={handleToggleTracking}
                        danger={isTracking}
                        size="large"
                        style={{ borderRadius: 8, fontWeight: 500 }}
                    >
                        {isTracking ? 'Stop Tracking' : 'Start Tracking'}
                    </Button>
                </Space>
            </div>

            {/* Stats Cards */}
            <Row gutter={[20, 20]} style={{ marginBottom: 24 }}>
                <Col xs={24} sm={12} lg={6}>
                    <StatCard
                        title="Active Time"
                        value={formatDuration(stats.totalActiveTime)}
                        icon={<ThunderboltOutlined style={{ fontSize: 22 }} />}
                        color="#22c55e"
                        subtext="Time spent actively using apps"
                    />
                </Col>
                <Col xs={24} sm={12} lg={6}>
                    <StatCard
                        title="AFK Time"
                        value={formatDuration(stats.totalAfkTime)}
                        icon={<PauseCircleOutlined style={{ fontSize: 22 }} />}
                        color="#f59e0b"
                        subtext="Time away from keyboard"
                    />
                </Col>
                <Col xs={24} sm={12} lg={6}>
                    <StatCard
                        title="Apps Used"
                        value={stats.totalApps}
                        icon={<AppstoreOutlined style={{ fontSize: 22 }} />}
                        color="#2851e6"
                        subtext="Unique applications"
                    />
                </Col>
                <Col xs={24} sm={12} lg={6}>
                    <StatCard
                        title="Most Used"
                        value={stats.topApp || '-'}
                        icon={<TrophyOutlined style={{ fontSize: 22 }} />}
                        color="#7c3aed"
                        subtext="Top application by time"
                    />
                </Col>
            </Row>

            {/* Timeline Section */}
            <Card 
                title={
                    <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
                        <DashboardOutlined style={{ color: '#2851e6' }} />
                        <span style={{ fontWeight: 600, color: '#171717' }}>Daily Timeline</span>
                    </div>
                }
                style={{ marginBottom: 24, borderRadius: 12, border: '1px solid #e5e5e5' }}
            >
                {getTimelineBlocks()}
                {timeline.length > 0 && getHourLabels()}
                {timeline.length > 0 && (
                    <div style={{ display: 'flex', gap: 24, marginTop: 16, fontSize: 13 }}>
                        <span style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
                            <span style={{ display: 'inline-block', width: 12, height: 12, background: '#22c55e', borderRadius: 2 }} />
                            <span style={{ color: '#525252', fontWeight: 500 }}>Active</span>
                        </span>
                        <span style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
                            <span style={{ display: 'inline-block', width: 12, height: 12, background: '#d4d4d4', borderRadius: 2 }} />
                            <span style={{ color: '#525252', fontWeight: 500 }}>AFK</span>
                        </span>
                    </div>
                )}
            </Card>

            {/* Bottom Section */}
            <Row gutter={[20, 20]}>
                <Col xs={24} lg={8}>
                    <Card 
                        title={
                            <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
                                <TrophyOutlined style={{ color: '#f59e0b' }} />
                                <span style={{ fontWeight: 600, color: '#171717' }}>Top Applications</span>
                            </div>
                        }
                        style={{ borderRadius: 12, border: '1px solid #e5e5e5' }}
                    >
                        {topApps.length > 0 ? (
                            <div style={{ display: 'flex', flexDirection: 'column', gap: 16 }}>
                                {topApps.map((app, idx) => (
                                    <div key={idx}>
                                        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 8 }}>
                                            <div style={{ display: 'flex', alignItems: 'center', gap: 10 }}>
                                                <div style={{ 
                                                    width: 28, 
                                                    height: 28, 
                                                    borderRadius: 6, 
                                                    background: idx === 0 ? '#fef3c7' : idx === 1 ? '#f3f4f6' : idx === 2 ? '#fef9c3' : '#f5f5f5',
                                                    display: 'flex',
                                                    alignItems: 'center',
                                                    justifyContent: 'center',
                                                    fontWeight: 700,
                                                    fontSize: 12,
                                                    color: idx === 0 ? '#92400e' : idx === 1 ? '#374151' : idx === 2 ? '#854d0e' : '#737373'
                                                }}>
                                                    {idx + 1}
                                                </div>
                                                <span style={{ fontWeight: 500, color: '#171717' }}>{app.appName}</span>
                                            </div>
                                            <span style={{ color: '#525252', fontWeight: 600, fontSize: 13 }}>
                                                {formatDuration(app.totalDuration)}
                                            </span>
                                        </div>
                                        <Progress 
                                            percent={Math.round(app.percentage)} 
                                            strokeColor={{
                                                '0%': '#2851e6',
                                                '100%': '#1d307a',
                                            }}
                                            trailColor="#f0f0f0"
                                            showInfo={false}
                                            strokeWidth={8}
                                        />
                                    </div>
                                ))}
                            </div>
                        ) : (
                            <Empty 
                                image={Empty.PRESENTED_IMAGE_SIMPLE} 
                                description="No application data"
                            />
                        )}
                    </Card>
                </Col>
                <Col xs={24} lg={16}>
                    <Card 
                        title={
                            <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
                                <AppstoreOutlined style={{ color: '#2851e6' }} />
                                <span style={{ fontWeight: 600, color: '#171717' }}>Activity Summary</span>
                            </div>
                        }
                        style={{ borderRadius: 12, border: '1px solid #e5e5e5' }}
                    >
                        <Table
                            dataSource={summary}
                            columns={summaryColumns}
                            rowKey="appName"
                            pagination={false}
                            size="middle"
                            scroll={{ y: 320 }}
                            locale={{
                                emptyText: (
                                    <Empty 
                                        image={Empty.PRESENTED_IMAGE_SIMPLE} 
                                        description="No activity summary data"
                                        style={{ margin: '20px 0' }}
                                    />
                                )
                            }}
                            style={{ borderRadius: 8 }}
                        />
                    </Card>
                </Col>
            </Row>
        </div>
    );
};

export default Activity;
