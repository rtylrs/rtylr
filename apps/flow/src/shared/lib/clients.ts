import { createOneKitClient } from '@voxire/api';
import { BusinessServiceClient } from '@voxire/voxie/hub/business/v1/client';
import { UserServiceClient } from '@voxire/voxie/hub/user/v1/client';
import { DashboardServiceClient } from '@voxire/voxie/flow/dashboard/v1/client';
import { WorkspaceServiceClient } from '@voxire/voxie/flow/workspace/v1/client';
import { ProjectServiceClient } from '@voxire/voxie/flow/project/v1/client';
import { TaskServiceClient } from '@voxire/voxie/flow/task/v1/client';
import { LabelServiceClient } from '@voxire/voxie/flow/label/v1/client';
import { MilestoneServiceClient } from '@voxire/voxie/flow/milestone/v1/client';
import { CycleServiceClient } from '@voxire/voxie/flow/cycle/v1/client';
import { ViewServiceClient } from '@voxire/voxie/flow/view/v1/client';
import { InsightsServiceClient } from '@voxire/voxie/flow/insights/v1/client';
import { NotificationServiceClient } from '@voxire/voxie/flow/notification/v1/client';
import { SettingServiceClient } from '@voxire/voxie/flow/setting/v1/client';

export const businessClient = createOneKitClient(BusinessServiceClient, 'flow');
export const userClient = createOneKitClient(UserServiceClient, 'flow');
export const dashboardClient = createOneKitClient(DashboardServiceClient, 'flow');
export const workspaceClient = createOneKitClient(WorkspaceServiceClient, 'flow');
export const projectClient = createOneKitClient(ProjectServiceClient, 'flow');
export const taskClient = createOneKitClient(TaskServiceClient, 'flow');
export const labelClient = createOneKitClient(LabelServiceClient, 'flow');
export const milestoneClient = createOneKitClient(MilestoneServiceClient, 'flow');
export const cycleClient = createOneKitClient(CycleServiceClient, 'flow');
export const viewClient = createOneKitClient(ViewServiceClient, 'flow');
export const insightsClient = createOneKitClient(InsightsServiceClient, 'flow');
export const notificationClient = createOneKitClient(NotificationServiceClient, 'flow');
export const settingClient = createOneKitClient(SettingServiceClient, 'flow');

// license, flow presence and hr compat are deliberately absent. Those three
// contracts do not exist in voxie and never have, so no client can be generated
// for them; importing them broke `vite build` for this app. Their call sites use
// the REST transport in ./api instead. See #894.
