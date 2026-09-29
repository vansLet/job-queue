package web

import "jobqueue/model"

func FromRequest(r TaskRequest) model.Task {
	return model.Task{
		Type:    r.Type,
		Payload: r.Payload,
	}
}
func ToRespone(t model.Task) ResponeTask {
	return ResponeTask{
		ID:       t.ID,
		Type:     t.Type,
		Payload:  t.Payload,
		Status:   t.Status.String(),
		CreateAt: t.CreateAt,
	}
}
