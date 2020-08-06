package modules

// Course 课程信息
type Class struct {
	Model
	Name string `json:"name"` // 名称（工业机器人1802）
	Code string `json:"code"` // 代码（例如1802）
	Description string `json:"description"` // 描述
	Students []*Id `json:"students"` // 学生
}

func (c *Class) Get(where map[string]interface{}) error {
	return db.Where(where).First(c).Error
}

func (c *Class) GetAutoCreate(where map[string]interface{}) error {
	return db.Where(where).FirstOrCreate(c).Error
}

func (c *Class) GetFind(where map[string]interface{}, isPreload ...bool) (results []*Class, err error) {
	results = make([]*Class, 0)
	if len(isPreload) > 1 && isPreload[0] {
		err = db.Where(where).Preload("Scores").Find(&results).Error
		return results, err
	}
	err = db.Where(where).Find(&results).Error
	return results, err
}