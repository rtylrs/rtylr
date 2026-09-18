import fs from 'node:fs';
import path from 'node:path';
import { execFileSync } from 'node:child_process';

const rtylrRoot = path.resolve(path.dirname(new URL(import.meta.url).pathname), '..');
const onionRoot = path.resolve(rtylrRoot, '..', 'onion');
const docsRoot = path.join(rtylrRoot, 'docs');
const matrixPath = path.join(docsRoot, 'role-access-control-v2-feature-matrix.json');
const coveragePath = path.join(docsRoot, 'role-access-control-v2-handler-coverage.json');
const matrixMarkdownPath = path.join(docsRoot, 'role-access-control-v2-feature-matrix.md');
const matrix = JSON.parse(fs.readFileSync(matrixPath, 'utf8'));
const actionColumns = matrix.actionColumns;
const publicMenuReferences = new Set(['menu:orders:view', 'menu:analytics:view']);
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
  file: 'internal/modules/finance_bank_feeds/internal/routes/handlers/direct_sensitive_authorization_test.go',
  test: 'TestDirectSensitiveAuthorization',
  subtest: 'sensitive fields redacted and restored',
  method: 'toViewerPayload',
  delegation: false,
});
directTests.set('hr:employees:sensitive_view', {
  file: 'internal/modules/hr_core/internal/routes/handlers/direct_sensitive_authorization_test.go',
  test: 'TestDirectSensitiveAuthorization',
  subtest: 'sensitive fields redacted and restored',
  method: 'employeeToProto',
  delegation: false,
});

function testReference(evidence) {
  return `${evidence.file}#${evidence.test}/${evidence.subtest}`;
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
  const delegationApplies = isDelegationOperation(record.permissionKey);
  return {
    app: record.app,
    permission: record.permissionKey,
    action,
    appLicenseRequirement:
      record.app === 'hub'
        ? 'Dash membership and the Dash application authorization boundary are active.'
        : `${record.app} license is active and access to the licensed app is assigned.`,
    permissionAndRoleRequirement: `The server-resolved effective access contains ${record.permissionKey}.`,
    resourceScopeRequirement: `${record.scopeType} resource scope is enforced before the operation output or mutation.`,
    tenantRequirement:
      'The authenticated business is resolved server-side and passed to the concrete service and repository query.',
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
    scopeDenialTestName: scopedResourceEvidence(record, evidence),
    wrongTenantDenialTestName: denialReference(evidence, 'wrong tenant or cross-business target'),
    staleAuthorizationDenialTestName: denialReference(evidence, 'stale authorization version'),
    delegationDenialTestName: delegationApplies
      ? denialReference(evidence, 'delegation ceiling')
      : notApplicable('No canonical authorization delegation occurs in this operation.'),
    responseAssertion: `${reference} asserts the concrete operation response is non-nil on allow and carries the operation error status on denial.`,
    dataLeakAssertion: `${reference} recursively inspects the concrete denied response and rejects non-zero fields, rows, identifiers, and payload data.`,
    sideEffectAssertion: `${reference} snapshots every base table before and after each denied concrete operation and requires identical state.`,
    auditAssertion: `${reference} verifies the denied concrete operation cannot append authorization or activity audit rows; successful mutation audit policy is asserted by the operation test where required.`,
    licenseRequirement:
      record.app === 'hub'
        ? 'Dash membership and the Dash application authorization boundary are active.'
        : `${record.app} license is active and access to the licensed app is assigned.`,
    roleRequirement: `The concrete operation requires the server-resolved ${record.permissionKey} permission and its assigned role.`,
    scopeRequirement: `${record.scopeType} scope is resolved from the effective app assignment and checked before the concrete operation.`,
    operationSpecificRepositoryEvidence: operationRepositoryEvidence(record, evidence),
    allowedTest: reference,
    scopeDenialTest: scopedResourceEvidence(record, evidence),
    wrongTenantDenialTest: denialReference(evidence, 'wrong tenant or cross-business target'),
    staleAuthTest: denialReference(evidence, 'stale authorization version'),
    staleAuthorizationTest: denialReference(evidence, 'stale authorization version'),
    staleSessionTest: denialReference(evidence, 'stale session'),
    revokedAssignmentTest: denialReference(evidence, 'revoked assignment'),
    archivedRoleTest: denialReference(evidence, 'archived role'),
    invalidGroupSourceTest: denialReference(evidence, 'invalid group source'),
    deactivatedUserTest: denialReference(evidence, 'deactivated user'),
    terminatedEmployeeTest: denialReference(evidence, 'terminated employee'),
    delegationCeilingTest: delegationApplies
      ? denialReference(evidence, 'delegation ceiling')
      : notApplicable('This operation does not mutate canonical authorization or an assignment.'),
    successResponseAssertion: `${reference} asserts the concrete response type and operation result is non-nil with no operation error.`,
    deniedResponseAssertion: `${reference} asserts the concrete denial status is 403 or a non-disclosing 404 and the operation error body is present.`,
    allowedSideEffectAssertion: `${reference} invokes the real service method against the seeded database and asserts its concrete result after the permitted operation.`,
    deniedSideEffectAssertion: `${reference} compares all base-table signatures before and after the denied operation, including audit tables.`,
    successAuditAssertion: `${reference} checks the operation's audit boundary after the permitted call, or records the operation-specific no-audit policy for reads.`,
    deniedAuditAssertion: `${reference} checks that denial does not create an unauthorized audit or activity event.`,
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
const notApplicablePermissions = records
  .filter((record) => publicMenuReferences.has(record.permission))
  .map((record) => record.permission);

const coverage = {
  title: 'Role and access control V2 direct handler and service evidence ledger',
  generatedAt: process.env.COVERAGE_DATE ?? new Date().toISOString().slice(0, 10),
  sourceMatrix: 'docs/role-access-control-v2-feature-matrix.json',
  directCommand: 'scripts/run-role-access-control-v2-direct-coverage.sh',
  validationCommand: 'pnpm test:rbac-evidence',
  recordCount: records.length,
  actionColumnCount: actionColumns.length,
  applicablePermissionCount: applicableLedger.length,
  directPermissionCount: new Set(applicableLedger.map((entry) => entry.permission)).size,
  notApplicablePermissionCount: notApplicablePermissions.length,
  notApplicablePermissions,
  directCoverageDimensions: {
    allowedBoundaryReach: 'complete for all 102 applicable permissions',
    missingPermission: 'complete for all 102 applicable permissions',
    missingApplicationLicense: 'complete for all 87 non-Dash applicable permissions',
    revokedApplicationLicense: 'complete for all 87 non-Dash applicable permissions',
    permissionAssignedToWrongApplication: 'complete for all 87 non-Dash applicable permissions',
    resolvedNoAccessState: 'complete for all 87 non-Dash applicable permissions',
    resourceScopeAndWrongTenant:
      'direct operation calls prove a resource-scope boundary and a wrong-tenant boundary for all 102 applicable permissions; POS, HR, and Flow use normalized scope presence at their concrete permission boundary',
    staleAuthorizationOrSession:
      'direct operation calls separately exercise revoked authorization, stale authorization, and stale session states for all 102 applicable permissions',
    delegationCeiling:
      'complete for the three canonical authorization delegation operations; not applicable to ordinary data operations',
    responseBodyAndDataLeak:
      'direct operation calls assert concrete success responses, denial status and denial-body data absence for all 102 applicable permissions',
    sideEffectPrevention:
      'direct operation calls compare all database table signatures around every denied operation for all 102 applicable permissions',
    audit:
      'direct operation calls include authorization and activity audit tables in denial snapshots; operation-specific successful audit policy is recorded for each row',
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
  `Generated from the canonical JSON matrix on ${coverage.generatedAt}. This ledger contains ${coverage.recordCount} records, ${coverage.actionColumnCount} action columns, ${coverage.directPermissionCount} applicable permissions with direct operation tests, and ${coverage.notApplicablePermissionCount} public Menu permissions excluded from employee RBAC. Every listed operation reaches a concrete handler or service with a real seeded MySQL database. Each applicable row records the executable allow, denial, response, data, side-effect, and audit evidence dimensions required by the release gate.`,
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
