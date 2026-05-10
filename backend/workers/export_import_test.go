package workers

import (
	"archive/zip"
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stashsphere/backend/factories"
	"github.com/stashsphere/backend/models"
	"github.com/stashsphere/backend/services"
	testcommon "github.com/stashsphere/backend/test_common"
	"github.com/stretchr/testify/assert"
)

func TestImportExportRoundtrip(t *testing.T) {
	ctx := context.Background()

	// Source instance: create data and export it

	db1, tearDown1, err := testcommon.CreateTestSchema()
	assert.NoError(t, err)
	t.Cleanup(func() { db1.Close() })
	t.Cleanup(tearDown1)

	imageStore1 := t.TempDir()
	imageTmp1 := t.TempDir()
	exportStore1 := t.TempDir()
	exportTmp1 := t.TempDir()

	imageService1, err := services.NewImageService(db1, imageStore1, imageTmp1)
	assert.NoError(t, err)

	userService1 := services.NewUserService(db1, false, "", 60, nil)
	ns1 := services.NewNotificationService(db1, services.NotificationData{
		FrontendUrl:  "https://example.com",
		InstanceName: "StashsphereTest",
	}, &services.TestEmailService{})
	thingService1 := services.NewThingService(db1, imageService1, ns1)
	listService1 := services.NewListService(db1, ns1)
	exportService1 := services.NewExportService(db1)

	userParams := factories.UserFactory.MustCreate().(*services.CreateUserParams)
	user1, err := userService1.CreateUser(ctx, *userParams)
	assert.NoError(t, err)

	pngFile, err := testcommon.Assets.Open("assets/test.png")
	assert.NoError(t, err)
	image1, err := imageService1.CreateImage(ctx, user1.ID, "test.png", pngFile)
	pngFile.Close()
	assert.NoError(t, err)

	thingParams := factories.ThingFactory.MustCreate().(*services.CreateThingParams)
	thingParams.OwnerId = user1.ID
	thingParams.ImagesIds = []string{image1.ID}
	thing1, err := thingService1.CreateThing(ctx, *thingParams)
	assert.NoError(t, err)

	listParams := factories.ListFactory.MustCreate().(*services.CreateListParams)
	listParams.OwnerId = user1.ID
	listParams.ThingIds = []string{thing1.ID}
	list1, err := listService1.CreateList(ctx, *listParams)
	assert.NoError(t, err)

	_, err = exportService1.CreateExport(ctx, user1.ID)
	assert.NoError(t, err)

	ew := NewExportWorker(db1, imageStore1, exportStore1, exportTmp1, 24*time.Hour, time.Minute)
	ew.process()

	done, err := exportService1.GetDoneExport(ctx, user1.ID)
	assert.NoError(t, err)
	assert.True(t, done.FilePath.Valid)

	zipPath := filepath.Join(exportStore1, done.FilePath.String)
	_, err = os.Stat(zipPath)
	assert.NoError(t, err, "exported zip must exist on disk")

	// Destination instance: simulate destroy + recreate, then import

	db2, tearDown2, err := testcommon.CreateTestSchema()
	assert.NoError(t, err)
	t.Cleanup(func() { db2.Close() })
	t.Cleanup(tearDown2)

	imageStore2 := t.TempDir()
	imageTmp2 := t.TempDir()
	importDir2 := t.TempDir()

	imageService2, err := services.NewImageService(db2, imageStore2, imageTmp2)
	assert.NoError(t, err)

	userService2 := services.NewUserService(db2, false, "", 60, nil)
	user2Params := factories.UserFactory.MustCreate().(*services.CreateUserParams)
	user2, err := userService2.CreateUser(ctx, *user2Params)
	assert.NoError(t, err)

	importService2, err := services.NewImportService(db2, importDir2)
	assert.NoError(t, err)

	zipFile, err := os.Open(zipPath)
	assert.NoError(t, err)
	defer zipFile.Close()

	_, err = importService2.QueueImport(ctx, user2.ID, zipFile, 500*1024*1024)
	assert.NoError(t, err)

	iw := NewImportWorker(db2, imageService2, importDir2, time.Minute)
	iw.process()

	// Verify import completed successfully
	imp, err := importService2.GetImport(ctx, user2.ID)
	assert.NoError(t, err)
	assert.Equal(t, models.ImportStatusDone, imp.Status)
	assert.False(t, imp.ErrorMSG.Valid, "import error: %s", imp.ErrorMSG.String)
	assert.Equal(t, 1, imp.ThingsImported.Int)
	assert.Equal(t, 1, imp.ListsImported.Int)
	assert.Equal(t, 1, imp.ImagesImported.Int)

	// Assert things were restored
	things2, err := models.Things(
		models.ThingWhere.OwnerID.EQ(user2.ID),
	).All(ctx, db2)
	assert.NoError(t, err)
	assert.Len(t, things2, 1)
	assert.Equal(t, thing1.Name, things2[0].Name)
	assert.Equal(t, thing1.Description, things2[0].Description)
	assert.Equal(t, thing1.PrivateNote, things2[0].PrivateNote)

	// Assert lists were restored
	lists2, err := models.Lists(
		models.ListWhere.OwnerID.EQ(user2.ID),
	).All(ctx, db2)
	assert.NoError(t, err)
	assert.Len(t, lists2, 1)
	assert.Equal(t, list1.Name, lists2[0].Name)

	// Assert images were restored
	images2, err := models.Images(
		models.ImageWhere.OwnerID.EQ(user2.ID),
	).All(ctx, db2)
	assert.NoError(t, err)
	assert.Len(t, images2, 1)
	assert.Equal(t, image1.Name, images2[0].Name)
	assert.Equal(t, image1.Mime, images2[0].Mime)
}

func TestExportWorkerPurgesErroredRows(t *testing.T) {
	ctx := context.Background()

	db, tearDown, err := testcommon.CreateTestSchema()
	assert.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	t.Cleanup(tearDown)

	userService := services.NewUserService(db, false, "", 60, nil)
	userParams := factories.UserFactory.MustCreate().(*services.CreateUserParams)
	user, err := userService.CreateUser(ctx, *userParams)
	assert.NoError(t, err)

	exportService := services.NewExportService(db)
	exp, err := exportService.CreateExport(ctx, user.ID)
	assert.NoError(t, err)

	// Mark it as an error and back-date created_at past the retention period.
	retention := time.Hour
	_, err = models.Exports(models.ExportWhere.ID.EQ(exp.ID)).UpdateAll(ctx, db, models.M{
		"status":     models.ExportStatusError,
		"error_msg":  "something went wrong",
		"created_at": time.Now().UTC().Add(-(retention + time.Minute)),
	})
	assert.NoError(t, err)

	ew := NewExportWorker(db, t.TempDir(), t.TempDir(), t.TempDir(), retention, time.Minute)
	ew.process()

	count, err := models.Exports(models.ExportWhere.ID.EQ(exp.ID)).Count(ctx, db)
	assert.NoError(t, err)
	assert.Equal(t, int64(0), count, "errored export row must be deleted after retention period")
}

func TestImportMissingCollection(t *testing.T) {
	ctx := context.Background()

	db, tearDown, err := testcommon.CreateTestSchema()
	assert.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	t.Cleanup(tearDown)

	importDir := t.TempDir()
	imageStore := t.TempDir()
	imageTmp := t.TempDir()

	imageService, err := services.NewImageService(db, imageStore, imageTmp)
	assert.NoError(t, err)

	userService := services.NewUserService(db, false, "", 60, nil)
	userParams := factories.UserFactory.MustCreate().(*services.CreateUserParams)
	user, err := userService.CreateUser(ctx, *userParams)
	assert.NoError(t, err)

	// Build a ZIP that contains a file but no collection.json.
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	w, err := zw.Create("readme.txt")
	assert.NoError(t, err)
	_, err = w.Write([]byte("no collection here"))
	assert.NoError(t, err)
	assert.NoError(t, zw.Close())

	importService, err := services.NewImportService(db, importDir)
	assert.NoError(t, err)
	_, err = importService.QueueImport(ctx, user.ID, &buf, 500*1024*1024)
	assert.NoError(t, err)

	iw := NewImportWorker(db, imageService, importDir, time.Minute)
	iw.process()

	imp, err := importService.GetImport(ctx, user.ID)
	assert.NoError(t, err)
	assert.Equal(t, models.ImportStatusError, imp.Status)
	assert.Contains(t, imp.ErrorMSG.String, "collection.json")
}
