package domain

import (
	"errors"
	"testing"
)

func int64Ptr(v int64) *int64 { return &v }

func TestPlaceUnderParent(t *testing.T) {
	const business, project = int64(7), int64(70)
	node := func(id int64, taskType string) HierarchyNode {
		return HierarchyNode{ID: id, ProjectID: project, BusinessID: business, TaskType: taskType}
	}
	epic := node(1, TaskTypeEpic)
	task := node(2, TaskTypeTask)
	taskUnderEpic := node(3, TaskTypeBug)
	taskUnderEpic.EpicID = int64Ptr(1)
	subtask := node(4, TaskTypeTask)
	subtask.ParentTaskID = int64Ptr(2)
	taskWithSubtasks := node(5, TaskTypeFeature)
	taskWithSubtasks.Subtasks = 2
	otherProject := node(6, TaskTypeTask)
	otherProject.ProjectID = project + 1
	otherBusinessEpic := node(8, TaskTypeEpic)
	otherBusinessEpic.BusinessID = business + 1
	subtaskOfChild := node(9, TaskTypeTask)
	subtaskOfChild.ParentTaskID = int64Ptr(10)

	cases := []struct {
		name       string
		child      HierarchyNode
		parent     HierarchyNode
		wantErr    error
		wantParent int64
		wantEpic   int64
	}{
		{name: "task under an epic", child: task, parent: epic, wantEpic: 1},
		{name: "new task under an epic", child: node(0, TaskTypeBug), parent: epic, wantEpic: 1},
		{name: "task with subtasks under an epic", child: taskWithSubtasks, parent: epic, wantEpic: 1},
		{name: "subtask promoted under an epic", child: subtask, parent: epic, wantEpic: 1},
		{name: "task becomes a subtask", child: node(10, TaskTypeTask), parent: task, wantParent: 2},
		{name: "new subtask of a task under an epic", child: node(0, TaskTypeTask), parent: taskUnderEpic, wantParent: 3},
		{name: "task under an epic becomes a subtask", child: taskUnderEpic, parent: task, wantParent: 2},
		{name: "epic under an epic", child: node(11, TaskTypeEpic), parent: epic, wantErr: ErrEpicCannotHaveParent},
		{name: "epic under a task", child: epic, parent: task, wantErr: ErrEpicCannotHaveParent},
		{name: "new epic under a task", child: node(0, TaskTypeEpic), parent: task, wantErr: ErrEpicCannotHaveParent},
		{name: "task under itself", child: task, parent: task, wantErr: ErrTaskParentCycle},
		{name: "task under its own subtask", child: node(10, TaskTypeTask), parent: subtaskOfChild, wantErr: ErrTaskParentCycle},
		{name: "subtask of a subtask", child: node(12, TaskTypeTask), parent: subtask, wantErr: ErrTaskParentTooDeep},
		{name: "new subtask of a subtask", child: node(0, TaskTypeTask), parent: subtask, wantErr: ErrTaskParentTooDeep},
		{name: "task with subtasks becomes a subtask", child: taskWithSubtasks, parent: task, wantErr: ErrTaskParentTooDeep},
		{name: "parent in another project", child: task, parent: otherProject, wantErr: ErrTaskParentOtherProject},
		{name: "parent in another business", child: task, parent: otherBusinessEpic, wantErr: ErrTaskParentNotFound},
		{name: "missing parent", child: task, parent: HierarchyNode{}, wantErr: ErrTaskParentNotFound},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := PlaceUnderParent(tc.child, tc.parent)
			if !errors.Is(err, tc.wantErr) {
				t.Fatalf("error = %v, want %v", err, tc.wantErr)
			}
			if tc.wantErr != nil {
				return
			}
			if (tc.wantParent == 0) != (got.ParentTaskID == nil) || (got.ParentTaskID != nil && *got.ParentTaskID != tc.wantParent) {
				t.Fatalf("parent = %v, want %d", got.ParentTaskID, tc.wantParent)
			}
			if (tc.wantEpic == 0) != (got.EpicID == nil) || (got.EpicID != nil && *got.EpicID != tc.wantEpic) {
				t.Fatalf("epic = %v, want %d", got.EpicID, tc.wantEpic)
			}
		})
	}
}

func TestValidateTaskTypeChange(t *testing.T) {
	plain := HierarchyNode{ID: 1, TaskType: TaskTypeTask}
	subtask := HierarchyNode{ID: 2, TaskType: TaskTypeTask, ParentTaskID: int64Ptr(1)}
	underEpic := HierarchyNode{ID: 3, TaskType: TaskTypeBug, EpicID: int64Ptr(9)}
	withSubtasks := HierarchyNode{ID: 4, TaskType: TaskTypeTask, Subtasks: 1}
	emptyEpic := HierarchyNode{ID: 5, TaskType: TaskTypeEpic}
	usedEpic := HierarchyNode{ID: 6, TaskType: TaskTypeEpic, EpicChildren: 3}

	cases := []struct {
		name    string
		task    HierarchyNode
		newType string
		wantErr error
	}{
		{"task to bug", plain, TaskTypeBug, nil},
		{"task to epic", plain, TaskTypeEpic, nil},
		{"subtask to feature", subtask, TaskTypeFeature, nil},
		{"subtask to epic", subtask, TaskTypeEpic, ErrEpicCannotHaveParent},
		{"task under an epic to epic", underEpic, TaskTypeEpic, ErrEpicCannotHaveParent},
		{"task with subtasks to epic", withSubtasks, TaskTypeEpic, ErrEpicHasSubtasks},
		{"empty epic to task", emptyEpic, TaskTypeTask, nil},
		{"epic with tasks to task", usedEpic, TaskTypeTask, ErrEpicHasChildren},
		{"epic with tasks stays an epic", usedEpic, TaskTypeEpic, nil},
		{"unknown type", plain, "story", ErrInvalidTaskType},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if err := ValidateTaskTypeChange(tc.task, tc.newType); !errors.Is(err, tc.wantErr) {
				t.Fatalf("error = %v, want %v", err, tc.wantErr)
			}
		})
	}
}
