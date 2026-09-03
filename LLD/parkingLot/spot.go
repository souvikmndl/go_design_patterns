package parkinglot

type SpotSize string

const (
	Small  SpotSize = "SMALL"
	Medium SpotSize = "MEDIUM"
	Large  SpotSize = "Large"
)

type Floor struct {
	floorNumber     int
	spots           []*Spot
	occupiedSpotIDs map[string]bool
}

func NewFloor(floorNumber int) *Floor {
	return &Floor{
		floorNumber:     floorNumber,
		spots:           []*Spot{},
		occupiedSpotIDs: make(map[string]bool),
	}
}

func (fl *Floor) GetFloorNumber() int {
	return fl.floorNumber
}

func (fl *Floor) GetFloorSpots() []*Spot {
	return fl.spots
}

func (fl *Floor) findAvailableSpot(size SpotSize) *Spot {
	for _, spot := range fl.spots {
		if spot.GetSpotSize() == size {
			if _, ok := fl.occupiedSpotIDs[spot.GetID()]; !ok {
				return spot
			}
		}
	}

	return nil
}

type Spot struct {
	id       string
	spotSize SpotSize
}

func NewSpot(id string, size SpotSize) *Spot {
	return &Spot{
		id:       id,
		spotSize: size,
	}
}

func (sp *Spot) GetID() string {
	return sp.id
}

func (sp *Spot) GetSpotSize() SpotSize {
	return sp.spotSize
}
