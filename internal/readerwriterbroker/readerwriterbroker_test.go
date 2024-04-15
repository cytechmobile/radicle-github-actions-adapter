package readerwriterbroker

import (
	"bytes"
	"context"
	"io"
	"log/slog"
	"os"
	"radicle-github-actions-adapter/app/broker"
	"reflect"
	"strings"
	"testing"
)

func TestNewReaderWriterBroker(t *testing.T) {
	type args struct {
		reader io.Reader
	}
	tests := []struct {
		name       string
		args       args
		wantWriter string
		want       *ReaderWriterBroker
	}{
		{
			name: "Test NewReaderWriterBroker",
			args: args{
				reader: strings.NewReader("test data"),
			},
			wantWriter: "",
			want: &ReaderWriterBroker{
				brokerReader: strings.NewReader("test data"),
				brokerWriter: &bytes.Buffer{},
				logger:       slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{})),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			writer := &bytes.Buffer{}
			got := NewReaderWriterBroker(tt.args.reader, writer, slog.New(slog.NewJSONHandler(os.Stderr,
				&slog.HandlerOptions{})))
			if gotWriter := writer.String(); gotWriter != tt.wantWriter {
				t.Errorf("NewReaderWriterBroker() gotWriter = %v, want %v", gotWriter, tt.wantWriter)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewReaderWriterBroker() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestReaderWriterBroker_ParseRequestMessage(t *testing.T) {
	type fields struct {
		BrokerReader io.Reader
		BrokerWriter io.Writer
	}
	type args struct {
		ctx context.Context
	}
	var tests = []struct {
		name    string
		fields  fields
		args    args
		want    *broker.RequestMessage
		wantErr bool
	}{
		{
			name: "Test valid push request to ParseRequestMessage",
			fields: fields{
				BrokerReader: strings.NewReader(`{"request": "trigger","event_type": "push","pusher": {"id": "did:key:z6MkltRpzcq2ybm13yQpyre58JUeMvZY6toxoZVpLZ8YabRa","alias": "node_alias"},"before": "<BEFORE_COMMIT>","after": "<AFTER_COMMIT>","commits": ["<SOME_OTHER_COMMIT_BEING_PUSHED>", "<AFTER_COMMIT>" ], "repository": { "id": "<RID>", "name": "heartwood", "description": "Radicle is a sovereign peer-to-peer network for code collaboration, built on top of Git.","private": false,"default_branch": "main","delegates": ["did:key:z6MkltRpzcq2ybm13yQpyre58JUeMvZY6toxoZVpLZ8YabRa", "did:key:z6MkltRpzcq2ybm13yQpyre58JUeMvZY6toxoZVpLZ8YabRb"]}}`),
				BrokerWriter: &bytes.Buffer{},
			},
			args: args{ctx: context.TODO()},
			want: &broker.RequestMessage{
				Repo:   "<RID>",
				Commit: "<AFTER_COMMIT>",
				PushEvent: &broker.RequestPushEventMessage{
					Request:   "trigger",
					EventType: "push",
					Pusher: broker.Pusher{
						ID:    "did:key:z6MkltRpzcq2ybm13yQpyre58JUeMvZY6toxoZVpLZ8YabRa",
						Alias: "node_alias",
					},
					Before:  "<BEFORE_COMMIT>",
					After:   "<AFTER_COMMIT>",
					Commits: []string{"<SOME_OTHER_COMMIT_BEING_PUSHED>", "<AFTER_COMMIT>"},
					Repository: broker.Repository{
						ID:            "<RID>",
						Name:          "heartwood",
						Description:   "Radicle is a sovereign peer-to-peer network for code collaboration, built on top of Git.",
						Private:       false,
						DefaultBranch: "main",
						Delegates:     []string{"did:key:z6MkltRpzcq2ybm13yQpyre58JUeMvZY6toxoZVpLZ8YabRa", "did:key:z6MkltRpzcq2ybm13yQpyre58JUeMvZY6toxoZVpLZ8YabRb"},
					},
				},
				PatchEvent: nil,
			},
			wantErr: false,
		},
		{
			name: "Test invalid request to ParseRequestMessage - invalid request",
			fields: fields{
				BrokerReader: strings.NewReader(`{"request": 1`),
				BrokerWriter: &bytes.Buffer{},
			},
			args:    args{ctx: context.TODO()},
			want:    nil,
			wantErr: true,
		},
		{
			name: "Test invalid push request to ParseRequestMessage - invalid request",
			fields: fields{
				BrokerReader: strings.NewReader(`{"request": "trigger","event_type": "push","pusher": {"id": 123,"alias": "node_alias"},"before": "<BEFORE_COMMIT>","after": "<AFTER_COMMIT>","commits": ["<SOME_OTHER_COMMIT_BEING_PUSHED>", "<AFTER_COMMIT>" ], "repository": { "id": "<RID>", "name": "heartwood", "description": "Radicle is a sovereign peer-to-peer network for code collaboration, built on top of Git.","private": false,"default_branch": "main","delegates": ["did:key:z6MkltRpzcq2ybm13yQpyre58JUeMvZY6toxoZVpLZ8YabRa", "did:key:z6MkltRpzcq2ybm13yQpyre58JUeMvZY6toxoZVpLZ8YabRb"]}}`),
				BrokerWriter: &bytes.Buffer{},
			},
			args:    args{ctx: context.TODO()},
			want:    nil,
			wantErr: true,
		},
		{
			name: "Test invalid request to ParseRequestMessage - invalid event_type",
			fields: fields{
				BrokerReader: strings.NewReader(`{"request": "trigger","event_type": "unknown","pusher": {"id": "did:key:z6MkltRpzcq2ybm13yQpyre58JUeMvZY6toxoZVpLZ8YabRa","alias": "node_alias"},"before": "<BEFORE_COMMIT>","after": "<AFTER_COMMIT>","commits": ["<SOME_OTHER_COMMIT_BEING_PUSHED>", "<AFTER_COMMIT>" ], "repository": { "id": "<RID>", "name": "heartwood", "description": "Radicle is a sovereign peer-to-peer network for code collaboration, built on top of Git.","private": false,"default_branch": "main","delegates": ["did:key:z6MkltRpzcq2ybm13yQpyre58JUeMvZY6toxoZVpLZ8YabRa", "did:key:z6MkltRpzcq2ybm13yQpyre58JUeMvZY6toxoZVpLZ8YabRb"]}}`),
				BrokerWriter: &bytes.Buffer{},
			},
			args:    args{ctx: context.TODO()},
			want:    nil,
			wantErr: true,
		},

		{
			name: "Test invalid request to ParseRequestMessage - invalid request",
			fields: fields{
				BrokerReader: strings.NewReader(`{"request": "some request","event_type": "push","pusher": {"id": "did:key:z6MkltRpzcq2ybm13yQpyre58JUeMvZY6toxoZVpLZ8YabRa","alias": "node_alias"},"before": "<BEFORE_COMMIT>","after": "<AFTER_COMMIT>","commits": ["<SOME_OTHER_COMMIT_BEING_PUSHED>", "<AFTER_COMMIT>" ], "repository": { "id": "<RID>", "name": "heartwood", "description": "Radicle is a sovereign peer-to-peer network for code collaboration, built on top of Git.","private": false,"default_branch": "main","delegates": ["did:key:z6MkltRpzcq2ybm13yQpyre58JUeMvZY6toxoZVpLZ8YabRa", "did:key:z6MkltRpzcq2ybm13yQpyre58JUeMvZY6toxoZVpLZ8YabRb"]}}`),
				BrokerWriter: &bytes.Buffer{},
			},
			args:    args{ctx: context.TODO()},
			want:    nil,
			wantErr: true,
		},

		{
			name: "Test valid patch request to ParseRequestMessage",
			fields: fields{
				BrokerReader: strings.NewReader(`{"request":"trigger","event_type":"patch","action":"created","patch":{"id":"<PATCH_ID>","author":{"id":"did:key:z6MkltRpzcq2ybm13yQpyre58JUeMvZY6toxoZVpLZ8YabRa","alias":"node_alias"},"title":"Add description in README","state":{"status":"Open","conflicts":[{"revision_id":"rev1","oid":"id1"}]},"before":"<BEFORE_COMMIT>","after":"<AFTER_COMMIT>","commits":["<SOME_OTHER_COMMIT_BEING_PUSHED>","<AFTER_COMMIT>"],"target":"delegates","labels":["small","goodFirstIssue","enhancement","bug"],"assignees":["did:key:z6MkltRpzcq2ybm13yQpyre58JUeMvZY6toxoZVpLZ8YabRa"],"revisions":[{"id":"41aafe22200464bf905b143d4233f7f1fa4a9123","author":{"id":"did:key:z6MkltRpzcq2ybm13yQpyre58JUeMvZY6toxoZVpLZ8YabRa","alias":"my_alias"},"description":"The revision description","base":"193ed2f675ac6b0d1ab79ed65057c8a56a4fab23","oid":"f0f5d38ffa8d54a7cc737fc4e75ab1e2e178eaa1","timestamp":1699437445}]},"repository":{"id":"<RID>","name":"heartwood","description":"Radicle is a sovereign peer-to-peer network for code collaboration, built on top of Git.","private":false,"default_branch":"main","delegates":["did:key:z6MkltRpzcq2ybm13yQpyre58JUeMvZY6toxoZVpLZ8YabRa","did:key:z6MkltRpzcq2ybm13yQpyre58JUeMvZY6toxoZVpLZ8YabRb"]}}`),
				BrokerWriter: &bytes.Buffer{},
			},
			args: args{ctx: context.TODO()},
			want: &broker.RequestMessage{
				Repo:   "<RID>",
				Commit: "<AFTER_COMMIT>",
				PatchEvent: &broker.RequestPatchEventMessage{
					Request:   "trigger",
					EventType: "patch",
					Action:    "created",
					Patch: broker.PatchDetails{
						ID: "<PATCH_ID>",
						Author: broker.Author{
							ID:    "did:key:z6MkltRpzcq2ybm13yQpyre58JUeMvZY6toxoZVpLZ8YabRa",
							Alias: "node_alias",
						},
						Title: "Add description in README",
						State: broker.PatchState{
							Status: "Open",
							Conflicts: []broker.PatchConflict{
								broker.PatchConflict{
									RevisionID: "rev1",
									Oid:        "id1",
								},
							},
						},
						Before:    "<BEFORE_COMMIT>",
						After:     "<AFTER_COMMIT>",
						Commits:   []string{"<SOME_OTHER_COMMIT_BEING_PUSHED>", "<AFTER_COMMIT>"},
						Target:    "delegates",
						Labels:    []string{"small", "goodFirstIssue", "enhancement", "bug"},
						Assignees: []string{"did:key:z6MkltRpzcq2ybm13yQpyre58JUeMvZY6toxoZVpLZ8YabRa"},
						Revisions: []broker.PatchRevision{
							{
								ID: "41aafe22200464bf905b143d4233f7f1fa4a9123",
								Author: broker.Author{
									ID:    "did:key:z6MkltRpzcq2ybm13yQpyre58JUeMvZY6toxoZVpLZ8YabRa",
									Alias: "my_alias",
								},
								Description: "The revision description",
								Base:        "193ed2f675ac6b0d1ab79ed65057c8a56a4fab23",
								Oid:         "f0f5d38ffa8d54a7cc737fc4e75ab1e2e178eaa1",
								Timestamp:   1699437445,
							},
						},
					},
					Repository: broker.Repository{
						ID:            "<RID>",
						Name:          "heartwood",
						Description:   "Radicle is a sovereign peer-to-peer network for code collaboration, built on top of Git.",
						Private:       false,
						DefaultBranch: "main",
						Delegates:     []string{"did:key:z6MkltRpzcq2ybm13yQpyre58JUeMvZY6toxoZVpLZ8YabRa", "did:key:z6MkltRpzcq2ybm13yQpyre58JUeMvZY6toxoZVpLZ8YabRb"},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "Test invalid patch request to ParseRequestMessage - invalid request",
			fields: fields{
				BrokerReader: strings.NewReader(`{"request":"trigger","event_type":"patch","action":"created","patch":{"id":123,"author":{"id":"did:key:z6MkltRpzcq2ybm13yQpyre58JUeMvZY6toxoZVpLZ8YabRa","alias":"node_alias"},"title":"Add description in README","state":{"status":"Open","conflicts":[{"revision_id":"rev1","oid":"id1"}]},"before":"<BEFORE_COMMIT>","after":"<AFTER_COMMIT>","commits":["<SOME_OTHER_COMMIT_BEING_PUSHED>","<AFTER_COMMIT>"],"target":"delegates","labels":["small","goodFirstIssue","enhancement","bug"],"assignees":["did:key:z6MkltRpzcq2ybm13yQpyre58JUeMvZY6toxoZVpLZ8YabRa"],"revisions":[{"id":"41aafe22200464bf905b143d4233f7f1fa4a9123","author":{"id":"did:key:z6MkltRpzcq2ybm13yQpyre58JUeMvZY6toxoZVpLZ8YabRa","alias":"my_alias"},"description":"The revision description","base":"193ed2f675ac6b0d1ab79ed65057c8a56a4fab23","oid":"f0f5d38ffa8d54a7cc737fc4e75ab1e2e178eaa1","timestamp":1699437445}]},"repository":{"id":"<RID>","name":"heartwood","description":"Radicle is a sovereign peer-to-peer network for code collaboration, built on top of Git.","private":false,"default_branch":"main","delegates":["did:key:z6MkltRpzcq2ybm13yQpyre58JUeMvZY6toxoZVpLZ8YabRa","did:key:z6MkltRpzcq2ybm13yQpyre58JUeMvZY6toxoZVpLZ8YabRb"]}}`),
				BrokerWriter: &bytes.Buffer{},
			},
			args:    args{ctx: context.TODO()},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sb := &ReaderWriterBroker{
				brokerReader: tt.fields.BrokerReader,
				brokerWriter: tt.fields.BrokerWriter,
				logger:       slog.New(slog.NewJSONHandler(os.Stderr, nil)),
			}
			got, err := sb.ParseRequestMessage(tt.args.ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseRequest() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseRequest() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestReaderWriterBroker_ServeResponse(t *testing.T) {
	type fields struct {
		BrokerReader io.Reader
		BrokerWriter io.Writer
	}
	type args struct {
		ctx             context.Context
		responseMessage broker.ResponseMessage
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []byte
		wantErr bool
	}{
		{
			name: "Test valid ServeResponse",
			fields: fields{
				BrokerReader: strings.NewReader("test data"),
				BrokerWriter: &bytes.Buffer{},
			},
			args: args{
				ctx: context.TODO(),
				responseMessage: broker.ResponseMessage{
					Response: "some_response",
					RunID:    &broker.RunID{ID: "550e8400-e29b-41d4-a716-446655440000"},
					Result:   "Completed",
				},
			},
			want: []byte(`{"response":"some_response","run_id":"550e8400-e29b-41d4-a716-446655440000",
			"result":"Completed"}`),
			wantErr: false,
		},
		{
			name: "Test valid ServeResponse omits empty run_id",
			fields: fields{
				BrokerReader: strings.NewReader("test data"),
				BrokerWriter: &bytes.Buffer{},
			},
			args: args{
				ctx: context.TODO(),
				responseMessage: broker.ResponseMessage{
					Response: "some_response",
					RunID:    nil,
					Result:   "Completed",
				},
			},
			want:    []byte(`{"response":"some_response","result":"Completed"}`),
			wantErr: false,
		},
		{
			name: "Test valid ServeResponse omits empty result",
			fields: fields{
				BrokerReader: strings.NewReader("test data"),
				BrokerWriter: &bytes.Buffer{},
			},
			args: args{
				ctx: context.TODO(),
				responseMessage: broker.ResponseMessage{
					Response: "some_response",
					RunID:    &broker.RunID{ID: "550e8400-e29b-41d4-a716-446655440000"},
					Result:   "",
				},
			},
			want:    []byte(`{"response":"some_response","run_id":{"id":"550e8400-e29b-41d4-a716-446655440000"}}`),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sb := &ReaderWriterBroker{
				brokerReader: tt.fields.BrokerReader,
				brokerWriter: tt.fields.BrokerWriter,
				logger:       slog.New(slog.NewJSONHandler(os.Stderr, nil)),
			}
			if err := sb.ServeResponse(tt.args.ctx, tt.args.responseMessage); (err != nil) != tt.wantErr {
				t.Errorf("ServeResponse() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestReaderWriterBroker_ServeErrorResponse(t *testing.T) {
	type fields struct {
		BrokerReader io.Reader
		BrokerWriter io.Writer
	}
	type args struct {
		ctx                  context.Context
		responseErrorMessage broker.ResponseMessage
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []byte
		wantErr bool
	}{
		{
			name: "Test valid ServeErrorResponse",
			fields: fields{
				BrokerReader: strings.NewReader("test data"),
				BrokerWriter: &bytes.Buffer{},
			},
			args: args{
				ctx: context.TODO(),
				responseErrorMessage: broker.ResponseMessage{
					Response: "some_response",
					Result:   "failure",
				},
			},
			want:    []byte(`{"response":"some_response","result":{"error":"some error"}}`),
			wantErr: false,
		},
		{
			name: "Test valid ServeErrorResponse with empty error",
			fields: fields{
				BrokerReader: strings.NewReader("test data"),
				BrokerWriter: &bytes.Buffer{},
			},
			args: args{
				ctx: context.TODO(),
				responseErrorMessage: broker.ResponseMessage{
					Response: "some_response",
					Result:   "failure",
				},
			},
			want:    []byte(`{"response":"some_response","result":{"error":""}}`),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sb := &ReaderWriterBroker{
				brokerReader: tt.fields.BrokerReader,
				brokerWriter: tt.fields.BrokerWriter,
				logger:       slog.New(slog.NewJSONHandler(os.Stderr, nil)),
			}
			if err := sb.ServeResponse(tt.args.ctx, tt.args.responseErrorMessage); (err != nil) != tt.wantErr {
				t.Errorf("ServeResponse() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
