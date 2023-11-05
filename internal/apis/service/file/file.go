package file

import (
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"infra-3.xyz/hyperdot-node/internal/apis/base"
	"infra-3.xyz/hyperdot-node/internal/clients"
)

const ServiceName = "file"

type Service struct {
	s3Cliet *clients.SimpleS3Cliet
}

// New user service
func New(s3Client *clients.SimpleS3Cliet) *Service {
	svc := &Service{
		s3Cliet: s3Client,
	}
	return svc
}

// @Summary Get file
// @Description Get file
// @Tags File apis
// @Accept application/json
// @Param file query string true "file name"
// @Success 200 {string} string "ok"
// @Router /file [get]
func (s *Service) GetFileHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		file := ctx.Query("file")
		if len(file) == 0 {
			base.ResponseErr(ctx, http.StatusBadRequest, "file is required")
			return
		}

		obj, err := s.s3Cliet.Get(ctx, "hyperdot", file)
		if err != nil {
			base.ResponseErr(ctx, http.StatusInternalServerError, err.Error())
			return
		}
		defer obj.Close()

		// get icon content type
		var contentType string
		seq := strings.Split(file, ".")
		if len(seq) == 0 {
			contentType = "application/octet-stream"
		} else {
			switch seq[len(seq)-1] {
			case "svg":
				contentType = "image/svg+xml"
			case "png":
				contentType = "image/png"
			case "jpg", "jpeg":
				contentType = "image/jpeg"
			default:
				contentType = "application/octet-stream"
			}
		}

		data, err := io.ReadAll(obj)
		if err != nil {
			base.ResponseErr(ctx, http.StatusInternalServerError, err.Error())
			return
		}

		ctx.Header("Content-Length", strconv.Itoa(len(data)))
		ctx.Header("Content-Type", contentType)
		if _, err := ctx.Writer.Write(data); err != nil {
			base.ResponseErr(ctx, http.StatusInternalServerError, err.Error())
			return
		}
	}
}

func (s *Service) Name() string {
	return ServiceName
}

func (s *Service) RouteTables() []base.RouteTable {
	group := "file"
	return []base.RouteTable{
		{
			Method:     "GET",
			Path:       group,
			Handler:    s.GetFileHandler(),
			AllowGuest: true,
			Regexp:     "",
		},
	}
}
