package modules

// Course 课程信息
type Course struct {
	Model
	Name string `json:"name"` // 名称
	Description string `json:"description"` // 描述
	Scores []*Score `json:"scores"` // 成绩
}

func (s *Course) Get(where map[string]interface{}) error {
	return db.Where(where).First(s).Error
}

func (s *Course) GetFind(where map[string]interface{}, isPreload ...bool) (results []*Course, err error) {
	results = make([]*Course, 0)
	if len(isPreload) > 1 && isPreload[0] {
		err = db.Where(where).Preload("Scores").Find(&results).Error
		return results, err
	}
	err = db.Where(where).Find(&results).Error
	return results, err
}

func (s *Course) GetAutoCreate(where map[string]interface{}) error {
	return db.Where(where).FirstOrCreate(s).Error
}