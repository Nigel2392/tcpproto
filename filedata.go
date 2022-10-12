package tcpproto

type FileData struct {
	Name     string
	Size     int
	Boundary string
	Present  bool
	Content  []byte
}

func (fdata *FileData) StartBoundary() []byte {
	return []byte("--" + fdata.Boundary + "--")

}

func (fdata *FileData) EndBoundary() []byte {
	return []byte("----" + fdata.Boundary + "----")
}
