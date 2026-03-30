package albumtypes

import "slices"

type Access struct {
	Type  string   `json:"type" validate:"oneof=public private shared"`
	Share []string `json:"share" validate:"required_if=Type shared,dive,email"`
}

/* Check if resource can be accessed */
func CanAccess(access Access, userEmail string, isOwner bool) bool {
	if isOwner {
		return true
	}

	switch access.Type {
	case "public":
		return true
	case "shared":
		return slices.Contains(access.Share, userEmail)
	default:
		return false
	}
}
