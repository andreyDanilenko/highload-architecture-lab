package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"worker-pool/internal/adapter/inbound/http/helpers"
	"worker-pool/internal/config"
	"worker-pool/internal/domain"
	"worker-pool/internal/usecase"
	"worker-pool/internal/work"
)

// WorkHandler submits tasks through different dispatcher strategies.
type WorkHandler struct {
	cfg      *config.Config
	naive    usecase.TaskDispatcher
	bounded  usecase.TaskDispatcher
	reliable usecase.TaskDispatcher
	advanced usecase.TaskDispatcher
}

// NewWorkHandler constructs handlers for work endpoints.
func NewWorkHandler(cfg *config.Config, naive, bounded, reliable, advanced usecase.TaskDispatcher) *WorkHandler {
	return &WorkHandler{
		cfg:      cfg,
		naive:    naive,
		bounded:  bounded,
		reliable: reliable,
		advanced: advanced,
	}
}

func taskIDFrom(r *http.Request) string {
	if id := r.URL.Query().Get("id"); id != "" {
		return id
	}
	return "task-" + strconv.FormatInt(time.Now().UnixNano(), 10)
}

func writeDispatch(w http.ResponseWriter, r *http.Request, d usecase.TaskDispatcher) {
	if r.Method != http.MethodPost {
		helpers.WriteError(w, http.StatusMethodNotAllowed, "method_not_allowed", "Method not allowed")
		return
	}
	out := d.Dispatch(r.Context(), domain.TaskID(taskIDFrom(r)))
	switch {
	case out.Err != nil:
		helpers.WriteError(w, http.StatusInternalServerError, "dispatch_error", out.Err.Error())
	case out.QueueFull:
		helpers.WriteError(w, http.StatusServiceUnavailable, "queue_full", "Worker queue is full")
	case out.Accepted:
		helpers.WriteJSON(w, http.StatusAccepted, map[string]string{"status": "accepted"})
	default:
		helpers.WriteError(w, http.StatusInternalServerError, "unknown", "Unexpected dispatch outcome")
	}
}

// Naive handles POST /work/naive.
func (h *WorkHandler) Naive(w http.ResponseWriter, r *http.Request) {
	writeDispatch(w, r, h.naive)
}

// Bounded handles POST /work/bounded.
func (h *WorkHandler) Bounded(w http.ResponseWriter, r *http.Request) {
	writeDispatch(w, r, h.bounded)
}

// Reliable handles POST /work/reliable.
func (h *WorkHandler) Reliable(w http.ResponseWriter, r *http.Request) {
	writeDispatch(w, r, h.reliable)
}

// Advanced handles POST /work/advanced.
func (h *WorkHandler) Advanced(w http.ResponseWriter, r *http.Request) {
	writeDispatch(w, r, h.advanced)
}

func newUploadJobID() (string, error) {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}
	return hex.EncodeToString(b[:]), nil
}

func (h *WorkHandler) videoUpload(w http.ResponseWriter, r *http.Request, d usecase.TaskDispatcher, poolMode bool) {
	if r.Method != http.MethodPost {
		helpers.WriteError(w, http.StatusMethodNotAllowed, "method_not_allowed", "Method not allowed")
		return
	}
	if poolMode {
		if h.cfg.PoolTaskMode != "upload" {
			helpers.WriteError(w, http.StatusBadRequest, "wrong_mode", "Set POOL_TASK_MODE=upload and restart to accept real video uploads on pool endpoints")
			return
		}
	} else {
		if h.cfg.NaiveMode != "upload" {
			helpers.WriteError(w, http.StatusBadRequest, "wrong_mode", "Set NAIVE_MODE=upload and restart to accept real video uploads on /work/naive/upload")
			return
		}
	}

	maxBytes := int64(h.cfg.UploadMaxMB) * 1024 * 1024
	r.Body = http.MaxBytesReader(w, r.Body, maxBytes)
	if err := r.ParseMultipartForm(maxBytes); err != nil {
		helpers.WriteError(w, http.StatusBadRequest, "bad_multipart", err.Error())
		return
	}
	file, hdr, err := r.FormFile("video")
	if err != nil {
		helpers.WriteError(w, http.StatusBadRequest, "missing_video", "multipart form field \"video\" (file) is required")
		return
	}
	defer file.Close()

	jobID, err := newUploadJobID()
	if err != nil {
		helpers.WriteError(w, http.StatusInternalServerError, "job_id", "could not generate job id")
		return
	}
	ext := filepath.Ext(hdr.Filename)
	if ext == "" {
		ext = ".bin"
	}
	if err := os.MkdirAll(h.cfg.VideoUploadDir, 0o755); err != nil {
		helpers.WriteError(w, http.StatusInternalServerError, "mkdir", err.Error())
		return
	}
	dest := filepath.Join(h.cfg.VideoUploadDir, jobID+ext)
	out, err := os.Create(dest)
	if err != nil {
		helpers.WriteError(w, http.StatusInternalServerError, "save", err.Error())
		return
	}
	if _, err := io.Copy(out, file); err != nil {
		_ = out.Close()
		_ = os.Remove(dest)
		helpers.WriteError(w, http.StatusInternalServerError, "save", err.Error())
		return
	}
	if err := out.Close(); err != nil {
		_ = os.Remove(dest)
		helpers.WriteError(w, http.StatusInternalServerError, "save", err.Error())
		return
	}

	dispatchOut := d.Dispatch(r.Context(), domain.TaskID(jobID))
	switch {
	case dispatchOut.Err != nil:
		_ = os.Remove(dest)
		helpers.WriteError(w, http.StatusInternalServerError, "dispatch_error", dispatchOut.Err.Error())
	case dispatchOut.QueueFull:
		_ = os.Remove(dest)
		helpers.WriteError(w, http.StatusServiceUnavailable, "queue_full", "Worker queue is full")
	case dispatchOut.Accepted:
		helpers.WriteJSON(w, http.StatusAccepted, map[string]string{
			"status":     "accepted",
			"job_id":     jobID,
			"result_url": "/work/result/" + jobID,
		})
	default:
		_ = os.Remove(dest)
		helpers.WriteError(w, http.StatusInternalServerError, "unknown", "Unexpected dispatch outcome")
	}
}

// NaiveUpload handles POST /work/naive/upload (multipart field "video").
func (h *WorkHandler) NaiveUpload(w http.ResponseWriter, r *http.Request) {
	h.videoUpload(w, r, h.naive, false)
}

// BoundedUpload handles POST /work/bounded/upload.
func (h *WorkHandler) BoundedUpload(w http.ResponseWriter, r *http.Request) {
	h.videoUpload(w, r, h.bounded, true)
}

// ReliableUpload handles POST /work/reliable/upload.
func (h *WorkHandler) ReliableUpload(w http.ResponseWriter, r *http.Request) {
	h.videoUpload(w, r, h.reliable, true)
}

// AdvancedUpload handles POST /work/advanced/upload.
func (h *WorkHandler) AdvancedUpload(w http.ResponseWriter, r *http.Request) {
	h.videoUpload(w, r, h.advanced, true)
}

// VideoResult serves GET /work/result/{id} — transcoded .mp4 when ready.
func (h *WorkHandler) VideoResult(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		helpers.WriteError(w, http.StatusMethodNotAllowed, "method_not_allowed", "Method not allowed")
		return
	}
	id := r.PathValue("id")
	if id == "" {
		helpers.WriteError(w, http.StatusBadRequest, "missing_id", "job id required")
		return
	}
	safe := work.SafeTaskFilePart(id)
	path := filepath.Join(h.cfg.VideoOutputDir, safe+".mp4")
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			helpers.WriteError(w, http.StatusNotFound, "not_ready", "Result not found yet or transcoding failed; retry later")
			return
		}
		helpers.WriteError(w, http.StatusInternalServerError, "stat", err.Error())
		return
	}
	w.Header().Set("Content-Type", "video/mp4")
	http.ServeFile(w, r, path)
}
