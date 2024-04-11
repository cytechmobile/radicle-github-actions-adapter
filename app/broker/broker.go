package broker

import (
	"context"
	"fmt"
)

type RequestMessageType string

const (
	RequestMessageTypePush  RequestMessageType = "push"
	RequestMessageTypePatch RequestMessageType = "patch"
)

type RequestTypeMessage struct {
	Request   string             `json:"request"`
	EventType RequestMessageType `json:"event_type"`
}

type RequestMessage struct {
	Repo       string                    `json:"repo"`
	Commit     string                    `json:"commit"`
	PushEvent  *RequestPushEventMessage  `json:"push_event"`
	PatchEvent *RequestPatchEventMessage `json:"patch_event"`
}

func (rm *RequestMessage) String() string {
	return fmt.Sprintf("RequestMessage{Repo:%+v, Commit:%+v, PushEvent:%+v, PatchEvent:%+v}", rm.Repo, rm.Commit,
		*rm.PushEvent, *rm.PatchEvent)
}

type RequestPushEventMessage struct {
	Request    string     `json:"request"`
	EventType  string     `json:"event_type"`
	Pusher     Pusher     `json:"pusher"`
	Before     string     `json:"before"`
	After      string     `json:"after"`
	Commits    []string   `json:"commits"`
	Repository Repository `json:"repository"`
}

type RequestPatchEventMessage struct {
	Request    string       `json:"request"`
	EventType  string       `json:"event_type"`
	Action     string       `json:"action"`
	Patch      PatchDetails `json:"patch"`
	Repository Repository   `json:"repository"`
}

type PatchDetails struct {
	ID        string          `json:"id"`
	Author    Author          `json:"author"`
	Title     string          `json:"title"`
	State     PatchState      `json:"state"`
	Before    string          `json:"before"`
	After     string          `json:"after"`
	Commits   []string        `json:"commits"`
	Target    string          `json:"target"`
	Labels    []string        `json:"labels"`
	Assignees []string        `json:"assignees"`
	Revisions []PatchRevision `json:"revisions"`
}

type PatchState struct {
	Status    string          `json:"status"`
	Conflicts []PatchConflict `json:"conflicts"`
}

type PatchConflict struct {
	RevisionID string `json:"revision_id"`
	Oid        string `json:"oid"`
}

type PatchRevision struct {
	ID          string `json:"id"`
	Author      Author `json:"author"`
	Description string `json:"description"`
	Base        string `json:"base"`
	Oid         string `json:"oid"`
	Timestamp   int64  `json:"timestamp"`
}

type Repository struct {
	ID            string   `json:"id"`
	Name          string   `json:"name"`
	Description   string   `json:"description"`
	Private       bool     `json:"private"`
	DefaultBranch string   `json:"default_branch"`
	Delegates     []string `json:"delegates"`
}

type Pusher struct {
	ID    string `json:"id"`
	Alias string `json:"alias"`
}

type Author struct {
	ID    string `json:"id"`
	Alias string `json:"alias"`
}

type ResponseMessage struct {
	Response      string            `json:"response"`
	RunID         *RunID            `json:"run_id,omitempty"`
	Result        string            `json:"result,omitempty"`
	ResultDetails []WorkflowDetails `json:"-"`
}

func (rm *ResponseMessage) String() string {
	return fmt.Sprintf("ResponseMessage{Response:%+v, RunID:%+v, Result:%+v, ResultDetails:%+v}", rm.Response,
		*rm.RunID, rm.Result, rm.ResultDetails)
}

type WorkflowDetails struct {
	WorkflowID        string             `json:"workflow_id"`
	WorkflowName      string             `json:"workflow_name"`
	WorkflowResult    string             `json:"workflow_result"`
	WorkflowArtifacts []WorkflowArtifact `json:"workflow_artifacts"`
}

type WorkflowArtifact struct {
	Id     string
	Name   string
	Url    string
	ApiUrl string
}

type ResponseErrorMessage struct {
	Response string       `json:"response"`
	Result   ErrorMessage `json:"result,omitempty"`
}

type ErrorMessage struct {
	Error string `json:"error"`
}

type RunID struct {
	ID string `json:"id,omitempty"`
}

// Broker should be implemented to get access to the broker's data
type Broker interface {
	ParseRequestMessage(ctx context.Context) (*RequestMessage, error)
	ServeResponse(ctx context.Context, responseMessage ResponseMessage) error
	ServeErrorResponse(ctx context.Context, responseErrorMessage ResponseErrorMessage) error
}
