package podman

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/oars-sigs/oars-cloud/core"
)

func (c *client) ImagePull(ctx context.Context, svc *core.ContainerService) error {
	if svc.ImagePullPolicy == "" {
		svc.ImagePullPolicy = core.ImagePullIfNotPresent
	}
	isExist, err := c.ImageExists(ctx, svc.Image)
	if err != nil {
		return err
	}
	if svc.ImagePullPolicy == core.ImagePullAlways || !isExist {
		res, err := c.Post(ctx, fmt.Sprintf("/libpod/images/pull?reference=%s", svc.Image), nil)
		if err != nil {
			return err
		}

		defer res.Body.Close()
		if res.StatusCode != http.StatusOK {
			body, _ := ioutil.ReadAll(res.Body)
			return fmt.Errorf("unknown error, status code: %d: %s", res.StatusCode, body)
		}

		dec := json.NewDecoder(res.Body)
		var report ImagePullReport
		for {
			if err = dec.Decode(&report); err == io.EOF {
				break
			} else if err != nil {
				return fmt.Errorf("Error reading response: %w", err)
			}

			if report.Error != "" {
				return errors.New(report.Error)
			}
		}
	}

	return nil
}

func (c *client) ImageExists(ctx context.Context, nameWithTag string) (bool, error) {

	res, err := c.Get(ctx, fmt.Sprintf("/libpod/images/%s/exists", nameWithTag))
	if err != nil {
		return false, err
	}

	defer res.Body.Close()

	if res.StatusCode == http.StatusNotFound {
		return false, nil
	}
	if res.StatusCode == http.StatusNoContent {
		return true, nil
	}
	return false, fmt.Errorf("unknown error, status code: %d", res.StatusCode)
}

type ImagePullReport struct {
	// Stream used to provide output from c/image
	Stream string `json:"stream,omitempty"`
	// Error contains text of errors from c/image
	Error string `json:"error,omitempty"`
	// Images contains the ID's of the images pulled
	Images []string `json:"images,omitempty"`
	// ID contains image id (retained for backwards compatibility)
	ID string `json:"id,omitempty"`
}
