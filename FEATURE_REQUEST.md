# Feature Request: Profile-Based Service Isolation

## 📋 Summary

Implement profile-based service isolation to ensure that uptime statistics, service metrics, and service management operations are properly scoped to individual profiles rather than being shared globally across all profiles.

## 🎯 Problem Statement

Currently, the uptime statistics feature (recently added) and other service-related data appear to be shared globally across all profiles. This creates several issues:

1. **Data Contamination**: Services with the same name in different profiles share uptime statistics
2. **Profile Independence**: Profiles should be completely isolated environments
3. **Multi-Environment Support**: Users cannot properly differentiate between dev/staging/production environments
4. **Service Conflicts**: Services in different profiles may interfere with each other's metrics

## 💡 Proposed Solution

Implement complete profile-based isolation for all service-related functionality:

### Core Changes Required:

#### 1. Database Schema Updates
```sql
-- Add profile_id foreign key to services table
ALTER TABLE services ADD COLUMN profile_id TEXT REFERENCES profiles(id);

-- Add profile_id to uptime statistics tracking
ALTER TABLE service_metrics ADD COLUMN profile_id TEXT REFERENCES profiles(id);

-- Add composite indexes for performance
CREATE INDEX idx_services_profile_id ON services(profile_id);
CREATE INDEX idx_service_metrics_profile_service ON service_metrics(profile_id, service_id);
```

#### 2. API Endpoints Scoping
- All service endpoints should accept/require profile context
- Service operations (start/stop/restart) should be profile-scoped
- Metrics and uptime statistics should filter by active profile
- Service discovery should only scan directories within profile's `projectsDir`

#### 3. Frontend Profile Context
- Ensure all service operations pass current profile ID
- Service cards and metrics should only show data for active profile
- Profile switching should completely refresh service data
- Service topology should only show services within the active profile

#### 4. Service Management Isolation
- Service ports should be managed per-profile (avoid conflicts)
- Environment variables should be profile-scoped (already implemented)
- Service dependencies should only reference services within the same profile
- Log aggregation should be profile-aware

## 🔧 Implementation Details

### Backend Changes

1. **Service Manager Updates**
   ```go
   type ServiceManager struct {
       profileID string
       services  map[string]*Service
       // ... other fields
   }
   
   func NewServiceManager(cfg *config.Config, db *database.Database, profileID string) (*ServiceManager, error)
   ```

2. **Database Queries**
   - Add `WHERE profile_id = ?` to all service-related queries
   - Update service creation to include profile ID
   - Modify uptime statistics collection to be profile-aware

3. **API Handler Updates**
   ```go
   // Extract profile ID from request context or headers
   func (h *Handler) getProfileContext(r *http.Request) string {
       // Implementation to get active profile ID
   }
   
   // Update all service endpoints to use profile context
   func (h *Handler) GetServices(w http.ResponseWriter, r *http.Request) {
       profileID := h.getProfileContext(r)
       services := h.serviceManager.GetServicesByProfile(profileID)
       // ...
   }
   ```

### Frontend Changes

1. **Profile Context Provider**
   ```typescript
   interface ProfileContextType {
       activeProfileId: string;
       switchProfile: (profileId: string) => void;
       services: Service[]; // Only services for active profile
   }
   ```

2. **Service API Updates**
   ```typescript
   // All service API calls should include profile context
   const getServices = async (profileId: string): Promise<Service[]> => {
       const response = await fetch(`/api/profiles/${profileId}/services`);
       return response.json();
   };
   ```

3. **Component Updates**
   - Update `useServiceManagement` hook to be profile-aware
   - Modify service cards to display profile-scoped metrics
   - Update topology view to show only profile services

## 📊 Expected Benefits

1. **True Environment Isolation**: Dev/staging/production profiles are completely separate
2. **Accurate Metrics**: Uptime statistics reflect actual service performance per environment
3. **Reduced Conflicts**: Services in different profiles don't interfere with each other
4. **Better User Experience**: Clear separation of concerns between different environments
5. **Scalability**: Support for many profiles without data mixing

## 🧪 Testing Requirements

### Unit Tests
- Service isolation within profiles
- Profile switching doesn't leak data
- Metrics are correctly scoped

### Integration Tests
- Multiple profiles with same service names
- Profile switching preserves service states
- Database queries properly filter by profile ID

### User Acceptance Tests
- Create two profiles with identical service names
- Verify uptime statistics are separate
- Confirm service operations only affect active profile
- Test profile switching maintains isolation

## 📈 Migration Strategy

1. **Phase 1**: Add profile_id columns with NULL allowed
2. **Phase 2**: Migrate existing services to default profile
3. **Phase 3**: Make profile_id NOT NULL
4. **Phase 4**: Update all API endpoints
5. **Phase 5**: Update frontend components
6. **Phase 6**: Add comprehensive tests

## 🏷️ Labels

- `enhancement`
- `backend`
- `frontend` 
- `database`
- `profiles`
- `service-management`
- `uptime-statistics`

## 👤 Contributor Notes

This feature request addresses architectural improvements needed after the recent uptime statistics implementation. The goal is to ensure that profiles provide true environment isolation as originally intended.

**Priority**: High - This is fundamental to the multi-profile architecture

**Estimated Effort**: Large - Requires database schema changes, API updates, and frontend modifications

**Dependencies**: None - Can be implemented independently

---

*This feature request was generated based on the recent uptime statistics implementation and the need for proper profile-based service isolation.*