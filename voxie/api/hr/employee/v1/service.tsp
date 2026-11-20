import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace hr.employee.v1;

/** EmployeeService defines HR employee endpoints. */
@route("/hr/employee/v1")
interface EmployeeService {
  /** ListEmployees returns employees with optional filters. */
  @get
  @route("/employees")
  listEmployees(...ListEmployeesRequest): ListEmployeesResponse;

  /** CreateEmployee creates a new employee. */
  @post
  @route("/employees")
  createEmployee(...CreateEmployeeRequest): CreateEmployeeResponse;

  /** UpdateEmployee updates an existing employee. */
  @put
  @route("/employees/{id}")
  updateEmployee(...UpdateEmployeeRequest): UpdateEmployeeResponse;

  /** InviteEmployee sends an employee invitation. */
  @post
  @route("/employees/invite")
  inviteEmployee(...InviteEmployeeRequest): InviteEmployeeResponse;

  /** AcceptInvite accepts an employee invitation. */
  @post
  @route("/employees/invite/accept")
  acceptInvite(...AcceptInviteRequest): AcceptInviteResponse;
}
