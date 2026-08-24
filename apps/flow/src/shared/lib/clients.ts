import { createOneKitClient } from '@voxire/api';
import { BusinessServiceClient } from '@voxire/voxie/hub/business/v1/client';
import { UserServiceClient } from '@voxire/voxie/hub/user/v1/client';
import { NotificationServiceClient } from '@voxire/voxie/flow/notification/v1/client';
import { SettingServiceClient } from '@voxire/voxie/flow/setting/v1/client';
// v2: business-scoped contracts (workspace removal)
import { ProjectServiceClient as ProjectServiceV2 } from '@voxire/voxie/flow/project/v2/client';
import { TaskServiceClient as TaskServiceV2 } from '@voxire/voxie/flow/task/v2/client';
import { DashboardServiceClient as DashboardServiceV2 } from '@voxire/voxie/flow/dashboard/v2/client';
import { LabelServiceClient as LabelServiceV2 } from '@voxire/voxie/flow/label/v2/client';
import { MilestoneServiceClient as MilestoneServiceV2 } from '@voxire/voxie/flow/milestone/v2/client';
import { CycleServiceClient as CycleServiceV2 } from '@voxire/voxie/flow/cycle/v2/client';
import { ViewServiceClient as ViewServiceV2 } from '@voxire/voxie/flow/view/v2/client';
import { InsightsServiceClient as InsightsServiceV2 } from '@voxire/voxie/flow/insights/v2/client';
import { TimeServiceClient as TimeServiceV2 } from '@voxire/voxie/flow/time/v2/client';
import { ProfileServiceClient as ProfileClient } from '@voxire/voxie/flow/profile/v1/client';
import { StatusServiceClient } from '@voxire/voxie/flow/status/v2/client';

export const businessClient = createOneKitClient(BusinessServiceClient, 'flow');
export const userClient = createOneKitClient(UserServiceClient, 'flow');
export const notificationClient = createOneKitClient(NotificationServiceClient, 'flow');
export const settingClient = createOneKitClient(SettingServiceClient, 'flow');

// ── v2 instances (Phase 2 swaps in; v1 stays until Phase 3 removes) ────────
export const projectV2 = createOneKitClient(ProjectServiceV2, 'flow');
export const taskV2 = createOneKitClient(TaskServiceV2, 'flow');
export const dashboardV2 = createOneKitClient(DashboardServiceV2, 'flow');
export const labelV2 = createOneKitClient(LabelServiceV2, 'flow');
export const milestoneV2 = createOneKitClient(MilestoneServiceV2, 'flow');
export const cycleV2 = createOneKitClient(CycleServiceV2, 'flow');
export const viewV2 = createOneKitClient(ViewServiceV2, 'flow');
export const insightsV2 = createOneKitClient(InsightsServiceV2, 'flow');
export const timeV2 = createOneKitClient(TimeServiceV2, 'flow');
export const profileClient = createOneKitClient(ProfileClient, 'flow');
export const statusClient = createOneKitClient(StatusServiceClient, 'flow');

// license, flow presence and hr compat are deliberately absent. Those three
// contracts do not exist in voxie and never have, so no client can be generated
// for them; importing them broke `vite build` for this app. Their call sites use
// the REST transport in ./api instead. See #894.
