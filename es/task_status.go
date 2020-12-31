package es

import "github.com/tidwall/gjson"

func newTaskStatus(body string) TaskStatus {
	var completed bool
	var done int64
	var total int64
	var failure TaskStatusFailure

	if parsed := gjson.Get(body, "completed"); parsed.Exists() {
		completed = parsed.Bool()
	}

	if parsed := gjson.Get(body, "task.status"); parsed.Exists() {
		created := parsed.Get("created").Int()
		updated := parsed.Get("updated").Int()
		deleted := parsed.Get("deleted").Int()

		done = created + updated + deleted
		total = parsed.Get("total").Int()
	}

	if parsed := gjson.Get(body, "response.failures.0"); parsed.Exists() {
		failure = TaskStatusFailure{
			Id:          parsed.Get("id").String(),
			CauseType:   parsed.Get("cause.type").String(),
			CauseReason: parsed.Get("cause.reason").String(),
		}
	}

	return TaskStatus{
		IsCompleted: completed,
		Done:        done,
		Total:       total,
		Failure:     failure,
	}
}

type TaskStatus struct {
	IsCompleted bool
	Done        int64
	Total       int64
	Failure     TaskStatusFailure
}

type TaskStatusFailure struct {
	Id          string
	CauseType   string
	CauseReason string
}
