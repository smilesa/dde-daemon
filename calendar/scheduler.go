package calendar

import (
	"time"

	"github.com/jinzhu/gorm"
	"github.com/linuxdeepin/go-dbus-factory/org.freedesktop.notifications"
	dbus "pkg.deepin.io/lib/dbus1"
	"pkg.deepin.io/lib/dbusutil"
)

type Scheduler struct {
	db            *gorm.DB
	notifications *notifications.Notifications
	timerGroup    timerGroup
	changeChan    chan struct{}
	quitChan      chan struct{}
	methods       *struct {
		GetJobs   func() `in:"startYear,startMonth,startDay,endYear,endMonth,endDay" out:"jobs"`
		GetJob    func() `in:"id" out:"job"`
		DeleteJob func() `in:"id"`
		UpdateJob func() `in:"jobInfo"`
		CreateJob func() `in:"jobInfo" out:"id"`

		GetTypes   func() `out:"types"`
		GetType    func() `in:"id" out:"type"`
		DeleteType func() `in:"id"`
		UpdateType func() `in:"typeInfo"`
		CreateType func() `in:"typeInfo" out:"id"`

		DebugRemindJob func() `in:"id"`
	}
}

func newScheduler(db *gorm.DB, service *dbusutil.Service) *Scheduler {
	sessionBus := service.Conn()
	m := &Scheduler{
		db:            db,
		changeChan:    make(chan struct{}),
		quitChan:      make(chan struct{}),
		notifications: notifications.NewNotifications(sessionBus),
	}
	return m
}

func (s *Scheduler) GetInterfaceName() string {
	return dbusInterface
}

func (s *Scheduler) GetJobs(startYear, startMonth, startDay, endYear, endMonth, endDay int32) (string, *dbus.Error) {
	start := newTimeYMDHM(int(startYear), time.Month(startMonth), int(startDay), 0, 0)
	end := newTimeYMDHM(int(endYear), time.Month(endMonth), int(endDay), 0, 0)
	jobs, err := s.getJobs(start, end)
	if err != nil {
		return "", dbusutil.ToError(err)
	}
	result, err := toJson(jobs)
	return result, dbusutil.ToError(err)
}

type DateJobsWrap struct {
	Date string
	Jobs []*JobJSON
}

func (s *Scheduler) getJobs(start, end time.Time) ([]DateJobsWrap, error) {
	var allJobs []*Job
	err := s.db.Find(&allJobs).Error
	if err != nil {
		return nil, err
	}

	t0 := time.Now()
	var result []DateJobsWrap
	err = iterDays(start, end, func(t time.Time) error {
		date := timeToDate(t)
		jobs, err := getDateJobs(allJobs, date)
		if err != nil {
			return err
		}

		jobs1 := make([]*JobJSON, len(jobs))
		for i, j := range jobs {
			jj, err := j.toJobJSON()
			if err != nil {
				return err
			}
			jobs1[i] = jj
		}

		result = append(result, DateJobsWrap{
			Date: timeToDate(t).String(),
			Jobs: jobs1,
		})
		return nil
	})
	if err != nil {
		return nil, err
	}
	logger.Debug("cost time:", time.Since(t0))
	return result, nil
}

func (s *Scheduler) getJob(id uint) (*JobJSON, error) {
	var job Job
	err := s.db.First(&job, id).Error
	if err != nil {
		return nil, err
	}
	return job.toJobJSON()
}

func (s *Scheduler) deleteJob(id uint) error {
	var job Job
	err := s.db.Select("id").First(&job, id).Error
	if err != nil {
		return err
	}

	return s.db.Unscoped().Delete(&job).Error
}

func (s *Scheduler) updateJob(job *Job) error {
	err := job.validate()
	if err != nil {
		return err
	}
	var job0 Job
	err = s.db.Find(&job0, job.ID).Error
	if err != nil {
		return err
	}

	diffMap := make(map[string]interface{})

	if job0.Type != job.Type {
		diffMap["Type"] = job.Type
	}
	if job0.Title != job.Title {
		diffMap["Title"] = job.Title
	}
	if job0.Description != job.Description {
		diffMap["Description"] = job.Description
	}
	if job0.AllDay != job.AllDay {
		diffMap["AllDay"] = job.AllDay
	}
	if !job0.Start.Equal(job.Start) {
		diffMap["Start"] = job.Start
	}
	if !job0.End.Equal(job.End) {
		diffMap["End"] = job.End
	}
	if job0.RRule != job.RRule {
		diffMap["RRule"] = job.RRule
	}
	if job0.Remind != job.Remind {
		diffMap["Remind"] = job.Remind
	}
	if job0.Ignore != job.Ignore {
		diffMap["Ignore"] = job.Ignore
	}

	if len(diffMap) > 0 {
		err = s.db.Model(job).Updates(diffMap).Error
	}
	return err
}

func (s *Scheduler) createJob(job *Job) error {
	err := job.validate()
	if err != nil {
		return err
	}
	job.ID = 0

	err = s.db.Create(job).Error
	return err
}

func (s *Scheduler) getTypes() ([]*JobTypeJSON, error) {
	var types []JobType
	err := s.db.Find(&types).Error
	if err != nil {
		return nil, err
	}

	result := make([]*JobTypeJSON, len(types))
	for idx, t := range types {
		result[idx] = t.toJobTypeJSON()
	}
	return result, nil
}

func (s *Scheduler) getType(id uint) (*JobTypeJSON, error) {
	var jobType JobType
	err := s.db.First(&jobType, id).Error
	if err != nil {
		return nil, err
	}
	return jobType.toJobTypeJSON(), nil
}

func (s *Scheduler) deleteType(id uint) error {
	var jobType JobType
	err := s.db.Select("id").First(&jobType, id).Error
	if err != nil {
		return err
	}

	return s.db.Delete(&jobType).Error
}

func (s *Scheduler) createType(jobType *JobType) error {
	err := jobType.validate()
	if err != nil {
		return err
	}
	jobType.ID = 0
	return s.db.Create(jobType).Error
}

func (s *Scheduler) updateType(jobType *JobType) error {
	err := jobType.validate()
	if err != nil {
		return err
	}
	var jobType0 JobType
	err = s.db.Find(&jobType0, jobType.ID).Error
	if err != nil {
		return err
	}

	diffMap := make(map[string]interface{})

	if jobType0.Name != jobType.Name {
		diffMap["Name"] = jobType.Name
	}
	if jobType0.Color != jobType.Color {
		diffMap["Color"] = jobType.Color
	}

	if len(diffMap) > 0 {
		err = s.db.Model(jobType).Updates(diffMap).Error
	}
	return err
}

type timerGroup struct {
	timers []*time.Timer
}

func (tg *timerGroup) addJob(job *Job, m *Scheduler) {
	now := time.Now()
	duration := job.remindTime.Sub(now)
	logger.Debugf("notify job %d %q after %v", job.ID, job.Title, duration)
	tg.timers = append(tg.timers, time.AfterFunc(duration, func() {
		m.remindJob(job)
	}))
}

func (tg *timerGroup) reset() {
	for _, timer := range tg.timers {
		timer.Stop()
	}
	tg.timers = nil
}

func (s *Scheduler) remindJob(job *Job) {
	layout := "2006-01-02 15:04"
	body := job.Start.Format(layout) + " ~ " + job.End.Format(layout)
	logger.Debug("remind:", job.Title, body)
	id, err := s.notifications.Notify(0, "dde-daemon", 0,
		"", job.Title,
		body, nil, nil, 0)
	if err != nil {
		logger.Warning(err)
	}
	logger.Debug("id:", id)
}

func (s *Scheduler) DebugRemindJob(id int64) *dbus.Error {
	var job Job
	err := s.db.First(&job, id).Error
	if err != nil {
		return dbusutil.ToError(err)
	}

	s.remindJob(&job)
	return nil
}

func (s *Scheduler) notifyJobsChange() {
	s.changeChan <- struct{}{}
}

func (s *Scheduler) startRemindLoop() {
	const interval = 10 * time.Minute
	ticker := time.NewTicker(interval)

	setTimerGroup := func(tr TimeRange) {
		s.timerGroup.reset()
		jobs, err := s.getRemindJobs(tr)
		if err != nil {
			logger.Warning(err)
		}
		for _, job := range jobs {
			s.timerGroup.addJob(job, s)
		}
	}

	go func() {
		for {
			select {
			case now := <-ticker.C:
				tr := TimeRange{
					start: now,
					end:   now.Add(interval),
				}
				setTimerGroup(tr)

			case <-s.changeChan:
				now := time.Now()
				tr := TimeRange{
					start: now,
					end:   now.Add(interval),
				}
				setTimerGroup(tr)

			case <-s.quitChan:
				return
			}
		}
	}()
}

func (s *Scheduler) getRemindJobs(tr TimeRange) ([]*Job, error) {
	var allJobs []*Job
	err := s.db.Find(&allJobs, "remind IS NOT NULL AND remind != '' ").Error
	if err != nil {
		return nil, err
	}

	start := setClock(tr.start, Clock{})
	end := start.AddDate(0, 0, 8)

	var result []*Job
	err = iterDays(start, end, func(t time.Time) error {

		date := timeToDate(t)
		jobs, err := getDateJobs(allJobs, date)
		if err != nil {
			return err
		}

		for _, job := range jobs {
			remindT, err := parseRemind(job.Start, job.Remind)
			if err != nil {
				continue
			}
			job.remindTime = remindT
			if (tr.start.Before(remindT) || tr.start.Equal(remindT)) &&
				(remindT.Before(tr.end) || remindT.Equal(tr.end)) {
				// tr.start <= remindT <= tr.end
				result = append(result, job)
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}