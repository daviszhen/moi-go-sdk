package sdk

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"
)

// PipelineFile represents a single file to be uploaded when creating a GenAI pipeline.
type PipelineFile struct {
	FileName string   // FileName is the name of the file
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
	if envelope.Code != "" && envelope.Code != "OK" {
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
