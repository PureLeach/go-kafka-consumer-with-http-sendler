package models

type User struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type Notification struct {
	Msg1 string `string:"msg1"`
	Msg2 string `string:"msg2"`
	// From    User   `json:"from"`
	// To      User   `json:"to"`
	// Message string `json:"message"`
}
