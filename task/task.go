package task

import (
	"github.com/google/uuid"
	"github.com/moonstripe/workflow-dag/graph"
)

type ActionInput map[string]string

type ActionStatus int

const (
	Queued ActionStatus = iota
	Running
	Waiting
	Finished
	Failed
)

func (aS ActionStatus) String() string {
	switch aS {
	case Queued:
		return "queued"
	case Running:
		return "running"
	case Waiting:
		return "waiting"
	case Finished:
		return "finished"
	case Failed:
		return "failed"
	default:
		return "unknown state"
	}
}

type ActionOutput struct {
	ActionId uuid.UUID         `json:"action_id"`
	Status   ActionStatus      `json:"status"`
	Data     map[string]string `json:"data"`
}

type Actionable interface {
	Conduct(aI ActionInput) ActionOutput
	String() string
}

type Task struct {
	TaskID  uuid.UUID    `json:"task_id"`
	Actions []Actionable `json:"actions"`
}

func (t Task) Id() uuid.UUID {
	return t.TaskID
}

func (t Task) Label() string {
	return t.TaskID.String()[:8]
}

type TaskGraph struct {
	tasks []graph.Node
	adj   map[graph.Node][]graph.Node
}
