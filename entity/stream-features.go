package entity

type Features struct {
	Mechanisms []string `xml:"mechanisms>mechanism"`
	dumbProducer
}
