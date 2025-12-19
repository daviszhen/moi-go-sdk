package sdk

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

// PipelineFile represents a single file to be uploaded when creating a GenAI pipeline.
type PipelineFile struct {
	FileName string    // FileName is the name of the file
	Reader   io.Reader // Reader provides the file content
}

// CreateGenAIPipeline creates a new GenAI pipeline with optional file uploads.
//
// If files are provided, they will be uploaded as part of the pipeline creation.
// If no files are provided, only the pipeline configuration is sent.
//
// Example:
//
//	file, _ := os.Open("data.csv")
//	defer file.Close()
//
//	resp, err := client.CreateGenAIPipeline(ctx, &sdk.GenAICreatePipelineRequest{
//		Name: "my-pipeline",
//		// ... other config
//	}, []sdk.PipelineFile{
//		{FileName: "data.csv", Reader: file},
//	})
//	if err != nil {
//		return err
//	}
//	fmt.Printf("Created pipeline ID: %s\n", resp.PipelineID)
func (c *RawClient) CreateGenAIPipeline(ctx context.Context, req *GenAICreatePipelineRequest, files []PipelineFile, opts ...CallOption) (*GenAICreatePipelineResponse, error) {
	if len(files) == 0 {
		if req == nil {
			return nil, ErrNilRequest
		}
		var resp GenAICreatePipelineResponse
		if err := c.postJSON(ctx, "/v1/genai/pipeline", req, &resp, opts...); err != nil {
			return nil, err
		}
		return &resp, nil
	}

	if req == nil {
		return nil, ErrNilRequest
	}

	pr, pw := io.Pipe()
	writer := multipart.NewWriter(pw)
	contentType := writer.FormDataContentType()

	go func() {
		defer pw.Close()
		defer writer.Close()

		payload, err := json.Marshal(req)
		if err != nil {
			pw.CloseWithError(err)
			return
		}
		if err := writer.WriteField("payload", string(payload)); err != nil {
			pw.CloseWithError(err)
			return
		}
		if len(req.FileNames) > 0 {
			for _, name := range req.FileNames {
				if err := writer.WriteField("file_names", name); err != nil {
					pw.CloseWithError(err)
					return
				}
			}
		}

		for i, file := range files {
			if file.Reader == nil {
				pw.CloseWithError(fmt.Errorf("file reader at index %d is nil", i))
				return
			}
			filename := file.FileName
			if strings.TrimSpace(filename) == "" {
				filename = fmt.Sprintf("file_%d", i)
			}
			part, err := writer.CreateFormFile("files", filename)
			if err != nil {
				pw.CloseWithError(err)
				return
			}
			if _, err := io.Copy(part, file.Reader); err != nil {
				pw.CloseWithError(err)
				return
			}
		}
	}()

	callOpts := newCallOptions(opts...)
	resp, err := c.doRaw(ctx, http.MethodPost, "/v1/genai/pipeline", pr, callOpts, func(r *http.Request) {
		r.Header.Set(headerContentType, contentType)
		r.Header.Set(headerAccept, mimeJSON)
	})
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var envelope apiEnvelope
	if err := json.NewDecoder(resp.Body).Decode(&envelope); err != nil {
		return nil, err
	}
	// Check for error code (case-insensitive comparison)
	// Some services return "ok" (lowercase) while others return "OK" (uppercase)
	if envelope.Code != "" && strings.ToUpper(envelope.Code) != "OK" {
		return nil, &APIError{
			Code:       envelope.Code,
			Message:    envelope.Msg,
			RequestID:  envelope.RequestID,
			HTTPStatus: resp.StatusCode,
		}
	}
	var pipelineResp GenAICreatePipelineResponse
	if len(envelope.Data) > 0 && string(envelope.Data) != "null" {
		if err := json.Unmarshal(envelope.Data, &pipelineResp); err != nil {
			return nil, err
		}
	}
	return &pipelineResp, nil
}

// GetGenAIJob retrieves detailed information about a GenAI job.
//
// Returns the job status, results, and other metadata.
//
// Example:
//
//	resp, err := client.GetGenAIJob(ctx, "job-id-123")
//	if err != nil {
//		return err
//	}
//	fmt.Printf("Job Status: %s\n", resp.Status)
func (c *RawClient) GetGenAIJob(ctx context.Context, jobID string, opts ...CallOption) (*GenAIGetJobDetailResponse, error) {
	if strings.TrimSpace(jobID) == "" {
		return nil, fmt.Errorf("jobID cannot be empty")
	}
	var resp GenAIGetJobDetailResponse
	path := fmt.Sprintf("/v1/genai/jobs/%s", url.PathEscape(jobID))
	if err := c.getJSON(ctx, path, &resp, opts...); err != nil {
		return nil, err
	}
	return &resp, nil
}

// DownloadGenAIResult downloads a file result from a GenAI job.
//
// Returns a FileStream that must be closed by the caller. The stream contains
// the file content that can be read directly.
//
// Example:
//
//	stream, err := client.DownloadGenAIResult(ctx, "file-id-123")
//	if err != nil {
//		return err
//	}
//	defer stream.Close()
//
//	data, err := io.ReadAll(stream.Body)
//	if err != nil {
//		return err
//	}
//	fmt.Printf("Downloaded %d bytes\n", len(data))
func (c *RawClient) DownloadGenAIResult(ctx context.Context, fileID string, opts ...CallOption) (*FileStream, error) {
	if strings.TrimSpace(fileID) == "" {
		return nil, fmt.Errorf("fileID cannot be empty")
	}
	callOpts := newCallOptions(opts...)
	path := fmt.Sprintf("/v1/genai/results/file/%s", url.PathEscape(fileID))
	resp, err := c.doRaw(ctx, http.MethodGet, path, nil, callOpts, nil)
	if err != nil {
		return nil, err
	}
	return &FileStream{
		Body:       resp.Body,
		Header:     resp.Header.Clone(),
		StatusCode: resp.StatusCode,
	}, nil
}

// CreateWorkflow creates a new workflow.
//
// This method creates a workflow using workflow metadata, which includes:
// - Workflow name
// - Source volume names/IDs
// - Target volume ID/name
// - Process mode (interval and offset)
// - File types
// - Workflow definition (nodes and connections)
//
// Example:
//
//	resp, err := client.CreateWorkflow(ctx, &sdk.WorkflowMetadata{
//		Name: "my-workflow",
//		SourceVolumeIDs: []string{"vol-123"},
//		TargetVolumeID: "vol-456",
//		FileTypes: []int{1, 2, 3},
//		ProcessMode: &sdk.ProcessMode{
//			Interval: 3600,
//			Offset:   0,
//		},
//		Workflow: &sdk.CatalogWorkflow{
//			Nodes: []sdk.CatalogWorkflowNode{
//				{
//					ID:   "node1",
//					Type: "ParseNode",
//					InitParameters: map[string]map[string]interface{}{},
//				},
//			},
//			Connections: []sdk.CatalogWorkflowConnection{
//				{
//					Sender:   "node1",
//					Receiver: "node2",
//				},
//			},
//		},
//	})
//	if err != nil {
//		return err
//	}
//	fmt.Printf("Created workflow ID: %s\n", resp.ID)
func (c *RawClient) CreateWorkflow(ctx context.Context, req *WorkflowMetadata, opts ...CallOption) (*WorkflowCreateResponse, error) {
	if req == nil {
		return nil, ErrNilRequest
	}
	// Ensure required fields are initialized to avoid serializing them as null
	// The server requires these fields to be present even if empty
	if req.SourceVolumeNames == nil {
		req.SourceVolumeNames = []string{}
	}
	if req.SourceVolumeIDs == nil {
		req.SourceVolumeIDs = []string{}
	}
	if req.ProcessMode == nil {
		req.ProcessMode = &ProcessMode{
			Interval: -1, // Default: trigger on file load
			Offset:   0,
		}
	}
	if req.FileTypes == nil {
		req.FileTypes = []int{}
	}
	// Ensure all workflow nodes have InitParameters initialized to empty map
	// to avoid serializing them as null
	if req.Workflow != nil {
		for i := range req.Workflow.Nodes {
			if req.Workflow.Nodes[i].InitParameters == nil {
				req.Workflow.Nodes[i].InitParameters = map[string]map[string]interface{}{}
			}
		}
	}
	var resp WorkflowCreateResponse
	if err := c.postJSON(ctx, "/v1/genai/workflow", req, &resp, opts...); err != nil {
		return nil, err
	}
	return &resp, nil
}

// ListWorkflowJobs lists workflow jobs with optional filtering and pagination.
//
// This method calls the workflow-be API endpoint /byoa/api/v1/workflow_job to retrieve
// a list of workflow jobs. The request supports filtering by workflow ID, source file ID, and status,
// as well as pagination.
//
// Parameters:
//   - req: the list request with optional filters and pagination parameters
//
// Returns:
//   - *WorkflowJobListResponse: the response containing the list of jobs and total count
//   - error: any error that occurred
//
// Example:
//
//	resp, err := client.ListWorkflowJobs(ctx, &sdk.WorkflowJobListRequest{
//		WorkflowID: "workflow-123",
//		Status:     "running",
//		Page:       1,
//		PageSize:   20,
//	})
//	if err != nil {
//		return err
//	}
//	for _, job := range resp.List {
//		fmt.Printf("Job: %s, Status: %s\n", job.JobID, job.Status)
//	}
func (c *RawClient) ListWorkflowJobs(ctx context.Context, req *WorkflowJobListRequest, opts ...CallOption) (*WorkflowJobListResponse, error) {
	if req == nil {
		return nil, ErrNilRequest
	}

	// Build query parameters
	query := url.Values{}
	if req.WorkflowID != "" {
		query.Set("workflow_id", req.WorkflowID)
	}
	if req.SourceFileID != "" {
		query.Set("source_file_id", req.SourceFileID)
	}
	if req.Status != "" {
		query.Set("status", req.Status)
	}
	if req.Page > 0 {
		query.Set("page", strconv.Itoa(req.Page))
	}
	if req.PageSize > 0 {
		query.Set("page_size", strconv.Itoa(req.PageSize))
	}

	// Use raw response structure to match API format
	type rawResponse struct {
		Jobs  []workflowJobRaw `json:"jobs"`
		Total int              `json:"total"`
	}

	rawResp := rawResponse{
		Jobs:  []workflowJobRaw{},
		Total: 0,
	}
	path := "/byoa/api/v1/workflow_job"
	if len(query) > 0 {
		path += "?" + query.Encode()
	}
	if err := c.getJSON(ctx, path, &rawResp, opts...); err != nil {
		return nil, err
	}

	// Convert raw jobs to WorkflowJob format
	jobs := make([]WorkflowJob, len(rawResp.Jobs))
	for i, rawJob := range rawResp.Jobs {
		jobs[i] = WorkflowJob{
			JobID:        rawJob.ID,
			WorkflowID:   rawJob.WorkflowID,
			SourceFileID: req.SourceFileID,                 // Populate from request filter
			Status:       WorkflowJobStatus(rawJob.Status), // Convert int to WorkflowJobStatus
			StartTime:    rawJob.StartTime,
		}
		// Handle end_time (can be null)
		if rawJob.EndTime != nil {
			jobs[i].EndTime = *rawJob.EndTime
		}
		// Try to extract source_file_id from description if available
		if jobs[i].SourceFileID == "" && rawJob.Description != nil {
			if triggerTaskID, ok := rawJob.Description["triggerTaskID"]; ok {
				// Convert to string if it's a number
				if idStr, ok := triggerTaskID.(string); ok {
					jobs[i].SourceFileID = idStr
				} else if idNum, ok := triggerTaskID.(float64); ok {
					jobs[i].SourceFileID = strconv.FormatFloat(idNum, 'f', -1, 64)
				}
			}
		}
	}

	resp := WorkflowJobListResponse{
		Jobs:  jobs,
		Total: rawResp.Total,
	}
	// Ensure Jobs is never nil
	if resp.Jobs == nil {
		resp.Jobs = []WorkflowJob{}
	}
	return &resp, nil
}
