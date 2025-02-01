package resources

type Actions struct {
	CanEdit   bool `json:"canEdit"`
	CanDelete bool `json:"canDelete"`
	CanShare  bool `json:"canShare"`
}
