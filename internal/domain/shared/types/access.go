package types

import "slices"

// Access represents album access level
type Access string

const (
	AccessPublic  Access = "public"
	AccessPrivate Access = "private"
	AccessShared  Access = "shared"
)

/* Check if resource can be accessed */
func (a Access) CanAccess(sharedEmails []string, userEmail string, isOwner bool) bool {
	if isOwner {
		return true
	}

	switch a {
	case AccessPublic:
		return true
	case AccessShared:
		return slices.Contains(sharedEmails, userEmail)
	default:
		return false
	}
}
