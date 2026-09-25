package api

// OneKit's generated Register*ServiceServer helpers take `any` and assert the
// server interface at runtime, so a contract that gains an RPC still compiles
// and only fails when module.go registers the service during boot. These
// assertions move that failure to build time. Every service registered in
// module.go must appear here.

import (
	flowcyclev1 "github.com/voxire/voxie/gen/go/flow/cycle/v1"
	flowcyclev2 "github.com/voxire/voxie/gen/go/flow/cycle/v2"
	flowdashboardv1 "github.com/voxire/voxie/gen/go/flow/dashboard/v1"
	flowdashboardv2 "github.com/voxire/voxie/gen/go/flow/dashboard/v2"
	flowinsightsv1 "github.com/voxire/voxie/gen/go/flow/insights/v1"
	flowinsightsv2 "github.com/voxire/voxie/gen/go/flow/insights/v2"
	flowlabelv1 "github.com/voxire/voxie/gen/go/flow/label/v1"
	flowlabelv2 "github.com/voxire/voxie/gen/go/flow/label/v2"
	flowmilestonev1 "github.com/voxire/voxie/gen/go/flow/milestone/v1"
	flowmilestonev2 "github.com/voxire/voxie/gen/go/flow/milestone/v2"
	flownotificationv1 "github.com/voxire/voxie/gen/go/flow/notification/v1"
	flowprofilev1 "github.com/voxire/voxie/gen/go/flow/profile/v1"
	flowprojectv1 "github.com/voxire/voxie/gen/go/flow/project/v1"
	flowprojectv2 "github.com/voxire/voxie/gen/go/flow/project/v2"
	flowsettingv1 "github.com/voxire/voxie/gen/go/flow/setting/v1"
	flowstatusv2 "github.com/voxire/voxie/gen/go/flow/status/v2"
	flowtaskv1 "github.com/voxire/voxie/gen/go/flow/task/v1"
	flowtaskv2 "github.com/voxire/voxie/gen/go/flow/task/v2"
	flowtimev1 "github.com/voxire/voxie/gen/go/flow/time/v1"
	flowtimev2 "github.com/voxire/voxie/gen/go/flow/time/v2"
	flowviewv1 "github.com/voxire/voxie/gen/go/flow/view/v1"
	flowviewv2 "github.com/voxire/voxie/gen/go/flow/view/v2"
	flowworkspacev1 "github.com/voxire/voxie/gen/go/flow/workspace/v1"
)

var (
	_ flowworkspacev1.WorkspaceServiceServer       = (*WorkspaceAPI)(nil)
	_ flowprojectv1.ProjectServiceServer           = (*ProjectAPI)(nil)
	_ flowprojectv2.ProjectServiceServer           = (*ProjectAPIV2)(nil)
	_ flowlabelv1.LabelServiceServer               = (*LabelAPI)(nil)
	_ flowlabelv2.LabelServiceServer               = (*LabelAPIV2)(nil)
	_ flowstatusv2.StatusServiceServer             = (*StatusAPIV2)(nil)
	_ flowtaskv1.TaskServiceServer                 = (*TaskAPI)(nil)
	_ flowtaskv2.TaskServiceServer                 = (*TaskAPIV2)(nil)
	_ flowdashboardv1.DashboardServiceServer       = (*DashboardAPI)(nil)
	_ flowdashboardv2.DashboardServiceServer       = (*DashboardAPIV2)(nil)
	_ flowinsightsv1.InsightsServiceServer         = (*InsightsAPI)(nil)
	_ flowinsightsv2.InsightsServiceServer         = (*InsightsAPIV2)(nil)
	_ flowcyclev1.CycleServiceServer               = (*CycleAPI)(nil)
	_ flowcyclev2.CycleServiceServer               = (*CycleAPIV2)(nil)
	_ flowmilestonev1.MilestoneServiceServer       = (*MilestoneAPI)(nil)
	_ flowmilestonev2.MilestoneServiceServer       = (*MilestoneAPIV2)(nil)
	_ flowviewv1.ViewServiceServer                 = (*ViewAPI)(nil)
	_ flowviewv2.ViewServiceServer                 = (*ViewAPIV2)(nil)
	_ flownotificationv1.NotificationServiceServer = (*NotificationAPI)(nil)
	_ flowprofilev1.ProfileServiceServer           = (*ProfileAPI)(nil)
	_ flowsettingv1.SettingServiceServer           = (*SettingAPI)(nil)
	_ flowtimev1.TimeServiceServer                 = (*WorkTimeAPI)(nil)

	_ flowtimev2.TimeServiceServer = (*WorkTimeAPIV2)(nil)
)
