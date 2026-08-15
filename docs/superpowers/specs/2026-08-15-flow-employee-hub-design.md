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

## Two real gaps this surfaced, neither introduced by this work

Investigating what Flow's front door would actually call found the same
class of bug already fixed once today in Flow's own backend - twice, in HR,
independently.

### 1. `createLeaveRequest` trusts a caller-supplied `employee_id`

`CreateLeaveRequestRequest.employee_id` (`packages/voxie/api/hr/leave/v1/models.onk:105-118`)
travels straight through to the write with no check that it belongs to the
caller. `create_leave_request_handler.go:59-72` only verifies the caller is
authenticated, never that `req.EmployeeID` is theirs. Exactly the shape of
the hole closed in `WorkTimeAPI.scope()` earlier today (`3d754ec7`): any
authenticated user in the business can submit a leave request as any other
employee.

### 2. `listLeaveRequests`'s filters are declared but not implemented

The contract has optional `employee_id` and `status` query filters
(`models.onk:87-92`), but `list_leave_requests_handler.go:19-22` and
`domain.ListLeaveRequestsParams` (`types.go:59-62`) only take `limit`/`offset` -
there is no filtering code at all. Today, "list my leave requests" is not
actually expressible: the endpoint returns every request in the tenant,
unscoped, to whoever calls it.

### 3. `listStaff` returns salary and emergency-contact data to any authenticated caller

`staff_service.go:25-41` resolves only the business id and checks the caller
isn't anonymous - no role check, no department scoping. The `Staff` message
(`models.onk:4-31`) carries `salary`, `hourly_rate`, `overtime_rate`, `address`,
`emergency_contact_name`, `emergency_contact_phone` in the same payload as
name and email, with no leaner shape for a casual directory. This is
plausibly fine today because only HR-app users, already trusted with HR data,
call it. A Directory screen inside Flow would hand the same payload to every
employee in the business who opens it.

None of these are Flow's bugs to have introduced and none are new - they were
already true before this document existed. They are listed here because
building Leave and Directory on top of them, unfixed, would make Flow the
thing that turns a dormant hole into an actively-exercised one. Proposed
build order below treats fixing them as a prerequisite, not an afterthought,
the same way the standalone `scope()` fix shipped ahead of anything that
needed it.

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

Blocked on gap 3 above. Once resolved - either a leaner response shape that
drops salary/rate/emergency-contact fields, or a genuinely separate
"directory" projection - this is a straightforward business-wide list backed
by `hr.staff.v1`'s existing `listStaff`, gated by the new configurable
permission. No new backend concept needed beyond the field-shape fix and the
permission itself.

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

1. **Fix the two HR gaps**, on their own, the same way `scope()` shipped
   alone today: deny a cross-employee `employee_id` on
   `createLeaveRequest`/`listLeaveRequests` until real authorization exists,
   implement the declared list filters for real, add the cancel RPC, and
   resolve the `listStaff` over-exposure. This is HR's contract and HR's
   backend - it lands as its own separate, clearly-labeled change against
   HR's code, reviewed as the authorization-model change it is, independent
   of who ends up writing it.
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
the fix ships as its own separate, clearly-labeled change against HR's code

- never folded quietly into a Flow PR - because it alters another app's
  authorization model and deserves review as exactly that, independent of who
  ends up writing it or when.
