package integration_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stashsphere/backend/cmd"
	"github.com/stashsphere/backend/factories"
	"github.com/stashsphere/backend/services"
	testcommon "github.com/stashsphere/backend/test_common"
	"github.com/stretchr/testify/assert"
)

type paginatedPage struct {
	Things         []json.RawMessage `json:"things"`
	Lists          []json.RawMessage `json:"lists"`
	TotalCount     uint64            `json:"totalCount"`
	TotalPageCount uint64            `json:"totalPageCount"`
}

func getPage(t *testing.T, e *echo.Echo, cookies []*http.Cookie, path string) paginatedPage {
	req := httptest.NewRequest(http.MethodGet, path, nil)
	for _, cookie := range cookies {
		req.AddCookie(cookie)
	}
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	var page paginatedPage
	assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &page))
	return page
}

func TestPaginateFalseReturnsOnePage(t *testing.T) {
	db, tearDown, err := testcommon.CreateTestSchema()
	assert.NoError(t, err)
	t.Cleanup(tearDown)
	t.Cleanup(func() { db.Close() })

	e, _, err := cmd.SetupWithDB(db, oidcBaseConfig(t), false, false, "")
	assert.NoError(t, err)

	ctx := context.Background()
	userService := services.NewUserService(db, false, "", 60, nil)
	notificationService := services.NewNotificationService(db, services.NotificationData{
		FrontendUrl:  "https://example.com",
		InstanceName: "StashsphereTest",
	}, &services.TestEmailService{})
	imageService, err := services.NewTmpImageService(db)
	assert.NoError(t, err)
	t.Cleanup(func() { os.Remove(imageService.StorePath()) })
	thingService := services.NewThingService(db, imageService, notificationService)
	listService := services.NewListService(db, notificationService)

	userParams := factories.UserFactory.MustCreate().(*services.CreateUserParams)
	user, err := userService.CreateUser(ctx, *userParams)
	assert.NoError(t, err)
	for range 3 {
		thingParams := factories.ThingFactory.MustCreate().(*services.CreateThingParams)
		thingParams.OwnerId = user.ID
		_, err = thingService.CreateThing(ctx, *thingParams)
		assert.NoError(t, err)
		listParams := factories.ListFactory.MustCreate().(*services.CreateListParams)
		listParams.OwnerId = user.ID
		_, err = listService.CreateList(ctx, *listParams)
		assert.NoError(t, err)
	}

	login := httptest.NewRequest(http.MethodPost, "/api/user/login",
		strings.NewReader(`{"email":"`+userParams.Email+`","password":"`+userParams.Password+`"}`))
	login.Header.Set("Content-Type", "application/json")
	loginRec := httptest.NewRecorder()
	e.ServeHTTP(loginRec, login)
	assert.Equal(t, http.StatusOK, loginRec.Code)
	cookies := loginRec.Result().Cookies()

	// paginated, one per page
	things := getPage(t, e, cookies, "/api/things?perPage=1")
	assert.Len(t, things.Things, 1)
	assert.Equal(t, uint64(3), things.TotalPageCount)
	lists := getPage(t, e, cookies, "/api/lists?perPage=1")
	assert.Len(t, lists.Lists, 1)
	assert.Equal(t, uint64(3), lists.TotalPageCount)

	// without pagination everything is on a single page, whatever perPage says
	for _, path := range []string{"/api/things?paginate=false", "/api/things?paginate=false&perPage=1"} {
		things = getPage(t, e, cookies, path)
		assert.Len(t, things.Things, 3, path)
		assert.Equal(t, uint64(3), things.TotalCount, path)
		assert.Equal(t, uint64(1), things.TotalPageCount, path)
	}
	for _, path := range []string{"/api/lists?paginate=false", "/api/lists?paginate=false&perPage=1"} {
		lists = getPage(t, e, cookies, path)
		assert.Len(t, lists.Lists, 3, path)
		assert.Equal(t, uint64(3), lists.TotalCount, path)
		assert.Equal(t, uint64(1), lists.TotalPageCount, path)
	}
}
