package models

type Picture Video

// PictureDescriptor describes the Picture table (returns name and filterset)
func PictureDescriptor() Descriptor {
	d := VideoDescriptor()
	d.TableName = "pictures"
	return d
}
