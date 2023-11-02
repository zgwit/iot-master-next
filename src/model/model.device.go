package model

type DeviceType struct {
	CreateTime int64 `form:"create_time" bson:"create_time" json:"create_time"`

	Name    string `form:"name" bson:"name" json:"name"`
	Id      string `form:"id" bson:"id" json:"id"`
	ModelId string `form:"model_id" bson:"model_id" json:"model_id"`

	Labels []NameValueType `form:"labels" bson:"labels" json:"labels"`

	Config interface{} `form:"config" bson:"config" json:"config"`
}

type DevciceByCreateTime []DeviceType

func (a DevciceByCreateTime) Len() int           { return len(a) }
func (a DevciceByCreateTime) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a DevciceByCreateTime) Less(i, j int) bool { return a[i].CreateTime > a[j].CreateTime }
