package os

type OSInfo struct {
	ID      string // ubuntu, debian, centos, rocky
	IDLike  string // debian, rhel
	Version string // 22.04
	Major   int    // 22
	Family  string // debian | rhel
}
