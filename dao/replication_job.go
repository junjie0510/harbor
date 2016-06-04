package dao

import (
	"fmt"

	"strings"

	"github.com/astaxie/beego/orm"
	"github.com/vmware/harbor/models"
)

// AddRepTarget ...
func AddRepTarget(target models.RepTarget) (int64, error) {
	o := orm.NewOrm()
	return o.Insert(&target)
}

// GetRepTarget ...
func GetRepTarget(id int64) (*models.RepTarget, error) {
	o := orm.NewOrm()
	t := models.RepTarget{ID: id}
	err := o.Read(&t)
	if err == orm.ErrNoRows {
		return nil, nil
	}
	return &t, err
}

// DeleteRepTarget ...
func DeleteRepTarget(id int64) error {
	o := orm.NewOrm()
	_, err := o.Delete(&models.RepTarget{ID: id})
	return err
}

// UpdateRepTarget ...
func UpdateRepTarget(target models.RepTarget) error {
	o := orm.NewOrm()
	if len(target.Password) != 0 {
		_, err := o.Update(&target)
		return err
	}

	_, err := o.Update(&target, "URL", "Name", "Username")
	return err
}

// GetAllRepTargets ...
func GetAllRepTargets() ([]*models.RepTarget, error) {
	o := orm.NewOrm()
	qs := o.QueryTable(&models.RepTarget{})
	var targets []*models.RepTarget
	_, err := qs.All(&targets)
	return targets, err
}

// AddRepPolicy ...
func AddRepPolicy(policy models.RepPolicy) (int64, error) {
	o := orm.NewOrm()
	sqlTpl := `insert into replication_policy (name, project_id, target_id, enabled, description, cron_str, start_time, creation_time, update_time ) values (?, ?, ?, ?, ?, ?, %s, NOW(), NOW())`
	var sql string
	if policy.Enabled == 1 {
		sql = fmt.Sprintf(sqlTpl, "NOW()")
	} else {
		sql = fmt.Sprintf(sqlTpl, "NULL")
	}
	p, err := o.Raw(sql).Prepare()
	if err != nil {
		return 0, err
	}
	r, err := p.Exec(policy.Name, policy.ProjectID, policy.TargetID, policy.Enabled, policy.Description, policy.CronStr)
	if err != nil {
		return 0, err
	}
	id, err := r.LastInsertId()
	return id, err
}

// GetRepPolicy ...
func GetRepPolicy(id int64) (*models.RepPolicy, error) {
	o := orm.NewOrm()
	p := models.RepPolicy{ID: id}
	err := o.Read(&p)
	if err == orm.ErrNoRows {
		return nil, nil
	}
	return &p, err
}

// GetRepPolicyByProject ...
func GetRepPolicyByProject(projectID int64) ([]*models.RepPolicy, error) {
	var res []*models.RepPolicy
	o := orm.NewOrm()
	_, err := o.QueryTable("replication_policy").Filter("project_id", projectID).All(&res)
	return res, err
}

// DeleteRepPolicy ...
func DeleteRepPolicy(id int64) error {
	o := orm.NewOrm()
	_, err := o.Delete(&models.RepPolicy{ID: id})
	return err
}

// UpdateRepPolicyEnablement ...
func UpdateRepPolicyEnablement(id int64, enabled int) error {
	o := orm.NewOrm()
	p := models.RepPolicy{
		ID:      id,
		Enabled: enabled}
	_, err := o.Update(&p, "Enabled")

	return err
}

// EnableRepPolicy ...
func EnableRepPolicy(id int64) error {
	return UpdateRepPolicyEnablement(id, 1)
}

// DisableRepPolicy ...
func DisableRepPolicy(id int64) error {
	return UpdateRepPolicyEnablement(id, 0)
}

// AddRepJob ...
func AddRepJob(job models.RepJob) (int64, error) {
	o := orm.NewOrm()
	if len(job.Status) == 0 {
		job.Status = models.JobPending
	}
	if len(job.TagList) > 0 {
		job.Tags = strings.Join(job.TagList, ",")
	}
	return o.Insert(&job)
}

// GetRepJob ...
func GetRepJob(id int64) (*models.RepJob, error) {
	o := orm.NewOrm()
	j := models.RepJob{ID: id}
	err := o.Read(&j)
	if err == orm.ErrNoRows {
		return nil, nil
	}
	genTagListForJob(&j)
	return &j, nil
}

// GetRepJobByPolicy ...
func GetRepJobByPolicy(policyID int64) ([]*models.RepJob, error) {
	var res []*models.RepJob
	_, err := repJobPolicyIDQs(policyID).All(&res)
	genTagListForJob(res...)
	return res, err
}

// GetRepJobToStop get jobs that are possibly being handled by workers of a certain policy.
func GetRepJobToStop(policyID int64) ([]*models.RepJob, error) {
	var res []*models.RepJob
	_, err := repJobPolicyIDQs(policyID).Filter("status__in", models.JobPending, models.JobRunning).All(&res)
	genTagListForJob(res...)
	return res, err
}

func repJobPolicyIDQs(policyID int64) orm.QuerySeter {
	o := orm.NewOrm()
	return o.QueryTable("replication_job").Filter("policy_id", policyID)
}

// DeleteRepJob ...
func DeleteRepJob(id int64) error {
	o := orm.NewOrm()
	_, err := o.Delete(&models.RepJob{ID: id})
	return err
}

// UpdateRepJobStatus ...
func UpdateRepJobStatus(id int64, status string) error {
	o := orm.NewOrm()
	j := models.RepJob{
		ID:     id,
		Status: status,
	}
	num, err := o.Update(&j, "Status")
	if num == 0 {
		err = fmt.Errorf("Failed to update replication job with id: %d %s", id, err.Error())
	}
	return err
}

func genTagListForJob(jobs ...*models.RepJob) {
	for _, j := range jobs {
		if len(j.Tags) > 0 {
			j.TagList = strings.Split(j.Tags, ",")
		}
	}
}