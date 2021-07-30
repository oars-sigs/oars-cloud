package podman

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"time"

	"github.com/oars-sigs/oars-cloud/core"
)

// ExecConfig contains the configuration of an exec session
type ExecConfig struct {
	// Command the the command that will be invoked in the exec session.
	// Must not be empty.
	Command []string `json:"Cmd"`
	// DetachKeys are keys that will be used to detach from the exec
	// session.
	DetachKeys string `json:"DetachKeys,omitempty"`
	// Environment is a set of environment variables that will be set for
	// the first process started by the exec session.
	Environment map[string]string `json:"Env,omitempty"`
	// The user, and optionally, group to run the exec process inside the container.
	// Format is one of: user, user:group, uid, or uid:gid."
	User string `json:"User,omitempty"`
	// WorkDir is the working directory for the first process that will be
	// launched by the exec session.
	// If set to "" the exec session will be started in / within the
	// container.
	WorkDir string `json:"WorkingDir,omitempty"`
	// Tty is whether the exec session will allocate a pseudoterminal.
	Tty bool `json:"Tty,omitempty"`
	// AttachStdin is whether the STDIN stream will be forwarded to the exec
	// session's first process when attaching. Only available if Terminal is
	// false.
	AttachStdin bool `json:"AttachStdin,omitempty"`
	// AttachStdout is whether the STDOUT stream will be forwarded to the
	// exec session's first process when attaching. Only available if
	// Terminal is false.
	AttachStdout bool `json:"AttachStdout,omitempty"`
	// AttachStderr is whether the STDERR stream will be forwarded to the
	// exec session's first process when attaching. Only available if
	// Terminal is false.
	AttachStderr bool `json:"AttachStderr,omitempty"`
	// Privileged is whether the exec session will be privileged - that is,
	// will be granted additional capabilities.
	Privileged bool `json:"Privileged,omitempty"`
}

// ExecSessionResponse contains the ID of a newly created exec session
type ExecSessionResponse struct {
	ID string
}

// ExecCreate creates an exec session to run a command inside a running container
func (c *client) ExecCreate(ctx context.Context, name string, config ExecConfig) (string, error) {
	jsonString, err := json.Marshal(config)
	if err != nil {
		return "", err
	}

	res, err := c.Post(ctx, fmt.Sprintf("/v3.0.0/libpod/containers/%s/exec", name), bytes.NewBuffer(jsonString))
	if err != nil {
		return "", err
	}

	defer res.Body.Close()

	if res.StatusCode != http.StatusCreated {
		body, _ := ioutil.ReadAll(res.Body)
		return "", fmt.Errorf("unknown error, status code: %d: %s", res.StatusCode, body)
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return "", err
	}
	execResponse := &ExecSessionResponse{}
	err = json.Unmarshal(body, execResponse)
	if err != nil {
		return "", err
	}
	return execResponse.ID, err
}

// ExecStartRequest prepares to stream a exec session
type ExecStartRequest struct {

	// streams
	Stdin  io.Reader
	Stdout io.Writer
	Stderr io.Writer

	// Tty indicates whether pseudo-terminal is to be allocated
	Tty bool

	// AttachOutput is whether to attach to STDOUT
	// If false, stdout will not be attached
	AttachOutput bool
	// AttachError is whether to attach to STDERR
	// If false, stdout will not be attached
	AttachError bool
	// AttachInput is whether to attach to STDIN
	// If false, stdout will not be attached
	AttachInput bool
}

// DemuxHeader reads header for stream from server multiplexed stdin/stdout/stderr/2nd error channel
func DemuxHeader(r io.Reader, buffer []byte) (fd, sz int, err error) {
	n, err := io.ReadFull(r, buffer[0:8])
	if err != nil {
		return
	}
	if n < 8 {
		err = io.ErrUnexpectedEOF
		return
	}

	fd = int(buffer[0])
	if fd < 0 || fd > 3 {
		err = fmt.Errorf(`channel "%d" found, 0-3 supported`, fd)
		return
	}

	sz = int(binary.BigEndian.Uint32(buffer[4:8]))
	return
}

// DemuxFrame reads contents for frame from server multiplexed stdin/stdout/stderr/2nd error channel
func DemuxFrame(r io.Reader, buffer []byte, length int) (frame []byte, err error) {
	if len(buffer) < length {
		buffer = append(buffer, make([]byte, length-len(buffer)+1)...)
	}

	n, err := io.ReadFull(r, buffer[0:length])
	if err != nil {
		return nil, nil
	}
	if n < length {
		err = io.ErrUnexpectedEOF
		return
	}

	return buffer[0:length], nil
}

// HijackedResponse holds connection information for a hijacked request.
type HijackedResponse struct {
	Conn net.Conn
	Body io.ReadCloser
}

// Close closes the hijacked connection and reader.
func (h *HijackedResponse) Close() error {
	h.Body.Close()
	return h.Conn.Close()
}

func (h *HijackedResponse) Write(p []byte) (n int, err error) {
	return h.Conn.Write(p)
}

func (h *HijackedResponse) Read(p []byte) (n int, err error) {
	return h.Conn.Read(p)
}

// ExecStartAndAttach starts and attaches to a given exec session.
func (c *client) ExecStart(ctx context.Context, sessionID string, options ExecStartRequest) (core.ExecResp, error) {
	client := new(http.Client)
	*client = *c.httpClient
	client.Timeout = 0

	var socket net.Conn
	socketSet := false
	dialContext := client.Transport.(*http.Transport).DialContext
	t := &http.Transport{
		DialContext: func(ctx context.Context, network, address string) (net.Conn, error) {
			c, err := dialContext(ctx, network, address)
			if err != nil {
				return nil, err
			}
			if !socketSet {
				socket = c
				socketSet = true
			}
			return c, err
		},
		IdleConnTimeout: time.Duration(0),
	}
	client.Transport = t

	// Detach is always false.
	// podman reference doc states that "true" is not supported
	execStartReq := struct {
		Detach bool `json:"Detach"`
		Tty    bool
		H      int `json:"h"`
		W      int `json:"w"`
	}{
		Detach: false,
		Tty:    true,
		H:      763,
		W:      1024,
	}
	jsonBytes, err := json.Marshal(execStartReq)
	if err != nil {
		return socket, err
	}
	req, err := http.NewRequestWithContext(ctx, "POST", fmt.Sprintf("%s/v3.0.0/libpod/exec/%s/start", c.baseUrl, sessionID), bytes.NewBuffer(jsonBytes))
	if err != nil {
		return socket, err
	}
	req.Header.Set("Content-Type", "application/json")
	res, err := client.Do(req)
	if err != nil {
		return socket, err
	}
	//defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		res.Body.Close()
		return socket, err
	}

	return &HijackedResponse{Conn: socket, Body: res.Body}, nil
}
