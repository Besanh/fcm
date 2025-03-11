package repositories

import (
	"context"
	"fcm/models"
	"fcm/pkgs/mongodb"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type (
	IDevicesNotification interface {
		IRepoGeneric[models.DevicesNotification]
		SelectInfo(ctx context.Context, param models.DevicesNotificationQueryParam, limit, offset int64) (total int64, entries []*models.DevicesNotification, err error)
		CountUnreadNotifications(ctx context.Context, param models.DevicesNotificationQueryParam) (total int64, err error)
	}
	DevicesNotification struct {
		RepoGeneric[models.DevicesNotification]
	}
)

func NewDevicesNotification(db mongodb.IMongoDBClient) IDevicesNotification {
	return &DevicesNotification{
		RepoGeneric: RepoGeneric[models.DevicesNotification]{
			DB:         db,
			Collection: "devices_notifications",
		},
	}
}

func (repo *DevicesNotification) SelectInfo(ctx context.Context, param models.DevicesNotificationQueryParam, limit, offset int64) (total int64, entries []*models.DevicesNotification, err error) {
	entries = make([]*models.DevicesNotification, 0)
	filters := make(bson.D, 0)

	// TODO: Add more filter
	if len(param.DeviceId_Eq) > 0 {
		filters = append(filters, primitive.E{Key: "device_id", Value: param.DeviceId_Eq})
	}
	if len(param.Type_Eq) > 0 {
		filters = append(filters, primitive.E{Key: "type", Value: param.Type_Eq})
	}
	if !param.Time_Gte.IsZero() {
		filters = append(filters, primitive.E{Key: "created_at", Value: bson.M{"$gte": param.Time_Gte}})
	}
	if !param.Time_Lte.IsZero() {
		filters = append(filters, primitive.E{Key: "created_at", Value: bson.M{"$lte": param.Time_Lte}})
	}

	cur, err := repo.GetCollection().Find(ctx, filters, options.Find().
		SetSort(bson.D{{Key: "created_at", Value: -1}}).
		SetLimit(limit).
		SetSkip(offset))
	if err != nil {
		return
	}
	defer cur.Close(ctx)

	for cur.Next(ctx) {
		entry := new(models.DevicesNotification)
		if err = cur.Decode(entry); err != nil {
			return
		}
		entries = append(entries, entry)
	}
	if err = cur.Err(); err != nil {
		return
	}
	total, err = repo.GetCollection().CountDocuments(ctx, filters)
	if err != nil {
		return
	}
	return
}

func (repo *DevicesNotification) CountUnreadNotifications(ctx context.Context, param models.DevicesNotificationQueryParam) (total int64, err error) {
	filters := make(bson.D, 0)

	// TODO: Add more filter
	if len(param.DeviceId_Eq) > 0 {
		filters = append(filters, primitive.E{Key: "device_id", Value: param.DeviceId_Eq})
	}
	if len(param.Type_Eq) > 0 {
		filters = append(filters, primitive.E{Key: "type", Value: param.Type_Eq})
	}
	if !param.Time_Gte.IsZero() {
		filters = append(filters, primitive.E{Key: "created_at", Value: bson.M{"$gte": param.Time_Gte}})
	}
	if !param.Time_Lte.IsZero() {
		filters = append(filters, primitive.E{Key: "created_at", Value: bson.M{"$lte": param.Time_Lte}})
	}

	filters = append(filters, primitive.E{Key: "read_at", Value: bson.M{"$eq": nil}})
	total, err = repo.GetCollection().CountDocuments(ctx, filters)
	if err != nil {
		return
	}
	return
}
