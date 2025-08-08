# Service Uptime Statistics Dashboard

This feature adds comprehensive uptime tracking and statistics for all services in the Vertex Service Manager.

## Features Added

### Backend Changes

1. **Extended ServiceMetrics Model** (`internal/models/service_dependency.go`)
   - Added `UptimeStatistics` struct with metrics like:
     - Total restarts count
     - Uptime percentage for 24h and 7d periods
     - Mean Time Between Failures (MTBF)
     - Total downtime tracking

2. **Uptime Tracking Service** (`internal/services/uptime_tracking.go`)
   - Singleton uptime tracker that records service state changes
   - Calculates uptime percentages and downtime durations
   - Tracks service restart events and failure patterns

3. **Integration with Service Lifecycle**
   - Added uptime event recording in service start/stop/restart operations
   - Integrated with monitoring system to update uptime stats
   - Automatic tracking of process failures and recoveries

4. **New API Endpoints** (`internal/handlers/uptime_handler.go`)
   - `GET /api/uptime/statistics` - Returns uptime stats for all services
   - `GET /api/uptime/statistics/{id}` - Returns uptime stats for specific service

### Frontend Changes

1. **New React Component** (`web/src/components/UptimeStatistics/UptimeStatisticsDashboard.tsx`)
   - Dashboard displaying service uptime statistics
   - Real-time updates every 30 seconds
   - Color-coded uptime percentages (green >99%, yellow >95%, red <95%)
   - Formatted duration displays for downtime and MTBF

2. **Updated Navigation** (`web/src/components/Sidebar/Sidebar.tsx`)
   - Added "Uptime Stats" menu item with clock icon

3. **Type Definitions** (`web/src/types.ts`)
   - Added `UptimeStatistics` interface
   - Updated `ServiceMetrics` to include uptime stats

## API Usage

### Get All Services Uptime Statistics
```
GET /api/uptime/statistics
```

Response:
```json
{
  "statistics": {
    "service-uuid": {
      "serviceName": "example-service",
      "serviceId": "service-uuid",
      "port": 8080,
      "status": "running",
      "healthStatus": "healthy",
      "stats": {
        "totalRestarts": 3,
        "uptimePercentage24h": 99.5,
        "uptimePercentage7d": 98.2,
        "mtbf": 86400000000000,
        "lastDowntime": "2025-08-07T10:30:00Z",
        "totalDowntime24h": 120000000000,
        "totalDowntime7d": 1800000000000
      }
    }
  },
  "summary": {
    "totalServices": 5,
    "runningServices": 4,
    "unhealthyServices": 1
  }
}
```

### Get Specific Service Uptime Statistics
```
GET /api/uptime/statistics/{service-id}
```

## Dashboard Features

- **Summary Cards**: Quick overview of total, running, and unhealthy services
- **Detailed Table**: Per-service uptime metrics including:
  - 24-hour and 7-day uptime percentages
  - Total restart count
  - Mean Time Between Failures (MTBF)
  - Total downtime in the last 24 hours
- **Color Coding**: Visual indicators for uptime health
- **Auto-Refresh**: Updates every 30 seconds
- **Duration Formatting**: Human-readable time formats (e.g., "2d 4h 30m")

## Benefits

1. **Operational Insights**: Quickly identify problematic services
2. **SLA Monitoring**: Track service availability against targets
3. **Historical Trends**: Understand service reliability patterns
4. **Proactive Maintenance**: Identify services requiring attention
5. **Performance Tracking**: Monitor improvements over time

## Implementation Notes

- Uses singleton pattern for uptime tracker to ensure data consistency
- Integrates seamlessly with existing monitoring infrastructure
- Minimal performance impact with efficient event recording
- Automatically handles service lifecycle events
- Memory-efficient with event history limits (1000 events per service)
