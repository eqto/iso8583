package message

//Signon ...
type Signon struct {
	Status string `bit:"39"`

	/*
		001: Signon
		002: Signoff
		201: Cutoff
		301: Echo
	*/
	NetworkCode int `bit:"70"`
}
