# Flow: from project tool to employee hub

Date: 2026-08-15
Status: **information architecture settled, awaiting build sign-off**
Builds on: `2026-08-14-flow-time-section-design.md`, which this does not replace -
the daily time-logging work in that document is one piece of what follows.

## What this is

Flow stops being only a project-management tool and becomes the place an
employee goes for everything they need to report to or ask of the company -
not by rebuilding HR's or Finance's systems a second time, but by putting a
front door on them. Flow reads and writes into the systems that already own
this data; those systems stay the system of record. This is the same
judgment call the time-tracking work already made once, generalized: Flow
writes `hr_time_entries` rather than inventing a second time store, and this
document extends that principle to leave and to staff data.

## Decided

**Scope.** Two groups, not one flat list:

- **My space** - the employee's own: My day (existing), Leave, Profile.
- **Company** - everyone's: the existing workspace Team/Reports screens,
  plus a new company-wide Directory and a new "who's in today" presence view.

**Expenses are out of this round.** Finance's `Expense` record has no
`employee_id` field at all - only `owner_id` (the tenant). There is no way
today to know who submitted a given expense, so a "your expenses" view is not
buildable without a contract change to another app's data model first. That
change is real work with its own owner and is not folded into this.

**Front door, not a second system.** For leave: Flow calls HR's real
`hr.leave.v1` service - list requests, submit one, read the balance. Flow
never stores a leave request. For the directory and profile: Flow calls HR's
real `hr.staff.v1` service. Nothing employee-facing in this document creates
a new table anywhere for data HR already owns.

**Directory and presence access is a real per-business setting**, not a
hardcoded rule - consistent with this repo's configurability rule (root
CLAUDE.md section 8): some businesses want every employee to see the company
roster, some don't, and that has to be an actual toggle a business owner
sets, not a decision baked into the code.

**Nav shape**, sidebar groups:

```
Projects            (existing, unchanged)
My space
  My day             (existing /time)
  Leave
  Profile
Company
  Team               (existing /time/team - workspace-scoped, unchanged)
  Reports            (existing /time/reports - workspace-scoped, unchanged)
  Directory          (new - business-wide)
  Who's in today     (new - business-wide)
Settings             (existing)
```

`Team` and `Reports` do not move to the new business-wide permission model -
they stay exactly as built, gated by `RequireWorkspaceRole`, because they
answer "who's on my team" (a workspace concept). `Directory` and `Who's in
today` answer "who works here at all" (a business concept) and need the new,
separate, configurable permission decided above. Putting all four under one
"Company" heading is a nav grouping, not a claim that they share one
permission check - they don't, and the build order below keeps that
distinction real in code, not just in the sidebar label.

## One real gap, one narrower than first thought, one that turned out not to exist

Starting the build, actually reading the code before touching it (rather than
trusting an earlier delegated investigation) found that two of the three
things reported here were wrong. Corrected below, with the earlier claim left
visible so the mistake stays legible rather than quietly disappearing.

### 1. `createLeaveRequest` trusts a caller-supplied `employee_id` - confirmed real

`CreateLeaveRequestRequest.employee_id` (`packages/voxie/api/hr/leave/v1/models.onk:105-118`)
travels straight through to the write with no check that it belongs to the
caller. `create_leave_request_handler.go:59-72` only verifies the caller is
authenticated, never that `req.EmployeeID` is theirs, and
`Service.CreateLeaveRequest` (`application/service.go:96-169`) never checks it
either. Exactly the shape of the hole closed in `WorkTimeAPI.scope()` earlier
today (`3d754ec7`): any authenticated user in the business can submit a leave
request as any other employee.

The fix does not need to be invented from scratch: `Service.ListLeaveRequests`
already has the right pattern next to it (see below) - resolve the caller's
own employee id, and only let a caller with the approval permission act on
someone else's.

### 2. `listLeaveRequests`'s filters - this was wrong, no gap exists

The original claim was that `employee_id`/`status` filters were declared in
the contract but never implemented. Reading `Service.ListLeaveRequests`
(`application/service.go:171-171` onward) before writing anything found a
real, already-wired `scopeLeaveRequests` helper
(`application/service.go:266-285`): it resolves the caller's own employee id
via `GetEmployeeIDByUserID`, and only lets the client-supplied `employee_id`
through when the caller holds `permissionApproveTimeOff` - otherwise the
request is forced to the caller's own employee id regardless of what was
asked for. This is exactly the authorization this document assumed was
missing. Nothing to fix here.

### 3. `listStaff` - real, but narrower: compensation is already gated, contact details are not

The original claim was that salary and emergency-contact data were both
exposed with no gate at all. `staff_service.go:38` and the dedicated
`canViewCompensation` helper (`api/permissions.go`) already fail closed on
`salary`/`hourly_rate`/`overtime_rate`, resolved the same way as the leave
fix above - via `GetEmployeeIDByUserID` plus a real permission check
(`manage_employees`/`manage_payroll`). That part already works.

What is still real: `newStaffPayload` (`staff_service.go:196-201`) only zeroes
the three compensation fields. `address`, `emergency_contact_name`, and
`emergency_contact_phone` pass through unconditionally to any authenticated
caller in the business, with no role check at all - `resolveBusinessID`
(`staff_service.go:183-194`) checks only that the caller is authenticated.
Narrower than first written, but a home address and an emergency contact are
still not something every employee should see about every colleague by
default.

Neither of these two real items is a bug this project introduced, and neither
is new - both predate this document. They stay listed as build prerequisites,
not deferred cleanup, because Flow becoming the front door means real traffic
starts flowing through paths that were previously only reachable from inside
the HR app.

## What each screen actually does

### Leave

Two views, both hitting `hr.leave.v1` directly:

- **List mine.** Once (1) below lands, `listLeaveRequests` can actually be
  called with `employee_id` scoped to the caller. Shows status, dates,
  leave type, and the balance for that type (`listLeaveBalances` already
  returns `entitled_days`/`used_days`/`pending_days`/`remaining_days` per
  type - this exists today, nothing new needed for balance).
- **Submit one.** `createLeaveRequest` needs `leave_type_id`, `start_date`,
  `end_date`, `total_days`, optional `notes`. Leave types are already
  business-configurable (`hr.leave.v1`'s `listLeaveTypes`), so the type
  picker is a real config-driven list, not a hardcoded enum - consistent
  with this repo's "never hardcode a business rule" rule.
- **Cancel one.** No cancel RPC exists in the contract today - decided below
  that this ships as a small addition alongside the other HR fixes, scoped
  to the caller's own request and only while it is still `pending`.

### Profile

Both halves the earlier round asked for:

- **HR record**, read-only: `hr.staff.v1`/`hr.employee.v1` plus the ad hoc
  (non-`.onk`, hand-rolled REST) `/hr/staff/detail` endpoint for photo/bio.
  That second endpoint being outside the generated-client pattern is a
  rougher integration than the rest of this document - still callable via
  `safe.get` (the sanctioned `safeFetch` path), just manually typed rather
  than contract-generated. Editing still only happens in HR; Flow's profile
  is read-only, full stop.
- **Flow activity**, Flow's own: hours logged (already have this data),
  leave balance and recent requests (once Leave above exists). Not a new
  concept, just the two halves of this document's own data pulled into one
  view.

### Directory

Blocked on gap 3 above (the narrower, corrected version: `address` and the
two emergency-contact fields, not salary - that part is already safe). Once
`newStaffPayload` gates those the same way it already gates compensation,
this is a straightforward business-wide list backed by `hr.staff.v1`'s
existing `listStaff`, gated by the new configurable permission. No new
backend concept needed beyond that field-gating fix and the permission
itself.

### Who's in today

The one genuinely new backend capability in this document. Everything built
today for `/time/team` (`ListWorkDaysForWorkspace`) resolves through one
workspace's membership; this needs the same resolution for every employee in
the business regardless of workspace, gated by the new business-wide
permission rather than `RequireWorkspaceRole`. Decided below: this is
**presence only** - mode (onsite/remote/hybrid), on-leave, or not-started-yet

- not the full `WorkDay` shape. No clocked/logged/gap minutes: a company-wide
  view watched by every employee is a different exposure than the
  workspace-admin-gated Team view, and showing colleagues' exact hours to
  people with no management relationship over them is not what "who's in
  today" is asking for.

## Log Hours, from the screenshot

Separately from the front-door work above, confirmed scope for the logging
screen itself:

- **A dedicated screen**, not the inline row at the bottom of My Day today.
- **Project → work type → configurable sub-task buttons.** The buttons
  (`DEV Web Development`, `DEV Web Maintenance`, etc. in the reference) are
  per-business configuration, not a fixed list - this is exactly what
  `flow_work_types` (workspace-scoped, two-level, admin-maintained tree) was
  already built for in the time-tracking amendment. No new schema.
- **Two description fields.** A first field is preset automatically from
  whichever sub-task button was picked (the reference shows "Build / revamp
  websites / Create / Edit web sections" - a breadcrumb of the classification
  chosen, not typed), and a second free-text field for what was actually
  done. Today's single free-text `description` field on `WorkDayLine`
  already exists and becomes the second field; the first is computed
  client-side from the work-type tree, not a new stored value.
- **No third location option.** Onsite/Remote/Hybrid stays as built.

## Build order

1. **Fix the two real HR gaps**, on their own, the same way `scope()` shipped
   alone today: deny a cross-employee `employee_id` on `createLeaveRequest`
   using the same `GetEmployeeIDByUserID` + approval-permission pattern
   `listLeaveRequests` already uses, add the cancel RPC, and gate
   `address`/`emergency_contact_*` in `listStaff` the same way compensation
   is already gated. This is HR's contract and HR's backend - it lands as
   its own separate, clearly-labeled change against HR's code, reviewed as
   the authorization-model change it is, independent of who ends up writing
   it.
2. **Log Hours rebuild** - independently shippable, no dependency on step 1.
3. **Leave** - depends on step 1's fixes to be safe to ship at all.
4. **Profile** - depends on nothing new; can land any time after step 2.
5. **Directory** - depends on step 1's field-shape fix.
6. **Who's in today** - depends on the new business-wide permission and a
   new backend query; the least-specified item here and the one most likely
   to need its own round of questions when it's actually picked up.

## Decisions on the three items left open

### 1. Cancel a pending leave request: add it, scoped tight

A self-service leave feature with no way to withdraw your own request will
surface as a real complaint the first time someone submits one by mistake or
their plans change - "submit" without "undo" is half a feature. Add a small
`cancelLeaveRequest` RPC alongside the other HR fixes: the caller may only
cancel their own request (same ownership check as the `employee_id` fix
above), and only while `status` is still `pending`. Once a request is
approved, cancelling it can have payroll or scheduling consequences a bare
status flip should not trigger silently - that path, if it's ever needed,
is a request to a manager, not a self-service button.

### 2. Who's in today: presence only, no minutes

The existing Team view is workspace-admin-gated - only someone who manages
that specific team sees it. A company-wide version watched by every employee
is a different exposure entirely: showing exact clocked/logged/gap minutes
for every colleague, including people with no management relationship to
each other, is peer surveillance wearing the costume of reference info. The
feature's own stated purpose - "who's onsite/remote/on leave right now" - is
about mode and status, not performance. Ship it thin: mode
(onsite/remote/hybrid), on-leave, or not-started-yet. No minutes, no gap
figures. Widening it later is a real, separate decision if it turns out to
be genuinely wanted - not the default to ship with.

### 3. Who owns fixing the two HR gaps: not mine to assign, but the process is

Which person at Voxire signs off on a change to HR's authorization surface
is not something to decide from inside a design doc. What can be decided:
the fix ships as its own separate, clearly-labeled change against HR's code -
never folded quietly into a Flow PR - because it alters another app's
authorization model and deserves review as exactly that, independent of who
ends up writing it or when.
