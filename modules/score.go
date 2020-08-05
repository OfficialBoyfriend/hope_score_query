package modules

import "errors"

// Score 成绩信息
type Score struct {
	Model
	IdId uint `json:"id_id"` // 所属学生
	CourseId uint `json:"course_id"` // 所属课程
	Result float64 `json:"result"` // 成绩
	ExaminationTime string `json:"examination_time"` // 考试时间
	Credit float64 `json:"credit"` // 学分
	EduType string `json:"edu_type"` // 修读方式
}

func (s *Score) Create() error {
	if !db.NewRecord(s) {
		return errors.New("检测到错误，主键已存在")
	}
	return db.Create(s).Error
}

func (s *Score) GetAutoCreate(where map[string]interface{}) error {
	return db.Where(where).FirstOrCreate(s).Error
}

func (s *Score) GetAutoCreateUseStruct() error {
	return db.FirstOrCreate(s, s).Error
}

func (s *Score) GetFind(where map[string]interface{}) (results []*Score, err error) {
	results = make([]*Score, 0)
	err = db.Where(where).Find(&results).Error
	return results, err
}

func (s Score) GetTop(num uint, where map[string]interface{}) (results []*Score, err error) {
	results = make([]*Score, 0)
	err = db.Order("result desc").Limit(num).Where(where).Find(&results).Error
	return results, err
}