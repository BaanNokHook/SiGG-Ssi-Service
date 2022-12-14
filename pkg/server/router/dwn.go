package router

import (
	"context"
	"fmt"
	"net/http"

	"github.com/TBD54566975/ssi-sdk/credential/manifest"

	"github.com/tbd54566975/ssi-service/pkg/dwn"
	"github.com/tbd54566975/ssi-service/pkg/service/dwn"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	dwnpkg "github.com/tbd54566975/ssi-service/pkg/dwn"
	"github.com/tbd54566975/ssi-service/pkg/server/framework"
	svcframework "github.com/tbd54566975/ssi-service/pkg/service/framework"
)

type DWNRouter struct {
	service *dwn.Service
}

func NewDWNRouter(s svcframework.Service) (*DWNRouter, error) {
	if s == nil {
		return nil, errors.New("service cannot be nil")
	}
	dwnService, ok := s.(*dwn.Service)
	if !ok {
		return nil, fmt.Errorf("could not create dwn router with service type: %s", s.Type())
	}

	return &DWNRouter{
		service: dwnService,
	}, nil

}

type PublishManifestRequest struct {
	ManifestID string `json:"manifestID" validate:"required"`
}

func (req PublishManifestRequest) ToServiceRequest() dwn.PublishManifestRequest {
	return dwn.PublishManifestRequest{
		ManifestID: req.ManifestID,
	}
}

type PublishManifestResponse struct {
	manifest    manifest.CredentialManifest    `json:"manifest" validate:"required"`
	DWNResponse dwnpkg.PublishManifestResponse `json:"swnResponse" validate:"required"`
}

// PublishManifest godoc
// @Summary      Publish Manifest to DWN
// @Description  Publish Manifest to DWN
// @Tags         DWNAPI
// @Accept       json
// @Produce      json
// @Param        request  body      PublishManifestRequest  true  "request body"
// @Success      201      {object}  PublishManifestResponse
// @Failure      400      {string}  string  "Bad request"
// @Failure      500      {string}  string  "Internal server error"
// @Router       /v1/dwn/manifest [put]

func (dwnr DWNRouter) PublishManifest(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	if dwnr.service.Config().DWNEndpoint == "" {
		errMsg := "could not publish manifest to dwn because dwn endpoint is not configured"
		logrus.Error(errMsg)
		return framework.NewRequestError(errors.New(errMsg), http.StatusInternalServerError)
	}

	var request PublishManifestRequest
	if err := framework.Decode(r, &request); err != nil {
		errMsg := "invalid publish manifest message request"
		logrus.WithError(err).Error(errMsg)
		return framework.NewRequestError(errors.Wrap(err, errMsg), http.StatusBadRequest)
	}

	req := request.ToServiceRequest()
	publishManifestResponse, err := dwnr.service.GetManifest(req)

	if err != nil || publishManifestResponse.Manifest.IsEmpty() {
		errMsg := "could not retrieve manifest"
		logrus.WithError(err).Error(errMsg)
		return framework.NewRequestError(errors.Wrap(err, errMsg), http.StatusInternalServerError)
	}

	dwnResp, err := dwnpkg.PublishManifest(dwnr.service.Config().DWNEndpoint, publishManifestResponse.Manifest)

	if err != nil {
		errMsg := "could not publish manifest to DWN"
		logrus.WithError(err).Error(errMsg)
		return framework.NewRequestError(errors.Wrap(err, errMsg), http.StatusInternalServerError)
	}

	resp := PublishManifestResponse{Manifest: publishManifestResponse.Manifest, DWNResponse: *dwnResp}
	return framework.Respond(ctx, w, resp, http.StatusAccepted)
}
