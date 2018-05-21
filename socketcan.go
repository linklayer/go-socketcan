package socketcan

type CanFrame struct {
	ArbId    int
	Dlc      byte
	Data     []byte
	Extended bool
}
