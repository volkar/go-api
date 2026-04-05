package cache

import (
	"github.com/google/uuid"
)

var (
	UserMapperPrefix  = "c:u_m:"
	UserEntityPrefix  = "c:u:"
	AlbumMapperPrefix = "c:a_m:"
	AlbumEntityPrefix = "c:a:"
)

/* Cache key "c:u:USER_ID" */
func keyUserEntity(userID uuid.UUID) string {
	return UserEntityPrefix + userID.String()
}

/* Cache key "c:u_m:USER_SLUG" */
func keyUserMapper(userSlug string) string {
	return UserMapperPrefix + userSlug
}

/* Cache key "c:a:ALBUM_ID" */
func keyAlbumEntity(albumID uuid.UUID) string {
	return AlbumEntityPrefix + albumID.String()
}

/* Cache key "c:a_m:USER_ID/ALBUM_SLUG" */
func keyAlbumMapper(userID uuid.UUID, albumSlug string) string {
	return AlbumMapperPrefix + userID.String() + "/" + albumSlug
}
