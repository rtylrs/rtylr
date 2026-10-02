export function flowAuthFields(memberType: 'owner' | 'admin' | 'member' = 'owner') {
  return {
    member_type: memberType,
    effective_access: {
      business_id: '1',
      is_owner: memberType === 'owner',
      is_admin: memberType === 'admin',
      dash_permissions: [],
      subscribed_apps: ['flow'],
      licenses: [],
      authorization_version: '1',
      apps: [
        {
          app_id: 'flow',
          licensed: true,
          has_access: true,
          permissions: ['flow:tasks:view', 'flow:projects:view'],
          permission_scopes: [],
          sources: ['explicit'],
        },
      ],
    },
  };
}
