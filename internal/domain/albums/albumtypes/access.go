package albumtypes

import "slices"

/* Check if resource can be accessed */
func CanAccess(access string, sharedEmails []string, userEmail string, isOwner bool) bool {
	if isOwner {
		return true
	}

	switch access {
	case "public":
		return true
	case "shared":
		return slices.Contains(sharedEmails, userEmail)
	default:
		return false
	}
}
