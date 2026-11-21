package domain

import "errors"

// Typed relations between two tasks of one business. They are stored in
// flow_task_dependencies, one row per relation with a fixed direction:
//
//	blocks      task_id is blocked by depends_on_task_id
//	duplicates  task_id duplicates depends_on_task_id
//	clones      task_id clones depends_on_task_id
//	relates     symmetric; the lower task id is stored in task_id
//
// Reads present a row from either end, so "A blocks B" shows on B as
// "blocked by A".
const (
	RelationBlocks     = "blocks"
	RelationRelates    = "relates"
	RelationDuplicates = "duplicates"
	RelationClones     = "clones"

	RelationOutward = "outward"
	RelationInward  = "inward"
)

var (
	ErrRelationInvalidType      = errors.New("relation type must be one of blocks, relates, duplicates, clones")
	ErrRelationInvalidDirection = errors.New("relation direction must be outward or inward")
	ErrRelationSelf             = errors.New("a task cannot be related to itself")
	ErrRelationExists           = errors.New("these tasks already have this relation")
	ErrRelationCycle            = errors.New("this relation would make the tasks block each other")
	ErrRelationNotFound         = errors.New("relation not found")
)

// TaskRelationEdge is a relation as stored.
type TaskRelationEdge struct {
	ID              int64
	TaskID          int64
	DependsOnTaskID int64
	LinkType        string
}

// NewTaskRelationEdge turns "taskID <relation> otherTaskID", read from taskID's
// side, into the stored row. An empty direction is outward.
func NewTaskRelationEdge(taskID, otherTaskID int64, relationType, direction string) (TaskRelationEdge, error) {
	if direction == "" {
		direction = RelationOutward
	}
	if direction != RelationOutward && direction != RelationInward {
		return TaskRelationEdge{}, ErrRelationInvalidDirection
	}
	if taskID == otherTaskID {
		return TaskRelationEdge{}, ErrRelationSelf
	}
	subject, object := taskID, otherTaskID
	if direction == RelationInward {
		subject, object = otherTaskID, taskID
	}
	switch relationType {
	case RelationBlocks:
		// subject blocks object, so the object is the one that depends.
		return TaskRelationEdge{TaskID: object, DependsOnTaskID: subject, LinkType: relationType}, nil
	case RelationDuplicates, RelationClones:
		return TaskRelationEdge{TaskID: subject, DependsOnTaskID: object, LinkType: relationType}, nil
	case RelationRelates:
		low, high := taskID, otherTaskID
		if low > high {
			low, high = high, low
		}
		return TaskRelationEdge{TaskID: low, DependsOnTaskID: high, LinkType: relationType}, nil
	}
	return TaskRelationEdge{}, ErrRelationInvalidType
}

// DirectionFrom reports how the relation reads from viewerTaskID's side.
func (e TaskRelationEdge) DirectionFrom(viewerTaskID int64) string {
	switch e.LinkType {
	case RelationRelates:
		return RelationOutward
	case RelationBlocks:
		if viewerTaskID == e.DependsOnTaskID {
			return RelationOutward
		}
		return RelationInward
	}
	if viewerTaskID == e.TaskID {
		return RelationOutward
	}
	return RelationInward
}

// OtherTaskID is the end of the relation that is not viewerTaskID.
func (e TaskRelationEdge) OtherTaskID(viewerTaskID int64) int64 {
	if viewerTaskID == e.TaskID {
		return e.DependsOnTaskID
	}
	return e.TaskID
}

// RelationKey names a type and direction as one label key.
func RelationKey(relationType, direction string) string {
	inward := direction == RelationInward
	switch relationType {
	case RelationBlocks:
		if inward {
			return "blocked_by"
		}
		return "blocks"
	case RelationRelates:
		return "relates_to"
	case RelationDuplicates:
		if inward {
			return "duplicated_by"
		}
		return "duplicates"
	case RelationClones:
		if inward {
			return "cloned_by"
		}
		return "clones"
	}
	return ""
}

// RelatedTask is the task at the other end of a relation.
type RelatedTask struct {
	ID               int64
	ProjectID        int64
	Identifier       string
	Title            string
	ProjectName      string
	Status           string
	StatusName       string
	StatusColor      string
	StatusBucket     string
	AssigneeID       *int64
	AssigneeName     string
	AssigneeInitials string
}

// TaskRelation is one relation read from a task's side.
type TaskRelation struct {
	ID        int64
	Type      string
	Direction string
	Task      RelatedTask
}

type TaskRelationGroup struct {
	Type      string
	Direction string
	Relation  string
	Relations []TaskRelation
}

var relationGroupOrder = []struct{ relationType, direction string }{
	{RelationBlocks, RelationOutward},
	{RelationBlocks, RelationInward},
	{RelationRelates, RelationOutward},
	{RelationDuplicates, RelationOutward},
	{RelationDuplicates, RelationInward},
	{RelationClones, RelationOutward},
	{RelationClones, RelationInward},
}

// GroupTaskRelations buckets relations by type and direction in a fixed order
// and drops empty groups. Order inside a group is the input order.
func GroupTaskRelations(relations []TaskRelation) []TaskRelationGroup {
	groups := make([]TaskRelationGroup, 0, len(relationGroupOrder))
	for _, slot := range relationGroupOrder {
		group := TaskRelationGroup{
			Type:      slot.relationType,
			Direction: slot.direction,
			Relation:  RelationKey(slot.relationType, slot.direction),
		}
		for _, relation := range relations {
			if relation.Type == slot.relationType && relation.Direction == slot.direction {
				group.Relations = append(group.Relations, relation)
			}
		}
		if len(group.Relations) > 0 {
			groups = append(groups, group)
		}
	}
	return groups
}

// TaskRef is the minimum needed to publish a task event.
type TaskRef struct {
	ID        int64
	ProjectID int64
}

// TaskRelationChange is the result of adding or removing a relation: its id and
// both ends, so the caller can tell each end's subscribers.
type TaskRelationChange struct {
	RelationID int64
	Tasks      [2]TaskRef
}

// UnblockedTask is a task whose last open blocker has just closed.
type UnblockedTask struct {
	ID           int64
	ProjectID    int64
	WorkspaceID  int64
	OwnerID      int64
	Title        string
	AssigneeID   int64
	ProjectName  string
	ProjectColor string
}
