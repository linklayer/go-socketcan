package socketcan

type CanFrame struct {
	ArbId    uint32
	Dlc      byte
	Data     []byte
	Extended bool
}
