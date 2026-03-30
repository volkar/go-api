package cache

import (
	"fmt"

	"github.com/google/uuid"
)

/* Cache key "cache:mapper:SLUG" */
func keyMapper(slug string) string {
	return fmt.Sprintf("cache:mapper:%s", slug)
}

/* Cache key "cache:user:UUID" */
func keyUser(id uuid.UUID) string {
	return fmt.Sprintf("cache:user:%s", id.String())
}

/* Cache key "cache:album:USER_SLUG/ALBUM_SLUG" */
func keyAlbum(user_slug string, album_slug string) string {
	return fmt.Sprintf("cache:album:%s/%s", user_slug, album_slug)
}

/* Cache key "cache:list:UUID" */
func keyAlbumList(id uuid.UUID) string {
	return fmt.Sprintf("cache:list:%s", id.String())
}

/* Cache key "cache:deleted:UUID" */
func keyDeletedAlbums(id uuid.UUID) string {
	return fmt.Sprintf("cache:deleted:%s", id.String())
}
