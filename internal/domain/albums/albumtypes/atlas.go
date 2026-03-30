package albumtypes

type AtlasItem struct {
	Type string `json:"type" validate:"required,oneof=title text"`
	Src  string `json:"src" validate:"required,min=1"`
}

type Atlas []AtlasItem
