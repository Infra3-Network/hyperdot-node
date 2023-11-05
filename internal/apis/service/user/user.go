package user

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	_ "gorm.io/driver/postgres"
	"gorm.io/gorm"

	"infra-3.xyz/hyperdot-node/internal/apis/base"
	"infra-3.xyz/hyperdot-node/internal/clients"
	"infra-3.xyz/hyperdot-node/internal/dataengine"
	"infra-3.xyz/hyperdot-node/internal/datamodel"
	"infra-3.xyz/hyperdot-node/internal/utils"
)

const (
	ServiceName      = "user"
	PasswordProvider = "password"
)

// Service user service
type Service struct {
	db            *gorm.DB
	s3Cliet       *clients.SimpleS3Cliet
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

	var data ResponseGetUserData
	if !rows.Next() {
		base.ResponseErr(ctx, http.StatusOK, "user not found")
		return
	}

	if err := s.db.ScanRows(rows, &data); err != nil {
		base.ResponseErr(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	ctx.JSON(http.StatusOK, ResponseGetUser{
		Data: data,
		BaseResponse: base.BaseResponse{
			Success: true,
		},
	})
}

// GetCurrentUserHandler Get the current logined user.
// @Summary Get the current logined user.
// @Description Get the current logined user.
// @Tags user apis
// @Accept application/json
// @Produce application/json
// @Security ApiKeyAuth
// @Success 200 {object} ResponseGetUser
// @Router /user [get]
func (s *Service) GetCurrentUserHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		userId, err := base.GetCurrentUserId(ctx)
		if err != nil {
			base.ResponseErr(ctx, http.StatusBadRequest, err.Error())
			return
		}

		s.getUserInternalHandler(userId, ctx)
	}
}

// GetUserHandler Get user by id.
// @Summary Get user by id.
// @Description Get user by id.
// @Tags user apis
// @Accept application/json
// @Produce application/json
// @Security ApiKeyAuth
// @Param id path int true "user id"
// @Success 200 {object} ResponseGetUser
// @Router /user/{id} [get]
func (s *Service) GetUserHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		id, err := base.GetUintParam(ctx, "id")
		if err != nil {
			base.ResponseErr(ctx, http.StatusBadRequest, err.Error())
			return
		}

		s.getUserInternalHandler(id, ctx)
	}
}

// UpdateUserHandler Update logined user info.
// @Summary Update user info.
// @Description Update user info.
// @Tags user apis
// @Accept application/json
// @Produce application/json
// @Security ApiKeyAuth
// @Param Authorization header string true "token"
// @Param body body datamodel.UserModel true "update user request"
// @Success 200 {object} ResponseUpdateUser
// @Router /user [put]
func (s *Service) UpdateUserHandler() gin.HandlerFunc {
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

		if len(request.Username) > 0 {
			if request.Username == user.Username {
				base.ResponseErr(ctx, http.StatusOK, "username not change")
				return
			}
			user.Username = request.Username
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

		ctx.JSON(http.StatusOK, ResponseUpdateUser{
			Data: user,
			BaseResponse: base.BaseResponse{
				Success: true,
			},
		})

	}
}

// UpdateEmailHandler Update logined user email.
// @Summary Update user email.
// @Description Update user email.
// @Tags user apis
// @Accept application/json
// @Produce application/json
// @Security ApiKeyAuth
// @Param Authorization header string true "token"
// @Param body body RequestUpdateEmail true "update email request"
// @Success 200 {object} ResponseUpdateUser
// @Router /user/email [put]
func (s *Service) UpdateEmailHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		userId, err := base.GetCurrentUserId(ctx)
		if err != nil {
			base.ResponseErr(ctx, http.StatusBadRequest, err.Error())
			return
		}

		var request RequestUpdateEmail
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

		ctx.JSON(http.StatusOK, ResponseUpdateUser{
			Data: user,
			BaseResponse: base.BaseResponse{
				Success: true,
			},
		})
	}
}

// UpdatePasswordHandler Update logined user password.
// @Summary Update user password.
// @Description Update user password.
// @Tags user apis
// @Accept application/json
// @Produce application/json
// @Security ApiKeyAuth
// @Param Authorization header string true "token"
// @Param body body RequestUpdatePassword true "update password request"
// @Success 200 {object} ResponseUpdateUser
// @Router /user/password [put]
func (s *Service) UpdatePasswordHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		userId, err := base.GetCurrentUserId(ctx)
		if err != nil {
			base.ResponseErr(ctx, http.StatusBadRequest, err.Error())
			return
		}

		var request RequestUpdatePassword
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
		if !utils.VerifyPassword(user.EncryptedPassword, request.CurrentPassword) {
			base.ResponseErr(ctx, http.StatusBadRequest, "current password not match")
			return
		}

		// check new password is same current password

		if utils.VerifyPassword(user.EncryptedPassword, request.NewPassword) {
			base.ResponseErr(ctx, http.StatusBadRequest, "password not change")
			return
		}

		// update new password
		encryptedPassword, err := utils.GeneratePassword(request.NewPassword)
		if err != nil {
			base.ResponseErr(ctx, http.StatusBadRequest, err.Error())
			return
		}

		if err := s.db.Model(&user).Update("encrypted_password", encryptedPassword).Error; err != nil {
			base.ResponseErr(ctx, http.StatusInternalServerError, err.Error())
			return
		}
		ctx.JSON(http.StatusOK, ResponseUpdateUser{
			Data: user,
			BaseResponse: base.BaseResponse{
				Success: true,
			},
		})
	}
}

// UploadAvatarHandler Upload user avatar.
// @Summary Upload user avatar.
// @Description Upload user avatar.
// @Tags user apis
// @Accept multipart/form-data
// @Produce application/json
// @Security ApiKeyAuth
// @Param Authorization header string true "token"
// @Param avatar formData file true "avatar file"
// @Success 200 {object} ResponseUploadAvatar
// @Router /user/avatar/upload [post]
func (s *Service) UploadAvatarHandler() gin.HandlerFunc {
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

		ctx.JSON(http.StatusOK, ResponseUploadAvatar{
			BaseResponse: base.BaseResponse{
				Success: true,
			},
			Data: ResponseUploadAvatarData{
				Key:     uploadInfo.Key,
				Filsize: uploadInfo.Size,
			},
		})
	}
}

// GetAvatarHandler Get user avatar.
// @Summary Get user avatar.
// @Description Get user avatar.
// @Tags user apis
// @Accept application/json
// @Produce image/jpeg
// @Security ApiKeyAuth
// @Param Authorization header string true "token"
// @Router /user/avatar [get]
func (s *Service) GetAvatarHandler() gin.HandlerFunc {
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

// CreateAccountHandler Create account by username and password.
// @Summary Create account by username and password.
// @Description Create account by username and password.
// @Tags user apis
// @Accept application/json
// @Produce application/json
// @Param body body RequestCreateAccount true "create account request"
// @Success 200 {object} ResponseCreateAccount
// @Router /user/auth/createAccount [post]
func (s *Service) CreateAccountHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var request RequestCreateAccount
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
			if len(request.Username) == 0 {
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
			result := s.db.Where("username = ? OR email = ?", request.Username, request.Email).First(&existingUser)
			if result.Error == nil {
				if existingUser.Username == request.Username {
					base.ResponseErr(ctx, http.StatusBadRequest, "the user %s already exists", request.Username)
				} else if existingUser.Email == request.Email {
					base.ResponseErr(ctx, http.StatusBadRequest, "the email %s already exists", request.Email)
				}
				return
			}
			if result.Error != nil && !errors.Is(result.Error, gorm.ErrRecordNotFound) {
				base.ResponseErr(ctx, http.StatusInternalServerError, result.Error.Error())
				return
			}

			encryptedPassword, err := utils.GeneratePassword(request.Password)
			if err != nil {
				base.ResponseErr(ctx, http.StatusBadRequest, err.Error())
				return
			}

			user := datamodel.UserModel{
				UserBasic: datamodel.UserBasic{
					Provider:          request.Provider,
					Email:             request.Email,
					Username:          request.Username,
					EncryptedPassword: encryptedPassword,
				},
			}

			// write user to db
			if err := s.db.Create(&user).Error; err != nil {
				base.ResponseErr(ctx, http.StatusInternalServerError, err.Error())
				return
			}

			base.ResponseWithData(ctx, ResponseCreateAccount{})
			return

		case "github":
			// TODO: support login by github
			base.ResponseErr(ctx, http.StatusBadRequest, "unsupported provider %s", request.Provider)
			return

		default:
			base.ResponseErr(ctx, http.StatusBadRequest, "unsupported provider %s", request.Provider)
			return
		}
	}
}

// LoginHandle Login by username and password.
// @Summary Login by username and password.
// @Description Login by username and password.
// @Tags user apis
// @Accept application/json
// @Produce application/json
// @Param body body RequestLogin true "login request"
// @Success 200 {object} ResponseLogin
// @Router /user/auth/login [post]
func (s *Service) LoginHandle() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var request RequestLogin
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

			if !utils.VerifyPassword(existingUser.EncryptedPassword, request.Password) {
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

			base.ResponseWithData(ctx, ResponseLogin{
				Algorithm: "HS256",
				Token:     signing,
			})

		default:
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "Unsupported provider"})
			return
		}
	}
}

// Name service name
func (s *Service) Name() string {
	return ServiceName
}

// RouteTables route tables
func (s *Service) RouteTables() []base.RouteTable {
	group := "user"
	return []base.RouteTable{
		{
			Method:  "GET",
			Path:    group,
			Handler: s.GetCurrentUserHandler(),
		},

		{
			Method:  "GET",
			Path:    group + "/:id",
			Handler: s.GetUserHandler(),
		},
		{
			Method:  "PUT",
			Path:    group,
			Handler: s.UpdateUserHandler(),
		},

		{
			Method:  "PUT",
			Path:    group + "/email",
			Handler: s.UpdateEmailHandler(),
		},
		{
			Method:  "PUT",
			Path:    group + "/password",
			Handler: s.UpdatePasswordHandler(),
		},
		{
			Method:  "POST",
			Path:    group + "/avatar/upload",
			Handler: s.UploadAvatarHandler(),
		},
		{
			Method:  "GET",
			Path:    group + "/avatar",
			Handler: s.GetAvatarHandler(),
		},

		{
			Method:     "POST",
			Path:       group + "/auth/createAccount",
			Handler:    s.CreateAccountHandler(),
			AllowGuest: true,
			Regexp:     "",
		},
		{
			Method:     "POST",
			Path:       group + "/auth/login",
			Handler:    s.LoginHandle(),
			AllowGuest: true,
			Regexp:     "",
		},
	}
}
