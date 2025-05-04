package api

import (
	"context"
	"os"
	"os/user"
	"slices"

	"github.com/danielgtaylor/huma/v2"
)

type statusRequest struct {
	Fname string `json:"fname"`
}
type statusResponse struct {
	Body map[string]any `json:"body"`
}

func (a *Service) status(ctx context.Context, req *statusRequest) (*statusResponse, error) {
	wd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	usr, err := user.Current()
	if err != nil {
		return nil, err
	}

	envData := os.Environ()
	slices.Sort(envData)

	return &statusResponse{Body: map[string]any{
		"user": usr.Username,
		"wd":   wd,
		"env":  envData,
	}}, nil
}

func (a *Service) setupApiStatus(humaApi huma.API) {
	huma.Register(humaApi, huma.Operation{
		OperationID: "apiV1StatusGet",
		Method:      "GET",
		Path:        "/api/v1/status",
		Description: "retrieves general status of this service",
	}, a.status)
}
