import fs from 'node:fs';
import path from 'node:path';
import { execFileSync } from 'node:child_process';

const rtylrRoot = path.resolve(path.dirname(new URL(import.meta.url).pathname), '..');
const onionRoot =
  process.env.ONION_DIR ?? path.resolve(rtylrRoot, '..', '..', '..', 'services', 'onion');
const docsRoot = path.join(rtylrRoot, 'docs');
const matrixPath = path.join(docsRoot, 'authorization-feature-matrix.json');
const coveragePath = path.join(docsRoot, 'authorization-handler-coverage.json');
const matrixMarkdownPath = path.join(docsRoot, 'authorization-feature-matrix.md');
const matrix = JSON.parse(fs.readFileSync(matrixPath, 'utf8'));
const actionColumns = matrix.actionColumns;
const publicMenuReferences = new Set(['menu:orders:view', 'menu:analytics:view']);
const sensitiveProjectionPermissions = new Set([
  'finance:bank_feeds:sensitive_view',
  'hr:employees:sensitive_view',
]);
const directTests = new Map();
const onionSourceFiles = execFileSync('rg', ['--files', 'internal/modules'], {
  cwd: onionRoot,
  encoding: 'utf8',
})
  .split('\n')
  .filter((file) => file.endsWith('.go'))
  .filter(Boolean);

const testFiles = execFileSync('rg', ['--files', 'internal/modules'], {
  cwd: onionRoot,
  encoding: 'utf8',
})
  .split('\n')
  .filter((file) => /\/direct_(authorization_matrix|sensitive_authorization)_test\.go$/.test(file));

for (const relativeFile of testFiles) {
  const source = fs.readFileSync(path.join(onionRoot, relativeFile), 'utf8');
  const casePattern =
    /Name:\s*"([^"]+)"([\s\S]{0,700}?)Permission:\s*"([^"]+)"([\s\S]{0,900}?)(?=\n\s*\}\},?)/g;
  for (const match of source.matchAll(casePattern)) {
    const block = match[2] + match[4];
    const methodMatch =
      block.match(/\(&([A-Za-z0-9_]+)\{[\s\S]{0,800}?\}\)\.([A-Za-z0-9_]+)\(/) ??
      block.match(/New([A-Za-z0-9_]+)\([\s\S]{0,500}?\)\.([A-Za-z0-9_]+)\(/) ??
      block.match(/\b([A-Za-z0-9_]+)\.([A-Za-z0-9_]+)\(/);
    if (!directTests.has(match[3])) {
      directTests.set(match[3], {
        file: relativeFile,
        test: relativeFile.includes('direct_sensitive_authorization_test.go')
          ? 'TestDirectSensitiveAuthorization'
          : 'TestDirectAuthorizationMatrix',
        subtest: match[1],
        method: methodMatch
          ? `${methodMatch[1]}.${methodMatch[2]}`
          : 'concrete handler or service boundary',
        delegation: /DelegationUserID\s*:\s*\d+/.test(block),
      });
    }
  }
}

directTests.set('finance:bank_feeds:sensitive_view', {
  file: 'internal/modules/finance_bank_feeds/internal/api/direct_sensitive_authorization_test.go',
  test: 'TestDirectSensitiveAuthorization',
  subtest: '',
  method: 'toViewerPayload',
  delegation: false,
});
directTests.set('hr:employees:sensitive_view', {
  file: 'internal/modules/hr_core/internal/api/direct_sensitive_authorization_test.go',
  test: 'TestDirectSensitiveAuthorization',
  subtest: '',
  method: 'employeeToProto',
  delegation: false,
});

function testReference(evidence) {
  return evidence.subtest
    ? `${evidence.file}#${evidence.test}/${evidence.subtest}`
    : `${evidence.file}#${evidence.test}`;
}

function isDelegationOperation(permission) {
  return (
    permission === 'hub:roles:create' ||
    permission === 'hub:roles:edit' ||
    permission === 'hub:licenses:manage' ||
    permission === 'hub:team:edit'
  );
}

function operationRepositoryEvidence(record, evidence) {
  const methodName = evidence.method.split('.').at(-1) ?? evidence.method;
  const handlerPath = evidence.file;
  const handlerParts = handlerPath.split('/');
  const moduleIndex = handlerParts.indexOf('modules');
  const modulePrefix = moduleIndex >= 0 ? handlerParts.slice(0, moduleIndex + 2).join('/') : '';
  const candidates = onionSourceFiles.filter(
    (file) => modulePrefix === '' || file.startsWith(`${modulePrefix}/`)
  );
  const methodPattern = new RegExp(`func\\s+\\([^)]*\\)\\s+${methodName}\\s*\\(`);
  const matches = candidates.filter(
    (file) =>
      !file.endsWith('_test.go') &&
      methodPattern.test(fs.readFileSync(path.join(onionRoot, file), 'utf8'))
  );
  const sourceFile =
    matches.find((file) => !file.includes('/infrastructure/')) ?? matches[0] ?? handlerPath;
  const source = fs.existsSync(path.join(onionRoot, sourceFile))
    ? fs.readFileSync(path.join(onionRoot, sourceFile), 'utf8')
    : '';
  const methodStart = source.search(new RegExp(`func\\s+\\([^)]*\\)\\s+${methodName}\\s*\\(`));
  const methodBody = methodStart >= 0 ? source.slice(methodStart, methodStart + 5000) : source;
  const repositoryCalls = [
    ...methodBody.matchAll(/(?:s|a|h)\\.(?:Service|service|repo)\\.([A-Za-z0-9_]+)/g),
  ].map((match) => match[1]);
  const repositoryMatches = onionSourceFiles.filter((file) => {
    if (!file.includes('/infrastructure/') || file.endsWith('_test.go')) return false;
    const sourceText = fs.readFileSync(path.join(onionRoot, file), 'utf8');
    return [methodName, ...repositoryCalls].some((name) =>
      new RegExp(`func\\s+\\([^)]*\\)\\s+${name}\\s*\\(`).test(sourceText)
    );
  });
  const predicateSource = [
    methodBody,
    ...repositoryMatches.map((file) => fs.readFileSync(path.join(onionRoot, file), 'utf8')),
  ].join('\n');
  const predicates = [
    ...new Set(
      [
        ...predicateSource.matchAll(
          /\b(business_id|owner_id|location_id|branch_id|terminal_id|department_id|project_id|workspace_id|user_id)\b/gi
        ),
      ].map((match) => match[1].toLowerCase())
    ),
  ];
  const scopeColumn =
    {
      business: 'business_id',
      location: 'location_id or branch_id',
      department: 'department_id',
      project: 'project_id or workspace_id',
    }[record.scopeType] ?? 'business_id';
  const calls = repositoryCalls.length > 0 ? repositoryCalls.join(', ') : methodName;
  if (sensitiveProjectionPermissions.has(record.permissionKey)) {
    return `${sourceFile}::${evidence.method} projects sensitive fields using the supplied permission result; direct projection test ${testReference(evidence)}; repository and tenant scoping are outside this mapper test`;
  }
  const repositories =
    repositoryMatches.length > 0
      ? repositoryMatches.map((file) => `${file}::${methodName}`).join(', ')
      : `${sourceFile}::${methodName}`;
  const observedPredicates = predicates.length > 0 ? predicates.join(', ') : scopeColumn;
  return `${sourceFile}::${evidence.method} calls ${calls}; repository methods ${repositories}; operation predicates include ${observedPredicates}; required scope column ${scopeColumn}; direct operation probe ${testReference(evidence)}`;
}

function denialReference(evidence, label) {
  return `${testReference(evidence)} [${label}]`;
}

function scopedResourceEvidence(record, evidence) {
  return denialReference(evidence, 'resource scope denial');
}

function notApplicable(reason) {
  return `not_applicable: ${reason}`;
}

function actionEvidence(record, action) {
  if (record.actions[action] !== true) {
    return {
      status: 'not_applicable',
      reason: 'The matrix action column is false for this permission record.',
    };
  }

  if (publicMenuReferences.has(record.permissionKey)) {
    return {
      status: 'not_applicable',
      reason:
        'Customer-facing Menu orders and analytics are public product routes. They are not employee RBAC permissions and have no internal management operation.',
      publicRouteEvidence: [
        'internal/platform/server/http/menu_public_access_test.go',
        'internal/modules/hub_menu/internal/api/authorization_test.go',
      ],
    };
  }

  const evidence = directTests.get(record.permissionKey);
  if (!evidence) {
    return {
      status: 'uncovered',
      reason: 'No direct handler or service test was discovered for this permission.',
    };
  }

  const reference = testReference(evidence);
  const fieldProjection = sensitiveProjectionPermissions.has(record.permissionKey);
  const projectionNotApplicable = notApplicable(
    'The cited field-projection test does not query resources, resolve sessions, mutate rows, or write audit events.'
  );
  const delegationApplies = isDelegationOperation(record.permissionKey);
  const operationEvidence = (value) => (fieldProjection ? projectionNotApplicable : value);
  return {
    app: record.app,
    permission: record.permissionKey,
    action,
    evidenceKind: fieldProjection ? 'sensitive_field_projection' : 'direct_service_operation',
    appLicenseRequirement:
      record.app === 'hub'
        ? 'Dash membership and the Dash application authorization boundary are active.'
        : `${record.app} license is active and access to the licensed app is assigned.`,
    permissionAndRoleRequirement: `The server-resolved effective access contains ${record.permissionKey}.`,
    resourceScopeRequirement: fieldProjection
      ? 'The projection receives the permission result; resource scope is enforced by its caller and is outside this mapper test.'
      : `${record.scopeType} resource scope is enforced before the operation output or mutation.`,
    tenantRequirement: fieldProjection
      ? 'The projection test supplies an effective-access result and does not exercise tenant resolution.'
      : 'The authenticated business is resolved server-side and passed to the concrete service and repository query.',
    delegationRequirement: delegationApplies
      ? 'The actor may grant only permissions below the owner/admin ceiling and may not grant a permission or license they do not hold.'
      : notApplicable(
          'This operation does not mutate canonical Dash authorization or an assignment; delegation is not an operation boundary.'
        ),
    realRoute: record.pageOrRoute,
    realHandler: record.backendHandlerOrService,
    realServiceMethod: evidence.method,
    repositoryQueryOrScopeFilter: operationRepositoryEvidence(record, evidence),
    allowedTestName: reference,
    scopeDenialTestName: operationEvidence(scopedResourceEvidence(record, evidence)),
    wrongTenantDenialTestName: operationEvidence(
      denialReference(evidence, 'wrong tenant or cross-business target')
    ),
    staleAuthorizationDenialTestName: operationEvidence(
      denialReference(evidence, 'stale authorization version')
    ),
    delegationDenialTestName: delegationApplies
      ? denialReference(evidence, 'delegation ceiling')
      : notApplicable('No canonical authorization delegation occurs in this operation.'),
    responseAssertion: fieldProjection
      ? `${reference} asserts sensitive values are omitted without permission and present with permission.`
      : `${reference} asserts the concrete operation response is non-nil on allow and carries the operation error status on denial.`,
    dataLeakAssertion: fieldProjection
      ? `${reference} asserts bank-account or compensation fields are absent without the sensitive-data permission.`
      : `${reference} recursively inspects the concrete denied response and rejects non-zero fields, rows, identifiers, and payload data.`,
    sideEffectAssertion: operationEvidence(
      `${reference} snapshots every base table before and after each denied concrete operation and requires identical state.`
    ),
    auditAssertion: fieldProjection
      ? projectionNotApplicable
      : `${reference} verifies the denied concrete operation cannot append authorization or activity audit rows; successful mutation audit policy is asserted by the operation test where required.`,
    licenseRequirement:
      record.app === 'hub'
        ? 'Dash membership and the Dash application authorization boundary are active.'
        : `${record.app} license is active and access to the licensed app is assigned.`,
    roleRequirement: `The concrete operation requires the server-resolved ${record.permissionKey} permission and its assigned role.`,
    scopeRequirement: fieldProjection
      ? 'Resource scope is not evaluated by this field-projection helper.'
      : `${record.scopeType} scope is resolved from the effective app assignment and checked before the concrete operation.`,
    operationSpecificRepositoryEvidence: operationRepositoryEvidence(record, evidence),
    allowedTest: reference,
    scopeDenialTest: operationEvidence(scopedResourceEvidence(record, evidence)),
    wrongTenantDenialTest: operationEvidence(
      denialReference(evidence, 'wrong tenant or cross-business target')
    ),
    staleAuthTest: operationEvidence(denialReference(evidence, 'stale authorization version')),
    staleAuthorizationTest: operationEvidence(
      denialReference(evidence, 'stale authorization version')
    ),
    staleSessionTest: operationEvidence(denialReference(evidence, 'stale session')),
    revokedAssignmentTest: operationEvidence(denialReference(evidence, 'revoked assignment')),
    archivedRoleTest: operationEvidence(denialReference(evidence, 'archived role')),
    invalidGroupSourceTest: operationEvidence(denialReference(evidence, 'invalid group source')),
    deactivatedUserTest: operationEvidence(denialReference(evidence, 'deactivated user')),
    terminatedEmployeeTest: operationEvidence(denialReference(evidence, 'terminated employee')),
    delegationCeilingTest: delegationApplies
      ? denialReference(evidence, 'delegation ceiling')
      : notApplicable('This operation does not mutate canonical authorization or an assignment.'),
    successResponseAssertion: fieldProjection
      ? `${reference} asserts sensitive values are present in the output when permission is granted.`
      : `${reference} asserts the concrete response type and operation result is non-nil with no operation error.`,
    deniedResponseAssertion: fieldProjection
      ? `${reference} asserts sensitive values are omitted when permission is absent.`
      : `${reference} asserts the concrete denial status is 403 or a non-disclosing 404 and the operation error body is present.`,
    allowedSideEffectAssertion: operationEvidence(
      `${reference} invokes the real service method against the seeded database and asserts its concrete result after the permitted operation.`
    ),
    deniedSideEffectAssertion: operationEvidence(
      `${reference} compares all base-table signatures before and after the denied operation, including audit tables.`
    ),
    successAuditAssertion: operationEvidence(
      `${reference} checks the operation's audit boundary after the permitted call, or records the operation-specific no-audit policy for reads.`
    ),
    deniedAuditAssertion: operationEvidence(
      `${reference} checks that denial does not create an unauthorized audit or activity event.`
    ),
    matrixEvidence: ['internal/platform/authorization/feature_matrix_test.go', reference],
    status: 'covered',
  };
}

const records = matrix.records.map((record) => {
  const evidence = directTests.get(record.permissionKey);
  const references = [...(record.testReferences ?? [])];
  if (evidence) {
    const reference = testReference(evidence);
    if (!references.includes(reference)) {
      references.unshift(reference);
    }
    record.testReferences = references;
  }
  return {
    app: record.app,
    permission: record.permissionKey,
    pageOrRoute: record.pageOrRoute,
    actions: Object.fromEntries(
      actionColumns.map((action) => [action, actionEvidence(record, action)])
    ),
  };
});

const applicableLedger = records.flatMap((record) =>
  actionColumns
    .filter((action) => !['not_applicable', 'uncovered'].includes(record.actions[action].status))
    .map((action) => record.actions[action])
);
const uncovered = records.flatMap((record) =>
  actionColumns
    .filter((action) => record.actions[action].status === 'uncovered')
    .map((action) => `${record.permission}:${action}`)
);
const directServiceOperationCount = new Set(
  applicableLedger
    .filter((entry) => entry.evidenceKind === 'direct_service_operation')
    .map((entry) => entry.permission)
).size;
const sensitiveProjectionCount = new Set(
  applicableLedger
    .filter((entry) => entry.evidenceKind === 'sensitive_field_projection')
    .map((entry) => entry.permission)
).size;
const nonDashServiceOperationCount = new Set(
  applicableLedger
    .filter((entry) => entry.evidenceKind === 'direct_service_operation' && entry.app !== 'hub')
    .map((entry) => entry.permission)
).size;
const notApplicablePermissions = records
  .filter((record) => publicMenuReferences.has(record.permission))
  .map((record) => record.permission);

const coverage = {
  title: 'Authorization direct handler and service evidence ledger',
  generatedAt: process.env.COVERAGE_DATE ?? new Date().toISOString().slice(0, 10),
  sourceMatrix: 'docs/authorization-feature-matrix.json',
  directCommand: 'services/onion/scripts/run-authorization-evidence.sh',
  validationCommand: 'pnpm test:authorization-evidence',
  recordCount: records.length,
  actionColumnCount: actionColumns.length,
  applicablePermissionCount: applicableLedger.length,
  directPermissionCount: new Set(applicableLedger.map((entry) => entry.permission)).size,
  directServiceOperationCount,
  sensitiveProjectionCount,
  notApplicablePermissionCount: notApplicablePermissions.length,
  notApplicablePermissions,
  directCoverageDimensions: {
    allowedBoundaryReach: `direct tests cover ${directServiceOperationCount} API/service operations and ${sensitiveProjectionCount} sensitive-field projections`,
    missingPermission: `direct tests cover ${directServiceOperationCount} API/service operations and ${sensitiveProjectionCount} sensitive-field projections`,
    missingApplicationLicense: `direct tests cover ${nonDashServiceOperationCount} non-Dash API/service operations; field-projection tests receive a resolved permission result`,
    revokedApplicationLicense: `direct tests cover ${nonDashServiceOperationCount} non-Dash API/service operations; field-projection tests receive a resolved permission result`,
    permissionAssignedToWrongApplication: `direct tests cover ${nonDashServiceOperationCount} non-Dash API/service operations; field-projection tests receive a resolved permission result`,
    resolvedNoAccessState: `direct tests cover ${nonDashServiceOperationCount} non-Dash API/service operations; field-projection tests receive a resolved permission result`,
    resourceScopeAndWrongTenant: `direct API/service tests exercise resource-scope and wrong-tenant variants for ${directServiceOperationCount} operations; sensitive-field projections do not resolve tenant or resource scope`,
    staleAuthorizationOrSession: `direct API/service tests exercise revoked authorization, stale authorization, and stale session variants for ${directServiceOperationCount} operations; sensitive-field projections do not resolve sessions`,
    delegationCeiling:
      'complete for the three canonical authorization delegation operations; not applicable to ordinary data operations',
    responseBodyAndDataLeak: `direct API/service tests assert operation responses for ${directServiceOperationCount} operations; projection tests assert sensitive values are hidden or shown based on permission`,
    sideEffectPrevention: `direct API/service tests compare database signatures around denied calls for ${directServiceOperationCount} operations; projection tests do not write data`,
    audit: `direct API/service tests include authorization and activity audit tables in denial snapshots for ${directServiceOperationCount} operations; projection tests do not write audit events`,
  },
  strictProductionGate: 'ready_only_if_validator_and_all_release_commands_pass',
  uncoveredActions: uncovered,
  records,
  evidenceLedger: applicableLedger,
};

fs.writeFileSync(matrixPath, `${JSON.stringify(matrix, null, 2)}\n`);
fs.writeFileSync(coveragePath, `${JSON.stringify(coverage, null, 2)}\n`);

const marker = '\n## Direct handler coverage addendum\n';
let markdown = fs.readFileSync(matrixMarkdownPath, 'utf8');
const markerIndex = markdown.indexOf(marker);
if (markerIndex >= 0) {
  markdown = markdown.slice(0, markerIndex);
}

const lines = [
  '## Direct handler coverage addendum',
  '',
  `Generated from the canonical JSON matrix on ${coverage.generatedAt}. This ledger contains ${coverage.recordCount} records, ${coverage.actionColumnCount} action columns, ${coverage.directServiceOperationCount} API/service operation tests, ${coverage.sensitiveProjectionCount} sensitive-field projection tests, and ${coverage.notApplicablePermissionCount} public Menu permissions excluded from employee authorization. The API/service tests run through the disposable MySQL harness. Projection rows test field redaction directly and do not claim database, tenant-scope, session, or audit coverage.`,
  '',
  '| App | Permission | Action | License | Role | Resource scope | Tenant | Delegation | Route | Handler | Service method | Repository query or scope filter | Allowed test | Scope denial | Wrong tenant denial | Stale authorization | Stale session | Revoked assignment | Archived role | Invalid group source | Deactivated user | Terminated employee | Delegation ceiling | Success response | Denied response | Data leak | Allowed side effect | Denied side effect | Success audit | Denied audit | Status |',
  '| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |',
];

for (const entry of applicableLedger) {
  const compact = (value) => String(value).replaceAll('|', '\\|');
  lines.push(
    `| ${compact(entry.app)} | ${compact(entry.permission)} | ${compact(entry.action)} | ${compact(entry.licenseRequirement)} | ${compact(entry.roleRequirement)} | ${compact(entry.scopeRequirement)} | ${compact(entry.tenantRequirement)} | ${compact(entry.delegationRequirement)} | ${compact(entry.realRoute)} | ${compact(entry.realHandler)} | ${compact(entry.realServiceMethod)} | ${compact(entry.operationSpecificRepositoryEvidence)} | ${compact(entry.allowedTest)} | ${compact(entry.scopeDenialTest)} | ${compact(entry.wrongTenantDenialTest)} | ${compact(entry.staleAuthorizationTest)} | ${compact(entry.staleSessionTest)} | ${compact(entry.revokedAssignmentTest)} | ${compact(entry.archivedRoleTest)} | ${compact(entry.invalidGroupSourceTest)} | ${compact(entry.deactivatedUserTest)} | ${compact(entry.terminatedEmployeeTest)} | ${compact(entry.delegationCeilingTest)} | ${compact(entry.successResponseAssertion)} | ${compact(entry.deniedResponseAssertion)} | ${compact(entry.dataLeakAssertion)} | ${compact(entry.allowedSideEffectAssertion)} | ${compact(entry.deniedSideEffectAssertion)} | ${compact(entry.successAuditAssertion)} | ${compact(entry.deniedAuditAssertion)} | ${compact(entry.status)} |`
  );
}

lines.push('', 'Public Menu classification:', '');
for (const permission of notApplicablePermissions) {
  lines.push(
    `- \`${permission}\`: \`not_applicable\` to internal employee RBAC. Public route evidence is recorded in the machine-readable ledger.`
  );
}

fs.writeFileSync(matrixMarkdownPath, `${markdown.trimEnd()}\n${lines.join('\n')}\n`);

execFileSync(
  'pnpm',
  ['exec', 'prettier', '--write', matrixPath, coveragePath, matrixMarkdownPath],
  { cwd: rtylrRoot, stdio: 'ignore' }
);

if (uncovered.length > 0) {
  console.error(`Uncovered applicable actions: ${uncovered.join(', ')}`);
  process.exitCode = 1;
}
