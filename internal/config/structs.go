package config

type Config struct {
	Collectors Collectors `mapstructure:"collectors"`
}

type Collectors struct {
	Carpark Carpark `mapstructure:"carpark"`
}

type Carpark struct {
	FacilityIDs []string `mapstructure:"facility_ids"`
}
