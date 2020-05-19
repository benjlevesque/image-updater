package gcr

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/bigkevmcd/image-hooks/pkg/hooks"
)

// Parse takes an http.Request and parses it into a Quay.io Push hook if
// possible.
func Parse(req *http.Request) (hooks.PushEvent, error) {
	// TODO: LimitReader
	data, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return nil, err
	}
	h := &ContainerRegistryNotification{}
	err = json.Unmarshal(data, h)
	if err != nil {
		return nil, err
	}
	if h.Tag == "" {
		return nil, nil
	}
	return h, nil
}

// ContainerRegistryNotification is a struct for GCR's notifications.
type ContainerRegistryNotification struct {
	Action string `json:"action"`
	Digest string `json:"digest"`
	Tag    string `json:"tag"`
}

// PushedImageURL is an implementation of the hooks.PushEvent interface.
func (p ContainerRegistryNotification) PushedImageURL() string {
	return p.Tag
}

// EventRepository is an implementation of the hooks.PushEvent interface.
func (p ContainerRegistryNotification) EventRepository() string {
	parts := strings.Split(p.Digest, "/")
	project, imageDigest := parts[1], parts[2]
	image := strings.Split(imageDigest, "@")[0]
	return fmt.Sprintf("%s/%s", project, image)
}
