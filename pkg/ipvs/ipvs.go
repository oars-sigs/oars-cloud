package ipvs

//Service ...
type Service struct {
	Address   string
	Protocol  string
	Port      uint16
	Scheduler string
	Flags     ServiceFlags
	Timeout   uint32
}

//Destination ...
type Destination struct {
	Address string
	Port    uint16
	Weight  int
}

//ServiceFlags ...
type ServiceFlags uint32

const (
	// FlagPersistent specify IPVS service session affinity
	FlagPersistent = 0x1
	// FlagHashed specify IPVS service hash flag
	FlagHashed = 0x2
)
