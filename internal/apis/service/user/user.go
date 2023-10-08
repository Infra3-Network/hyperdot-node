package user

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"infra-3.xyz/hyperdot-node/internal/clients"
	"infra-3.xyz/hyperdot-node/internal/dataengine"

	"github.com/gin-gonic/gin"
	_ "gorm.io/driver/postgres"
	"gorm.io/gorm"

	"infra-3.xyz/hyperdot-node/internal/apis/base"
	"infra-3.xyz/hyperdot-node/internal/datamodel"
)

const (
	ServiceName      = "user"
	PasswordProvider = "password"
)

type Service struct {
	db      *gorm.DB
	s3Cliet *clients.SimpleS3Cliet
	// auth          *auth.Auth
	authProviders map[string]bool
	engines       map[string]dataengine.QueryEngine
}

// New user service
func New(db *gorm.DB, engines map[string]dataengine.QueryEngine, s3Client *clients.SimpleS3Cliet) *Service {
	svc := &Service{
		db:      db,
		engines: engines,
		s3Cliet: s3Client,
		authProviders: map[string]bool{
			PasswordProvider: true,
		},
	}
	return svc
}

func (s *Service) getUserInternalHandler(id uint, ctx *gin.Context) {
	sql := `SELECT u.id, u.uid, u.username, u.email, u.encrypted_password, u.bio, u.icon_url, u.twitter, u.github, u.telgram, u.discord, u.location, 
						u.confirmed_at, u.created_at, u.updated_at, us.stars, us.queries, us.dashboards FROM hyperdot_user as u LEFT JOIN 
						hyperdot_user_statistics as us ON u.id = us.user_id WHERE u.id = ?`

	rows, err := s.db.Raw(sql, id).Rows()
	if err != nil {
		base.ResponseErr(ctx, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()

	var data GetUserResponseData
	if !rows.Next() {
		base.ResponseErr(ctx, http.StatusOK, "user not found")
		return
	}

	if err := s.db.ScanRows(rows, &data); err != nil {
		base.ResponseErr(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	ctx.JSON(http.StatusOK, GetUserResponse{
		Data: data,
		BaseResponse: base.BaseResponse{
			Success: true,
		},
	})
}

func (s *Service) getCurrentUserHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		userId, err := base.GetCurrentUserId(ctx)
		if err != nil {
			base.ResponseErr(ctx, http.StatusBadRequest, err.Error())
			return
		}

		s.getUserInternalHandler(userId, ctx)
	}
}

func (s *Service) getUserHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		id, err := base.GetUintParam(ctx, "id")
		if err != nil {
			base.ResponseErr(ctx, http.StatusBadRequest, err.Error())
			return
		}

		s.getUserInternalHandler(id, ctx)
	}
}

func (s *Service) updateUserHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		userId, err := base.GetCurrentUserId(ctx)
		if err != nil {
			base.ResponseErr(ctx, http.StatusBadRequest, err.Error())
			return
		}

		var request datamodel.UserModel
		if err := ctx.ShouldBindJSON(&request); err != nil {
			base.ResponseErr(ctx, http.StatusBadRequest, err.Error())
			return
		}

		var user datamodel.UserModel
		result := s.db.Where("id = ?", userId).First(&user)
		if result.Error != nil {
			if errors.Is(result.Error, gorm.ErrRecordNotFound) {
				base.ResponseErr(ctx, http.StatusOK, "user not found")
				return
			}

			base.ResponseErr(ctx, http.StatusInternalServerError, result.Error.Error())
			return
		}

		if len(request.Bio) > 0 {
			user.Bio = request.Bio
		}

		if len(request.Twitter) > 0 {
			user.Twitter = request.Twitter
		}

		if len(request.Github) > 0 {
			user.Github = request.Github
		}

		if len(request.Telgram) > 0 {
			user.Telgram = request.Telgram
		}

		if len(request.Discord) > 0 {
			user.Discord = request.Discord
		}

		if len(request.Location) > 0 {
			user.Location = request.Location
		}

		if err := s.db.Save(&user).Error; err != nil {
			base.ResponseErr(ctx, http.StatusInternalServerError, err.Error())
			return
		}

		ctx.JSON(http.StatusOK, UpdateUserResponse{
			Data: user,
			BaseResponse: base.BaseResponse{
				Success: true,
			},
		})

	}
}

func (s *Service) updateEmailHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		userId, err := base.GetCurrentUserId(ctx)
		if err != nil {
			base.ResponseErr(ctx, http.StatusBadRequest, err.Error())
			return
		}

		var request UpdateEmailRequest
		if err := ctx.ShouldBindJSON(&request); err != nil {
			base.ResponseErr(ctx, http.StatusBadRequest, err.Error())
			return
		}

		if len(request.NewEmail) == 0 {
			base.ResponseErr(ctx, http.StatusBadRequest, "New email is required")
			return
		}

		var user datamodel.UserModel
		result := s.db.Where("id = ?", userId).First(&user)
		if result.Error != nil {
			if errors.Is(result.Error, gorm.ErrRecordNotFound) {
				base.ResponseErr(ctx, http.StatusOK, "user not found")
				return
			}

			base.ResponseErr(ctx, http.StatusInternalServerError, result.Error.Error())
			return
		}

		user.Email = request.NewEmail
		if err := s.db.Save(&user).Error; err != nil {
			base.ResponseErr(ctx, http.StatusInternalServerError, err.Error())
			return
		}

		ctx.JSON(http.StatusOK, UpdateUserResponse{
			Data: user,
			BaseResponse: base.BaseResponse{
				Success: true,
			},
		})
	}
}

func (s *Service) updatePasswordHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		userId, err := base.GetCurrentUserId(ctx)
		if err != nil {
			base.ResponseErr(ctx, http.StatusBadRequest, err.Error())
			return
		}

		var request UpdatePasswordRequest
		if err := ctx.ShouldBindJSON(&request); err != nil {
			base.ResponseErr(ctx, http.StatusBadRequest, err.Error())
			return
		}

		if len(request.CurrentPassword) == 0 {
			base.ResponseErr(ctx, http.StatusBadRequest, "current password is required")
			return
		}
		if len(request.NewPassword) == 0 {
			base.ResponseErr(ctx, http.StatusBadRequest, "new password is required")
			return
		}

		var user datamodel.UserModel
		result := s.db.Where("id = ?", userId).First(&user)
		if result.Error != nil {
			if errors.Is(result.Error, gorm.ErrRecordNotFound) {
				base.ResponseErr(ctx, http.StatusOK, "user not found")
				return
			}

			base.ResponseErr(ctx, http.StatusInternalServerError, result.Error.Error())
			return
		}

		// check current pasword
		if !verifyPassword(user.EncryptedPassword, request.CurrentPassword) {
			base.ResponseErr(ctx, http.StatusBadRequest, "current password not match")
			return
		}

		// check new password is same current password
		if verifyPassword(user.EncryptedPassword, request.NewPassword) {
			base.ResponseErr(ctx, http.StatusBadRequest, "password not change")
			return
		}

		// update new password
		encryptedPassword, err := generatePassword(request.NewPassword)
		if err != nil {
			base.ResponseErr(ctx, http.StatusBadRequest, err.Error())
			return
		}

		if err := s.db.Model(&user).Update("encrypted_password", encryptedPassword).Error; err != nil {
			base.ResponseErr(ctx, http.StatusInternalServerError, err.Error())
			return
		}
		ctx.JSON(http.StatusOK, UpdateUserResponse{
			Data: user,
			BaseResponse: base.BaseResponse{
				Success: true,
			},
		})
	}
}

func (s *Service) uploadAvatarHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		userId, err := base.GetCurrentUserId(ctx)
		if err != nil {
			base.ResponseErr(ctx, http.StatusBadRequest, err.Error())
			return
		}

		file, err := ctx.FormFile("avatar")
		if err != nil {
			base.ResponseErr(ctx, http.StatusBadRequest, err.Error())
			return
		}

		src, err := file.Open()
		if err != nil {
			base.ResponseErr(ctx, http.StatusInternalServerError, err.Error())
			return
		}
		defer src.Close()

		filePath := fmt.Sprintf("avatars/user-%d-%s", userId, file.Filename)
		if err := s.s3Cliet.MakeBucket(ctx, "hyperdot"); err != nil {
			base.ResponseErr(ctx, http.StatusInternalServerError, err.Error())
			return
		}

		uploadInfo, err := s.s3Cliet.Put(ctx, "hyperdot", filePath, src, file.Size)
		if err != nil {
			base.ResponseErr(ctx, http.StatusInternalServerError, err.Error())
			return
		}

		result := s.db.Model(&datamodel.UserModel{}).Where("id = ?", userId).Update("icon_url", uploadInfo.Key)
		if result.Error != nil {
			base.ResponseErr(ctx, http.StatusInternalServerError, result.Error.Error())
			return
		}

		base.ResponseWithData(ctx, gin.H{
			"object_size": uploadInfo.Size,
			"object_key":  uploadInfo.Key,
		})
	}
}

func (s *Service) getAvatarHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		userId, err := base.GetCurrentUserId(ctx)
		if err != nil {
			base.ResponseErr(ctx, http.StatusBadRequest, err.Error())
			return
		}

		var user datamodel.UserModel
		result := s.db.Where("id = ?", userId).First(&user)
		if result.Error != nil {
			if errors.Is(result.Error, gorm.ErrRecordNotFound) {
				base.ResponseErr(ctx, http.StatusOK, "user not found")
				return
			}

			base.ResponseErr(ctx, http.StatusInternalServerError, result.Error.Error())
			return
		}

		if len(user.IconUrl) == 0 {
			base.ResponseErr(ctx, http.StatusOK, "user avatar not found")
			return
		}

		obj, err := s.s3Cliet.Get(ctx, "hyperdot", user.IconUrl)
		if err != nil {
			base.ResponseErr(ctx, http.StatusInternalServerError, err.Error())
			return
		}
		defer obj.Close()

		// get icon content type
		var contentType string
		seq := strings.Split(user.IconUrl, ".")
		if len(seq) == 0 {
			contentType = "image/jpeg"
		} else {
			contentType = "image/" + seq[len(seq)-1]
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

		ctx.Status(http.StatusOK)
	}
}

func (s *Service) createAccountHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var request CreateAccountRequest
		if err := ctx.ShouldBindJSON(&request); err != nil {
			base.ResponseErr(ctx, http.StatusBadRequest, err.Error())
			return
		}

		if len(request.Provider) == 0 {
			request.Provider = PasswordProvider
		}

		if enable, ok := s.authProviders[request.Provider]; !ok || !enable {
			base.ResponseErr(ctx, http.StatusBadRequest, "unsupported provider %s", request.Provider)
			return
		}

		switch request.Provider {
		case "password":
			if len(request.UserId) == 0 {
				base.ResponseErr(ctx, http.StatusBadRequest, "user name is required")
				return
			}
			if len(request.Email) == 0 {
				base.ResponseErr(ctx, http.StatusBadRequest, "email is required")
				return
			}
			if len(request.Password) == 0 {
				base.ResponseErr(ctx, http.StatusBadRequest, "password is required")
				return
			}

			var existingUser datamodel.UserModel
			result := s.db.Where("username = ? OR email = ?", request.UserId, request.Email).First(&existingUser)
			if result.Error == nil {
				if existingUser.UID == request.UserId {
					base.ResponseErr(ctx, http.StatusBadRequest, "the user %s already exists", request.UserId)
				} else if existingUser.Email == request.Email {
					base.ResponseErr(ctx, http.StatusBadRequest, "the email %s already exists", request.Email)
				}
				return
			}
			if result.Error != nil && !errors.Is(result.Error, gorm.ErrRecordNotFound) {
				base.ResponseErr(ctx, http.StatusInternalServerError, result.Error.Error())
				return
			}

			encryptedPassword, err := generatePassword(request.Password)
			if err != nil {
				base.ResponseErr(ctx, http.StatusBadRequest, err.Error())
				return
			}

			user := datamodel.UserModel{
				UserBasic: datamodel.UserBasic{
					Provider: request.Provider,
					// UID:      "", TODO, by uuid
					Email:             request.Email,
					Username:          request.UserId,
					EncryptedPassword: encryptedPassword,
				},
			}

			// write user to db
			if err := s.db.Create(&user).Error; err != nil {
				base.ResponseErr(ctx, http.StatusInternalServerError, err.Error())
				return
			}

		case "github":
			// 实现GitHub注册逻辑
			// 根据GitHub提供的信息创建用户账户

		default:
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "Unsupported provider"})
			return
		}

		ctx.JSON(http.StatusOK, gin.H{"message": "Account created successfully"})
	}
}

func (s *Service) loginHandle() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var request LoginRequest
		if err := ctx.ShouldBindJSON(&request); err != nil {
			base.ResponseErr(ctx, http.StatusBadRequest, err.Error())
			return
		}

		if len(request.Provider) == 0 {
			request.Provider = PasswordProvider
		}

		if enable, ok := s.authProviders[request.Provider]; !ok || !enable {
			base.ResponseErr(ctx, http.StatusBadRequest, "unsupported provider %s", request.Provider)
			return
		}
		switch request.Provider {
		case "password":
			if len(request.UserId) == 0 && len(request.Email) == 0 {
				base.ResponseErr(ctx, http.StatusBadRequest, "user name or email is required")
				return
			}

			if len(request.Password) == 0 {
				base.ResponseErr(ctx, http.StatusBadRequest, "password is required")
				return
			}
			var existingUser datamodel.UserModel
			result := s.db.Where("username = ? OR email = ?", request.UserId, request.Email).First(&existingUser)
			if result.Error != nil {
				if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
					base.ResponseErr(ctx, http.StatusInternalServerError, result.Error.Error())
					return
				}

				base.ResponseErr(ctx, http.StatusOK, "user %s or email %s not found", request.UserId, request.Email)
				return
			}

			if !verifyPassword(existingUser.EncryptedPassword, request.Password) {
				base.ResponseErr(ctx, http.StatusOK, "password not match")
				return
			}

			expireAt := base.TokenDefaultExpireTime()
			signing, err := base.GenerateJwtToken(existingUser.ToClaims(), expireAt)
			if err != nil {
				base.ResponseErr(ctx, http.StatusInternalServerError, err.Error())
				return
			}

			http.SetCookie(ctx.Writer, &http.Cookie{
				Name:    "token",
				Value:   url.QueryEscape(signing),
				Expires: expireAt,
			})

			ctx.JSON(http.StatusOK, LoginResponse{
				BaseResponse: base.BaseResponse{
					Success: true,
				},
				Data: LoginResponseData{
					Algorithm: "HS256",
					Token:     signing,
				},
			})

		default:
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "Unsupported provider"})
			return
		}
	}
}

func (s *Service) Name() string {
	return ServiceName
}

func (s *Service) RouteTables() []base.RouteTable {
	group := "user"
	return []base.RouteTable{
		{
			Method:  "GET",
			Path:    group,
			Handler: s.getCurrentUserHandler(),
		},

		{
			Method:  "GET",
			Path:    group + "/:id",
			Handler: s.getUserHandler(),
		},
		{
			Method:  "PUT",
			Path:    group,
			Handler: s.updateUserHandler(),
		},

		{
			Method:  "PUT",
			Path:    group + "/email",
			Handler: s.updateEmailHandler(),
		},
		{
			Method:  "PUT",
			Path:    group + "/password",
			Handler: s.updatePasswordHandler(),
		},
		{
			Method:  "POST",
			Path:    group + "/avatar/upload",
			Handler: s.uploadAvatarHandler(),
		},
		{
			Method:  "GET",
			Path:    group + "/avatar",
			Handler: s.getAvatarHandler(),
		},

		{
			Method:  "POST",
			Path:    group + "/auth/createAccount",
			Handler: s.createAccountHandler(),
		},
		{
			Method:  "POST",
			Path:    group + "/auth/login",
			Handler: s.loginHandle(),
		},
	}
}
