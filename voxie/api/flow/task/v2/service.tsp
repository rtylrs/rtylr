import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace flow.task.v2;

@route("/flow/task/v2")
interface TaskService {
  @get
  @route("/tasks")
  listTasks(...ListTasksRequest): ListTasksResponse;

  @get
  @route("/tasks/my")
  getMyTasks(...GetMyTasksRequest): GetMyTasksResponse;

  @get
  @route("/tasks/my-project")
  getMyTasksProject(...GetMyTasksProjectRequest): GetMyTasksProjectResponse;

  @get
  @route("/tasks/search")
  searchTasks(...SearchTasksRequest): SearchTasksResponse;

  @post
  @route("/tasks")
  createTask(...CreateTaskRequest): CreateTaskResponse;

  @get
  @route("/tasks/{id}")
  getTaskDetail(...GetTaskDetailRequest): GetTaskDetailResponse;

  @put
  @route("/tasks/{id}")
  updateTask(...UpdateTaskRequest): UpdateTaskResponse;

  @patch
  @route("/tasks/{id}/status")
  updateTaskStatus(...UpdateTaskStatusRequest): UpdateTaskStatusResponse;

  @delete
  @route("/tasks/{id}")
  deleteTask(...DeleteTaskRequest): DeleteTaskResponse;

  @post
  @route("/tasks/subtasks")
  createSubtask(...CreateSubtaskRequest): CreateSubtaskResponse;

  @get
  @route("/tasks/{task_id}/subtasks")
  getSubtasks(...GetSubtasksRequest): GetSubtasksResponse;

  @post
  @route("/tasks/{task_id}/comments")
  addComment(...AddCommentRequest): AddCommentResponse;

  @delete
  @route("/tasks/comments/{id}")
  deleteComment(...DeleteCommentRequest): DeleteCommentResponse;

  @put
  @route("/tasks/comments/{id}")
  updateComment(...UpdateCommentRequest): UpdateCommentResponse;

  @get
  @route("/tasks/{task_id}/activity")
  getTaskActivity(...GetTaskActivityRequest): GetTaskActivityResponse;

  @post
  @route("/tasks/attachments")
  createAttachment(...CreateAttachmentRequest): CreateAttachmentResponse;

  @get
  @route("/tasks/{task_id}/attachments")
  getAttachments(...GetAttachmentsRequest): GetAttachmentsResponse;

  @delete
  @route("/tasks/attachments/{id}")
  deleteAttachment(...DeleteAttachmentRequest): DeleteAttachmentResponse;

  /**
   * createTaskLink adds a pasted URL (PR, issue, design, preview, error) to a
   * task. The server recognises the link type and, for public pages, fetches
   * the title once. Adding a URL the task already has returns that link.
   */
  @post
  @route("/tasks/links")
  createTaskLink(...CreateTaskLinkRequest): CreateTaskLinkResponse;

  @get
  @route("/tasks/{task_id}/links")
  listTaskLinks(...ListTaskLinksRequest): ListTaskLinksResponse;

  @delete
  @route("/tasks/links/{id}")
  deleteTaskLink(...DeleteTaskLinkRequest): DeleteTaskLinkResponse;

  /**
   * setTaskFieldValues writes or clears several custom field values of one
   * task in a single transaction. Needs the right to edit the task. The task
   * id travels in the body, like setTaskStatus, because a path of
   * /tasks/{task_id}/fields would collide with PUT /tasks/comments/{id}.
   */
  @put
  @route("/tasks/field-values")
  setTaskFieldValues(...SetTaskFieldValuesRequest): SetTaskFieldValuesResponse;

  @post
  @route("/tasks/{task_id}/time-logs")
  createTimeLog(...CreateTimeLogRequest): CreateTimeLogResponse;

  @get
  @route("/tasks/{task_id}/time-logs")
  getTimeLogs(...GetTimeLogsRequest): GetTimeLogsResponse;

  @delete
  @route("/tasks/time-logs/{id}")
  deleteTimeLog(...DeleteTimeLogRequest): DeleteTimeLogResponse;

  @post
  @route("/tasks/{task_id}/dependencies")
  addTaskDependency(...AddTaskDependencyRequest): AddTaskDependencyResponse;

  @delete
  @route("/tasks/{task_id}/dependencies/{depends_on_task_id}")
  removeTaskDependency(
    ...RemoveTaskDependencyRequest,
  ): RemoveTaskDependencyResponse;

  /**
   * addTaskRelation links two tasks with a typed relation: blocks, relates,
   * duplicates or clones. A self link is 400, an existing relation is 409 and
   * a blocks relation that would close a cycle is 409.
   */
  @post
  @route("/tasks/{task_id}/relations")
  addTaskRelation(...AddTaskRelationRequest): AddTaskRelationResponse;

  /**
   * listTaskRelations returns a task's relations grouped by type and
   * direction. addTaskDependency rows appear here as blocked_by.
   */
  @get
  @route("/tasks/{task_id}/relations")
  listTaskRelations(...ListTaskRelationsRequest): ListTaskRelationsResponse;

  @delete
  @route("/tasks/{task_id}/relations/{relation_id}")
  removeTaskRelation(...RemoveTaskRelationRequest): RemoveTaskRelationResponse;

  /**
   * setTaskParent moves a task under an epic or a task, or clears both. The
   * hierarchy is epic, task, subtask: the parent must be in the same project,
   * an epic takes no parent, and nothing may sit under a subtask.
   */
  @patch
  @route("/tasks/{id}/parent")
  setTaskParent(...SetTaskParentRequest): SetTaskParentResponse;

  /**
   * listEpicRollups summarises every epic of one project: its tasks by status
   * bucket, their estimates and the time logged on them.
   */
  @get
  @route("/tasks/epic-rollups")
  listEpicRollups(...ListEpicRollupsRequest): ListEpicRollupsResponse;

  /**
   * duplicateTask copies a task with its description, fields, labels, milestone
   * and subtasks into the same project, in the default status, and links the
   * copy to the original.
   */
  @post
  @route("/tasks/{id}/duplicate")
  duplicateTask(...DuplicateTaskRequest): DuplicateTaskResponse;

  /**
   * importTasks creates many tasks in one project. Every row is checked first;
   * if any row is invalid nothing is created and the response names the rows.
   */
  @post
  @route("/tasks/import")
  importTasks(...ImportTasksRequest): ImportTasksResponse;
}
