package snapshot

type SnapshotInfo struct {
	Name string
	Uuid string
}

func (s *SnapshotInfo) DomainID() string {
	return s.Uuid
}

func NewSnapshotInfo(name string, uuid string) (*SnapshotInfo, error) {
	snp := new(SnapshotInfo)
	snp.Name = name
	snp.Uuid = uuid

	return snp, nil
}
