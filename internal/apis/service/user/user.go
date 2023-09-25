package user

import (
	"errors"
	"net/http"
	"net/url"

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
	db *gorm.DB
	// auth          *auth.Auth
	authProviders map[string]bool
	engines       map[string]dataengine.QueryEngine
}

// New user service
func New(db *gorm.DB, engines map[string]dataengine.QueryEngine) *Service {
	svc := &Service{db: db,
		engines: engines,
		authProviders: map[string]bool{
			PasswordProvider: true,
		},
	}
	if err := svc.init(); err != nil {
		panic(err)
	}
	return svc
}

// init user service
func (s *Service) init() error {
	// Migrate AuthIdentity model, AuthIdentity will be used to save auth info, like username/password, oauth token, you could change that.
	if err := s.db.AutoMigrate(&datamodel.UserModel{}); err != nil {
		return err
	}
	if err := s.db.AutoMigrate(&datamodel.UserQueryModel{}); err != nil {
		return err
	}

	return nil
	// Register Auth providers
	// Allow use username/password
	//s.auth.RegisterProvider(password.New(&password.Config{}))
	//
	//// Allow use Github
	//s.auth.RegisterProvider(github.New(&github.Config{
	//	ClientID:     "github client id",
	//	ClientSecret: "github client secret",
	//}))
	//
	//// Allow use Google
	//s.auth.RegisterProvider(google.New(&google.Config{
	//	ClientID:       "google client id",
	//	ClientSecret:   "google client secret",
	//	AllowedDomains: []string{}, // Accept all domains, instead you can pass a whitelist of acceptable domains
	//}))
	//
	//// Allow use Facebook
	//s.auth.RegisterProvider(facebook.New(&facebook.Config{
	//	ClientID:     "facebook client id",
	//	ClientSecret: "facebook client secret",
	//}))
	//
	//// Allow use Twitter
	//s.auth.RegisterProvider(twitter.New(&twitter.Config{
	//	ClientID:     "twitter client id",
	//	ClientSecret: "twitter client secret",
	//}))
}

func (s *Service) getUserHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		v, ok := ctx.Get("user_id")
		if !ok {
			base.ResponseErr(ctx, http.StatusUnauthorized, "Unauthorized")
			return
		}
		userId := v.(uint)
		var user datamodel.UserModel
		result := s.db.First(&user, userId)
		if result.Error != nil {
			if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
				base.ResponseErr(ctx, http.StatusInternalServerError, result.Error.Error())
				return
			}

			base.ResponseErr(ctx, http.StatusOK, "user not found")
			return
		}
		user.EncryptedPassword = ""

		ctx.JSON(http.StatusOK, GetUserResponse{
			UserModel: user,
			BaseResponse: base.BaseResponse{
				Success: true,
			},
		})
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
			Handler: s.getUserHandler(),
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
