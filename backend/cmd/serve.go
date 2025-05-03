package cmd

import (
	"crypto/ed25519"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path"
	"reflect"
	"strings"
	"time"

	"github.com/go-playground/locales/en"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	en_translations "github.com/go-playground/validator/v10/translations/en"
	"github.com/golang-jwt/jwt/v5"
	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/stashsphere/backend/config"
	"github.com/stashsphere/backend/crypto"
	"github.com/stashsphere/backend/handlers"
	inv_middleware "github.com/stashsphere/backend/middleware"
	"github.com/stashsphere/backend/operations"
	"github.com/stashsphere/backend/services"
	"github.com/stashsphere/backend/utils"
	"github.com/volatiletech/sqlboiler/v4/boil"

	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/confmap"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
	_ "github.com/lib/pq"
	"github.com/spf13/cobra"
)

type CustomValidator struct {
	validator *validator.Validate
	trans     *ut.Translator
}

func (cv *CustomValidator) Validate(i interface{}) error {
	if err := cv.validator.Struct(i); err != nil {
		validationErrors := err.(validator.ValidationErrors)
		errors := utils.StashsphereValidationError{
			Errors: make(map[string]string),
		}
		for _, fieldErr := range validationErrors {
			errors.Errors[fieldErr.Field()] = fieldErr.Translate(*cv.trans)
		}
		return errors
	}
	return nil
}

func Serve(config config.StashsphereServeConfig, debug bool) error {
	consoleOutput := zerolog.ConsoleWriter{Out: os.Stderr}
	loggerOutput := consoleOutput
	logger := zerolog.New(loggerOutput)
	log.Logger = logger

	boil.DebugMode = debug

	if config.PrivateKey == "" {
		log.Warn().Msg("No private key provided, generating one. Cookies won't work after restart")
		generatedKey, err := crypto.GenerateEd25519StringKey()
		if err != nil {
			return err
		}
		config.PrivateKey = generatedKey
	}

	if config.InviteEnabled {
		log.Info().Msgf("Invite enabled and code required")
	} else {
		log.Info().Msgf("Invite disabled and no code required")
	}

	privateKey, err := crypto.LoadEd22519PrivateKeyFromString(config.PrivateKey)
	if err != nil {
		log.Fatal().Msgf("error loading private key from config: %v", err)
	}
	publicKey := privateKey.Public().(ed25519.PublicKey)

	dbOptions := fmt.Sprintf("user=%s dbname=%s host=%s", config.User, config.Name, config.Host)
	if config.Password != nil {
		dbOptions = fmt.Sprintf("%s password=%s", dbOptions, *config.Password)
	}
	if config.Port != nil {
		dbOptions = fmt.Sprintf("%s port=%d", dbOptions, *config.Port)
	}
	if config.SslMode != nil {
		dbOptions = fmt.Sprintf("%s sslmode=%s", dbOptions, *config.SslMode)
	}

	db, err := sql.Open("postgres", dbOptions)
	if err != nil {
		return err
	}

	defer func() {
		if err := db.Close(); err != nil {
			panic(err)
		}
	}()

	e := echo.New()

	e.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogURI:    true,
		LogStatus: true,
		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
			logger.Info().
				Str("URI", v.URI).
				Int("status", v.Status).
				Msg("request")

			return nil
		},
	}))

	en := en.New()
	uni := ut.New(en, en)
	trans, _ := uni.GetTranslator("en")
	validate := validator.New()
	en_translations.RegisterDefaultTranslations(validate, trans)

	// https://github.com/go-playground/validator/issues/861#issuecomment-976696946
	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("form"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})
	authService := services.NewAuthService(db, privateKey, publicKey, 6*time.Hour, config.ApiDomain)
	userService := services.NewUserService(db, config.InviteEnabled, config.InviteCode)
	notificationService := services.NewNotificationService(db)
	imageService, err := services.NewImageService(db, config.ImagePath)
	if err != nil {
		return err
	}
	cacheService, err := services.NewCacheService(config.ImageCachePath)
	if err != nil {
		return err
	}
	thingService := services.NewThingService(db, imageService)
	listService := services.NewListService(db)
	propertyService := services.NewPropertyService(db)
	searchService := services.NewSearchService(db, thingService, listService)
	shareService := services.NewShareService(db)
	friendService := services.NewFriendService(db, notificationService)

	e.Validator = &CustomValidator{validator: validate, trans: &trans}
	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Output: loggerOutput,
	}))
	e.Use(middleware.Recover())
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:     config.AllowedDomains,
		AllowCredentials: true,
	}))
	e.Use(echojwt.WithConfig(echojwt.Config{
		SigningKey:    publicKey,
		TokenLookup:   "cookie:stashsphere-access",
		SigningMethod: "EdDSA",
		ErrorHandler: func(c echo.Context, err error) error {
			var extratorErr *echojwt.TokenExtractionError
			var parsingErr *echojwt.TokenParsingError
			if err == echojwt.ErrJWTMissing || errors.As(err, &extratorErr) || errors.As(err, &parsingErr) {
				return nil
			}
			return err
		},
		ContinueOnIgnoredError: true,
		NewClaimsFunc: func(c echo.Context) jwt.Claims {
			return &operations.ApplicationClaims{}
		},
		ContextKey: "token",
	}))
	e.Use(inv_middleware.ExtractClaims("token"))
	e.Use(inv_middleware.HeadToGetMiddleware)
	e.HTTPErrorHandler = inv_middleware.CreateStashSphereHTTPErrorHandler(e)
	// TODO add refresh token middleware: check whether accessToken is less than 15min of lifetime, try to access refreshtoken, if validate
	// set new access and refresh token

	loginHandler := handlers.NewLoginHandler(authService)
	registerHandler := handlers.NewRegisterHandler(userService)
	thingHandler := handlers.NewThingHandler(thingService, listService, propertyService)
	listHandler := handlers.NewListHandler(listService)
	imageHandler := handlers.NewImageHandler(imageService, cacheService)
	searchHandler := handlers.NewSearchHandler(searchService, listService)
	profileHandler := handlers.NewProfileHandler(userService)
	userHandler := handlers.NewUserHandler(userService)
	shareHandler := handlers.NewShareHandler(shareService)
	friendHandler := handlers.NewFriendHandler(friendService)
	notificationHandler := handlers.NewNotificationHandler(notificationService)

	a := e.Group("/api")
	userGroup := a.Group("/user")
	userGroup.POST("/login", loginHandler.LoginHandlerPost)
	userGroup.DELETE("/logout", loginHandler.LogoutHandlerDelete)
	userGroup.POST("/register", registerHandler.RegisterHandlerPost)
	userGroup.GET("/profile", profileHandler.ProfileHandlerGet)
	userGroup.PATCH("/profile", profileHandler.ProfileHandlerPatch)

	usersGroup := a.Group("/users")
	usersGroup.GET("", userHandler.Index)
	usersGroup.GET("/:userId", userHandler.Get)

	thingsGroup := a.Group("/things")
	thingsGroup.GET("", thingHandler.ThingHandlerIndex)
	thingsGroup.POST("", thingHandler.ThingHandlerPost)
	thingsGroup.PATCH("/:thingId", thingHandler.ThingHandlerPatch)
	thingsGroup.GET("/:thingId", thingHandler.ThingHandlerShow)
	// TODO
	// thingsGroup.DELETE("/:thingId", thingHandler.ThingHandlerDelete)

	listsGroup := a.Group("/lists")
	listsGroup.GET("", listHandler.ListHandlerIndex)
	listsGroup.POST("", listHandler.ListHandlerPost)
	listsGroup.GET("/:listId", listHandler.ListHandlerShow)
	listsGroup.PATCH("/:listId", listHandler.ListHandlerPatch)

	imageGroup := a.Group("/images")
	imageGroup.GET("", imageHandler.ImageHandlerIndex)
	imageGroup.POST("", imageHandler.ImageHandlerPost)
	imageGroup.PATCH("/:imageId", imageHandler.ImageHandlerPatch)
	imageGroup.DELETE("/:imageId", imageHandler.ImageHandlerDelete)

	shareGroup := a.Group("/shares")
	shareGroup.POST("", shareHandler.ShareHandlerPost)
	shareGroup.GET("/:shareId", shareHandler.ShareHandlerGet)
	shareGroup.DELETE("/:shareId", shareHandler.ShareHandlerDelete)

	friendGroup := a.Group("/friends")
	friendGroup.GET("", friendHandler.FriendsIndex)
	friendGroup.DELETE("/:friendId", friendHandler.FriendDelete)

	friendRequestGroup := a.Group("/friend_requests")
	friendRequestGroup.GET("", friendHandler.FriendRequestIndex)
	friendRequestGroup.POST("", friendHandler.FriendRequestPost)
	friendRequestGroup.DELETE("/:requestId", friendHandler.FriendRequestDelete)
	friendRequestGroup.PATCH("/:requestId", friendHandler.FriendRequestUpdate)

	notificationsGroup := a.Group("/notifications")
	notificationsGroup.GET("", notificationHandler.Index)
	notificationsGroup.PATCH("/:notificationId", notificationHandler.Acknowledge)

	a.GET("/search", searchHandler.SearchHandlerGet)

	e.GET("/assets/:hash", imageHandler.ImageHandlerGet)
	e.HEAD("/assets/:hash", imageHandler.ImageHandlerGet)

	return e.Start(config.ListenAddress)
}

var serveCommand = &cobra.Command{
	Use:   "serve",
	Short: "Serve the Stashsphere API",
	RunE: func(cmd *cobra.Command, args []string) error {
		configPaths, _ := cmd.Flags().GetStringSlice("conf")
		debug, _ := cmd.Flags().GetBool("debug")

		var config config.StashsphereServeConfig

		stateDir := os.Getenv("STATE_DIRECTORY")
		if stateDir == "" {
			stateDir = "."
		}
		cacheDir := os.Getenv("CACHE_DIRECTORY")
		if cacheDir == "" {
			cacheDir = "."
		}
		imagePath := path.Join(stateDir, "image_store")
		imageCachePath := path.Join(cacheDir, "image_cache")

		k := koanf.New(".")
		k.Load(confmap.Provider(map[string]interface{}{
			"database.user":   "stashsphere",
			"database.name":   "stashsphere",
			"database.host":   "127.0.0.1",
			"listenAddress":   ":8081",
			"auth.privateKey": "",
			"imagePath":       imagePath,
			"imageCachePath":  imageCachePath,
			"invites.enabled": false,
			"invites.code":    "",
			"domains.allowed": []string{"http://localhost"},
			"domains.own":     []string{"localhost"},
		}, "."), nil)

		for _, configPath := range configPaths {
			if err := k.Load(file.Provider(configPath), yaml.Parser()); err != nil {
				log.Fatal().Msgf("error loading config: %v", err)
			}
			k.UnmarshalWithConf("", &config, koanf.UnmarshalConf{Tag: "koanf", FlatPaths: true})
		}

		return Serve(config, debug)
	},
}

func init() {
	serveCommand.Flags().StringSlice("conf", []string{"stashsphere.yaml"}, "path to one or more .yaml config files")
	serveCommand.Flags().Bool("debug", false, "enable debug mode")
	rootCmd.AddCommand(serveCommand)
}
