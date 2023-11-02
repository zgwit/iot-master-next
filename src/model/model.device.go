package model

type DeviceType struct {
	CreateTime string `form:"create_time" bson:"create_time" json:"create_time"`

	Name    string `form:"name" bson:"name" json:"name"`
	Id      string `form:"id" bson:"id" json:"id"`
	ModelId string `form:"model_id" bson:"model_id" json:"model_id"`

	Labels []NameValueType `form:"labels" bson:"labels" json:"labels"`

	Config interface{} `form:"config" bson:"config" json:"config"`
}
