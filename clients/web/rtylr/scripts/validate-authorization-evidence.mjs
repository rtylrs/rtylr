import fs from 'node:fs';
import path from 'node:path';
import { execFileSync } from 'node:child_process';

const root = path.resolve(path.dirname(new URL(import.meta.url).pathname), '..');
const docs = path.join(root, 'docs');
const matrix = JSON.parse(
  fs.readFileSync(path.join(docs, 'authorization-feature-matrix.json'), 'utf8')
);
const coverage = JSON.parse(
  fs.readFileSync(path.join(docs, 'authorization-handler-coverage.json'), 'utf8')
);
const failures = [];
const actionColumns = matrix.actionColumns;
const publicMenu = new Set(['menu:orders:view', 'menu:analytics:view']);

function fail(message) {
  failures.push(message);
}

if (matrix.records.length !== 104)
  fail(`expected 104 matrix records, got ${matrix.records.length}`);
if (actionColumns.length !== 22) fail(`expected 22 action columns, got ${actionColumns.length}`);
if (coverage.recordCount !== 104) fail(`coverage recordCount is ${coverage.recordCount}`);
if (coverage.actionColumnCount !== 22)
  fail(`coverage actionColumnCount is ${coverage.actionColumnCount}`);

const applicableRecords = matrix.records.filter((record) => !publicMenu.has(record.permissionKey));
const publicRecords = matrix.records.filter((record) => publicMenu.has(record.permissionKey));
if (applicableRecords.length !== 102)
  fail(`expected 102 applicable records, got ${applicableRecords.length}`);
if (publicRecords.length !== 2) fail(`expected 2 public Menu records, got ${publicRecords.length}`);
if (
  JSON.stringify(publicRecords.map((record) => record.permissionKey).sort()) !==
  JSON.stringify([...publicMenu].sort())
) {
  fail('public Menu classification does not match the approved two-row set');
}

const onionRoot =
  process.env.ONION_DIR ?? path.resolve(root, '..', '..', '..', 'services', 'onion');
if (!fs.existsSync(onionRoot)) {
  console.error(`onion checkout not found at ${onionRoot}; set ONION_DIR to its path`);
  process.exit(1);
}
const directHarnessPath = path.join(onionRoot, 'internal/testutil/direct_authorization.go');
const directHarness = fs.existsSync(directHarnessPath)
  ? fs.readFileSync(directHarnessPath, 'utf8')
  : '';
for (const requiredOperationBoundary of [
  'resource scope denial',
  'wrong tenant or cross-business target',
  'stale authorization version',
  'stale session',
  'revoked assignment',
  'archived role',
  'invalid group source',
  'deactivated user',
  'terminated employee',
  'delegation ceiling',
]) {
  if (!directHarness.includes(requiredOperationBoundary)) {
    fail(`direct authorization harness is missing ${requiredOperationBoundary} operation coverage`);
  }
}
const directFiles = execFileSync('rg', ['--files', 'internal/modules'], {
  cwd: onionRoot,
  encoding: 'utf8',
})
  .trim()
  .split('\n')
  .filter((file) => /\/direct_(authorization_matrix|sensitive_authorization)_test\.go$/.test(file))
  .filter(Boolean);
if (directFiles.length !== 36)
  fail(`expected 36 direct operation test files, got ${directFiles.length}`);

for (const record of applicableRecords) {
  const refs = record.testReferences ?? [];
  const directReference = refs.find((ref) => ref.includes('direct_'));
  const directFile = directReference ? path.join(onionRoot, directReference.split('#')[0]) : '';
  const directTest = directReference?.split('#')[1]?.split('/')[0] ?? '';
  if (
    !directReference ||
    !fs.existsSync(directFile) ||
    !fs.readFileSync(directFile, 'utf8').includes(`func ${directTest}(`)
  ) {
    fail(`${record.permissionKey} lacks an operation-level direct test reference`);
  }
}

const ledger = coverage.evidenceLedger ?? [];
if (ledger.length !== 102) fail(`expected 102 evidence ledger rows, got ${ledger.length}`);
if ((coverage.uncoveredActions ?? []).length !== 0)
  fail(`uncovered actions remain: ${coverage.uncoveredActions.join(', ')}`);
if (coverage.directPermissionCount !== 102)
  fail(`expected 102 direct permissions, got ${coverage.directPermissionCount}`);
if (coverage.notApplicablePermissionCount !== 2)
  fail(`expected 2 not-applicable permissions, got ${coverage.notApplicablePermissionCount}`);

const incompleteRows = ledger.filter((entry) => entry.status !== 'covered');
if (incompleteRows.length > 0) {
  fail(
    `${incompleteRows.length} applicable ledger rows are not release-complete: ${incompleteRows
      .map((entry) => `${entry.permission}:${entry.action}`)
      .join(', ')}`
  );
}

const requiredFields = [
  'app',
  'permission',
  'action',
  'appLicenseRequirement',
  'permissionAndRoleRequirement',
  'resourceScopeRequirement',
  'tenantRequirement',
  'delegationRequirement',
  'realRoute',
  'realHandler',
  'realServiceMethod',
  'repositoryQueryOrScopeFilter',
  'allowedTestName',
  'scopeDenialTestName',
  'wrongTenantDenialTestName',
  'staleAuthorizationDenialTestName',
  'delegationDenialTestName',
  'responseAssertion',
  'dataLeakAssertion',
  'sideEffectAssertion',
  'auditAssertion',
  'matrixEvidence',
  'status',
];

const operationEvidenceFields = [
  'allowedTest',
  'scopeDenialTest',
  'wrongTenantDenialTest',
  'staleAuthorizationTest',
  'staleSessionTest',
  'revokedAssignmentTest',
  'archivedRoleTest',
  'invalidGroupSourceTest',
  'deactivatedUserTest',
  'terminatedEmployeeTest',
  'delegationCeilingTest',
  'successResponseAssertion',
  'deniedResponseAssertion',
  'dataLeakAssertion',
  'allowedSideEffectAssertion',
  'deniedSideEffectAssertion',
  'successAuditAssertion',
  'deniedAuditAssertion',
];

const projectionNotApplicableFields = new Set([
  'scopeDenialTest',
  'wrongTenantDenialTest',
  'staleAuthorizationTest',
  'staleSessionTest',
  'revokedAssignmentTest',
  'archivedRoleTest',
  'invalidGroupSourceTest',
  'deactivatedUserTest',
  'terminatedEmployeeTest',
  'allowedSideEffectAssertion',
  'deniedSideEffectAssertion',
  'successAuditAssertion',
  'deniedAuditAssertion',
]);

function isNotApplicable(value) {
  return String(value).startsWith('not_applicable:');
}

function isDirectEvidence(value) {
  return String(value).includes('direct_');
}

function validateOperationEvidence(entry, field) {
  const value = entry[field];
  if (isNotApplicable(value)) {
    const isFieldProjection = entry.evidenceKind === 'sensitive_field_projection';
    if (
      field !== 'delegationCeilingTest' &&
      !(isFieldProjection && projectionNotApplicableFields.has(field))
    ) {
      fail(`${entry.permission}:${entry.action} may not mark ${field} not_applicable`);
    }
    return;
  }
  if (!isDirectEvidence(value)) {
    fail(`${entry.permission}:${entry.action} ${field} is not direct operation evidence`);
  }
}

for (const entry of ledger) {
  for (const field of requiredFields) {
    const value = entry[field];
    if (
      value === undefined ||
      value === null ||
      value === '' ||
      (Array.isArray(value) && value.length === 0)
    ) {
      fail(`${entry.permission ?? 'unknown'}:${entry.action ?? 'unknown'} has empty ${field}`);
    }
  }
  if (!String(entry.allowedTestName).includes('direct_'))
    fail(`${entry.permission}:${entry.action} allowed evidence is not direct`);
  for (const field of [
    'scopeDenialTestName',
    'wrongTenantDenialTestName',
    'staleAuthorizationDenialTestName',
  ]) {
    if (
      !String(entry[field]).includes('direct_') &&
      !String(entry[field]).startsWith('not_applicable:')
    ) {
      fail(`${entry.permission}:${entry.action} ${field} is not operation-level evidence`);
    }
  }
  if (entry.delegationRequirement.startsWith('not_applicable:')) {
    if (!String(entry.delegationDenialTestName).startsWith('not_applicable:')) {
      fail(
        `${entry.permission}:${entry.action} has delegation evidence without an applicable delegation requirement`
      );
    }
  } else if (!String(entry.delegationDenialTestName).includes('direct_')) {
    fail(`${entry.permission}:${entry.action} lacks direct delegation-ceiling evidence`);
  }
  if (entry.status !== 'covered') {
    fail(`${entry.permission}:${entry.action} status must be covered for the release gate`);
  }
  for (const field of operationEvidenceFields) validateOperationEvidence(entry, field);
  if (
    !String(entry.operationSpecificRepositoryEvidence).includes('.go::') ||
    !String(entry.operationSpecificRepositoryEvidence).includes('direct_')
  ) {
    fail(`${entry.permission}:${entry.action} lacks operation-specific repository evidence`);
  }
  const repositoryReference = String(entry.operationSpecificRepositoryEvidence).split('::')[0];
  if (
    !repositoryReference.startsWith('internal/modules/') ||
    !fs.existsSync(path.join(onionRoot, repositoryReference))
  ) {
    fail(
      `${entry.permission}:${entry.action} repository evidence does not point to an existing Onion source file`
    );
  }
  const forbiddenEvidence = [
    'scope-validated',
    'repository filters scope',
    'full operation-specific',
    'remains required',
    'not proven',
    'not_proven:',
  ];
  for (const field of ['operationSpecificRepositoryEvidence', ...operationEvidenceFields]) {
    const text = String(entry[field]).toLowerCase();
    if (forbiddenEvidence.some((phrase) => text.includes(phrase))) {
      fail(`${entry.permission}:${entry.action} contains incomplete evidence language in ${field}`);
    }
  }
  if (
    entry.matrixEvidence.some((reference) => reference.includes('feature_matrix_test.go')) &&
    entry.matrixEvidence.length === 1
  ) {
    fail(`${entry.permission}:${entry.action} relies only on central policy coverage`);
  }
}

for (const record of publicRecords) {
  const evidence = coverage.records.find(
    (candidate) => candidate.permission === record.permissionKey
  );
  if (!evidence || evidence.actions.view.status !== 'not_applicable') {
    fail(`${record.permissionKey} is not explicitly marked not_applicable`);
  }
}

if (failures.length > 0) {
  console.error(`Authorization evidence validation failed with ${failures.length} issue(s):`);
  for (const failure of failures) console.error(`- ${failure}`);
  process.exitCode = 1;
} else {
  console.log(
    `Authorization evidence validation passed: ${coverage.recordCount} records, ${coverage.actionColumnCount} actions, ${coverage.directServiceOperationCount} API/service operation tests, ${coverage.sensitiveProjectionCount} sensitive-field projection tests, ${coverage.notApplicablePermissionCount} public Menu not-applicable records, and no uncovered applicable cells.`
  );
}
