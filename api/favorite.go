package api

type FavoritePage struct {
	Data       []Favorite `json:"data"`
	NextCursor string     `json:"nextCursor,omitempty"`
}

type Favorite struct {
	ID string `json:"id"`
}
