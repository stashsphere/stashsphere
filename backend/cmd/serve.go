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

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/go-fuego/fuego"
	"github.com/go-fuego/fuego/extra/fuegoecho"
	"github.com/go-fuego/fuego/option"
	"github.com/go-fuego/fuego/param"
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
	ss_middleware "github.com/stashsphere/backend/middleware"
	"github.com/stashsphere/backend/operations"
	"github.com/stashsphere/backend/resources"
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

func setup(config config.StashSphereServeConfig, debug bool, serveOpenAPI bool, openAPIPath string) (*echo.Echo, *fuego.Engine, *sql.DB, error) {
	consoleOutput := zerolog.ConsoleWriter{Out: os.Stderr}
	loggerOutput := consoleOutput
	logger := zerolog.New(loggerOutput).With().Timestamp().Logger()
	log.Logger = logger

	boil.DebugMode = debug

	if config.Auth.PrivateKey == "" {
		log.Warn().Msg("No private key provided, generating one. Cookies won't work after restart")
		generatedKey, err := crypto.GenerateEd25519StringKey()
		if err != nil {
			return nil, nil, nil, err
		}
		config.Auth.PrivateKey = generatedKey
	}

	if config.Invites.Enabled {
		log.Info().Msgf("Invite enabled and code required")
	} else {
		log.Info().Msgf("Invite disabled and no code required")
	}

	privateKey, err := crypto.LoadEd22519PrivateKeyFromString(config.Auth.PrivateKey)
	if err != nil {
		log.Fatal().Msgf("error loading private key from config: %v", err)
	}
	publicKey := privateKey.Public().(ed25519.PublicKey)

	dbOptions := fmt.Sprintf("user=%s dbname=%s host=%s", config.Database.User, config.Database.Name, config.Database.Host)
	if config.Database.Password != nil {
		dbOptions = fmt.Sprintf("%s password=%s", dbOptions, *config.Database.Password)
	}
	if config.Database.Port != nil {
		dbOptions = fmt.Sprintf("%s port=%d", dbOptions, *config.Database.Port)
	}
	if config.Database.SslMode != nil {
		dbOptions = fmt.Sprintf("%s sslmode=%s", dbOptions, *config.Database.SslMode)
	}

	db, err := sql.Open("postgres", dbOptions)
	if err != nil {
		return nil, nil, nil, err
	}

	e := echo.New()
	e.HideBanner = true
	e.HidePort = true

	engine := fuego.NewEngine(
		fuego.WithOpenAPIConfig(fuego.OpenAPIConfig{
			Disabled:         !serveOpenAPI,
			DisableSwaggerUI: !serveOpenAPI,
			DisableLocalSave: false,
			DisableMessages:  false,
			PrettyFormatJSON: true,
			SwaggerURL:       "/swagger",
			SpecURL:          "/swagger/openapi.json",
			JSONFilePath:     openAPIPath,
			UIHandler:        fuego.DefaultOpenAPIHandler,
		}),
	)

	engine.OpenAPI.Description().Components.SecuritySchemes = openapi3.SecuritySchemes{
		"cookieAuth": &openapi3.SecuritySchemeRef{
			Value: openapi3.NewSecurityScheme().
				WithType("apiKey").
				WithIn("cookie").
				WithName("stashsphere-access").
				WithDescription("JWT access token stored in HTTP-only cookie"),
		},
	}
	engine.OpenAPI.Description().Info = &openapi3.Info{
		Title:   "OpenAPI Documentation for Stashsphere",
		Version: "0.9.0",
	}

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
	authService := services.NewAuthService(db, privateKey, publicKey, 6*time.Hour, 24*7*time.Hour, config.Domains.ApiDomain)
	userService := services.NewUserService(db, config.Invites.Enabled, config.Invites.InviteCode)

	emailService := services.NewEmailService(config.Email)
	notificationService := services.NewNotificationService(db,
		services.NotificationData{
			FrontendUrl:  config.FrontendUrl,
			InstanceName: config.InstanceName,
		}, emailService)
	imageService, err := services.NewImageService(db, config.Image.Path)
	if err != nil {
		return nil, nil, nil, err
	}
	cacheService, err := services.NewCacheService(config.Image.CachePath)
	if err != nil {
		return nil, nil, nil, err
	}
	thingService := services.NewThingService(db, imageService)
	listService := services.NewListService(db, notificationService)
	propertyService := services.NewPropertyService(db)
	searchService := services.NewSearchService(db, thingService, listService)
	shareService := services.NewShareService(db, notificationService)
	friendService := services.NewFriendService(db, notificationService)
	cartService := services.NewCartService(db)

	e.Validator = &CustomValidator{validator: validate, trans: &trans}
	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Output: loggerOutput,
		Format: `{"level":"info", "time":"${time_rfc3339_nano}","id":"${id}","remote_ip":"${remote_ip}",` +
			`"host":"${host}","method":"${method}","uri":"${uri}","user_agent":"${user_agent}",` +
			`"status":${status},"error":"${error}","latency":${latency},"latency_human":"${latency_human}"` +
			`,"bytes_in":${bytes_in},"bytes_out":${bytes_out}}` + "\n",
	}))
	e.Use(middleware.Recover())
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:     config.Domains.AllowedDomains,
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
			return &operations.AccessClaims{}
		},
		ContextKey: "token",
	}))
	e.Use(ss_middleware.ExtractClaims("token"))
	e.Use(ss_middleware.HeadToGetMiddleware)
	e.HTTPErrorHandler = ss_middleware.CreateStashSphereHTTPErrorHandler(e)

	loginHandler := handlers.NewLoginHandler(authService)
	registerHandler := handlers.NewRegisterHandler(userService)
	thingHandler := handlers.NewThingHandler(thingService, listService)
	listHandler := handlers.NewListHandler(listService)
	imageHandler := handlers.NewImageHandler(imageService, cacheService)
	searchHandler := handlers.NewSearchHandler(searchService, listService, propertyService)
	profileHandler := handlers.NewProfileHandler(userService)
	userHandler := handlers.NewUserHandler(userService)
	shareHandler := handlers.NewShareHandler(shareService)
	friendHandler := handlers.NewFriendHandler(friendService)
	notificationHandler := handlers.NewNotificationHandler(notificationService)
	cartHandler := handlers.NewCartHandler(cartService)

	a := e.Group("/api")
	userGroup := a.Group("/user")
	usersGroup := a.Group("/users")
	thingsGroup := a.Group("/things")
	listsGroup := a.Group("/lists")
	imageGroup := a.Group("/images")
	shareGroup := a.Group("/shares")
	friendGroup := a.Group("/friends")
	friendRequestGroup := a.Group("/friend_requests")
	notificationsGroup := a.Group("/notifications")
	cartGroup := a.Group("/cart")

	// user group
	commonUserOptions := option.Group(
		option.Tags("Auth"),
	)
	fuegoecho.PostEcho(engine, userGroup, "/login", loginHandler.LoginHandlerPost,
		option.Summary("Login"),
		option.Description("Login and obtain cookies / JWT"),
		option.RequestBody(
			fuego.RequestBody{
				Type:         handlers.LoginPostParams{},
				ContentTypes: []string{"application/json"},
			},
		),
		option.AddResponse(
			200,
			"Successful login",
			fuego.Response{
				Type:         utils.NoContent{},
				ContentTypes: []string{""},
			},
		),
		option.ResponseHeader("Set-Cookie", "JWT Cookies", fuego.ParamExample("all tokens", "stashsphere-access=eyJhbGciOiJFZERTQSIsInR5cCI6IkpXVCJ9.eyJ1c2VySWQiOiJ2R1l6emg1Rzk0RklCSWtrOUZmX1kiLCJlbWFpbCI6InRlc3RAa2xhbmRlc3QuaW4iLCJuYW1lIjoiVGVzdGVyIiwiaXNzIjoiaW52ZW50b3J5Iiwic3ViIjoiYWNjZXNzIiwiZXhwIjoxNzY0MzU1NDU1LCJuYmYiOjE3NjQzMzM4NTUsImlhdCI6MTc2NDMzMzg1NX0.FYxYK-ROe2cN4Iu1oGjbUlz0FM5Y4h2yPGQ2Vli_fY1_iUPgQvw31wgOJfJ-md-Fj1oIdWJ5zmjsbgqdhHfABA; Path=/; Max-Age=21600; HttpOnly; SameSite=Strictstashsphere-info=eyJhbGciOiJub25lIiwidHlwIjoiSldUIn0.eyJ1c2VySWQiOiJ2R1l6emg1Rzk0RklCSWtrOUZmX1kiLCJlbWFpbCI6InRlc3RAa2xhbmRlc3QuaW4iLCJuYW1lIjoiVGVzdGVyIiwiaXNzIjoiaW52ZW50b3J5Iiwic3ViIjoiYWNjZXNzIiwiZXhwIjoxNzY0MzU1NDU1LCJuYmYiOjE3NjQzMzM4NTUsImlhdCI6MTc2NDMzMzg1NX0.; Path=/; Max-Age=21600; SameSite=Strictstashsphere-refresh=eyJhbGciOiJFZERTQSIsInR5cCI6IkpXVCJ9.eyJ1c2VySWQiOiJ2R1l6emg1Rzk0RklCSWtrOUZmX1kiLCJpc3MiOiJpbnZlbnRvcnkiLCJzdWIiOiJyZWZyZXNoIiwiZXhwIjoxNzY0OTM4NjU1LCJuYmYiOjE3NjQzMzM4NTUsImlhdCI6MTc2NDMzMzg1NX0.2KAKPkjEPItcEfuwO3gXgyMljLQMTmbhDTEaj8xuupxurKQO145PJjl-L0giv6sQr31SQB_OfNRq_BAWej6iCQ; Path=/api/user/refresh; Max-Age=604800; HttpOnly; SameSite=Strictstashsphere-refresh-info=eyJhbGciOiJub25lIiwidHlwIjoiSldUIn0.eyJ1c2VySWQiOiJ2R1l6emg1Rzk0RklCSWtrOUZmX1kiLCJpc3MiOiJpbnZlbnRvcnkiLCJzdWIiOiJyZWZyZXNoIiwiZXhwIjoxNzY0OTM4NjU1LCJuYmYiOjE3NjQzMzM4NTUsImlhdCI6MTc2NDMzMzg1NX0.; Path=/; Max-Age=604800; SameSite=Strict")),
		commonUserOptions,
	)
	fuegoecho.PostEcho(engine, userGroup, "/refresh", loginHandler.LoginHandlerRefreshPost,
		option.Summary("Refresh Access Token"),
		option.Description("Refresh access token using refresh token cookie. Requires stashsphere-refresh cookie."),
		option.Cookie("stashsphere-refresh", "JWT refresh token", param.Required()),
		option.AddResponse(
			200,
			"Successful token refresh",
			fuego.Response{
				Type:         utils.NoContent{},
				ContentTypes: []string{""},
			},
		),
		option.AddResponse(
			401,
			"Unauthorized - refresh token missing or invalid",
			fuego.Response{
				Type:         ss_middleware.ErrorResponse{},
				ContentTypes: []string{"application/json"},
			},
		),
		option.ResponseHeader("Set-Cookie", "Updated JWT Cookies", param.Example("access and refresh tokens", "stashsphere-access=...; stashsphere-info=...; stashsphere-refresh=...; stashsphere-refresh-info=...")),
		commonUserOptions,
	)
	fuegoecho.DeleteEcho(engine, userGroup, "/logout", loginHandler.LogoutHandlerDelete,
		option.Summary("Logout"),
		option.Description("Logout and clear authentication cookies"),
		option.Security(openapi3.SecurityRequirement{"cookieAuth": []string{}}),
		option.Cookie("stashsphere-access", "JWT access token", param.Required()),
		option.AddResponse(
			200,
			"Successful logout",
			fuego.Response{
				Type:         utils.NoContent{},
				ContentTypes: []string{""},
			},
		),
		option.ResponseHeader("Set-Cookie", "Cleared JWT Cookies", param.Example("expired cookies", "stashsphere-access=; Max-Age=0; stashsphere-refresh=; Max-Age=0")),
		commonUserOptions,
	)
	fuegoecho.PostEcho(engine, userGroup, "/register", registerHandler.RegisterHandlerPost,
		option.Summary("Register"),
		option.Description("Register a new user account"),
		option.RequestBody(
			fuego.RequestBody{
				Type:         handlers.RegisterPostParams{},
				ContentTypes: []string{"application/json"},
			},
		),
		option.AddResponse(
			200,
			"Successful registration",
			fuego.Response{
				Type:         utils.NoContent{},
				ContentTypes: []string{""},
			},
		),
		option.AddResponse(
			400,
			"Invalid parameters or invite code",
			fuego.Response{
				Type:         ss_middleware.ErrorResponse{},
				ContentTypes: []string{"application/json"},
			},
		),
		commonUserOptions,
	)
	fuegoecho.GetEcho(engine, userGroup, "/profile", profileHandler.ProfileHandlerGet,
		option.Summary("Get Profile"),
		option.Description("Get current authenticated user's profile information"),
		option.Security(openapi3.SecurityRequirement{"cookieAuth": []string{}}),
		option.Cookie("stashsphere-access", "JWT access token", param.Required()),
		option.AddResponse(
			200,
			"User profile",
			fuego.Response{
				Type:         resources.Profile{},
				ContentTypes: []string{"application/json"},
			},
		),
		option.AddResponse(
			401,
			"Not authenticated",
			fuego.Response{
				Type:         ss_middleware.ErrorResponse{},
				ContentTypes: []string{"application/json"},
			},
		),
		commonUserOptions,
	)
	fuegoecho.PatchEcho(engine, userGroup, "/profile", profileHandler.ProfileHandlerPatch,
		option.Summary("Update Profile"),
		option.Description("Update current authenticated user's profile information"),
		option.Security(openapi3.SecurityRequirement{"cookieAuth": []string{}}),
		option.Cookie("stashsphere-access", "JWT access token", param.Required()),
		option.RequestBody(
			fuego.RequestBody{
				Type:         handlers.ProfileUpdateParams{},
				ContentTypes: []string{"application/json"},
			},
		),
		option.AddResponse(
			200,
			"Updated user profile",
			fuego.Response{
				Type:         resources.Profile{},
				ContentTypes: []string{"application/json"},
			},
		),
		option.AddResponse(
			400,
			"Invalid parameters",
			fuego.Response{
				Type:         ss_middleware.ErrorResponse{},
				ContentTypes: []string{"application/json"},
			},
		),
		option.AddResponse(
			401,
			"Not authenticated",
			fuego.Response{
				Type:         ss_middleware.ErrorResponse{},
				ContentTypes: []string{"application/json"},
			},
		),
		commonUserOptions,
	)

	// users group
	commonUsersOptions := option.Group(
		option.Tags("Users"),
		option.Security(openapi3.SecurityRequirement{"cookieAuth": []string{}}),
		option.Cookie("stashsphere-access", "JWT access token", param.Required()),
	)
	fuegoecho.GetEcho(engine, usersGroup, "", userHandler.Index,
		option.Summary("List Users"),
		option.Description("Get list of all users in the system"),
		option.AddResponse(
			200,
			"List of users",
			fuego.Response{
				Type:         []resources.UserProfile{},
				ContentTypes: []string{"application/json"},
			},
		),
		option.AddResponse(
			401,
			"Not authenticated",
			fuego.Response{
				Type:         ss_middleware.ErrorResponse{},
				ContentTypes: []string{"application/json"},
			},
		),
		commonUsersOptions,
	)
	fuegoecho.GetEcho(engine, usersGroup, "/:userId", userHandler.Get,
		option.Summary("Get User"),
		option.Description("Get a specific user's profile by ID"),
		option.Path("userId", "User ID", param.Required(), param.Example("example user ID", "vGYzzh5G94FIBIkk9Ff_Y")),
		option.AddResponse(
			200,
			"User profile",
			fuego.Response{
				Type:         resources.UserProfile{},
				ContentTypes: []string{"application/json"},
			},
		),
		option.AddResponse(
			401,
			"Not authenticated",
			fuego.Response{
				Type:         ss_middleware.ErrorResponse{},
				ContentTypes: []string{"application/json"},
			},
		),
		option.AddResponse(
			404,
			"User not found",
			fuego.Response{
				Type:         ss_middleware.ErrorResponse{},
				ContentTypes: []string{"application/json"},
			},
		),
		commonUsersOptions,
	)

	// things group
	commonThingsOptions := option.Group(
		option.Tags("Things"),
		option.Security(openapi3.SecurityRequirement{"cookieAuth": []string{}}),
		option.Cookie("stashsphere-access", "JWT access token", param.Required()),
	)
	fuegoecho.GetEcho(engine, thingsGroup, "", thingHandler.ThingHandlerIndex,
		option.Summary("List Things"),
		option.Description("Get paginated list of things owned by or shared with the authenticated user"),
		option.Query("page", "Page number for pagination (0-indexed)", param.Example("page 0", "0")),
		option.Query("perPage", "Items per page (default: 50)", param.Example("50 items", "50")),
		option.Query("filterOwnerId", "Filter by owner user ID (can be repeated)", param.Example("owner ID", "abc123")),
		option.AddResponse(
			200,
			"Paginated list of things",
			fuego.Response{
				Type:         resources.PaginatedThings{},
				ContentTypes: []string{"application/json"},
			},
		),
		option.AddResponse(
			401,
			"Not authenticated",
			fuego.Response{
				Type:         ss_middleware.ErrorResponse{},
				ContentTypes: []string{"application/json"},
			},
		),
		commonThingsOptions,
	)
	fuegoecho.PostEcho(engine, thingsGroup, "", thingHandler.ThingHandlerPost,
		option.Summary("Create Thing"),
		option.Description("Create a new thing with properties, images, and sharing settings"),
		option.RequestBody(
			fuego.RequestBody{
				Type:         handlers.NewThingParams{},
				ContentTypes: []string{"application/json"},
			},
		),
		option.AddResponse(
			201,
			"Thing created successfully",
			fuego.Response{
				Type:         resources.ReducedThing{},
				ContentTypes: []string{"application/json"},
			},
		),
		option.AddResponse(
			400,
			"Invalid parameters",
			fuego.Response{
				Type:         ss_middleware.ErrorResponse{},
				ContentTypes: []string{"application/json"},
			},
		),
		option.AddResponse(
			401,
			"Not authenticated",
			fuego.Response{
				Type:         ss_middleware.ErrorResponse{},
				ContentTypes: []string{"application/json"},
			},
		),
		commonThingsOptions,
	)
	fuegoecho.GetEcho(engine, thingsGroup, "/summary", thingHandler.ThingHandlerSummary,
		option.Summary("Get Things Summary"),
		option.Description("Get summary statistics of things owned by the authenticated user"),
		option.AddResponse(
			200,
			"Things summary",
			fuego.Response{
				Type:         services.ThingsForUserSummary{},
				ContentTypes: []string{"application/json"},
			},
		),
		option.AddResponse(
			401,
			"Not authenticated",
			fuego.Response{
				Type:         ss_middleware.ErrorResponse{},
				ContentTypes: []string{"application/json"},
			},
		),
		commonThingsOptions,
	)
	fuegoecho.PatchEcho(engine, thingsGroup, "/:thingId", thingHandler.ThingHandlerPatch,
		option.Summary("Update Thing"),
		option.Description("Update an existing thing's properties, images, and metadata"),
		option.Path("thingId", "Thing ID", param.Required(), param.Example("example thing ID", "thing123")),
		option.RequestBody(
			fuego.RequestBody{
				Type:         handlers.UpdateThingParams{},
				ContentTypes: []string{"application/json"},
			},
		),
		option.AddResponse(
			200,
			"Thing updated successfully",
			fuego.Response{
				Type:         resources.Thing{},
				ContentTypes: []string{"application/json"},
			},
		),
		option.AddResponse(
			400,
			"Invalid parameters",
			fuego.Response{
				Type:         ss_middleware.ErrorResponse{},
				ContentTypes: []string{"application/json"},
			},
		),
		option.AddResponse(
			401,
			"Not authenticated",
			fuego.Response{
				Type:         ss_middleware.ErrorResponse{},
				ContentTypes: []string{"application/json"},
			},
		),
		option.AddResponse(
			404,
			"Thing not found",
			fuego.Response{
				Type:         ss_middleware.ErrorResponse{},
				ContentTypes: []string{"application/json"},
			},
		),
		commonThingsOptions,
	)
	fuegoecho.GetEcho(engine, thingsGroup, "/:thingId", thingHandler.ThingHandlerShow,
		option.Summary("Get Thing"),
		option.Description("Get detailed information about a specific thing"),
		option.Path("thingId", "Thing ID", param.Required(), param.Example("example thing ID", "thing123")),
		option.AddResponse(
			200,
			"Thing details",
			fuego.Response{
				Type:         resources.Thing{},
				ContentTypes: []string{"application/json"},
			},
		),
		option.AddResponse(
			401,
			"Not authenticated",
			fuego.Response{
				Type:         ss_middleware.ErrorResponse{},
				ContentTypes: []string{"application/json"},
			},
		),
		option.AddResponse(
			404,
			"Thing not found",
			fuego.Response{
				Type:         ss_middleware.ErrorResponse{},
				ContentTypes: []string{"application/json"},
			},
		),
		commonThingsOptions,
	)
	fuegoecho.DeleteEcho(engine, thingsGroup, "/:thingId", thingHandler.ThingHandlerDelete,
		option.Summary("Delete Thing"),
		option.Description("Delete a thing owned by the authenticated user"),
		option.Path("thingId", "Thing ID", param.Required(), param.Example("example thing ID", "thing123")),
		option.AddResponse(
			204,
			"Thing deleted successfully",
			fuego.Response{
				Type:         utils.NoContent{},
				ContentTypes: []string{""},
			},
		),
		option.AddResponse(
			401,
			"Not authenticated",
			fuego.Response{
				Type:         ss_middleware.ErrorResponse{},
				ContentTypes: []string{"application/json"},
			},
		),
		option.AddResponse(
			404,
			"Thing not found",
			fuego.Response{
				Type:         ss_middleware.ErrorResponse{},
				ContentTypes: []string{"application/json"},
			},
		),
		commonThingsOptions,
	)

	// lists group
	commonListsOptions := option.Group(
		option.Tags("Lists"),
		option.Security(openapi3.SecurityRequirement{"cookieAuth": []string{}}),
		option.Cookie("stashsphere-access", "JWT access token", param.Required()),
	)
	fuegoecho.GetEcho(engine, listsGroup, "", listHandler.ListHandlerIndex,
		option.Summary("List Lists"),
		option.Description("Get paginated list of lists owned by or shared with the authenticated user"),
		option.Query("page", "Page number for pagination (0-indexed)", param.Example("page 0", "0")),
		option.Query("perPage", "Items per page (default: 50)", param.Example("50 items", "50")),
		option.AddResponse(
			200,
			"Paginated list of lists",
			fuego.Response{
				Type:         resources.PaginatedLists{},
				ContentTypes: []string{"application/json"},
			},
		),
		option.AddResponse(
			401,
			"Not authenticated",
			fuego.Response{
				Type:         ss_middleware.ErrorResponse{},
				ContentTypes: []string{"application/json"},
			},
		),
		commonListsOptions,
	)
	fuegoecho.PostEcho(engine, listsGroup, "", listHandler.ListHandlerPost,
		option.Summary("Create List"),
		option.Description("Create a new list containing specified things"),
		option.RequestBody(
			fuego.RequestBody{
				Type:         handlers.NewListParams{},
				ContentTypes: []string{"application/json"},
			},
		),
		option.AddResponse(
			201,
			"List created successfully",
			fuego.Response{
				Type:         resources.ReducedList{},
				ContentTypes: []string{"application/json"},
			},
		),
		option.AddResponse(
			400,
			"Invalid parameters",
			fuego.Response{
				Type:         ss_middleware.ErrorResponse{},
				ContentTypes: []string{"application/json"},
			},
		),
		option.AddResponse(
			401,
			"Not authenticated",
			fuego.Response{
				Type:         ss_middleware.ErrorResponse{},
				ContentTypes: []string{"application/json"},
			},
		),
		commonListsOptions,
	)
	fuegoecho.GetEcho(engine, listsGroup, "/:listId", listHandler.ListHandlerShow,
		option.Summary("Get List"),
		option.Description("Get detailed information about a specific list including all contained things"),
		option.Path("listId", "List ID", param.Required(), param.Example("example list ID", "list123")),
		option.AddResponse(
			200,
			"List details",
			fuego.Response{
				Type:         resources.List{},
				ContentTypes: []string{"application/json"},
			},
		),
		option.AddResponse(
			401,
			"Not authenticated",
			fuego.Response{
				Type:         ss_middleware.ErrorResponse{},
				ContentTypes: []string{"application/json"},
			},
		),
		option.AddResponse(
			404,
			"List not found",
			fuego.Response{
				Type:         ss_middleware.ErrorResponse{},
				ContentTypes: []string{"application/json"},
			},
		),
		commonListsOptions,
	)
	fuegoecho.PatchEcho(engine, listsGroup, "/:listId", listHandler.ListHandlerPatch,
		option.Summary("Update List"),
		option.Description("Update an existing list's name, things, and sharing settings"),
		option.Path("listId", "List ID", param.Required(), param.Example("example list ID", "list123")),
		option.RequestBody(
			fuego.RequestBody{
				Type:         handlers.UpdateListParams{},
				ContentTypes: []string{"application/json"},
			},
		),
		option.AddResponse(
			200,
			"List updated successfully",
			fuego.Response{
				Type:         resources.List{},
				ContentTypes: []string{"application/json"},
			},
		),
		option.AddResponse(
			400,
			"Invalid parameters",
			fuego.Response{
				Type:         ss_middleware.ErrorResponse{},
				ContentTypes: []string{"application/json"},
			},
		),
		option.AddResponse(
			401,
			"Not authenticated",
			fuego.Response{
				Type:         ss_middleware.ErrorResponse{},
				ContentTypes: []string{"application/json"},
			},
		),
		option.AddResponse(
			404,
			"List not found",
			fuego.Response{
				Type:         ss_middleware.ErrorResponse{},
				ContentTypes: []string{"application/json"},
			},
		),
		commonListsOptions,
	)
	fuegoecho.DeleteEcho(engine, listsGroup, "/:listId", listHandler.ListHandlerDelete,
		option.Summary("Delete List"),
		option.Description("Delete a list owned by the authenticated user"),
		option.Path("listId", "List ID", param.Required(), param.Example("example list ID", "list123")),
		option.AddResponse(
			204,
			"List deleted successfully",
			fuego.Response{
				Type:         utils.NoContent{},
				ContentTypes: []string{""},
			},
		),
		option.AddResponse(
			401,
			"Not authenticated",
			fuego.Response{
				Type:         ss_middleware.ErrorResponse{},
				ContentTypes: []string{"application/json"},
			},
		),
		option.AddResponse(
			404,
			"List not found",
			fuego.Response{
				Type:         ss_middleware.ErrorResponse{},
				ContentTypes: []string{"application/json"},
			},
		),
		commonListsOptions,
	)

	// image group
	commonImagesOptions := option.Group(
		option.Tags("Images"),
		option.Security(openapi3.SecurityRequirement{"cookieAuth": []string{}}),
		option.Cookie("stashsphere-access", "JWT access token", param.Required()),
	)
	fuegoecho.GetEcho(engine, imageGroup, "", imageHandler.ImageHandlerIndex,
		option.Summary("List Images"),
		option.Description("Get paginated list of images owned by the authenticated user"),
		option.Query("page", "Page number for pagination (0-indexed)", param.Example("page 0", "0")),
		option.Query("perPage", "Items per page (default: 50)", param.Example("50 items", "50")),
		option.Query("onlyUnassigned", "Filter to show only unassigned images", param.Example("only unassigned", "true")),
		option.AddResponse(
			200,
			"Paginated list of images",
			fuego.Response{
				Type:         resources.PaginatedImages{},
				ContentTypes: []string{"application/json"},
			},
		),
		option.AddResponse(
			401,
			"Not authenticated",
			fuego.Response{
				Type:         ss_middleware.ErrorResponse{},
				ContentTypes: []string{"application/json"},
			},
		),
		commonImagesOptions,
	)
	fuegoecho.PostEcho(engine, imageGroup, "", imageHandler.ImageHandlerPost,
		option.Summary("Upload Image"),
		option.Description("Upload a new image file. Content-Type must be multipart/form-data with 'file' field."),
		option.AddResponse(
			201,
			"Image uploaded successfully",
			fuego.Response{
				Type:         resources.ReducedImage{},
				ContentTypes: []string{"application/json"},
			},
		),
		option.AddResponse(
			400,
			"Invalid file or parameters",
			fuego.Response{
				Type:         ss_middleware.ErrorResponse{},
				ContentTypes: []string{"application/json"},
			},
		),
		option.AddResponse(
			401,
			"Not authenticated",
			fuego.Response{
				Type:         ss_middleware.ErrorResponse{},
				ContentTypes: []string{"application/json"},
			},
		),
		commonImagesOptions,
	)
	fuegoecho.PatchEcho(engine, imageGroup, "/:imageId", imageHandler.ImageHandlerPatch,
		option.Summary("Modify Image"),
		option.Description("Modify an image by rotating it (90, 180, or 270 degrees)"),
		option.Path("imageId", "Image ID", param.Required(), param.Example("example image ID", "image123")),
		option.RequestBody(
			fuego.RequestBody{
				Type:         handlers.ImageModifyParams{},
				ContentTypes: []string{"application/json"},
			},
		),
		option.AddResponse(
			201,
			"Image modified successfully",
			fuego.Response{
				Type:         resources.ReducedImage{},
				ContentTypes: []string{"application/json"},
			},
		),
		option.AddResponse(
			400,
			"Invalid parameters",
			fuego.Response{
				Type:         ss_middleware.ErrorResponse{},
				ContentTypes: []string{"application/json"},
			},
		),
		option.AddResponse(
			401,
			"Not authenticated",
			fuego.Response{
				Type:         ss_middleware.ErrorResponse{},
				ContentTypes: []string{"application/json"},
			},
		),
		option.AddResponse(
			404,
			"Image not found",
			fuego.Response{
				Type:         ss_middleware.ErrorResponse{},
				ContentTypes: []string{"application/json"},
			},
		),
		commonImagesOptions,
	)
	fuegoecho.DeleteEcho(engine, imageGroup, "/:imageId", imageHandler.ImageHandlerDelete,
		option.Summary("Delete Image"),
		option.Description("Delete an image owned by the authenticated user"),
		option.Path("imageId", "Image ID", param.Required(), param.Example("example image ID", "image123")),
		option.AddResponse(
			200,
			"Image deleted successfully",
			fuego.Response{
				Type:         resources.ReducedImage{},
				ContentTypes: []string{"application/json"},
			},
		),
		option.AddResponse(
			401,
			"Not authenticated",
			fuego.Response{
				Type:         ss_middleware.ErrorResponse{},
				ContentTypes: []string{"application/json"},
			},
		),
		option.AddResponse(
			404,
			"Image not found",
			fuego.Response{
				Type:         ss_middleware.ErrorResponse{},
				ContentTypes: []string{"application/json"},
			},
		),
		commonImagesOptions,
	)

	// shares group
	commonSharesOptions := option.Group(
		option.Tags("Shares"),
		option.Security(openapi3.SecurityRequirement{"cookieAuth": []string{}}),
		option.Cookie("stashsphere-access", "JWT access token", param.Required()),
	)
	fuegoecho.PostEcho(engine, shareGroup, "", shareHandler.ShareHandlerPost,
		option.Summary("Create Share"),
		option.Description("Share a thing or list with another user"),
		option.RequestBody(
			fuego.RequestBody{
				Type:         handlers.NewShareParams{},
				ContentTypes: []string{"application/json"},
			},
		),
		option.AddResponse(
			201,
			"Share created successfully",
			fuego.Response{
				Type:         resources.Share{},
				ContentTypes: []string{"application/json"},
			},
		),
		option.AddResponse(
			400,
			"Invalid parameters",
			fuego.Response{
				Type:         ss_middleware.ErrorResponse{},
				ContentTypes: []string{"application/json"},
			},
		),
		option.AddResponse(
			401,
			"Not authenticated",
			fuego.Response{
				Type:         ss_middleware.ErrorResponse{},
				ContentTypes: []string{"application/json"},
			},
		),
		commonSharesOptions,
	)
	fuegoecho.GetEcho(engine, shareGroup, "/:shareId", shareHandler.ShareHandlerGet,
		option.Summary("Get Share"),
		option.Description("Get details of a specific share"),
		option.Path("shareId", "Share ID", param.Required(), param.Example("example share ID", "share123")),
		option.AddResponse(
			200,
			"Share details",
			fuego.Response{
				Type:         resources.Share{},
				ContentTypes: []string{"application/json"},
			},
		),
		option.AddResponse(
			401,
			"Not authenticated",
			fuego.Response{
				Type:         ss_middleware.ErrorResponse{},
				ContentTypes: []string{"application/json"},
			},
		),
		option.AddResponse(
			404,
			"Share not found",
			fuego.Response{
				Type:         ss_middleware.ErrorResponse{},
				ContentTypes: []string{"application/json"},
			},
		),
		commonSharesOptions,
	)
	fuegoecho.DeleteEcho(engine, shareGroup, "/:shareId", shareHandler.ShareHandlerDelete,
		option.Summary("Delete Share"),
		option.Description("Delete a share (unshare)"),
		option.Path("shareId", "Share ID", param.Required(), param.Example("example share ID", "share123")),
		option.AddResponse(
			200,
			"Share deleted successfully",
			fuego.Response{
				Type:         utils.NoContent{},
				ContentTypes: []string{""},
			},
		),
		option.AddResponse(
			401,
			"Not authenticated",
			fuego.Response{
				Type:         ss_middleware.ErrorResponse{},
				ContentTypes: []string{"application/json"},
			},
		),
		option.AddResponse(
			404,
			"Share not found",
			fuego.Response{
				Type:         ss_middleware.ErrorResponse{},
				ContentTypes: []string{"application/json"},
			},
		),
		commonSharesOptions,
	)

	// friends group
	commonFriendsOptions := option.Group(
		option.Tags("Friends"),
		option.Security(openapi3.SecurityRequirement{"cookieAuth": []string{}}),
		option.Cookie("stashsphere-access", "JWT access token", param.Required()),
	)
	fuegoecho.GetEcho(engine, friendGroup, "", friendHandler.FriendsIndex,
		option.Summary("List Friends"),
		option.Description("Get list of all friends for the authenticated user"),
		option.AddResponse(
			200,
			"List of friends",
			fuego.Response{
				Type:         resources.FriendShipsResponse{},
				ContentTypes: []string{"application/json"},
			},
		),
		option.AddResponse(
			401,
			"Not authenticated",
			fuego.Response{
				Type:         ss_middleware.ErrorResponse{},
				ContentTypes: []string{"application/json"},
			},
		),
		commonFriendsOptions,
	)
	fuegoecho.DeleteEcho(engine, friendGroup, "/:friendId", friendHandler.FriendDelete,
		option.Summary("Unfriend"),
		option.Description("Remove a friend connection (unfriend)"),
		option.Path("friendId", "Friend user ID", param.Required(), param.Example("example friend ID", "user123")),
		option.AddResponse(
			200,
			"Friend removed successfully",
			fuego.Response{
				Type:         utils.NoContent{},
				ContentTypes: []string{""},
			},
		),
		option.AddResponse(
			401,
			"Not authenticated",
			fuego.Response{
				Type:         ss_middleware.ErrorResponse{},
				ContentTypes: []string{"application/json"},
			},
		),
		option.AddResponse(
			404,
			"Friend not found",
			fuego.Response{
				Type:         ss_middleware.ErrorResponse{},
				ContentTypes: []string{"application/json"},
			},
		),
		commonFriendsOptions,
	)

	// friend_requests group
	commonFriendRequestsOptions := option.Group(
		option.Tags("Friend Requests"),
		option.Security(openapi3.SecurityRequirement{"cookieAuth": []string{}}),
		option.Cookie("stashsphere-access", "JWT access token", param.Required()),
	)
	fuegoecho.GetEcho(engine, friendRequestGroup, "", friendHandler.FriendRequestIndex,
		option.Summary("List Friend Requests"),
		option.Description("Get list of sent and received friend requests"),
		option.AddResponse(
			200,
			"List of friend requests",
			fuego.Response{
				Type:         resources.FriendRequestResponse{},
				ContentTypes: []string{"application/json"},
			},
		),
		option.AddResponse(
			401,
			"Not authenticated",
			fuego.Response{
				Type:         ss_middleware.ErrorResponse{},
				ContentTypes: []string{"application/json"},
			},
		),
		commonFriendRequestsOptions,
	)
	fuegoecho.PostEcho(engine, friendRequestGroup, "", friendHandler.FriendRequestPost,
		option.Summary("Send Friend Request"),
		option.Description("Send a friend request to another user"),
		option.RequestBody(
			fuego.RequestBody{
				Type:         handlers.NewFriendRequestParams{},
				ContentTypes: []string{"application/json"},
			},
		),
		option.AddResponse(
			201,
			"Friend request sent successfully",
			fuego.Response{
				Type:         resources.FriendRequest{},
				ContentTypes: []string{"application/json"},
			},
		),
		option.AddResponse(
			400,
			"Invalid parameters",
			fuego.Response{
				Type:         ss_middleware.ErrorResponse{},
				ContentTypes: []string{"application/json"},
			},
		),
		option.AddResponse(
			401,
			"Not authenticated",
			fuego.Response{
				Type:         ss_middleware.ErrorResponse{},
				ContentTypes: []string{"application/json"},
			},
		),
		commonFriendRequestsOptions,
	)
	fuegoecho.DeleteEcho(engine, friendRequestGroup, "/:requestId", friendHandler.FriendRequestDelete,
		option.Summary("Cancel Friend Request"),
		option.Description("Cancel a sent friend request"),
		option.Path("requestId", "Friend request ID", param.Required(), param.Example("example request ID", "request123")),
		option.AddResponse(
			200,
			"Friend request cancelled successfully",
			fuego.Response{
				Type:         utils.NoContent{},
				ContentTypes: []string{""},
			},
		),
		option.AddResponse(
			401,
			"Not authenticated",
			fuego.Response{
				Type:         ss_middleware.ErrorResponse{},
				ContentTypes: []string{"application/json"},
			},
		),
		option.AddResponse(
			404,
			"Friend request not found",
			fuego.Response{
				Type:         ss_middleware.ErrorResponse{},
				ContentTypes: []string{"application/json"},
			},
		),
		commonFriendRequestsOptions,
	)
	fuegoecho.PatchEcho(engine, friendRequestGroup, "/:requestId", friendHandler.FriendRequestUpdate,
		option.Summary("Respond to Friend Request"),
		option.Description("Accept or reject a received friend request"),
		option.Path("requestId", "Friend request ID", param.Required(), param.Example("example request ID", "request123")),
		option.RequestBody(
			fuego.RequestBody{
				Type:         handlers.UpdateFriendRequestParams{},
				ContentTypes: []string{"application/json"},
			},
		),
		option.AddResponse(
			200,
			"Friend request updated successfully",
			fuego.Response{
				Type:         resources.FriendRequest{},
				ContentTypes: []string{"application/json"},
			},
		),
		option.AddResponse(
			400,
			"Invalid parameters",
			fuego.Response{
				Type:         ss_middleware.ErrorResponse{},
				ContentTypes: []string{"application/json"},
			},
		),
		option.AddResponse(
			401,
			"Not authenticated",
			fuego.Response{
				Type:         ss_middleware.ErrorResponse{},
				ContentTypes: []string{"application/json"},
			},
		),
		option.AddResponse(
			404,
			"Friend request not found",
			fuego.Response{
				Type:         ss_middleware.ErrorResponse{},
				ContentTypes: []string{"application/json"},
			},
		),
		commonFriendRequestsOptions,
	)

	// notifications group
	commonNotificationsOptions := option.Group(
		option.Tags("Notifications"),
		option.Security(openapi3.SecurityRequirement{"cookieAuth": []string{}}),
		option.Cookie("stashsphere-access", "JWT access token", param.Required()),
	)
	fuegoecho.GetEcho(engine, notificationsGroup, "", notificationHandler.Index,
		option.Summary("List Notifications"),
		option.Description("Get paginated list of notifications for the authenticated user"),
		option.Query("page", "Page number for pagination (0-indexed)", param.Example("page 0", "0")),
		option.Query("perPage", "Items per page (default: 50)", param.Example("50 items", "50")),
		option.Query("onlyUnacknowledged", "Filter to show only unacknowledged notifications", param.Example("only unacknowledged", "true")),
		option.AddResponse(
			200,
			"Paginated list of notifications",
			fuego.Response{
				Type:         resources.PaginatedNotifications{},
				ContentTypes: []string{"application/json"},
			},
		),
		option.AddResponse(
			401,
			"Not authenticated",
			fuego.Response{
				Type:         ss_middleware.ErrorResponse{},
				ContentTypes: []string{"application/json"},
			},
		),
		commonNotificationsOptions,
	)
	fuegoecho.PatchEcho(engine, notificationsGroup, "/:notificationId", notificationHandler.Acknowledge,
		option.Summary("Acknowledge Notification"),
		option.Description("Mark a notification as acknowledged (read)"),
		option.Path("notificationId", "Notification ID", param.Required(), param.Example("example notification ID", "notif123")),
		option.AddResponse(
			200,
			"Notification acknowledged successfully",
			fuego.Response{
				Type:         utils.NoContent{},
				ContentTypes: []string{""},
			},
		),
		option.AddResponse(
			401,
			"Not authenticated",
			fuego.Response{
				Type:         ss_middleware.ErrorResponse{},
				ContentTypes: []string{"application/json"},
			},
		),
		option.AddResponse(
			404,
			"Notification not found",
			fuego.Response{
				Type:         ss_middleware.ErrorResponse{},
				ContentTypes: []string{"application/json"},
			},
		),
		commonNotificationsOptions,
	)

	// cart group
	commonCartOptions := option.Group(
		option.Tags("Cart"),
		option.Security(openapi3.SecurityRequirement{"cookieAuth": []string{}}),
		option.Cookie("stashsphere-access", "JWT access token", param.Required()),
	)
	fuegoecho.GetEcho(engine, cartGroup, "", cartHandler.Index,
		option.Summary("Get Cart"),
		option.Description("Get the current user's shopping cart contents"),
		option.AddResponse(
			200,
			"Cart contents",
			fuego.Response{
				Type:         resources.Cart{},
				ContentTypes: []string{"application/json"},
			},
		),
		option.AddResponse(
			401,
			"Not authenticated",
			fuego.Response{
				Type:         ss_middleware.ErrorResponse{},
				ContentTypes: []string{"application/json"},
			},
		),
		commonCartOptions,
	)
	fuegoecho.PutEcho(engine, cartGroup, "", cartHandler.Put,
		option.Summary("Update Cart"),
		option.Description("Replace the cart contents with a new list of thing IDs"),
		option.RequestBody(
			fuego.RequestBody{
				Type:         handlers.CartParams{},
				ContentTypes: []string{"application/json"},
			},
		),
		option.AddResponse(
			200,
			"Cart updated successfully",
			fuego.Response{
				Type:         resources.Cart{},
				ContentTypes: []string{"application/json"},
			},
		),
		option.AddResponse(
			400,
			"Invalid parameters",
			fuego.Response{
				Type:         ss_middleware.ErrorResponse{},
				ContentTypes: []string{"application/json"},
			},
		),
		option.AddResponse(
			401,
			"Not authenticated",
			fuego.Response{
				Type:         ss_middleware.ErrorResponse{},
				ContentTypes: []string{"application/json"},
			},
		),
		commonCartOptions,
	)

	// search group
	commonSearchOptions := option.Group(
		option.Tags("Search"),
		option.Security(openapi3.SecurityRequirement{"cookieAuth": []string{}}),
		option.Cookie("stashsphere-access", "JWT access token", param.Required()),
	)
	fuegoecho.GetEcho(engine, a, "/search", searchHandler.SearchHandlerGet,
		option.Summary("Search"),
		option.Description("Search across things and lists"),
		option.Query("query", "Search query string", param.Required(), param.Example("search term", "laptop")),
		option.AddResponse(
			200,
			"Search results",
			fuego.Response{
				Type:         resources.SearchResult{},
				ContentTypes: []string{"application/json"},
			},
		),
		option.AddResponse(
			401,
			"Not authenticated",
			fuego.Response{
				Type:         ss_middleware.ErrorResponse{},
				ContentTypes: []string{"application/json"},
			},
		),
		commonSearchOptions,
	)
	fuegoecho.GetEcho(engine, a, "/search/property_auto_complete", searchHandler.AutocompleteGet,
		option.Summary("Autocomplete thing properties"),
		option.Description("Autocomplete names and values of properties"),
		option.Query("name", "the name to auto-complete, won't auto-complete when value is provided", param.Required(), param.Example("name", "length")),
		option.Query("value", "the value to auto-complete", param.Example("value", "1300")),
		option.AddResponse(
			200,
			"Search results",
			fuego.Response{
				Type:         services.PropertyAutoCompleteResult{},
				ContentTypes: []string{"application/json"},
			},
		),
		option.AddResponse(
			401,
			"Not authenticated",
			fuego.Response{
				Type:         ss_middleware.ErrorResponse{},
				ContentTypes: []string{"application/json"},
			},
		),
		commonSearchOptions,
	)

	// assets group
	commonAssetsOptions := option.Group(
		option.Tags("Assets"),
		option.Security(openapi3.SecurityRequirement{"cookieAuth": []string{}}),
		option.Cookie("stashsphere-access", "JWT access token", param.Required()),
	)
	fuegoecho.GetEcho(engine, e, "/assets/:hash", imageHandler.ImageHandlerGet,
		option.Summary("Get Image Asset"),
		option.Description("Retrieve an image file by its hash. Supports optional resizing via width parameter."),
		option.Path("hash", "Image content hash", param.Required(), param.Example("example hash", "ABCDEF123456")),
		option.Query("width", "Resize image to specified width (20-8192 pixels, only for JPEG/PNG)", param.Example("resize to 800px", "800")),
		option.AddResponse(
			200,
			"Image file (binary data)",
			fuego.Response{
				Type:         []byte{},
				ContentTypes: []string{"image/jpeg", "image/png", "image/gif"},
			},
		),
		option.AddResponse(
			304,
			"Not Modified (ETag match)",
			fuego.Response{
				Type:         utils.NoContent{},
				ContentTypes: []string{""},
			},
		),
		option.AddResponse(
			401,
			"Not authenticated",
			fuego.Response{
				Type:         ss_middleware.ErrorResponse{},
				ContentTypes: []string{"application/json"},
			},
		),
		option.AddResponse(
			404,
			"Image not found",
			fuego.Response{
				Type:         ss_middleware.ErrorResponse{},
				ContentTypes: []string{"application/json"},
			},
		),
		option.ResponseHeader("ETag", "Entity tag for caching", param.Example("etag value", "ABCDEF123456")),
		option.ResponseHeader("Cache-Control", "Cache control directive", param.Example("no-cache", "no-cache")),
		option.Header("If-None-Match", "ETag for conditional request", param.Example("etag value", "ABCDEF123456")),
		commonAssetsOptions,
	)
	e.HEAD("/assets/:hash", imageHandler.ImageHandlerGet)

	engine.RegisterOpenAPIRoutes(&fuegoecho.OpenAPIHandler{Echo: e})

	return e, engine, db, nil
}

func Serve(config config.StashSphereServeConfig, debug bool, serveOpenAPI bool) error {
	echo, _, db, err := setup(config, debug, serveOpenAPI, "")
	if err != nil {
		return err
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Error().Err(err).Msg("failed to close database connection")
		}
	}()
	log.Info().Msgf("stashsphere listening on %s", config.ListenAddress)
	return echo.Start(config.ListenAddress)
}

var serveCommand = &cobra.Command{
	Use:   "serve",
	Short: "Serve the Stashsphere API",
	RunE: func(cmd *cobra.Command, args []string) error {
		configPaths, _ := cmd.Flags().GetStringSlice("conf")
		debug, _ := cmd.Flags().GetBool("debug")
		serveOpenAPI, _ := cmd.Flags().GetBool("serve-openapi")

		var config config.StashSphereServeConfig

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
			"database": map[string]interface{}{
				"user": "stashsphere",
				"name": "stashsphere",
				"host": "127.0.0.1",
			},
			"listenAddress": ":8081",
			"auth": map[string]interface{}{
				"privateKey": "",
			},
			"image": map[string]interface{}{
				"path":      imagePath,
				"cachePath": imageCachePath,
			},
			"invites": map[string]interface{}{
				"enabled": false,
				"code":    "",
			},
			"domains": map[string]interface{}{
				"allowed": []string{"http://localhost"},
				"own":     []string{"localhost"},
			},
			"frontendUrl":  "http://localhost",
			"instanceName": "stashsphereDev",
			"email": map[string]interface{}{
				"backend": "stdout",
			},
		}, "."), nil)

		for _, configPath := range configPaths {
			if err := k.Load(file.Provider(configPath), yaml.Parser()); err != nil {
				log.Fatal().Msgf("error loading config: %v", err)
			}
			k.UnmarshalWithConf("", &config, koanf.UnmarshalConf{Tag: "koanf", FlatPaths: false})
		}

		return Serve(config, debug, serveOpenAPI)
	},
}

func init() {
	serveCommand.Flags().StringSlice("conf", []string{"stashsphere.yaml"}, "path to one or more .yaml config files")
	serveCommand.Flags().Bool("debug", false, "enable debug mode")
	serveCommand.Flags().Bool("serve-openapi", false, "enable serving OpenAPI/Swagger UI at /swagger")
	rootCmd.AddCommand(serveCommand)
}
