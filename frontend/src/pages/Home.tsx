import React from 'react';
import { Card, Row, Col, Statistic, Typography, Tag, List, Progress } from 'antd';
import {
    CheckCircleOutlined,
    ClockCircleOutlined,
    WarningOutlined,
    FileTextOutlined
} from '@ant-design/icons';

const { Title, Text } = Typography;

interface Requirement {
    id: string;
    title: string;
    status: 'completed' | 'pending' | 'warning';
    progress: number;
}

const requirements: Requirement[] = [
    { id: '1', title: 'User Authentication', status: 'completed', progress: 100 },
    { id: '2', title: 'Dashboard Layout', status: 'completed', progress: 100 },
    { id: '3', title: 'API Integration', status: 'pending', progress: 45 },
    { id: '4', title: 'Data Validation', status: 'warning', progress: 20 },
    { id: '5', title: 'Testing Coverage', status: 'pending', progress: 60 },
];

const getStatusIcon = (status: string) => {
    switch (status) {
        case 'completed':
            return <CheckCircleOutlined style={{ color: '#52c41a' }} />;
        case 'warning':
            return <WarningOutlined style={{ color: '#faad14' }} />;
        default:
            return <ClockCircleOutlined style={{ color: '#7c3aed' }} />;
    }
};

const getStatusTag = (status: string) => {
    switch (status) {
        case 'completed':
            return <Tag color="success">Completed</Tag>;
        case 'warning':
            return <Tag color="warning">Needs Attention</Tag>;
        default:
            return <Tag color="purple">In Progress</Tag>;
    }
};

const Home: React.FC = () => {
    const completed = requirements.filter(r => r.status === 'completed').length;
    const pending = requirements.filter(r => r.status === 'pending').length;
    const warnings = requirements.filter(r => r.status === 'warning').length;

    return (
        <div className="p-6">
            <Title level={2} style={{ color: '#4c1d95', marginBottom: 24 }}>
                Status Requirements
            </Title>

            <Row gutter={[16, 16]} className="mb-6">
                <Col xs={24} sm={8}>
                    <Card>
                        <Statistic
                            title="Completed"
                            value={completed}
                            prefix={<CheckCircleOutlined style={{ color: '#52c41a' }} />}
                            valueStyle={{ color: '#52c41a' }}
                        />
                    </Card>
                </Col>
                <Col xs={24} sm={8}>
                    <Card>
                        <Statistic
                            title="In Progress"
                            value={pending}
                            prefix={<ClockCircleOutlined style={{ color: '#7c3aed' }} />}
                            valueStyle={{ color: '#7c3aed' }}
                        />
                    </Card>
                </Col>
                <Col xs={24} sm={8}>
                    <Card>
                        <Statistic
                            title="Needs Attention"
                            value={warnings}
                            prefix={<WarningOutlined style={{ color: '#faad14' }} />}
                            valueStyle={{ color: '#faad14' }}
                        />
                    </Card>
                </Col>
            </Row>

            <Card title={
                <span>
                    <FileTextOutlined style={{ marginRight: 8, color: '#7c3aed' }} />
                    Requirements List
                </span>
            }>
                <List
                    dataSource={requirements}
                    renderItem={(item) => (
                        <List.Item
                            actions={[getStatusTag(item.status)]}
                        >
                            <List.Item.Meta
                                avatar={getStatusIcon(item.status)}
                                title={<Text strong>{item.title}</Text>}
                                description={
                                    <Progress
                                        percent={item.progress}
                                        strokeColor={item.status === 'completed' ? '#52c41a' : '#7c3aed'}
                                        size="small"
                                    />
                                }
                            />
                        </List.Item>
                    )}
                />
            </Card>
        </div>
    );
};

export default Home;
