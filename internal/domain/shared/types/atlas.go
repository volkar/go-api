package types

// AtlasItem represents an item in the album's atlas
type AtlasItem struct {
	Type string `json:"type" validate:"required,oneof=title text"`
	Src  string `json:"src" validate:"required,min=1"`
}

// Atlas represents the album's atlas
type Atlas []AtlasItem
