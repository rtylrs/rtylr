import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace hr.department.v1;

/** DepartmentService defines HR department endpoints. */
@route("/hr/department/v1")
interface DepartmentService {
  /** ListDepartments returns departments for the active business. */
  @get
  @route("/departments")
  listDepartments(...ListDepartmentsRequest): ListDepartmentsResponse;

  /** CreateDepartment creates a new department. */
  @post
  @route("/departments")
  createDepartment(...CreateDepartmentRequest): CreateDepartmentResponse;

  /** UpdateDepartment updates an existing department. */
  @put
  @route("/departments/{id}")
  updateDepartment(...UpdateDepartmentRequest): UpdateDepartmentResponse;

  /** DeleteDepartment deletes a department. */
  @delete
  @route("/departments/{id}")
  deleteDepartment(...DeleteDepartmentRequest): DeleteDepartmentResponse;
}
