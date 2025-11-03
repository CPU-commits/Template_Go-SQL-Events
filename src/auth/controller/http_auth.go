package controller

import (
	"net/http"
	"strings"

	"github.com/CPU-commits/Template_Go-EventDriven/src/auth/dto"
	"github.com/CPU-commits/Template_Go-EventDriven/src/auth/service"
	"github.com/CPU-commits/Template_Go-EventDriven/src/cmd/http/utils"
	"github.com/CPU-commits/Template_Go-EventDriven/src/package/bus"
	util "github.com/CPU-commits/Template_Go-EventDriven/src/utils"
	"github.com/gin-gonic/gin"
	"github.com/markbates/goth"
)

type HttpAuthController struct {
	userService service.UserService
	authService service.AuthService
}

func NewAuthHttpController(bus bus.Bus) *HttpAuthController {
	return &HttpAuthController{
		userService: *service.NewUserService(
			sqlUserRepository,
		),
		authService: *service.NewAuthService(
			sqlAuthRepository,
			sqlUserRepository,
			bus,
		),
	}
}

// Register godoc
//
//	@Summary	Registrase dentro de la aplicación de AlpacaTatto
//	@Tags		auth
//	@Success	200		{object}	controller.LoginResponse
//	@Param		authDto	body		dto.RegisterDto			true	"name, username, email, password, role"
//	@Failure	503		{object}	utils.ProblemDetails	"Error con la base de datos"
//
//	@Failure	403		{object}	utils.ProblemDetails	"Credenciales inválidas"
//	@Failure	409		{object}	utils.ProblemDetails	"La sesión no existe. Probablemente porque la eliminaron"
//
//	@Router		/api/auth/register [post]
func (httpAuth *HttpAuthController) Register(c *gin.Context) {
	var registerDto *dto.RegisterDto

	if err := c.BindJSON(&registerDto); err != nil {
		utils.ResErrValidators(c, err)
		return
	}
	registerDto.Email = util.CleanAndLower(registerDto.Email)
	if err := httpAuth.authService.Register(registerDto, true); err != nil {
		utils.ResFromErr(c, err)
		return
	}

	c.JSON(http.StatusCreated, nil)
}

func (httpAuth *HttpAuthController) HandleAuthUser(c *gin.Context, user goth.User) error {
	externalAuthDto := &dto.ExternalAuthDTO{
		Email:     user.Email,
		Name:      user.Name,
		AvatarURL: user.AvatarURL,
		Provider:  c.Param("provider"),
	}

	externalAuthDto.Email = util.CleanAndLower(externalAuthDto.Email)
	userModel, idAuth, err := httpAuth.authService.RegisterOrLoginExternalAuth(externalAuthDto)
	if err != nil {
		return err
	}
	// Generate token and session
	sessionDto := dto.SessionDto{}
	tokenSession, err := sessionService.NewSession(sessionDto, idAuth, userModel.ID)
	if err != nil {
		return err
	}
	tokenAccess, err := sessionService.GenerateAccess(tokenSession, userModel)
	if err != nil {
		return err
	}

	c.SetSameSite(http.SameSiteNoneMode)
	c.SetCookie(
		"auth_access_token",
		tokenAccess,
		3600,
		"/",
		settingsData.CLIENT_DOMAIN,
		true,
		settingsData.GO_ENV == "prod",
	)
	c.SetCookie(
		"refresh_access_token",
		tokenSession,
		3600,
		"/",
		settingsData.CLIENT_DOMAIN,
		true,
		settingsData.GO_ENV == "prod",
	)

	return nil
}

// Login godoc
//
//	@Summary	Loggearse dentro de la aplicación de NeoHome
//	@Tags		auth
//	@Success	200		{object}	controller.LoginResponse
//	@Param		authDto	body		dto.AuthDto				true	"Password y username"
//	@Failure	503		{object}	utils.ProblemDetails	"Error con la base de datos"
//
//	@Failure	403		{object}	utils.ProblemDetails	"Credenciales inválidas"
//	@Failure	409		{object}	utils.ProblemDetails	"La sesión no existe. Probablemente porque la eliminaron"
//
//	@Router		/api/auth/login [post]
func (httpAuth *HttpAuthController) Login(c *gin.Context) {
	var authDto *dto.AuthDto

	if err := c.BindJSON(&authDto); err != nil {
		utils.ResErrValidators(c, err)
		return
	}
	authDto.Username = util.CleanAndLower(authDto.Username)

	user, idAuth, err := httpAuth.authService.LoginLocal(*authDto)
	if err != nil {
		utils.ResFromErr(c, err)
		return
	}
	// Generate token and session
	sessionDto := dto.SessionDto{}
	tokenSession, err := sessionService.NewSession(sessionDto, idAuth, user.ID)
	if err != nil {
		utils.ResFromErr(c, err)
		return
	}
	tokenAccess, err := sessionService.GenerateAccess(tokenSession, user)
	if err != nil {
		utils.ResFromErr(c, err)
		return
	}

	c.JSON(http.StatusOK, LoginResponse{
		AccessToken:  tokenAccess,
		RefreshToken: tokenSession,
		User:         user,
	})
}

// Refresh godoc
//
//	@Summary	Refrescar sesión
//	@Tags		auth
//	@Success	200			{object}	controller.LoginResponse
//	@Param		X-Refresh	header		string					true	"Token de refresco, es decir, de sesión"
//	@Failure	403			{object}	utils.ProblemDetails	"No está el token de refresco en el header X-Refresh"
//	@Failure	400			{object}	utils.ProblemDetails	"No es un token válido JWT"
//	@Failure	404			{object}	utils.ProblemDetails	"El token no tiene un usuario registrado en la BD"
//	@Failure	409			{object}	utils.ProblemDetails	"La sesión no existe. Probablemente porque la eliminaron"
//	@Router		/api/auth/refresh [post]
func (httpAuth *HttpAuthController) Refresh(c *gin.Context) {
	refreshToken := c.GetHeader("X-Refresh")
	if refreshToken == "" {
		utils.ResWithMessageID(c, "refresh_not_found", http.StatusForbidden)
		return
	}
	// Manage token
	token, err := utils.VerifyToken(refreshToken)
	if err != nil {
		utils.ResWithMessageID(c, "unauthorized", http.StatusBadRequest)
		return
	}
	metadata, err := utils.ExtractRefreshTokeMetadata(token)
	if err != nil {
		utils.ResWithMessageID(c, "unauthorized", http.StatusBadRequest)
		return
	}
	// Refresh session with user
	user, err := httpAuth.userService.GetUserById(int64(metadata.UID))
	if err != nil {
		utils.ResFromErr(c, err)
		return
	}
	oldSessionToken := strings.ReplaceAll(refreshToken, "Bearer ", "")

	tokenSession, err := sessionService.RefreshSession(
		oldSessionToken,
		user.ID,
	)
	if err != nil {
		utils.ResFromErr(c, err)
		return
	}
	tokenAccess, err := sessionService.GenerateAccess(
		tokenSession,
		user,
	)
	if err != nil {
		utils.ResFromErr(c, err)
		return
	}

	c.JSON(http.StatusOK, LoginResponse{
		AccessToken:  tokenAccess,
		RefreshToken: tokenSession,
		User:         user,
	})
}
