package model

type DeviceType struct {
	Name    string `form:"name" bson:"name" json:"name"`
	Id      string `form:"id" bson:"id" json:"id"`
	ModelId string `form:"model_id" bson:"model_id" json:"model_id"`

	Labels map[string]string `form:"labels" bson:"labels" json:"labels"`

	Config interface{} `form:"config" bson:"config" json:"config"`
}
