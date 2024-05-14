package database

import "go.mongodb.org/mongo-driver/bson"

func withoutDeleted(filter any) any {
	deletedFilter := bson.M{"deletedAt": nil}

	if filter == nil {
		return deletedFilter
	}

	return bson.M{"$and": bson.A{filter, deletedFilter}}
}
