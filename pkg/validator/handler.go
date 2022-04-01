package validator

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/giantswarm/microerror"
	"k8s.io/api/admission/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/apimachinery/pkg/types"

	"github.com/giantswarm/azure-admission-controller/v2/pkg/generic"
)

type Validator interface {
	Validate(ctx context.Context, request *v1beta1.AdmissionRequest) error
	Log(keyVals ...interface{})
}

var (
	scheme       = runtime.NewScheme()
	codecs       = serializer.NewCodecFactory(scheme)
	Deserializer = codecs.UniversalDeserializer()
)

func Handler(validator Validator) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		if request.Header.Get("Content-Type") != "application/json" {
			validator.Log("level", "error", "message", fmt.Sprintf("invalid content-type: %s", request.Header.Get("Content-Type")))
			writer.WriteHeader(http.StatusBadRequest)
			return
		}

		data, err := ioutil.ReadAll(request.Body)
		if err != nil {
			validator.Log("level", "error", "message", "unable to read request")
			writer.WriteHeader(http.StatusInternalServerError)
			return
		}

		review := v1beta1.AdmissionReview{}
		if _, _, err := Deserializer.Decode(data, nil, &review); err != nil {
			validator.Log("level", "error", "message", "unable to parse admission review request")
			writer.WriteHeader(http.StatusBadRequest)
			return
		}

		err = validator.Validate(request.Context(), review.Request)
		if err != nil {
			writeResponse(validator, writer, errorResponse(review.Request.UID, microerror.Mask(err)))
			return
		}

		writeResponse(validator, writer, &v1beta1.AdmissionResponse{
			Allowed: true,
			UID:     review.Request.UID,
		})
	}
}

func writeResponse(logger generic.Logger, writer http.ResponseWriter, response *v1beta1.AdmissionResponse) {
	resp, err := json.Marshal(v1beta1.AdmissionReview{
		TypeMeta: metav1.TypeMeta{
			Kind:       "AdmissionReview",
			APIVersion: "admission.k8s.io/v1",
		},
		Response: response,
	})
	if err != nil {
		logger.Log("level", "error", "message", "unable to serialize response", "stack", microerror.JSON(err))
		writer.WriteHeader(http.StatusInternalServerError)
	}

	if _, err := writer.Write(resp); err != nil {
		logger.Log("level", "error", "message", "unable to write response", "stack", microerror.JSON(err))
	}

	logger.Log("level", "info", "message", fmt.Sprintf("Validated request responded with result: %t", response.Allowed))
}

func errorResponse(uid types.UID, err error) *v1beta1.AdmissionResponse {
	return &v1beta1.AdmissionResponse{
		Allowed: false,
		UID:     uid,
		Result: &metav1.Status{
			Reason:  metav1.StatusReasonBadRequest,
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		},
	}
}
