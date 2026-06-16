package model

type Tag struct {
	IdModel
	Name         string                 `json:"name" gorm:"default:'';not null;"`
	UserId       uint                   `json:"user_id" gorm:"default:0;not null;index"`
	Color        uint                   `json:"color" gorm:"default:0;not null;"` //color is the color value of flutter, from 0x00000000 to 0xFFFFFFFF; the first two digits represent transparency, and the last 6 digits represent color, which can be converted to rgba
	CollectionId uint                   `json:"collection_id" gorm:"default:0;not null;index"`
	Collection   *AddressBookCollection `json:"collection,omitempty"`
	TimeModel
}

type TagList struct {
	Tags []*Tag `json:"list"`
	Pagination
}
